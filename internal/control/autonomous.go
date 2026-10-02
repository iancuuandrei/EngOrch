package control

import (
	"context"
	"errors"
	"fmt"

	"harness.local/engorch/internal/agenttree"
	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/effects"
)

// AutonomousRepair records a bounded, candidate-specific repair iteration.
// A candidate can consume at most one repair slot: retrying a failed runtime
// cannot silently spend the budget or create an unbounded loop.
type AutonomousRepair struct {
	Attempt     int    `json:"attempt"`
	CandidateID string `json:"candidate_id"`
}

func autonomousPolicyID(s Snapshot) (string, error) {
	if s.Creation.Execution == nil {
		return "", errors.New("autonomous execution was not enabled for this run")
	}
	if err := s.Creation.Execution.Validate(); err != nil {
		return "", err
	}
	return canonical.Hash("harness.execution-policy.v1", *s.Creation.Execution)
}

func autonomousFileAuthorization(s Snapshot, intent effects.Intent) (effects.Authorization, error) {
	policyID, err := autonomousPolicyID(s)
	if err != nil {
		return effects.Authorization{}, err
	}
	id, err := intent.ID()
	if err != nil {
		return effects.Authorization{}, err
	}
	return effects.Authorization{IntentID: id, Actor: "fabric:autonomous", Authority: "autonomous-v1", PolicyID: policyID}, nil
}

// autonomousFilesApplied identifies the restart boundary after a confirmed
// writer effect but before verification was planned. Comparing the complete
// prepared effect prevents an unrelated or older file receipt from advancing
// this run.
func autonomousFilesApplied(s Snapshot) bool {
	return s.FileOutcome == "CONFIRMED" && s.FileIntent != nil && s.WriterProposal != nil &&
		s.FileIntent.Prepared.Intent == s.WriterProposal.Prepared.Intent &&
		s.Candidate != nil && *s.Candidate == s.FileIntent.Prepared.Proposal.After
}

func autonomousRepairApplied(s Snapshot) bool {
	if !autonomousFilesApplied(s) || s.RepairCandidateID == "" {
		return false
	}
	beforeID, err := s.WriterProposal.Prepared.Proposal.Before.ID()
	return err == nil && beforeID == s.RepairCandidateID
}

// autonomousVerificationCovered reports whether the exact candidate already has
// verification evidence. A confirmed repair must be verified once before a new
// repair slot is admitted; an already-verified failure must admit a new slot
// instead of re-verifying the same candidate and discarding feedback.
func autonomousVerificationCovered(s Snapshot, candidateID string) bool {
	return s.Verification != nil && s.Verification.Plan.CandidateID == candidateID
}

// autonomousDispatchBlocked refuses to admit new provider or role work while
// earlier work remains pending or UNKNOWN. Reconciliation observes state; it
// never resends an uncertain effect.
func autonomousDispatchBlocked(s Snapshot) error {
	if s.Verification != nil && s.Verification.Pending {
		return errors.New("verification is pending and cannot be resent automatically")
	}
	for id, d := range s.AgentDispatch {
		if d.Observation == nil {
			return fmt.Errorf("agent dispatch %s is pending and cannot be resent automatically", id)
		}
		if d.Observation.Status == agenttree.StatusUnknown {
			return fmt.Errorf("agent dispatch %s is UNKNOWN and requires reconciliation", id)
		}
	}
	for _, m := range s.ModelAccess {
		if m.Terminal == nil {
			return errors.New("model access remains unresolved and cannot be resent automatically")
		}
	}
	if s.PlannerAccess != nil && s.Plan == nil {
		return errors.New("planner invocation unresolved; automatic retry denied")
	}
	if s.ExplorerHost != nil && s.ExplorerHost.RuntimeReceipt == nil {
		return errors.New("explorer runtime remains unresolved and cannot be resent automatically")
	}
	if s.WriterHost != nil && s.WriterHost.RuntimeReceipt == nil {
		return errors.New("writer runtime remains unresolved and cannot be resent automatically")
	}
	if s.ReviewHost != nil && s.ReviewHost.RuntimeReceipt == nil {
		return errors.New("review runtime remains unresolved and cannot be resent automatically")
	}
	if s.PlannerHost != nil && s.PlannerReceipt == nil {
		return errors.New("planner runtime remains unresolved and cannot be resent automatically")
	}
	return nil
}

