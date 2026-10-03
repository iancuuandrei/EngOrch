package control

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/fileeffects"
	"harness.local/engorch/internal/runtime"
	"harness.local/engorch/internal/taskscheduler"
	"harness.local/engorch/internal/worktree"
)

func repairDesignFixture(t *testing.T, scope, proposed []string, candidateSubstitution bool) (Snapshot, engineeringplan.Graph) {
	t.Helper()
	initialWrite := "bool.go"
	for _, allowed := range scope {
		if allowed == "src" {
			initialWrite = "src/bool.go"
		}
	}
	initial := engineeringplan.Graph{Version: engineeringplan.Version, Mode: engineeringplan.ModeGraph, Summary: "bounded repair", Tasks: []engineeringplan.Task{
		{ID: "impl", Kind: engineeringplan.Implementation, Title: "Initial implementation", ScopePaths: scope, WritePaths: []string{initialWrite}, ExpectedEvidence: []engineeringplan.Evidence{{Kind: "file", Description: "candidate"}}, EstimatedSeconds: 20},
		{ID: "verify", Kind: engineeringplan.Verification, Title: "Verify", Dependencies: []string{"impl"}, ScopePaths: scope, ExpectedEvidence: []engineeringplan.Evidence{{Kind: "test", Description: "native"}}, EstimatedSeconds: 10},
		{ID: "review", Kind: engineeringplan.Review, Title: "Review", Dependencies: []string{"verify"}, ScopePaths: scope, ExpectedEvidence: []engineeringplan.Evidence{{Kind: "review", Description: "native"}}, EstimatedSeconds: 10},
	}}
	acceptedJSON, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	initial.Tasks[0].Completed = true
	initial.Tasks[0].Attempts = []engineeringplan.Attempt{{ID: "attempt-1", Outcome: engineeringplan.AttemptCompleted}}
	initial.Tasks[1].Attempts = []engineeringplan.Attempt{{ID: "attempt-1", Outcome: engineeringplan.AttemptFailed}}
	failed := GraphTaskEvidence{TaskID: "verify", AttemptID: "attempt-1", Outcome: "failed", VerificationPlanID: strings.Repeat("e", 64)}
	graph, err := engineeringplan.RepairDesignExtension(initial, "verify", failed.VerificationPlanID, 1)
	if err != nil {
		t.Fatal(err)
	}
	var design, implementation engineeringplan.Task
	for _, task := range graph.Tasks {
		if task.Kind == engineeringplan.Design && task.ParentID == "verify" {
			design = task
		}
		if task.Kind == engineeringplan.Implementation && task.ParentID == "verify" {
			implementation = task
		}
	}
	for i := range graph.Tasks {
		if graph.Tasks[i].ID == design.ID {
			graph.Tasks[i].Completed = true
			graph.Tasks[i].Attempts = []engineeringplan.Attempt{{ID: "attempt-1", Outcome: engineeringplan.AttemptCompleted}}
		}
	}
	candidate := worktree.Candidate{Version: 1, WorktreeID: strings.Repeat("a", 64), Head: strings.Repeat("b", 40), IndexHash: strings.Repeat("c", 64), FilesHash: strings.Repeat("d", 64), FileCount: 1}
	candidateID, err := candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	recordCandidateID := candidateID
	if candidateSubstitution {
		recordCandidateID = strings.Repeat("f", 64)
	}
	output, err := json.Marshal(Exploration{CandidateID: recordCandidateID, Summary: "Generated wrapper owns methods", Paths: proposed})
	if err != nil {
		t.Fatal(err)
	}
	const explorerID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	creation := Creation{Objective: "Implement behavior without generated-code drift", Config: config.Config{PlannerContract: plannerContractGraphV3}, Execution: &ExecutionPolicy{Mode: "autonomous-v1", MaxRepairs: 2, Context: taskContextBoundedV1, GraphVersion: 1, MaxParallel: 1, RepairPlanningVersion: 1}}
	s := Snapshot{
		Creation: creation, PlanID: strings.Repeat("1", 64), Plan: &runtime.Result{Output: string(acceptedJSON)},
		Candidate: &candidate,
		Graph: &GraphState{PlanID: strings.Repeat("1", 64), Revision: 2, Graph: graph, Evidence: map[string]GraphTaskEvidence{
			"impl":    {TaskID: "impl", AttemptID: "attempt-1", Outcome: "completed"},
			"verify":  failed,
			design.ID: {TaskID: design.ID, AttemptID: "attempt-1", Outcome: "completed", ExplorerInvocationID: explorerID},
		}},
		Explorations: []ExplorerRecord{{Question: "", Invocation: runtime.Invocation{ID: explorerID}, Result: runtime.Result{Output: string(output)}}},
	}
	question, err := explorerQuestionForTask(s, design)
	if err != nil {
		t.Fatal(err)
	}
	s.Explorations[0].Question = question
	next, err := engineeringplan.RefineRepairWritePaths(graph, implementation.ID, proposed, scope)
	if err != nil {
		t.Fatal(err)
	}
	return s, next
}

