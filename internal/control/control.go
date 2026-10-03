package control

import (
	"errors"
	"strings"

	"harness.local/engorch/internal/access"
	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/codexhost"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/effects"
	"harness.local/engorch/internal/fileeffects"
	"harness.local/engorch/internal/journal"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/runtime"
	"harness.local/engorch/internal/worktree"
)

// Creation binds immutable inputs. Nonce distinguishes repeated human objectives.
type Creation struct {
	Version       int                 `json:"version"`
	Nonce         string              `json:"nonce"`
	Repository    repository.Identity `json:"repository"`
	Objective     string              `json:"objective"`
	Config        config.Config       `json:"config"`
	HostAdmission *HostAdmission      `json:"host_admission,omitempty"`
	// Execution is absent for the interactive v1 workflow. When present it is
	// immutable run input which allows the controller to make the narrowly
	// defined machine approvals below.
	Execution *ExecutionPolicy `json:"execution,omitempty"`
}

// ExecutionPolicy is an explicit, immutable opt-in to the bounded autonomous
// workflow. It never grants authority for network publication or other
// externally visible effects; its only machine authorization is an exact plan
// and candidate-bound regular-file proposal.
//
// Context is an optional immutable task-context mode. Empty preserves the
// legacy autonomous workflow byte-for-byte; "bounded-v1" enables bounded
// observable task context admission for native role execution. No other value
// is admitted.
//
// GraphVersion selects hierarchical task graph execution. Zero preserves the
// legacy sequential workflow byte-for-byte; 1 enables autonomous graph
// execution with bounded parallel read-only explorers. MaxParallel bounds
// explorer concurrency between 1 and 8 inclusive; zero preserves legacy
// identity and means sequential (1) for old runs.
type ExecutionPolicy struct {
	Mode         string `json:"mode"`
	MaxRepairs   int    `json:"max_repairs"`
	Context      string `json:"context,omitempty"`
	GraphVersion int    `json:"graph_version,omitempty"`
	MaxParallel  int    `json:"max_parallel,omitempty"`
	// RepairPlanningVersion opts new graph runs into candidate-bound design
	// tasks that refine never-started repair write paths within original scope.
	RepairPlanningVersion int `json:"repair_planning_version,omitempty"`
	// ParallelImplementationVersion opts new graph runs into a bounded static
	// cohort of at most two independent implementation writers. Their proposals
	// are collected on one candidate and applied through one aggregate effect.
	ParallelImplementationVersion int `json:"parallel_implementation_version,omitempty"`
}

// Validate admits only the bounded autonomous workflow with a repair budget
// between 0 and 8 inclusive. Context, when present, must be "bounded-v1".
// GraphVersion must be 0 (legacy) or 1 (graph); MaxParallel must be 0
// (legacy/unspecified) or 1..8. Graph version 1 requires an explicit 1..8
// bound; legacy runs keep MaxParallel 0 for identical canonical identity.
// RepairPlanningVersion 1 requires graph execution and a matching graph-v3
// planner contract at creation replay; zero preserves the prior repair recipe.
func (p ExecutionPolicy) Validate() error {
	if p.Mode != "autonomous-v1" || p.MaxRepairs < 0 || p.MaxRepairs > 8 {
		return errors.New("invalid execution policy")
	}
	if p.Context != "" && p.Context != taskContextBoundedV1 {
		return errors.New("invalid execution task context")
	}
	if p.GraphVersion != 0 && p.GraphVersion != 1 {
		return errors.New("invalid execution graph version")
	}
	if p.MaxParallel < 0 || p.MaxParallel > 8 {
		return errors.New("invalid execution parallelism")
	}
	if p.GraphVersion == 1 && (p.MaxParallel < 1 || p.MaxParallel > 8) {
		return errors.New("graph execution requires max parallel 1..8")
	}
	if p.GraphVersion == 0 && p.MaxParallel > 1 {
		return errors.New("sequential execution cannot request parallelism")
	}
	if p.RepairPlanningVersion != 0 && p.RepairPlanningVersion != 1 {
		return errors.New("invalid repair planning version")
	}
	if p.RepairPlanningVersion == 1 && p.GraphVersion != 1 {
		return errors.New("repair planning requires graph execution")
	}
	if p.ParallelImplementationVersion != 0 && p.ParallelImplementationVersion != 1 {
		return errors.New("invalid parallel implementation version")
	}
	if p.ParallelImplementationVersion == 1 && (p.GraphVersion != 1 || p.RepairPlanningVersion != 1 || p.Context != taskContextBoundedV1) {
		return errors.New("parallel implementation requires graph execution, bounded task context, and repair planning")
	}
	return nil
}