// autonomousWriterEffect advances the candidate-bound writer/proposal/
// authorization/application sequence shared by IMPLEMENTING and REPAIRING.
// It preserves every dispatch guard, the exact candidate binding, and all
// error handling of the inlined blocks it replaces.
//
// It returns proposalStaged=true when a fresh proposal was requested and the
// caller must continue the outer loop, or proposalStaged=false after the
// confirmed effect was applied and the caller proceeds to the normal next
// iteration.
func autonomousWriterEffect(ctx context.Context, path string, s Snapshot) (Snapshot, bool, error) {
	if s.WriterProposal == nil || s.WriterProposal.Prepared.Proposal.Before != *s.Candidate {
		if err := autonomousDispatchBlocked(s); err != nil {
			return s, false, err
		}
		if _, err := RunWriter(ctx, path); err != nil {
			latest, inspectErr := InspectOr(s, path, err)
			return latest, false, inspectErr
		}
		return s, true, nil
	}
	if err := autonomousDispatchBlocked(s); err != nil {
		return s, false, err
	}
	a, err := autonomousFileAuthorization(s, s.WriterProposal.Prepared.Intent)
	if err != nil {
		return s, false, err
	}
	if _, err := ApplyFiles(ctx, path, s.WriterProposal.Prepared, a); err != nil {
		latest, inspectErr := InspectOr(s, path, err)
		return latest, false, inspectErr
	}
	return s, false, nil
}