func repairDesignFakeResult(t *testing.T, invocation runtime.Invocation, output string) runtime.Result {
	t.Helper()
	fake := &runtime.Fake{}
	result, err := fake.Execute(context.Background(), invocation)
	if err != nil {
		t.Fatal(err)
	}
	if err := fake.Close(); err != nil {
		t.Fatal(err)
	}
	result.Output = output
	return result
}

func TestRepairDesignPathRevisionWriterAndFreshGatesReachReady(t *testing.T) {
	ctx := context.Background()
	c := graphCreation(t, 1)
	c.Config.Version = 1
	c.Config.PlannerContract = plannerContractGraphV3
	c.Config.ReviewerContract = "json-v1"
	c.Execution.RepairPlanningVersion = 1
	initial := engineeringplan.Graph{Version: engineeringplan.Version, Mode: engineeringplan.ModeGraph, Summary: "generated repair fixture", Tasks: []engineeringplan.Task{
		{ID: "impl", Kind: engineeringplan.Implementation, Title: "Initial implementation", ScopePaths: []string{"."}, WritePaths: []string{"file.txt"}, ExpectedEvidence: []engineeringplan.Evidence{{Kind: "file", Description: "candidate"}}, EstimatedSeconds: 20},
		{ID: "verify", Kind: engineeringplan.Verification, Title: "Verify", Dependencies: []string{"impl"}, ScopePaths: []string{"."}, ExpectedEvidence: []engineeringplan.Evidence{{Kind: "test", Description: "native"}}, EstimatedSeconds: 10},
		{ID: "review", Kind: engineeringplan.Review, Title: "Review", Dependencies: []string{"verify"}, ScopePaths: []string{"."}, ExpectedEvidence: []engineeringplan.Evidence{{Kind: "review", Description: "native"}}, EstimatedSeconds: 10},
	}}
	path, awaiting := graphAwaitingApprovalWithGraph(t, c, initial)
	machineAuthorizePlan(t, path, awaiting)
	if _, err := StartWorkspace(ctx, path); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureGraphRecorded(path, awaiting); err != nil {
		t.Fatal(err)
	}
	if err := maybeAdmitTaskContext(ctx, path, "writer", c.Objective); err != nil {
		t.Fatal(err)
	}
	initialCandidate := applyWriterOutput(t, path, "initial implementation\n")
	if _, _, err := autonomousGraphImplementing(ctx, path, initialCandidate); err != nil {
		t.Fatalf("initial writer and native verification failed: %v", err)
	}
	s, err := Inspect(path)
	if err != nil || s.State != "REVIEWING" {
		t.Fatalf("initial passing candidate did not reach review: %s %v", s.State, err)
	}
	if err := maybeAdmitTaskContext(ctx, path, "reviewer", c.Objective); err != nil {
		t.Fatal(err)
	}
	firstReview, err := PrepareReviewInvocation(path)
	if err != nil {
		t.Fatal(err)
	}
	oldCandidate, err := s.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	changes, err := canonical.Bytes(ReviewVerdict{CandidateID: oldCandidate, VerificationPlanID: s.Verification.PlanID, Decision: "changes_requested", Findings: []ReviewFinding{{Path: "file.txt", Message: "generated wrapper must retain methods"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RecordReview(path, ReviewRecord{Invocation: firstReview, Result: repairDesignFakeResult(t, firstReview, string(changes))}); err != nil {
		t.Fatal(err)
	}
	s, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err = autonomousGraphRepairing(ctx, path, s)
	if err != nil || s.RepairAttempts != 1 || s.Graph.Revision != 2 {
		t.Fatalf("failed review did not add bounded repair design: revision=%d slot=%d err=%v", s.Graph.Revision, s.RepairAttempts, err)
	}
	var design engineeringplan.Task
	for _, task := range s.Graph.Graph.Tasks {
		if task.Kind == engineeringplan.Design && task.ParentID == "review" {
			design = task
		}
	}
	if design.ID == "" {
		t.Fatal("repair design task missing")
	}
	question, err := explorerQuestionForTask(s, design)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AdmitTaskContext(ctx, path, "explorer", question); err != nil {
		t.Fatal(err)
	}
	s, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	designInvocation, err := explorerInvocation(s, question)
	if err != nil {
		t.Fatal(err)
	}
	staticTask := taskscheduler.TaskSpec{ID: design.ID, RunID: s.RunID, ControllerPath: path, Operation: taskscheduler.OperationExplorer, Input: question, InvocationID: designInvocation.ID}
	claim := taskscheduler.Claim{Task: staticTask}
	if !v1StaticGraphExplorerStateAllowed(s) || !isStaticGraphExplorerCohort(s, claim) {
		t.Fatal("opted-in V1 repair design was not admitted to its exact static explorer cohort")
	}
	if _, _, scheduled, err := scheduledInvocation(staticTask, nil); err != nil || scheduled.ID != designInvocation.ID {
		t.Fatalf("V1 repair design could not derive its exact scheduled invocation: %v", err)
	}
	legacy := s
	legacy.Creation.Execution = &ExecutionPolicy{Mode: "autonomous-v1", MaxRepairs: 2, GraphVersion: 1, MaxParallel: 1}
	if v1StaticGraphExplorerStateAllowed(legacy) || isStaticGraphExplorerCohort(legacy, claim) {
		t.Fatal("non-opted-in V1 repair was admitted to the repair-design scheduler exception")
	}
	initialState := s
	initialState.State = "IMPLEMENTING"
	initialState.Creation.Execution = &ExecutionPolicy{Mode: "autonomous-v1", MaxRepairs: 2, GraphVersion: 1, MaxParallel: 1}
	if !v1StaticGraphExplorerStateAllowed(initialState) {
		t.Fatal("legacy V1 implementation-phase explorer admission changed")
	}
	legacyQuestion, err := explorerQuestionForTask(initialState, design)
	if err != nil {
		t.Fatal(err)
	}
	legacyInvocation, err := explorerInvocation(initialState, legacyQuestion)
	if err != nil {
		t.Fatal(err)
	}
	legacyTask := staticTask
	legacyTask.Input = legacyQuestion
	legacyTask.InvocationID = legacyInvocation.ID
	if !isStaticGraphExplorerCohort(initialState, taskscheduler.Claim{Task: legacyTask}) {
		t.Fatal("legacy V1 implementation-phase static explorer cohort admission changed")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	pathsJSON, err := canonical.Bytes(Exploration{CandidateID: candidateID, Summary: "The extension file owns the generated methods and survives regeneration.", Paths: []string{"bool_ext.go"}})
	if err != nil {
		t.Fatal(err)
	}
	designRecord := ExplorerRecord{Question: question, Invocation: designInvocation, Result: repairDesignFakeResult(t, designInvocation, string(pathsJSON))}
	if _, err := RecordExploration(path, designRecord); err != nil {
		t.Fatalf("exact candidate-bound repair design was rejected: %v", err)
	}
	s, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recordGraphProgress(path, GraphProgress{Version: 1, PlanID: s.Graph.PlanID, Digest: s.Graph.Digest, TaskID: design.ID, AttemptID: "attempt-1", Outcome: "completed", ExplorerInvocationID: designInvocation.ID}); err != nil {
		t.Fatal(err)
	}
	s, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	// The repair loop itself consumes the persisted design and appends a
	// replay-validated refinement; it must not jump straight to writer dispatch.
	refined, err := autonomousGraphRepairing(ctx, path, s)
	if err != nil || refined.Graph.Revision != 3 {
		t.Fatalf("repair design did not refine its slot: revision=%d err=%v", refined.Graph.Revision, err)
	}
	impl, ok := refined.Graph.Graph.Task("impl-repair-1")
	if !ok || !stringListsEqual(impl.WritePaths, []string{"bool_ext.go"}) {
		t.Fatalf("repair write ownership differs from design: %+v", impl)
	}
	if err := maybeAdmitTaskContext(ctx, path, "writer", c.Objective); err != nil {
		t.Fatal(err)
	}
	writerInvocation, err := PrepareWriterInvocation(path)
	if err != nil {
		t.Fatal(err)
	}
	candidateID, err = refined.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	writerOutput, err := canonical.Bytes(WriterProposal{CandidateID: candidateID, Changes: []fileeffects.Change{{Path: "bool_ext.go", ContentBase64: content64("func (b Bool) MarshalText() ([]byte, error) { return []byte(b.String()), nil }\n")}}})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := RecordWriterProposal(ctx, path, writerInvocation, repairDesignFakeResult(t, writerInvocation, string(writerOutput)))
	if err != nil {
		t.Fatalf("refined in-scope writer proposal was rejected: %v", err)
	}
	refined, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := autonomousFileAuthorization(refined, proposal.Prepared.Intent)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyFiles(ctx, path, proposal.Prepared, authorization); err != nil {
		t.Fatal(err)
	}
	// Simulate restart at the confirmed file receipt before graph progress.
	s, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err = autonomousGraphRepairing(ctx, path, s)
	if err != nil || s.State != "REVIEWING" {
		t.Fatalf("confirmed repair did not resume through fresh verification: state=%s err=%v", s.State, err)
	}
	impl, _ = s.Graph.Graph.Task("impl-repair-1")
	if !impl.Completed || s.Graph.Evidence[impl.ID].CandidateID == "" {
		t.Fatal("restart advanced to fresh verification without implementation evidence")
	}
	if err := maybeAdmitTaskContext(ctx, path, "reviewer", c.Objective); err != nil {
		t.Fatal(err)
	}
	freshReview, err := PrepareReviewInvocation(path)
	if err != nil {
		t.Fatal(err)
	}
	candidateID, err = s.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	reviewOutput, err := canonical.Bytes(ReviewVerdict{CandidateID: candidateID, VerificationPlanID: s.Verification.PlanID, Decision: "approve", Findings: []ReviewFinding{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RecordReview(path, ReviewRecord{Invocation: freshReview, Result: repairDesignFakeResult(t, freshReview, string(reviewOutput))}); err != nil {
		t.Fatal(err)
	}
	if err := recordGraphReviewProgress(path); err != nil {
		t.Fatal(err)
	}
	ready, err := Inspect(path)
	if err != nil || ready.State != "READY" {
		t.Fatalf("fresh repair gates did not reach READY: state=%s err=%v", ready.State, err)
	}
	if err := requireGraphReady(ready); err != nil {
		t.Fatal("READY lacks complete actual repair evidence", err)
	}
}

func TestRepairWriteRefinementUsesExactCandidateBoundDesignPaths(t *testing.T) {
	path := "internal/gen-atomicwrapper/wrapper.tmpl"
	s, next := repairDesignFixture(t, []string{"."}, []string{path}, false)
	if err := validateRepairWriteRefinement(s, next); err != nil {
		t.Fatalf("recorded in-scope repair design rejected: %v", err)
	}
	changed := next
	changed.Tasks = append([]engineeringplan.Task(nil), next.Tasks...)
	for i := range changed.Tasks {
		if changed.Tasks[i].Kind == engineeringplan.Implementation && changed.Tasks[i].ParentID == "verify" {
			changed.Tasks[i].WritePaths = []string{"bool.go"}
		}
	}
	if err := validateRepairWriteRefinement(s, changed); err == nil || !strings.Contains(err.Error(), "exact recorded design") {
		t.Fatalf("refinement substituted different paths than the design result: %v", err)
	}
	substituted, substitutedNext := repairDesignFixture(t, []string{"."}, []string{path}, true)
	if err := validateRepairWriteRefinement(substituted, substitutedNext); err == nil || !strings.Contains(err.Error(), "current candidate") {
		t.Fatalf("design result for a different candidate was admitted: %v", err)
	}
}

func TestRepairPlanningPolicyRequiresMatchingImmutablePlannerContract(t *testing.T) {
	creation := Creation{Execution: &ExecutionPolicy{Mode: "autonomous-v1", MaxRepairs: 1, GraphVersion: 1, MaxParallel: 1, RepairPlanningVersion: 1}}
	if err := validateRepairPlanningBinding(creation); err == nil {
		t.Fatal("repair policy without versioned planner contract accepted")
	}
	creation.Config.PlannerContract = plannerContractGraphV3
	if err := validateRepairPlanningBinding(creation); err != nil {
		t.Fatal("matching repair planner contract rejected", err)
	}
	creation.Execution.RepairPlanningVersion = 0
	if err := validateRepairPlanningBinding(creation); err == nil {
		t.Fatal("new planner contract without immutable repair policy accepted")
	}
}
