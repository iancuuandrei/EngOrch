package memoryadmission

import "testing"

func TestNextReducesOnlyFutureWorkersAndUsesHysteresis(t *testing.T) {
	const maxWorkers = 4
	first, err := Next(nil, Observation{Status: ObservationObserved, Source: SourceLinuxProcMeminfo, AvailableMiB: 3_999, TotalMiB: 16_000}, maxWorkers, 1_000, 1_000, 500)
	if err != nil || first.Mode != ModePressured || first.EffectiveWorkers != 2 {
		t.Fatalf("unexpected baseline decision %+v, %v", first, err)
	}
	pressured, err := Next(&first, Observation{Status: ObservationObserved, Source: SourceLinuxProcMeminfo, AvailableMiB: 3_500, TotalMiB: 16_000}, maxWorkers, 1_000, 1_000, 500)
	if err != nil || pressured.Mode != ModePressured || pressured.EffectiveWorkers != 2 {
		t.Fatalf("expected two future workers under pressure: %+v, %v", pressured, err)
	}
	stillPressured, err := Next(&pressured, Observation{Status: ObservationObserved, Source: SourceLinuxProcMeminfo, AvailableMiB: 5_400, TotalMiB: 16_000}, maxWorkers, 1_000, 1_000, 500)
	if err != nil || stillPressured.Mode != ModePressured || stillPressured.EffectiveWorkers != 2 {
		t.Fatalf("hysteresis released before exit threshold: %+v, %v", stillPressured, err)
	}
	recovered, err := Next(&stillPressured, Observation{Status: ObservationObserved, Source: SourceLinuxProcMeminfo, AvailableMiB: 5_501, TotalMiB: 16_000}, maxWorkers, 1_000, 1_000, 500)
	if err != nil || recovered.Mode != ModeNormal || recovered.EffectiveWorkers != maxWorkers {
		t.Fatalf("pressure did not clear above exit threshold: %+v, %v", recovered, err)
	}
}

func TestUnavailableObservationFallsBackToSerialWithoutInventingZero(t *testing.T) {
	observation := unavailable()
	decision, err := Next(nil, observation, 3, 2_000, DefaultReserveMiB, DefaultHysteresisMiB)
	if err != nil || observation.AvailableMiB != 0 || observation.Status != ObservationUnavailable || decision.Mode != ModeUnknown || decision.EffectiveWorkers != 1 {
		t.Fatalf("unavailable observation was not conservative and honest: %+v %+v %v", observation, decision, err)
	}
}

func TestNextRejectsPolicyDriftWithinRun(t *testing.T) {
	first, err := Next(nil, Observation{Status: ObservationUnavailable, Reason: "unavailable"}, 3, 1_000, 1_000, 500)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Next(&first, Observation{Status: ObservationUnavailable, Reason: "unavailable"}, 2, 1_000, 1_000, 500); err == nil {
		t.Fatal("policy drift was accepted")
	}
}
