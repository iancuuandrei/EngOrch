package effects

import (
	"strings"
	"testing"
)

func autonomousIntent(t *testing.T, kind string) (Intent, string) {
	t.Helper()
	h := strings.Repeat("a", 64)
	i := Intent{Version: 1, RunID: h, PlanID: h, RepositoryID: h, Kind: kind, InputHash: h}
	id, err := i.ID()
	if err != nil {
		t.Fatal(err)
	}
	return i, id
}

func TestMachineAuthorizationIsFilesystemOnly(t *testing.T) {
	fs, fsID := autonomousIntent(t, "filesystem")
	policy := strings.Repeat("b", 64)
	good := Authorization{IntentID: fsID, Actor: "fabric:autonomous", Authority: "autonomous-v1", PolicyID: policy}
	if err := good.Validate(fs); err != nil {
		t.Fatalf("filesystem machine authorization rejected: %v", err)
	}
	other, otherID := autonomousIntent(t, "ri_producer")
	badKind := Authorization{IntentID: otherID, Actor: "fabric:autonomous", Authority: "autonomous-v1", PolicyID: policy}
	if err := badKind.Validate(other); err == nil {
		t.Fatal("machine authority admitted non-filesystem effect")
	}
	for _, tc := range []Authorization{
		{IntentID: fsID, Actor: "human", Authority: "autonomous-v1", PolicyID: policy},
		{IntentID: fsID, Actor: "fabric:autonomous", Authority: "other", PolicyID: policy},
		{IntentID: fsID, Actor: "fabric:autonomous", Authority: "autonomous-v1", PolicyID: ""},
		{IntentID: strings.Repeat("0", 64), Actor: "fabric:autonomous", Authority: "autonomous-v1", PolicyID: policy},
	} {
		if err := tc.Validate(fs); err == nil {
			t.Fatalf("invalid machine authorization admitted: %+v", tc)
		}
	}
	operator := Authorization{IntentID: fsID, Actor: "explicit-operator"}
	if err := operator.Validate(fs); err != nil {
		t.Fatalf("operator filesystem approval rejected: %v", err)
	}
	changed := fs
	changed.InputHash = strings.Repeat("c", 64)
	if err := operator.Validate(changed); err == nil {
		t.Fatal("operator approval survived input change")
	}
	if err := good.Validate(changed); err == nil {
		t.Fatal("machine approval survived input change")
	}
}

func TestMachineOutcomeNeverAuthorizesRetry(t *testing.T) {
	fs, fsID := autonomousIntent(t, "filesystem")
	if outcome, err := Outcome(fs, nil); err != nil || outcome != "UNKNOWN" {
		t.Fatalf("missing receipt must stay UNKNOWN: %s %v", outcome, err)
	}
	h := strings.Repeat("a", 64)
	confirmed := &Receipt{Version: 1, IntentID: fsID, Outcome: "CONFIRMED", ObservationHash: h}
	if outcome, err := Outcome(fs, confirmed); err != nil || outcome != "CONFIRMED" {
		t.Fatalf("valid confirmation rejected: %s %v", outcome, err)
	}
	changed := fs
	changed.InputHash = strings.Repeat("d", 64)
	if _, err := Outcome(changed, confirmed); err == nil {
		t.Fatal("receipt applied to another intent")
	}
	bad := *confirmed
	bad.ObservationHash = ""
	if _, err := Outcome(fs, &bad); err == nil {
		t.Fatal("confirmation without observation binding")
	}
}
