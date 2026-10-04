package taskscheduler

import (
	"errors"
	"testing"
)

func TestAppendClaimWithInspectRetainsUnknownAfterPostCommitInspectFailure(t *testing.T) {
	task := TaskSpec{ID: "task", RunID: digest('a'), ControllerPath: "C:\\repo\\run", Operation: OperationWriter, InvocationID: digest('b')}
	claim := Claim{Version: 1, ScheduleID: digest('c'), Task: task, Generation: 1, ControllerHead: digest('d')}
	persisted := false
	outcome, dispatch, _, err := appendClaimWithInspect(claim, func() error {
		persisted = true
		return errors.New("simulated lock-release failure after durable append")
	}, func() (TaskState, bool, error) {
		return TaskState{}, false, errors.New("simulated inspect lock failure")
	})
	if !persisted || outcome != AdmissionClaimUnknown || dispatch || err == nil {
		t.Fatalf("ambiguous claim append was treated as absent: persisted=%v outcome=%v dispatch=%v err=%v", persisted, outcome, dispatch, err)
	}
}

func TestAppendClaimWithInspectReleasesOnlyProvenAbsentClaim(t *testing.T) {
	task := TaskSpec{ID: "task", RunID: digest('a'), ControllerPath: "C:\\repo\\run", Operation: OperationWriter, InvocationID: digest('b')}
	claim := Claim{Version: 1, ScheduleID: digest('c'), Task: task, Generation: 1, ControllerHead: digest('d')}
	outcome, dispatch, _, err := appendClaimWithInspect(claim, func() error {
		return errors.New("simulated append failure before write")
	}, func() (TaskState, bool, error) {
		return TaskState{Status: StatusReady}, true, nil
	})
	if outcome != AdmissionClaimNotRecorded || dispatch || err == nil {
		t.Fatalf("confirmed absent claim did not remain an error: outcome=%v dispatch=%v err=%v", outcome, dispatch, err)
	}

	outcome, dispatch, state, err := appendClaimWithInspect(claim, func() error {
		return errors.New("simulated lock-release failure after durable append")
	}, func() (TaskState, bool, error) {
		return TaskState{Generation: 1, Claim: &claim, Status: StatusClaimed}, true, nil
	})
	if outcome != AdmissionClaimRecorded || dispatch || err != nil || state.Generation != claim.Generation {
		t.Fatalf("exact durable claim was not recognized without a duplicate dispatch: outcome=%v dispatch=%v state=%+v err=%v", outcome, dispatch, state, err)
	}
}
