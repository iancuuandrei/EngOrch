package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/control"
	"harness.local/engorch/internal/repository"
)

const defaultAutonomousMaxRepairs = 2

const defaultAutonomousMaxParallel = 3

// runCommand retains the original `run RUN` operation and adds the explicit
// autonomous objective form. The latter creates the immutable execution policy
// before any provider dispatch is considered by the controller.
func runCommand(ctx context.Context, root string, args []string, out io.Writer) error {
	if len(args) == 0 || args[0] != "--autonomous" {
		if len(args) != 1 {
			return errors.New("run requires one RUN or --autonomous OBJECTIVE")
		}
		return legacyRunCommand(ctx, root, args[0], out)
	}
	return autonomousRunCommand(ctx, root, args[1:], out)
}

func legacyRunCommand(ctx context.Context, root, id string, out io.Writer) error {
	p, err := runPath(root, id)
	if err != nil {
		return err
	}
	bound, err := control.Inspect(p)
	if err != nil {
		return err
	}
	if err := requireRunBinding(bound, root, id); err != nil {
		return err
	}
	s, err := control.StartWorkspace(ctx, p)
	if err != nil {
		return err
	}
	return output(out, s)
}

func autonomousRunCommand(ctx context.Context, root string, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("run --autonomous", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	maxRepairs := fs.Int("max-repairs", defaultAutonomousMaxRepairs, "maximum bounded repair attempts")
	maxParallel := fs.Int("max-parallel", defaultAutonomousMaxParallel, "maximum bounded parallel task workers (1 sequential, 2..8 parallel)")
	parallelWriters := fs.Bool("parallel-writers", false, "allow two independent initial implementation tasks when justified")
	prepareOnly := fs.Bool("prepare-only", false, "accept the graph and confirm its workspace, then return before explorer or writer dispatch")
	goalFile := fs.String("file", "", "read objective from file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *maxRepairs < 0 || *maxRepairs > 8 {
		return errors.New("max-repairs must be between 0 and 8")
	}
	if *maxParallel < 1 || *maxParallel > 8 {
		return errors.New("max-parallel must be between 1 and 8")
	}
	if *goalFile != "" {
		if fs.NArg() != 0 {
			return errors.New("run --autonomous --file takes no positional objective")
		}
		objective, err := planObjective(ctx, root, []string{"--file", *goalFile})
		if err != nil {
			return err
		}
		if err := validateAutonomousObjective(objective); err != nil {
			return err
		}
		return createAndRunAutonomous(ctx, root, objective, *maxRepairs, *maxParallel, *parallelWriters, *prepareOnly, out)
	}
	if fs.NArg() != 1 || fs.Arg(0) == "" {
		return errors.New("run --autonomous requires one objective or --file PATH")
	}
	if err := validateAutonomousObjective(fs.Arg(0)); err != nil {
		return err
	}
	return createAndRunAutonomous(ctx, root, fs.Arg(0), *maxRepairs, *maxParallel, *parallelWriters, *prepareOnly, out)
}

// validateAutonomousObjective rejects whitespace-only, invalid UTF-8 and
// oversized objectives before any durable run is created. The 256 KiB bound
// and nonempty UTF-8 conventions match planObjective goal-file validation.
func validateAutonomousObjective(objective string) error {
	if strings.TrimSpace(objective) == "" {
		return errors.New("autonomous objective must be nonempty text")
	}
	if !utf8.ValidString(objective) {
		return errors.New("autonomous objective must be valid UTF-8")
	}
	if len(objective) > 256<<10 {
		return errors.New("autonomous objective exceeds 256 KiB")
	}
	return nil
}

func createAndRunAutonomous(ctx context.Context, root, objective string, maxRepairs, maxParallel int, parallelWriters, prepareOnly bool, out io.Writer) error {
	if err := validateAutonomousObjective(objective); err != nil {
		return err
	}
	cfg, err := configuration(root)
	if err != nil {
		return err
	}
	identity, err := repository.Discover(ctx, root, cfg.Repository)
	if err != nil {
		return err
	}
	if filepath.Clean(identity.Root) != filepath.Clean(root) {
		return errors.New("root must be repository top level")
	}
	nonce := make([]byte, 16)
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	// New autonomous runs use hierarchical task graph execution with bounded
	// task context and design-planned repairs. The planner contract reserves a
	// conservative initial implementation scope; repair design may refine only
	// never-started repair writes within that scope. Nil/old policies remain
	// compatible and replay with their historical repair behavior.
	cfg.PlannerContract = "plan-graph-v5"
	parallelImplementationVersion := 0
	if parallelWriters {
		cfg.PlannerContract = "plan-graph-v6"
		parallelImplementationVersion = 1
		// Validate the opt-in against the configured implementation route before
		// creating a run or dispatching any planner intent.
		if err := cfg.Validate(); err != nil {
			return err
		}
	}
	if cfg.Reviewer != nil {
		cfg.ReviewerContract = "json-v1"
	}
	creation := control.Creation{
		Version:    1,
		Nonce:      hex.EncodeToString(nonce),
		Repository: identity,
		Objective:  objective,
		Config:     cfg,
		Execution:  &control.ExecutionPolicy{Mode: "autonomous-v1", MaxRepairs: maxRepairs, Context: "bounded-v1", GraphVersion: 1, MaxParallel: maxParallel, RepairPlanningVersion: 1, ParallelImplementationVersion: parallelImplementationVersion},
	}
	creation, err = bindCurrentHost(ctx, creation)
	if err != nil {
		return err
	}
	id, err := canonical.Hash("harness.run.v1", creation)
	if err != nil {
		return err
	}
	p, err := initializeRunPath(root, id, cfg, identity)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	if err = control.Append(p, "run.created", creation); err != nil {
		return err
	}
	if prepareOnly {
		s, err := control.PrepareAutonomous(ctx, p)
		if err != nil {
			return reportAutonomousFailure(out, p, id, err)
		}
		prepared, err := autonomousPreparedResult(s)
		if err != nil {
			return reportAutonomousFailure(out, p, id, err)
		}
		return output(out, prepared)
	}
	s, err := control.RunAutonomous(ctx, p)
	if err != nil {
		return reportAutonomousFailure(out, p, id, err)
	}
	return output(out, s)
}

type autonomousPrepared struct {
	Status        string `json:"status"`
	RunID         string `json:"run_id"`
	State         string `json:"state"`
	PlanID        string `json:"plan_id"`
	CandidateID   string `json:"candidate_id"`
	GraphDigest   string `json:"graph_digest"`
	GraphRevision int    `json:"graph_revision"`
}

func autonomousPreparedResult(s control.Snapshot) (autonomousPrepared, error) {
	if s.State != "IMPLEMENTING" || s.Workspace == nil || s.Candidate == nil || s.WorkspaceOutcome != "CONFIRMED" || s.Graph == nil || s.Graph.PlanID != s.PlanID || s.Graph.Revision != 1 {
		return autonomousPrepared{}, errors.New("autonomous preparation receipt is incomplete")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return autonomousPrepared{}, err
	}
	return autonomousPrepared{
		Status: "PREPARED", RunID: s.RunID, State: s.State, PlanID: s.PlanID,
		CandidateID: candidateID, GraphDigest: s.Graph.Digest, GraphRevision: s.Graph.Revision,
	}, nil
}

func autonomousResumeCommand(ctx context.Context, root string, args []string, out io.Writer) error {
	if len(args) > 1 {
		return errors.New("resume --autonomous accepts at most one RUN")
	}
	id := ""
	if len(args) == 1 {
		id = args[0]
	} else {
		var err error
		id, err = latestAutonomousRunID(root)
		if err != nil {
			return err
		}
	}
	p, err := runPath(root, id)
	if err != nil {
		return err
	}
	s, err := control.Inspect(p)
	if err != nil {
		return err
	}
	if err := requireRunBinding(s, root, id); err != nil {
		return err
	}
	if s.Creation.Execution == nil || s.Creation.Execution.Mode != "autonomous-v1" {
		return fmt.Errorf("run %s is not an autonomous run", id)
	}
	s, err = control.RunAutonomous(ctx, p)
	if err != nil {
		return reportAutonomousFailure(out, p, id, err)
	}
	return output(out, s)
}

type autonomousFailure struct {
	RunID         string `json:"run_id"`
	State         string `json:"state"`
	Phase         string `json:"phase"`
	BlockedReason string `json:"blocked_reason"`
}

// reportAutonomousFailure emits only the durable run identity and coarse
// controller state. Provider errors may carry credentials, so the JSON summary
// and the returned boundary error carry only the coarse blocked reason; the
// raw cause is never returned to the main stderr channel. Cancellation still
// reports through errors.Is by wrapping the corresponding sentinel.
func reportAutonomousFailure(out io.Writer, path, fallbackID string, cause error) error {
	summary := autonomousFailure{RunID: fallbackID, State: "UNKNOWN", Phase: "blocked", BlockedReason: autonomousBlockReason(cause)}
	if s, err := control.Inspect(path); err == nil {
		if s.RunID != "" {
			summary.RunID = s.RunID
		}
		summary.State = s.State
		summary.Phase = autonomousPhase(s.State)
	}
	boundary := sanitizedAutonomousError(summary, cause)
	if err := output(out, summary); err != nil {
		return errors.Join(boundary, err)
	}
	return boundary
}

// sanitizedAutonomousError binds only coarse durable identity to a useful
// message. It never wraps the raw provider cause, so credential-bearing detail
// cannot reach main stderr through the returned error.
func sanitizedAutonomousError(summary autonomousFailure, cause error) error {
	if errors.Is(cause, context.Canceled) {
		return fmt.Errorf("autonomous run %s blocked in %s (%s): %w", summary.RunID, summary.Phase, summary.BlockedReason, context.Canceled)
	}
	if errors.Is(cause, context.DeadlineExceeded) {
		return fmt.Errorf("autonomous run %s blocked in %s (%s): %w", summary.RunID, summary.Phase, summary.BlockedReason, context.DeadlineExceeded)
	}
	return fmt.Errorf("autonomous run %s blocked in %s (%s: %s)", summary.RunID, summary.Phase, summary.State, summary.BlockedReason)
}

func autonomousPhase(state string) string {
	switch state {
	case "OBJECTIVE", "PLANNING":
		return "planning"
	case "AWAITING_APPROVAL":
		return "approval"
	case "IMPLEMENTING":
		return "implementation"
	case "VERIFYING":
		return "verification"
	case "REVIEWING":
		return "review"
	case "REPAIRING":
		return "repair"
	case "READY":
		return "ready"
	default:
		return "blocked"
	}
}

func autonomousBlockReason(err error) string {
	if err == nil {
		return "unknown"
	}
	if errors.Is(err, control.ErrAutonomousVerificationNotRun) {
		return "verification_not_run"
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "unknown"), strings.Contains(message, "uncertain"), strings.Contains(message, "unresolved"):
		return "effect_requires_reconciliation"
	case strings.Contains(message, "repair bound"), strings.Contains(message, "repair budget"):
		return "repair_budget_exhausted"
	case strings.Contains(message, "cancel"):
		return "execution_interrupted"
	case strings.Contains(message, "pending"):
		return "pending_work_requires_observation"
	default:
		return "execution_blocked"
	}
}

func requireRunBinding(s control.Snapshot, root, id string) error {
	if s.RunID != id || filepath.Clean(s.Creation.Repository.Root) != filepath.Clean(root) {
		return errors.New("journal/run repository binding mismatch")
	}
	return nil
}
