package modelpolicy

import "testing"

func TestEvaluateCalibrationSelectsCandidateFromDisjointTaskEvidence(t *testing.T) {
	calibration := fixtureCalibration()
	selection, err := EvaluateCalibration(calibration)
	if err != nil {
		t.Fatal(err)
	}
	if !selection.Selected || selection.Profile.Name != "sol" || selection.Reason != "candidate_quality_improved_train_and_holdout" {
		t.Fatalf("selection = %+v", selection)
	}
	if selection.Training["sol"].Accepted != 1 || selection.Holdout["luna"].Failed != 1 || selection.Digest == "" {
		t.Fatalf("unexpected task counts: %+v", selection)
	}
	second, err := EvaluateCalibration(fixtureCalibration())
	if err != nil || second.Digest != selection.Digest {
		t.Fatalf("digest is not deterministic: %q %q %v", selection.Digest, second.Digest, err)
	}
}

func TestEvaluateCalibrationFallsBackForUnknownNotExercisedAndDrift(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Calibration)
		reason string
	}{
		{"unknown", func(c *Calibration) { c.Training[0].Outcome = OutcomeUnknown }, "fallback_unknown_outcome"},
		{"not-exercised", func(c *Calibration) {
			row := &c.Training[0]
			row.ObservedFixerProfiles = nil
			row.FixerInvocations = 0
			row.Outcome = OutcomeNotExercised
			row.Repairs = 3
		}, "fallback_fixer_not_exercised"},
		{"drift", func(c *Calibration) { c.Holdout[1].ObservedFixerProfiles[0] = c.Baseline }, "fallback_fixer_route_drift"},
	} {
		t.Run(test.name, func(t *testing.T) {
			selection, err := EvaluateCalibration(func() Calibration { c := fixtureCalibration(); test.mutate(&c); return c }())
			if err != nil || selection.Selected || selection.Profile.Name != "luna" || selection.Reason != test.reason {
				t.Fatalf("selection = %+v, err = %v", selection, err)
			}
		})
	}
}

func TestCalibrationAllowsMatchedArmsButRejectsTrainHoldoutOverlap(t *testing.T) {
	calibration := fixtureCalibration()
	if err := calibration.Validate(); err != nil {
		t.Fatalf("matched policy arms rejected: %v", err)
	}
	calibration.Holdout[0].ObjectiveHash = calibration.Training[0].ObjectiveHash
	if err := calibration.Validate(); err == nil {
		t.Fatal("train/holdout objective overlap accepted")
	}
}

func TestEvaluateCalibrationFallsBackForUnmatchedTaskPolicy(t *testing.T) {
	for _, mutate := range []func(*Calibration){
		func(c *Calibration) { c.Training[1].TaskID = "different-task" },
		func(c *Calibration) { c.Holdout[1].NonFixerPolicyID = hash('f') },
	} {
		calibration := fixtureCalibration()
		mutate(&calibration)
		selection, err := EvaluateCalibration(calibration)
		if err != nil || selection.Selected || selection.Reason != "fallback_unmatched_task_policy" {
			t.Fatalf("selection = %+v, err = %v", selection, err)
		}
	}
}

func TestCalibrationRejectsDuplicateTaskPolicyArm(t *testing.T) {
	calibration := fixtureCalibration()
	duplicate := calibration.Training[0]
	duplicate.RunID = "different-run"
	calibration.Training = append(calibration.Training, duplicate)
	if err := calibration.Validate(); err == nil {
		t.Fatal("duplicate task-policy arm accepted")
	}
}

func TestCalibrationRejectsInvalidUsage(t *testing.T) {
	calibration := fixtureCalibration()
	negative := int64(-1)
	calibration.Training[0].Usage = &EmpiricalUsage{CostMicroUSD: &negative}
	if err := calibration.Validate(); err == nil {
		t.Fatal("negative empirical usage accepted")
	}
	calibration = fixtureCalibration()
	cached, input := int64(2), int64(1)
	calibration.Training[0].Usage = &EmpiricalUsage{InputTokens: &input, CachedInputTokens: &cached}
	if err := calibration.Validate(); err == nil {
		t.Fatal("invalid cached subset accepted")
	}
}

func TestEvaluateCalibrationFallsBackWhenEvidenceIsBelowOperatorMinimum(t *testing.T) {
	calibration := fixtureCalibration()
	calibration.MinimumPerArm = 2
	selection, err := EvaluateCalibration(calibration)
	if err != nil || selection.Selected || selection.Reason != "fallback_insufficient_quality_evidence" {
		t.Fatalf("selection = %+v, err = %v", selection, err)
	}
}

func TestCalibrationDigestBindsTaskOutcomeWithoutCostInference(t *testing.T) {
	calibration := fixtureCalibration()
	input := int64(100)
	calibration.Training[0].Usage = &EmpiricalUsage{InputTokens: &input}
	first, err := calibration.Digest()
	if err != nil {
		t.Fatal(err)
	}
	calibration.Training[0].Outcome = OutcomeAccepted
	second, err := calibration.Digest()
	if err != nil || first == second {
		t.Fatalf("outcome mutation did not bind digest: %q %q %v", first, second, err)
	}
}

func TestObjectiveDigestAndDeclaredScopeAreExact(t *testing.T) {
	first, err := ObjectiveDigest("update the logger")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ObjectiveDigest("update the logger")
	if err != nil || first != second {
		t.Fatalf("objective digest is not stable: %q %q %v", first, second, err)
	}
	calibration := fixtureCalibration()
	calibration.Objectives = calibration.Objectives[:1]
	if err := calibration.Validate(); err == nil {
		t.Fatal("incomplete declared objective scope accepted")
	}
}

func fixtureCalibration() Calibration {
	baseline := Profile{Name: "luna", Runtime: "codex", Provider: "codex", Model: "gpt-6-luna", Effort: "high"}
	candidate := Profile{Name: "sol", Runtime: "codex", Provider: "codex", Model: "gpt-6-sol", Effort: "high"}
	row := func(cohort string, objective byte, profile Profile, outcome string, repairs int) TaskPolicyOutcome {
		return TaskPolicyOutcome{SourceID: "source-main", TaskID: cohort + "-task", RunID: cohort + "-" + profile.Name + "-run", ObjectiveHash: hash(objective), NonFixerPolicyID: hash('e'), AssignedProfile: profile, ObservedFixerProfiles: []Profile{profile}, FixerInvocations: 1, Outcome: outcome, Repairs: repairs}
	}
	return Calibration{Version: CalibrationVersion, FamilyID: "go-fixer", Role: "fixer", Objectives: []string{hash('a'), hash('c')}, Baseline: baseline, Candidate: candidate, MinimumPerArm: 1, Training: []TaskPolicyOutcome{row("train", 'a', baseline, OutcomeFailed, 0), row("train", 'a', candidate, OutcomeAccepted, 1)}, Holdout: []TaskPolicyOutcome{row("holdout", 'c', baseline, OutcomeFailed, 0), row("holdout", 'c', candidate, OutcomeAccepted, 2)}}
}

func hash(value byte) string {
	return string([]byte{value}) + "000000000000000000000000000000000000000000000000000000000000000"
}
