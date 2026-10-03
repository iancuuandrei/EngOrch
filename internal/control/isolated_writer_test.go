package control

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/controllerstate"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/runtime"
)

func isolatedWriterCreation(t *testing.T) Creation {
	t.Helper()
	c := graphCreation(t, 1)
	c.Execution.RepairPlanningVersion = 1
	c.Execution.IsolatedImplementationVersion = 1
	c.Config.PlannerContract = "plan-graph-v7"
	c.Config.WriterContract = "anchored-edits-v1"
	c.Config.ExplorerContract = "json-v2"
	profileID, err := canonical.Hash("harness.isolation-writer-profile.v1", *c.Config.Writer)
	if err != nil {
		t.Fatal(err)
	}
	route := engineeringplan.RuntimeResourceKey{ProfileID: profileID, Provider: c.Config.Writer.Provider, Model: c.Config.Writer.Model}
	c.Execution.IsolationEstimate = &IsolationEstimateTemplate{CPUMilli: 1, MemoryMiB: 1, VerificationSlots: 1, RuntimeSlots: 1}
	c.Execution.IsolationCapacity = &engineeringplan.ResourceCapacity{
		CPUMilli: 1, MemoryMiB: 1, VerificationSlots: 1, TotalRuntimeSlots: 1,
		ProviderSlots: []engineeringplan.ProviderSlotLimit{{Provider: route.Provider, Slots: 1}},
		ModelSlots:    []engineeringplan.ModelSlotLimit{{Model: engineeringplan.ProviderModelKey{Provider: route.Provider, Model: route.Model}, Slots: 1}},
		RuntimeSlots:  []engineeringplan.RuntimeSlotLimit{{Runtime: route, Slots: 1}},
	}
	c.Config.ControllerStateRoot = filepath.Join(t.TempDir(), "controller-state")
	if err := c.Execution.Validate(); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestIsolatedWriterInvocationUsesConfirmedChildBindings(t *testing.T) {
	c := isolatedWriterCreation(t)
	path, s := isolatedWriterGraphFixture(t, c)
	machineAuthorizePlan(t, path, s)
	if _, err := ensureGraphRecorded(path, s); err != nil {
		t.Fatal(err)
	}
	if _, err := StartWorkspace(context.Background(), path); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareGraphIsolationCohort(context.Background(), path); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTaskIsolation(context.Background(), path, "one"); err != nil {
		t.Fatal(err)
	}
	s, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	question, err := graphWriterTaskQuestion(s, "one")
	if err != nil {
		t.Fatal(err)
	}
	if err := maybeAdmitIsolatedWriterTaskContext(context.Background(), path, "one", question); err != nil {
		t.Fatal(err)
	}
	s, err = Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := writerInvocationForTask(s, "one")
	if err != nil {
		t.Fatal(err)
	}
	var input struct {
		CandidateID        string                           `json:"candidate_id"`
		TaskContext        *TaskContextRecord               `json:"task_context"`
		ImplementationTask *writerImplementationContext     `json:"implementation_task"`
		Isolation          *isolatedWriterInvocationContext `json:"isolation"`
	}
	if err := json.Unmarshal([]byte(invocation.Input), &input); err != nil {
		t.Fatal(err)
	}
	parentID, err := s.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	binding, err := isolatedWriterBindingForTask(s, "one")
	if err != nil {
		t.Fatal(err)
	}
	childID, err := binding.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	workspaceID, err := binding.Workspace.ID()
	if err != nil {
		t.Fatal(err)
	}
	if childID == parentID || input.CandidateID != childID || input.Isolation == nil || input.Isolation.TaskID != "one" || input.Isolation.WorkspaceID != workspaceID || input.Isolation.BaseCandidateID != parentID || input.Isolation.ChildCandidateID != childID {
		t.Fatalf("writer input was not bound to confirmed child: parent=%s child=%s input=%+v", parentID, childID, input)
	}
	if input.TaskContext == nil || input.TaskContext.IsolationTaskID != "one" || input.TaskContext.CandidateID != childID {
		t.Fatalf("writer task context was not child-bound: %+v", input.TaskContext)
	}
	if input.ImplementationTask == nil || input.ImplementationTask.ID != "one" || len(input.ImplementationTask.WritePaths) != 1 || input.ImplementationTask.WritePaths[0] != "file.txt" {
		t.Fatalf("writer task scope mismatch: %+v", input.ImplementationTask)
	}
	latest, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Candidate == nil {
		t.Fatal("parent candidate disappeared while building isolated invocation")
	}
	latestParentID, err := latest.Candidate.ID()
	if err != nil || latestParentID != parentID {
		t.Fatalf("isolated invocation changed parent candidate binding: got %s want %s (%v)", latestParentID, parentID, err)
	}
	if got, err := writerInvocationForTaskWithIsolation(s, "one", &isolatedWriterBinding{TaskID: "substituted"}); err == nil || got.ID != "" {
		t.Fatal("substituted child binding produced an invocation")
	}
}

func isolatedWriterGraphFixture(t *testing.T, c Creation) (string, Snapshot) {
	t.Helper()
	autonomousGitInit(t, c.Repository.Root)
	repo, err := repository.Discover(context.Background(), c.Repository.Root, c.Config.Repository)
	if err != nil {
		t.Fatal(err)
	}
	c.Repository = repo
	runID, err := canonical.Hash("harness.run.v1", c)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := controllerstate.Resolve(c.Config.ControllerStateRoot, c.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if err := controllerstate.Initialize(paths, c.Repository); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Runs, 0700); err != nil {
		t.Fatal(err)
	}
	path, err := paths.Run(runID)
	if err != nil {
		t.Fatal(err)
	}
	if err := Append(path, "run.created", c); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, "planning.started", struct{}{}); err != nil {
		t.Fatal(err)
	}
	graph := directGraphFixture()
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := plannerInvocationWithContextAndRecipe(c.Config, c.Objective, nil, c.Execution)
	if err != nil {
		t.Fatal(err)
	}
	model := invocation.Profile.Model
	if err := Append(path, "plan.recorded", runtime.Result{Version: 1, InvocationID: invocation.ID, Requested: invocation.Profile, ObservedModel: &model, Output: string(raw)}); err != nil {
		t.Fatal(err)
	}
	s, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, s
}