// EffectiveMaxParallel returns the bounded explorer concurrency for graph
// runs. Legacy (GraphVersion 0) is always sequential. Graph runs default to
// the validated MaxParallel bound.
func (p ExecutionPolicy) EffectiveMaxParallel() int {
	if p.GraphVersion != 1 {
		return 1
	}
	if p.MaxParallel < 1 {
		return 1
	}
	if p.MaxParallel > 8 {
		return 8
	}
	return p.MaxParallel
}

// GraphEnabled reports whether hierarchical task graph execution applies.
// Nil or legacy policies remain sequential.
func (p *ExecutionPolicy) GraphEnabled() bool {
	return p != nil && p.GraphVersion == 1
}

func parallelImplementationEnabled(s Snapshot) bool {
	return s.Creation.Execution != nil && s.Creation.Execution.ParallelImplementationVersion == 1
}

// MachineApproval records that the immutable execution policy, rather than a
// human, admitted the exact planner result.
type MachineApproval struct {
	PlanID   string `json:"plan_id"`
	PolicyID string `json:"policy_id"`
	Actor    string `json:"actor"`
}

// Snapshot is reconstructed state, never independent authority to append effects.
type Snapshot struct {
	CandidateIndex     *CandidateIndexObservation         `json:"candidate_index,omitempty"`
	PlannerAccess      *access.Intent                     `json:"planner_access,omitempty"`
	ModelAccess        []ModelAccessState                 `json:"model_access,omitempty"`
	Draft              *DraftState                        `json:"draft,omitempty"`
	Push               *PushState                         `json:"push,omitempty"`
	Explorations       []ExplorerRecord                   `json:"explorations,omitempty"`
	ExplorerHost       *ExplorerHostState                 `json:"explorer_host,omitempty"`
	ExplorerRuns       map[string]ExplorerHostState       `json:"explorer_runs,omitempty"`
	Commit             *CommitState                       `json:"commit,omitempty"`
	ReviewHost         *ReviewHostState                   `json:"review_host,omitempty"`
	Review             *ReviewRecord                      `json:"review,omitempty"`
	WriterHost         *WriterHostState                   `json:"writer_host,omitempty"`
	WriterProposal     *WriterRecord                      `json:"writer_proposal,omitempty"`
	GraphWriterHosts   map[string]WriterHostState         `json:"graph_writer_hosts,omitempty"`
	GraphWriterResults map[string]GraphWriterRecord       `json:"graph_writer_results,omitempty"`
	GraphWriterBatch   *GraphWriterBatchRecord            `json:"graph_writer_batch,omitempty"`
	RIProducer         *RIProducerState                   `json:"ri_producer"`
	RIPublish          *RIPublishState                    `json:"ri_publish"`
	RIImport           *RIImportState                     `json:"ri_import"`
	RILexical          *RILexicalState                    `json:"ri_lexical,omitempty"`
	RILexicalOverlay   *RILexicalOverlayState             `json:"ri_lexical_overlay,omitempty"`
	RunID              string                             `json:"run_id"`
	State              string                             `json:"state"`
	Creation           Creation                           `json:"creation"`
	PlanID             string                             `json:"plan_id"`
	Plan               *runtime.Result                    `json:"plan"`
	ApprovedBy         string                             `json:"approved_by"`
	MachineApproval    *MachineApproval                   `json:"machine_approval,omitempty"`
	RepairAttempts     int                                `json:"repair_attempts,omitempty"`
	RepairCandidateID  string                             `json:"repair_candidate_id,omitempty"`
	WorkspaceIntent    *worktree.Request                  `json:"workspace_intent"`
	Workspace          *worktree.Binding                  `json:"workspace"`
	Candidate          *worktree.Candidate                `json:"candidate"`
	WorkspaceOutcome   string                             `json:"workspace_outcome"`
	FileIntent         *FileIntent                        `json:"file_intent"`
	FileReceipt        *FileReceipt                       `json:"file_receipt"`
	FileOutcome        string                             `json:"file_outcome"`
	FileRecovery       *RecoveryIntent                    `json:"file_recovery"`
	Verification       *VerificationState                 `json:"verification"`
	PlannerHost        *codexhost.Launch                  `json:"planner_host"`
	PlannerHostReady   bool                               `json:"planner_host_ready"`
	PlannerHostReceipt *codexhost.Receipt                 `json:"planner_host_receipt"`
	PlannerReceipt     *PlannerReceipt                    `json:"planner_receipt"`
	PlannerProvider    *providerDispatchReceipt           `json:"planner_provider,omitempty"`
	ProviderRuntime    map[string]providerDispatchReceipt `json:"provider_runtime,omitempty"`
	AgentDispatch      map[string]AgentDispatchState      `json:"agent_dispatch,omitempty"`
	TaskContexts       []TaskContextRecord                `json:"task_contexts,omitempty"`
	Graph              *GraphState                        `json:"graph,omitempty"`
	Lifecycle          LifecycleState                     `json:"lifecycle"`
}

