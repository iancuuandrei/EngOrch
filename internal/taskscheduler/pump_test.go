package taskscheduler

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type skipClaimAdmissionAdapter struct{ *fakeAdapter }

func (skipClaimAdmissionAdapter) BeforeClaim(context.Context, TaskSpec) (func(AdmissionClaimOutcome) (func(), error), error) {
	return nil, nil
}

type refreshedAdmissionAdapter struct {
	probes     int
	callback   bool
	recorded   *AdmissionClaimOutcome
	dispatches int
}

func (a *refreshedAdmissionAdapter) Probe(_ context.Context, request ProbeRequest) (Evidence, error) {
	a.probes++
	status := StatusReady
	if a.probes > 1 {
		status = StatusPaused
	}
	return Evidence{RunID: request.Task.RunID, InvocationID: request.Task.InvocationID, ControllerHead: digest('b'), Status: status}, nil
}

func (a *refreshedAdmissionAdapter) Dispatch(_ context.Context, claim Claim) (Evidence, error) {
	a.dispatches++
	return admittedEvidence(claim.Task, StatusSucceeded, 'c'), nil
}

func (a *refreshedAdmissionAdapter) Reconcile(context.Context, Claim) (Evidence, error) {
	return Evidence{}, errors.New("unexpected reconcile")
}

func (a *refreshedAdmissionAdapter) BeforeClaim(context.Context, TaskSpec) (func(AdmissionClaimOutcome) (func(), error), error) {
	return func(recorded AdmissionClaimOutcome) (func(), error) {
		a.callback = true
		*a.recorded = recorded
		return nil, nil
	}, nil
}

func TestTickReleasesAdmissionWhenRefreshedProbeIsNoLongerReady(t *testing.T) {
	task := TaskSpec{ID: "task", RunID: digest('a'), ControllerPath: filepath.Join(t.TempDir(), "run"), Operation: OperationWriter, InvocationID: digest('1')}
	path := filepath.Join(t.TempDir(), "schedule")
	if _, err := Bind(path, scheduleDefinition(t, task)); err != nil {
		t.Fatal(err)
	}
	recorded := AdmissionClaimRecorded
	adapter := &refreshedAdmissionAdapter{recorded: &recorded}
	decision, err := Tick(context.Background(), path, adapter)
	if err != nil || decision.TaskID != "" || !adapter.callback || recorded != AdmissionClaimNotRecorded || adapter.dispatches != 0 {
		t.Fatalf("pre-claim reservation was not released: decision=%+v callback=%v recorded=%v dispatches=%d err=%v", decision, adapter.callback, recorded, adapter.dispatches, err)
	}
	snapshot, err := Inspect(path)
	if err != nil || snapshot.Tasks[task.ID].Generation != 0 || snapshot.Tasks[task.ID].Status != StatusReady {
		t.Fatalf("refresh stop still claimed the task: snapshot=%+v err=%v", snapshot.Tasks[task.ID], err)
	}
}

func TestTickSkipsCandidateWhenAdmissionGateFindsItStale(t *testing.T) {
	task := TaskSpec{ID: "task", RunID: digest('a'), ControllerPath: filepath.Join(t.TempDir(), "run"), Operation: OperationWriter, InvocationID: digest('1')}
	path := filepath.Join(t.TempDir(), "schedule")
	if _, err := Bind(path, scheduleDefinition(t, task)); err != nil {
		t.Fatal(err)
	}
	base := &fakeAdapter{evidence: map[string]Evidence{task.ID: readyEvidence(task, 'b')}, dispatches: map[string]int{}}
	decision, err := Tick(context.Background(), path, skipClaimAdmissionAdapter{fakeAdapter: base})
	if err != nil || decision.TaskID != "" {
		t.Fatalf("stale candidate was claimed: decision=%+v err=%v", decision, err)
	}
	snapshot, err := Inspect(path)
	if err != nil || snapshot.Tasks[task.ID].Generation != 0 || snapshot.Tasks[task.ID].Status != StatusReady || base.count(task.ID) != 0 {
		t.Fatalf("stale candidate produced a claim or dispatch: snapshot=%+v count=%d err=%v", snapshot.Tasks[task.ID], base.count(task.ID), err)
	}
}

