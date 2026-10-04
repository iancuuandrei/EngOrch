package control

import "testing"

func TestScheduledCorrectionLookupExcludesSerialCorrection(t *testing.T) {
	s := Snapshot{RoleCorrections: []RoleSemanticCorrection{
		{Attempt: 1},
		{Attempt: 2, ScheduledTaskID: "scheduled-task"},
	}}
	if _, ok := scheduledCorrectionForTask(s, ""); ok {
		t.Fatal("serial correction matched an empty scheduler task identity")
	}
	if correction, ok := scheduledCorrectionForTask(s, "scheduled-task"); !ok || correction.Attempt != 2 {
		t.Fatal("exact scheduled correction was not retained")
	}
	if _, ok := scheduledCorrectionForTask(s, "other-task"); ok {
		t.Fatal("unrelated scheduler task matched")
	}
}