// RunAutonomous advances an explicitly opted-in run through normal controller
// operations. Every operation retains its own intent-first and candidate-bound
// safeguards. In particular this function never recreates an UNKNOWN workspace
// or file effect, never repeats a pending verification, and never publishes,
// commits, or pushes.
//
// It is safe to call after process restart. A completed stage is inspected and
// skipped; an incomplete runtime is handed to that role's existing resume path.
func RunAutonomous(ctx context.Context, path string) (Snapshot, error) {
	for steps := 0; steps < 64; steps++ {
		if err := ctx.Err(); err != nil {
			s, inspectErr := Inspect(path)
			if inspectErr != nil {
				return s, errors.Join(err, inspectErr)
			}
			return s, err
		}
		s, err := Inspect(path)
		if err != nil {
			return s, err
		}
		if _, err := autonomousPolicyID(s); err != nil {
			return s, err
		}
		if err := RequireDispatchAllowed(s); err != nil {
			return s, err
		}
		switch s.State {
		case "OBJECTIVE", "PLANNING":
			if _, err := ResumePlanning(ctx, path); err != nil {
				return InspectOr(s, path, err)
			}
		case "AWAITING_APPROVAL":
			policyID, err := autonomousPolicyID(s)
			if err != nil {
				return s, err
			}
			if err := Append(path, "plan.autonomous-authorized", MachineApproval{PlanID: s.PlanID, PolicyID: policyID, Actor: "fabric:autonomous"}); err != nil {
				return InspectOr(s, path, err)
			}
		case "IMPLEMENTING":
			if s.Workspace == nil {
				if s.WorkspaceIntent != nil {
					if _, err := ReconcileWorkspace(ctx, path); err != nil {
						return InspectOr(s, path, fmt.Errorf("autonomous workspace remains unresolved: %w", err))
					}
				} else {
					if err := autonomousDispatchBlocked(s); err != nil {
						return s, err
					}
					if _, err := StartWorkspace(ctx, path); err != nil {
						return InspectOr(s, path, err)
					}
				}
				continue
			}
			if s.FileOutcome == "UNKNOWN" {
				if _, err := ReconcileFiles(ctx, path); err != nil {
					return InspectOr(s, path, fmt.Errorf("autonomous file effect remains UNKNOWN: %w", err))
				}
				continue
			}
			// A confirmed initial effect restarts directly at verification,
			// never with another writer request.
			if autonomousFilesApplied(s) {
				if err := autonomousDispatchBlocked(s); err != nil {
					return s, err
				}
				if _, err := Verify(ctx, path); err != nil {
					return InspectOr(s, path, err)
				}
				continue
			}
			if len(s.Explorations) == 0 {
				if err := autonomousDispatchBlocked(s); err != nil {
					return s, err
				}
				if _, err := RunExplorer(ctx, path, "Inspect the candidate for the approved objective. Identify the relevant files, tests, and implementation risks."); err != nil {
					return InspectOr(s, path, err)
				}
				continue
			}
			if ns, proposalStaged, err := autonomousWriterEffect(ctx, path, s); err != nil {
				return ns, err
			} else if proposalStaged {
				continue
			}
		case "VERIFYING":
			if s.Verification != nil && s.Verification.Pending {
				return s, errors.New("verification is pending and cannot be resent automatically")
			}
			if err := autonomousDispatchBlocked(s); err != nil {
				return s, err
			}
			if _, err := Verify(ctx, path); err != nil {
				return InspectOr(s, path, err)
			}
		case "REVIEWING":
			if err := autonomousDispatchBlocked(s); err != nil {
				return s, err
			}
			if _, err := RunReview(ctx, path); err != nil {
				return InspectOr(s, path, err)
			}
		case "REPAIRING":
			if s.Candidate == nil {
				return s, errors.New("repair requires a candidate")
			}
			candidateID, err := s.Candidate.ID()
			if err != nil {
				return s, err
			}
			if s.FileOutcome == "UNKNOWN" {
				if _, err := ReconcileFiles(ctx, path); err != nil {
					return InspectOr(s, path, fmt.Errorf("autonomous file effect remains UNKNOWN: %w", err))
				}
				continue
			}
			// Verify the exact confirmed repair before admitting another
			// slot. Checking the slot first would consume budget for an
			// already-repaired candidate and skip its verification.
			if autonomousRepairApplied(s) {
				if candidateID == s.RepairCandidateID {
					return s, errors.New("autonomous repair made no candidate progress")
				}
				if !autonomousVerificationCovered(s, candidateID) {
					if err := autonomousDispatchBlocked(s); err != nil {
						return s, err
					}
					if _, err := Verify(ctx, path); err != nil {
						return InspectOr(s, path, err)
					}
					continue
				}
			}
			if s.RepairCandidateID != "" && s.RepairCandidateID == candidateID && s.FileOutcome == "NOT_APPLIED" {
				return s, errors.New("autonomous repair made no candidate progress")
			}
			if s.RepairCandidateID != candidateID {
				if s.RepairAttempts >= s.Creation.Execution.MaxRepairs {
					return s, fmt.Errorf("autonomous repair bound reached (%d)", s.RepairAttempts)
				}
				if err := Append(path, "autonomous.repair-started", AutonomousRepair{Attempt: s.RepairAttempts + 1, CandidateID: candidateID}); err != nil {
					return InspectOr(s, path, err)
				}
				continue
			}
			// The repair was durably admitted for this candidate. The writer uses
			// existing verification/review feedback and its normal runtime resume.
			if ns, proposalStaged, err := autonomousWriterEffect(ctx, path, s); err != nil {
				return ns, err
			} else if proposalStaged {
				continue
			}
		case "READY":
			return s, nil
		default:
			return s, fmt.Errorf("autonomous execution cannot advance state %q", s.State)
		}
	}
	s, err := Inspect(path)
	if err != nil {
		return s, err
	}
	return s, errors.New("autonomous execution step limit reached")
}

// InspectOr keeps the durable snapshot available to callers when a normal
// operation returns an error after it may have persisted evidence.
func InspectOr(fallback Snapshot, path string, cause error) (Snapshot, error) {
	latest, err := Inspect(path)
	if err != nil {
		return fallback, errors.Join(cause, err)
	}
	return latest, cause
}