// Approval is explicit human input for one exact plan, not a model decision.
type Approval struct {
	PlanID string `json:"plan_id"`
	Actor  string `json:"actor"`
}

// Replay establishes all transition invariants from the complete journal. It
// rejects unknown events rather than returning a permissive partial state.
func Replay(events []journal.Event) (Snapshot, error) {
	s := Snapshot{}
	seenEffects := map[string]bool{}
	for _, e := range events {
		if err := admitLifecycleEvent(s, e.Kind); err != nil {
			return s, err
		}
		if !lifecycleControlEvent(e.Kind) && s.Draft != nil && s.Draft.LeaseRecovery != nil && s.Draft.LeaseRecovery.Outcome == "UNKNOWN" && e.Kind != "draft.observed" && e.Kind != "draft.lease-intent" && e.Kind != "draft.lease-observed" {
			return s, errors.New("draft lease recovery UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.Draft != nil && s.Draft.Outcome == "UNKNOWN" && e.Kind != "draft.observed" && e.Kind != "draft.lease-intent" && e.Kind != "draft.lease-observed" {
			return s, errors.New("draft UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.Push != nil && s.Push.LeaseRecovery != nil && s.Push.LeaseRecovery.Outcome == "UNKNOWN" && e.Kind != "push.observed" && e.Kind != "push.lease-intent" && e.Kind != "push.lease-observed" {
			return s, errors.New("push lease recovery UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.Push != nil && s.Push.Outcome == "UNKNOWN" && e.Kind != "push.observed" && e.Kind != "push.lease-intent" && e.Kind != "push.lease-observed" {
			return s, errors.New("push UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.Commit != nil && s.Commit.LeaseRecovery != nil && s.Commit.LeaseRecovery.Outcome == "UNKNOWN" && e.Kind != "commit.observed" && e.Kind != "commit.lease-observed" && e.Kind != "commit.lease-intent" {
			return s, errors.New("lease recovery UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.Commit != nil && s.Commit.Outcome == "UNKNOWN" && e.Kind != "commit.observed" && e.Kind != "commit.lease-intent" && e.Kind != "commit.lease-observed" && e.Kind != "commit.recovery-intent" {
			return s, errors.New("commit UNKNOWN; reconciliation required before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.RIProducer != nil && s.RIProducer.Outcome == "UNKNOWN" && s.RIProducer.Closure == nil && e.Kind != "ri.producer-observed" && e.Kind != "ri.producer-closed" {
			return s, errors.New("RI producer UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.RIPublish != nil && s.RIPublish.Outcome == "UNKNOWN" && e.Kind != "ri.publish-observed" && e.Kind != "ri.publish-recovery-intent" {
			return s, errors.New("RI publication UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.RIImport != nil && s.RIImport.Outcome == "UNKNOWN" && e.Kind != "ri.import-observed" {
			return s, errors.New("RI import UNKNOWN; reconcile before further events")
		}
		if !lifecycleControlEvent(e.Kind) && s.RILexical != nil && s.RILexical.Outcome == "UNKNOWN" && e.Kind != "ri.lexical-observed" {
			return s, errors.New("RI lexical UNKNOWN; reconciliation required")
		}
		if !lifecycleControlEvent(e.Kind) && s.RILexicalOverlay != nil && s.RILexicalOverlay.Outcome == "UNKNOWN" && e.Kind != "ri.overlay-observed" {
			return s, errors.New("RI overlay UNKNOWN; reconciliation required")
		}
		switch e.Kind {
		case "agent.dispatch-admitted", "agent.dispatch-observed":
			if err := replayAgentDispatch(&s, e); err != nil {
				return s, err
			}
		case "ri.overlay-intent", "ri.overlay-observed":
			if err := replayLexicalOverlay(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "ri.lexical-intent", "ri.lexical-observed":
			if err := replayRILexical(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "planning.access-intent":
			if err := replayPlanningAccess(&s, e); err != nil {
				return s, err
			}
		case "model.access-intent", "model.access-receipt":
			if err := replayModelAccess(&s, e); err != nil {
				return s, err
			}
		case "draft.lease-intent", "draft.lease-observed":
			if err := replayDraftLease(&s, e); err != nil {
				return s, err
			}
		case "draft.intent", "draft.observed":
			if err := replayDraft(&s, e); err != nil {
				return s, err
			}
		case "push.lease-intent", "push.lease-observed":
			if err := replayPushLease(&s, e); err != nil {
				return s, err
			}
		case "push.intent", "push.observed":
			if err := replayPush(&s, e); err != nil {
				return s, err
			}
		case "explorer.host-intent", "explorer.host-ready", "explorer.host-observed", "explorer.runtime-observed":
			if err := replayExplorerHost(&s, e); err != nil {
				return s, err
			}
			if e.Kind == "explorer.runtime-observed" {
				var receipt ExplorerRuntimeReceipt
				if err := canonical.Decode(e.Payload, &receipt); err != nil {
					return s, err
				}
				host, ok := explorerRunForInvocation(s, receipt.InvocationID)
				if !ok || host.RuntimeReceipt == nil {
					return s, errors.New("exact explorer runtime receipt required")
				}
				r := host.RuntimeReceipt
				if err := requireCompletedModelAccess(s, host.Intent.Invocation, r.JournalHead, r.ResultHash); err != nil {
					return s, err
				}
			}
		case "explorer.recorded":
			var record ExplorerRecord
			if err := canonical.Decode(e.Payload, &record); err != nil {
				return s, err
			}
			if err := replayExplorer(&s, record); err != nil {
				return s, err
			}
		case "commit.recovery-intent":
			if err := replayCommitRecovery(&s, e); err != nil {
				return s, err
			}
		case "commit.lease-intent", "commit.lease-observed":
			if err := replayCommitLease(&s, e); err != nil {
				return s, err
			}
		case "commit.observed":
			if err := replayCommitObservation(&s, e); err != nil {
				return s, err
			}
		case "commit.intent":
			if err := replayCommitIntent(&s, e); err != nil {
				return s, err
			}
		case "review.host-intent", "review.host-ready", "review.host-observed", "review.runtime-observed":
			if err := replayReviewHost(&s, e); err != nil {
				return s, err
			}
			if e.Kind == "review.runtime-observed" {
				r := s.ReviewHost.RuntimeReceipt
				if err := requireCompletedModelAccess(s, s.ReviewHost.Intent.Invocation, r.JournalHead, r.ResultHash); err != nil {
					return s, err
				}
			}
		case "review.recorded":
			if err := replayReview(&s, e); err != nil {
				return s, err
			}
		case "writer.host-intent", "writer.host-ready", "writer.host-observed", "writer.runtime-observed":
			if err := replayWriterHost(&s, e); err != nil {
				return s, err
			}
			if e.Kind == "writer.runtime-observed" {
				r := s.WriterHost.RuntimeReceipt
				if err := requireCompletedModelAccess(s, s.WriterHost.Intent.Invocation, r.JournalHead, r.ResultHash); err != nil {
					return s, err
				}
			}
		case "writer.proposed":
			if err := replayWriterProposal(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "graph.writer.host-intent", "graph.writer.host-ready", "graph.writer.host-observed", "graph.writer.runtime-observed":
			if err := replayGraphWriterHost(&s, e); err != nil {
				return s, err
			}
			if e.Kind == "graph.writer.runtime-observed" {
				var event GraphWriterHostEvent
				if err := canonical.Decode(e.Payload, &event); err != nil {
					return s, err
				}
				host, ok := s.GraphWriterHosts[event.TaskID]
				if !ok || host.RuntimeReceipt == nil {
					return s, errors.New("exact graph writer runtime receipt required")
				}
				if err := requireCompletedModelAccess(s, host.Intent.Invocation, host.RuntimeReceipt.JournalHead, host.RuntimeReceipt.ResultHash); err != nil {
					return s, err
				}
			}
		case "graph.writer.proposed":
			if err := replayGraphWriterProposal(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "graph.writer.batch-proposed":
			if err := replayGraphWriterBatch(&s, e); err != nil {
				return s, err
			}
		case "ri.producer-intent", "ri.producer-observed", "ri.producer-closed":
			if err := replayRIProducer(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "ri.publish-intent", "ri.publish-observed", "ri.publish-recovery-intent":
			if err := replayRIPublish(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "ri.import-intent", "ri.import-observed":
			if err := replayRIImport(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "run.created":
			if s.State != "" {
				return s, errors.New("run already exists")
			}
			var c Creation
			if err := canonical.Decode(e.Payload, &c); err != nil {
				return s, err
			}
			if c.Version != 1 || strings.TrimSpace(c.Nonce) == "" || len(c.Nonce) > 128 || c.Config.Planner.Role != "planner" {
				return s, errors.New("invalid creation")
			}
			if c.Execution != nil {
				if err := c.Execution.Validate(); err != nil {
					return s, err
				}
			}
			if err := validateRepairPlanningBinding(c); err != nil {
				return s, err
			}
			if err := c.Repository.Validate(); err != nil {
				return s, err
			}
			if err := c.Config.Validate(); err != nil {
				return s, err
			}
			if err := validateCreationHostAdmission(c); err != nil {
				return s, err
			}
			if c.Config.Repository != c.Repository.Name {
				return s, errors.New("configuration binding mismatch")
			}
			if _, err := plannerInvocation(c.Config, c.Objective); err != nil {
				return s, err
			}
			id, err := canonical.Hash("harness.run.v1", c)
			if err != nil {
				return s, err
			}
			s = Snapshot{RunID: id, State: "OBJECTIVE", Creation: c, WorkspaceOutcome: "NOT_RUN", Lifecycle: LifecycleState{Status: LifecycleActive}}
		case "run.pause-requested", "run.paused", "run.resumed", "run.cancel-requested", "run.cancelled":
			if err := replayLifecycle(&s, e); err != nil {
				return s, err
			}
		case "planning.started":
			if s.State != "OBJECTIVE" {
				return s, errors.New("planning transition rejected")
			}
			var empty struct{}
			if err := canonical.Decode(e.Payload, &empty); err != nil {
				return s, err
			}
			s.State = "PLANNING"
		case "plan.recorded":
			if s.State != "PLANNING" {
				return s, errors.New("plan transition rejected")
			}
			var r runtime.Result
			if err := canonical.Decode(e.Payload, &r); err != nil {
				return s, err
			}
			i, err := plannerInvocation(s.Creation.Config, s.Creation.Objective)
			if err != nil {
				return s, err
			}
			if err = runtime.ValidateResult(i, r, true); err != nil {
				return s, err
			}
			if s.Creation.Config.Version == 2 && s.Creation.Config.Planner.Runtime == "fake" {
				if err := validatePlanningAccessResult(s, r); err != nil {
					return s, err
				}
			}
			if s.Creation.Config.Planner.Runtime == "codex-app-server" {
				hash, err := canonical.Hash("harness.planner-result.v1", r)
				if err != nil {
					return s, err
				}
				if s.PlannerReceipt == nil || s.PlannerReceipt.ResultHash != hash {
					return s, errors.New("plan lacks exact runtime receipt")
				}
			}
			if s.Creation.Config.Planner.Runtime == "provider-api" || s.Creation.Config.Planner.Runtime == "opencode-http" {
				hash, err := canonical.Hash("harness.planner-result.v1", r)
				if err != nil {
					return s, err
				}
				if s.PlannerProvider == nil || s.PlannerProvider.ResultHash != hash || !sameCanonical(s.PlannerProvider.Result, r) {
					return s, errors.New("plan lacks exact direct provider receipt")
				}
			}
			id, err := canonical.Hash("harness.plan.v1", r)
			if err != nil {
				return s, err
			}
			s.Plan = &r
			s.PlanID = id
			s.State = "AWAITING_APPROVAL"
		case "plan.approved":
			if s.State != "AWAITING_APPROVAL" {
				return s, errors.New("approval transition rejected")
			}
			var a Approval
			if err := canonical.Decode(e.Payload, &a); err != nil {
				return s, err
			}
			if a.PlanID != s.PlanID || strings.TrimSpace(a.Actor) == "" || len(a.Actor) > 256 {
				return s, errors.New("approval does not bind exact plan and actor")
			}
			s.ApprovedBy = a.Actor
			s.State = "IMPLEMENTING"
		case "plan.autonomous-authorized":
			if s.State != "AWAITING_APPROVAL" || s.Creation.Execution == nil {
				return s, errors.New("autonomous approval transition rejected")
			}
			var a MachineApproval
			if err := canonical.Decode(e.Payload, &a); err != nil {
				return s, err
			}
			policyID, err := canonical.Hash("harness.execution-policy.v1", *s.Creation.Execution)
			if err != nil {
				return s, err
			}
			if a.PlanID != s.PlanID || a.PolicyID != policyID || a.Actor != "fabric:autonomous" {
				return s, errors.New("machine approval binding mismatch")
			}
			s.MachineApproval = &a
			s.ApprovedBy = a.Actor
			s.State = "IMPLEMENTING"
		case "autonomous.repair-started":
			if s.State != "REPAIRING" || s.Creation.Execution == nil {
				return s, errors.New("autonomous repair transition rejected")
			}
			var repair AutonomousRepair
			if err := canonical.Decode(e.Payload, &repair); err != nil {
				return s, err
			}
			if s.Candidate == nil {
				return s, errors.New("autonomous repair requires a candidate")
			}
			candidateID, err := s.Candidate.ID()
			if err != nil || repair.CandidateID != candidateID || repair.Attempt != s.RepairAttempts+1 || repair.Attempt > s.Creation.Execution.MaxRepairs || repair.CandidateID == s.RepairCandidateID {
				return s, errors.New("autonomous repair bound rejected")
			}
			s.RepairAttempts = repair.Attempt
			s.RepairCandidateID = repair.CandidateID
		case "candidate.index-observed":
			if err := replayCandidateIndex(&s, e.Payload); err != nil {
				return s, err
			}
		case "workspace.intent":
			if s.State != "IMPLEMENTING" || s.WorkspaceIntent != nil || s.Workspace != nil {
				return s, errors.New("workspace intent transition rejected")
			}
			var r worktree.Request
			if err := canonical.Decode(e.Payload, &r); err != nil {
				return s, err
			}
			if err := r.Validate(); err != nil {
				return s, err
			}
			expectedStateRoot, err := controllerNamespace(s.Creation)
			if err != nil || r.RunID != s.RunID || r.Source != s.Creation.Repository || r.CandidateIdentity != s.Creation.Config.CandidateIdentity || r.ControllerStateRoot != expectedStateRoot {
				return s, errors.New("workspace intent binding mismatch")
			}
			s.WorkspaceIntent = &r
			s.WorkspaceOutcome = "UNKNOWN"
		case "workspace.confirmed":
			if s.State != "IMPLEMENTING" || s.WorkspaceIntent == nil || s.Workspace != nil {
				return s, errors.New("workspace receipt transition rejected")
			}
			var receipt WorkspaceReceipt
			if err := canonical.Decode(e.Payload, &receipt); err != nil {
				return s, err
			}
			if receipt.Binding.Request != *s.WorkspaceIntent {
				return s, errors.New("workspace receipt substitution")
			}
			id, err := receipt.Binding.ID()
			if err != nil {
				return s, err
			}
			if _, err := receipt.Candidate.ID(); err != nil {
				return s, err
			}
			expectedVersion := 1
			if s.Creation.Config.CandidateIdentity == "semantic-index-v2" {
				expectedVersion = 2
			}
			if receipt.Candidate.Version != expectedVersion {
				return s, errors.New("candidate identity version differs from authorized config")
			}
			if receipt.Candidate.WorktreeID != id || receipt.Candidate.Head != s.Creation.Repository.Commit {
				return s, errors.New("workspace candidate binding mismatch")
			}
			s.Workspace = &receipt.Binding
			s.Candidate = &receipt.Candidate
			s.WorkspaceOutcome = "CONFIRMED"
		case "files.intent":
			if err := filesAllowed(s); err != nil {
				return s, err
			}
			var intent FileIntent
			if err := canonical.Decode(e.Payload, &intent); err != nil {
				return s, err
			}
			expected, err := preparedFiles(s, intent.Prepared.Proposal)
			if err != nil {
				return s, err
			}
			if expected.Intent != intent.Prepared.Intent {
				return s, errors.New("file effect intent binding mismatch")
			}
			if err = validateFileAuthorization(s, intent.Authorization, expected.Intent); err != nil {
				return s, err
			}
			id, err := expected.Intent.ID()
			if err != nil {
				return s, err
			}
			if seenEffects[id] {
				return s, errors.New("effect intent already attempted")
			}
			seenEffects[id] = true
			s.FileIntent = &intent
			s.FileRecovery = nil
			s.FileReceipt = nil
			s.FileOutcome = "UNKNOWN"
		case "files.recovery-intent":
			var intent RecoveryIntent
			if err := canonical.Decode(e.Payload, &intent); err != nil {
				return s, err
			}
			expected, err := preparedRecovery(s, intent.Prepared.Recovery)
			if err != nil {
				return s, err
			}
			if expected.Intent != intent.Prepared.Intent {
				return s, errors.New("recovery effect binding mismatch")
			}
			if err = intent.Authorization.Validate(expected.Intent); err != nil {
				return s, err
			}
			id, err := expected.Intent.ID()
			if err != nil {
				return s, err
			}
			if seenEffects[id] {
				return s, errors.New("recovery already attempted")
			}
			if s.FileReceipt == nil || s.FileReceipt.Observation.Candidate == nil || *s.FileReceipt.Observation.Candidate != intent.Prepared.Recovery.Observed {
				return s, errors.New("fresh reconciliation receipt required before recovery")
			}
			seenEffects[id] = true
			s.FileRecovery = &intent
			s.FileReceipt = nil
		case "files.observed":
			if s.FileIntent == nil || s.FileOutcome != "UNKNOWN" {
				return s, errors.New("no pending file effect")
			}
			var receipt FileReceipt
			if err := canonical.Decode(e.Payload, &receipt); err != nil {
				return s, err
			}
			if len(receipt.Observation.Error) > 2048 || receipt.Observation.Candidate == nil && receipt.Observation.Error == "" {
				return s, errors.New("invalid file observation")
			}
			if receipt.Observation.Candidate != nil {
				if _, err := receipt.Observation.Candidate.ID(); err != nil {
					return s, err
				}
			}
			h, err := canonical.Hash("harness.file-observation.v1", receipt.Observation)
			if err != nil {
				return s, err
			}
			if h != receipt.Receipt.ObservationHash {
				return s, errors.New("file observation hash mismatch")
			}
			outcome, err := effects.Outcome(activeFileIntent(s), &receipt.Receipt)
			if err != nil {
				return s, err
			}
			if outcome != fileeffects.Classify(s.FileIntent.Prepared.Proposal, receipt.Observation.Candidate) {
				return s, errors.New("file outcome not established by observation")
			}
			s.FileReceipt = &receipt
			s.FileOutcome = outcome
			if outcome == "CONFIRMED" {
				after := s.FileIntent.Prepared.Proposal.After
				s.Candidate = &after
			}
		case "planning.host-intent", "planning.host-ready", "planning.host-observed", "planning.runtime-observed":
			if err := replayPlanner(&s, e); err != nil {
				return s, err
			}
		case "planning.provider-observed":
			if err := replayPlannerProvider(&s, e); err != nil {
				return s, err
			}
		case "role.provider-observed":
			if err := replayRoleProvider(&s, e); err != nil {
				return s, err
			}
		case "verification.planned", "verification.started", "verification.observed", "verification.closed":
			if err := replayVerification(&s, e, seenEffects); err != nil {
				return s, err
			}
		case "task.context-admitted":
			if err := replayTaskContext(&s, e); err != nil {
				return s, err
			}
		case "graph.recorded", "graph.progress", "graph.revised":
			if err := replayGraph(&s, e); err != nil {
				return s, err
			}
		default:
			return s, errors.New("unknown controller event")
		}
	}
	return s, nil
}

func validateRepairPlanningBinding(c Creation) error {
	version := 0
	if c.Execution != nil {
		version = c.Execution.RepairPlanningVersion
	}
	parallelVersion := 0
	if c.Execution != nil {
		parallelVersion = c.Execution.ParallelImplementationVersion
	}
	contractV3 := c.Config.PlannerContract == plannerContractGraphV3
	contractV4 := c.Config.PlannerContract == "plan-graph-v4"
	if version == 1 && !(contractV3 && parallelVersion == 0 || contractV4 && parallelVersion == 1) ||
		version == 0 && (contractV3 || contractV4 || parallelVersion != 0) {
		return errors.New("repair planning policy and planner contract must be enabled together")
	}
	if (parallelVersion == 1) != contractV4 {
		return errors.New("parallel implementation policy and planner contract must be enabled together")
	}
	return nil
}

// Append validates semantic state inside journal.Append's exclusive lock. This
// prevents two callers from both authorizing transitions against a stale read.
func Append(path, kind string, payload any) error {
	if kind == "model.access-receipt" {
		return errors.New("model access receipts require runtime journal reconciliation")
	}
	if kind == "run.created" {
		if err := validateControllerCreationPath(path, payload); err != nil {
			return err
		}
	}
	_, err := journal.Append(path, kind, payload, func(events []journal.Event) error {
		s, err := Replay(events)
		if err != nil {
			return err
		}
		return validateControllerJournalPath(path, s.Creation, s.RunID)
	})
	return err
}

// Inspect validates integrity and semantic history before returning a snapshot.
func Inspect(path string) (Snapshot, error) {
	events, err := journal.Read(path)
	if err != nil {
		return Snapshot{}, err
	}
	s, err := Replay(events)
	if err != nil {
		return s, err
	}
	return s, validateControllerJournalPath(path, s.Creation, s.RunID)
}