func TestPumpDispatchesLateChildWhileParentWaits(t *testing.T) {
	parent := TaskSpec{ID: "parent", RunID: digest('a'), ControllerPath: filepath.Join(t.TempDir(), "run"), Operation: OperationPlanner, InvocationID: digest('1')}
	child := TaskSpec{ID: digest('2'), RunID: parent.RunID, ControllerPath: parent.ControllerPath, Operation: OperationExplorer, InvocationID: digest('3'), Input: "inspect"}
	path := filepath.Join(t.TempDir(), "schedule")
	if _, err := Bind(path, scheduleDefinition(t, parent)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	started, childDone := make(chan struct{}), make(chan struct{})
	adapter := &fakeAdapter{evidence: map[string]Evidence{parent.ID: readyEvidence(parent, 'b'), child.ID: readyEvidence(child, 'c')}, dispatches: map[string]int{}}
	adapter.dispatch = func(claim Claim) (Evidence, error) {
		if claim.Task.ID == parent.ID {
			close(started)
			select {
			case <-childDone:
			case <-ctx.Done():
				return Evidence{}, ctx.Err()
			}
		} else {
			close(childDone)
		}
		evidence := admittedEvidence(claim.Task, StatusSucceeded, 'e')
		adapter.set(claim.Task.ID, evidence)
		return evidence, nil
	}
	finished := make(chan error, 1)
	go func() {
		finished <- Pump(ctx, path, adapter, PumpOptions{Workers: 2, PollInterval: 10 * time.Millisecond})
	}()
	defer func() {
		cancel()
		select {
		case err := <-finished:
			if !errors.Is(err, context.Canceled) {
				t.Error(err)
			}
		case <-time.After(2 * time.Second):
			t.Error("pump did not stop")
		}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("parent did not start")
	}
	// The second worker must remain alive after finding no eligible work.
	time.Sleep(30 * time.Millisecond)
	if _, err := AddTask(path, DynamicTask{Task: child, ParentAgentID: digest('c'), AgentID: digest('d'), TurnID: child.ID}); err != nil {
		t.Fatal(err)
	}
	for {
		snapshot, err := Inspect(path)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Tasks[parent.ID].Status == StatusSucceeded && snapshot.Tasks[child.ID].Status == StatusSucceeded {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("late child did not unblock parent")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if adapter.count(parent.ID) != 1 || adapter.count(child.ID) != 1 {
		t.Fatal("duplicate dispatch")
	}
}

func TestPumpRejectsInvalidOptionsBeforeJournalAccess(t *testing.T) {
	for _, options := range []PumpOptions{{Workers: 0}, {Workers: 65}, {Workers: 1, PollInterval: time.Nanosecond}} {
		if err := Pump(context.Background(), "absent", &fakeAdapter{}, options); err == nil {
			t.Fatal("invalid options accepted")
		}
	}
}

func TestPumpStopsAllWorkersOnProbeFailure(t *testing.T) {
	task := TaskSpec{ID: "task", RunID: digest('a'), ControllerPath: filepath.Join(t.TempDir(), "run"), Operation: OperationPlanner, InvocationID: digest('1')}
	path := filepath.Join(t.TempDir(), "schedule")
	if _, err := Bind(path, scheduleDefinition(t, task)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	adapter := &fakeAdapter{evidence: map[string]Evidence{}, dispatches: map[string]int{}}
	if err := Pump(ctx, path, adapter, PumpOptions{Workers: 4}); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("probe error did not stop pump", err)
	}
	snapshot, err := Inspect(path)
	if err != nil || snapshot.Tasks[task.ID].Claim != nil || adapter.count(task.ID) != 0 {
		t.Fatal("failed probe admitted work", snapshot, err)
	}
}
