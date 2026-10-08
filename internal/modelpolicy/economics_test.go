package modelpolicy

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestBetaCDFMatchesAnalyticIntegerCases(t *testing.T) {
	cases := []struct {
		x     float64
		alpha int
		beta  int
		want  float64
	}{
		{0.5, 2, 1, 0.25},
		{0.5, 1, 2, 0.75},
		{0.3, 1, 1, 0.3},
		{0.0, 2, 2, 0},
		{1.0, 2, 2, 1},
	}
	for _, test := range cases {
		if got := betaCDF(test.x, test.alpha, test.beta); math.Abs(got-test.want) > 1e-12 {
			t.Fatalf("betaCDF(%v,%d,%d) = %v, want %v", test.x, test.alpha, test.beta, got, test.want)
		}
	}
}

func TestLowerBoundPPMMatchesAnalyticQuantiles(t *testing.T) {
	// Beta(1,1) prior only: quantile 0.05 -> 50000 ppm.
	if got := lowerBoundPPM(0, 0); got != 50000 {
		t.Fatalf("prior lower bound = %d, want 50000", got)
	}
	// Beta(2,1): CDF x^2, quantile sqrt(0.05) -> floor 223606.
	if got := lowerBoundPPM(1, 0); got != 223606 {
		t.Fatalf("Beta(2,1) lower bound = %d, want 223606", got)
	}
	// Beta(1,2): CDF 1-(1-x)^2, quantile 1-sqrt(0.95) -> floor 25320.
	if got := lowerBoundPPM(0, 1); got != 25320 {
		t.Fatalf("Beta(1,2) lower bound = %d, want 25320", got)
	}
	// Deterministic replay.
	if first, second := lowerBoundPPM(3, 2), lowerBoundPPM(3, 2); first != second {
		t.Fatalf("lower bound not deterministic: %d %d", first, second)
	}
	// Conservatism: reported ppm never overstates the quantile.
	for _, counts := range [][2]int{{1, 0}, {0, 1}, {3, 2}, {5, 5}} {
		ppm := lowerBoundPPM(counts[0], counts[1])
		alpha, beta := 1+counts[0], 1+counts[1]
		if betaCDF(float64(ppm)/MaxQualityPPM, alpha, beta) > economicsTail+1e-12 {
			t.Fatalf("ppm %d overstates Beta(%d,%d) quantile", ppm, alpha, beta)
		}
		if ppm < MaxQualityPPM && betaCDF(float64(ppm+1)/MaxQualityPPM, alpha, beta) <= economicsTail-1e-12 {
			t.Fatalf("ppm %d is not the tightest conservative bound for Beta(%d,%d)", ppm, alpha, beta)
		}
	}
}

func TestEvaluateEconomicsSelectsCheaperReliableCandidate(t *testing.T) {
	calibration := fixtureEconomicsCalibration()
	selection, err := EvaluateCalibration(calibration)
	if err != nil {
		t.Fatal(err)
	}
	if !selection.Selected || selection.Profile.Name != "sol" || selection.Reason != "candidate_cheaper_reliable_train_and_holdout" {
		t.Fatalf("selection = %+v", selection)
	}
	if selection.Version != CalibrationVersion2 || selection.Economics == nil {
		t.Fatalf("v2 selection lacks version/diagnostics: %+v", selection)
	}
	if selection.Economics.ConfidencePercent != EconomicsConfidencePercent || !selection.Economics.Training.CostsComplete {
		t.Fatalf("diagnostics incomplete: %+v", selection.Economics)
	}
	// Candidate lower bounds must meet floor and required in both cohorts.
	for _, cohort := range []EconomicsCohortDiagnostics{selection.Economics.Training, selection.Economics.Holdout} {
		if cohort.CandidateLowerPPM < cohort.RequiredPPM || cohort.CandidateLowerPPM < selection.Economics.QualityFloorPPM {
			t.Fatalf("candidate bound below requirement: %+v", cohort)
		}
	}
	second, err := EvaluateCalibration(fixtureEconomicsCalibration())
	if err != nil || second.Digest != selection.Digest || second.Economics.Training != selection.Economics.Training {
		t.Fatalf("v2 digest/diagnostics not deterministic: %q %q %v", selection.Digest, second.Digest, err)
	}
}

func TestEvaluateEconomicsDoesNotRequireStrictImprovement(t *testing.T) {
	// Equal observed quality (both 1 accepted / 0 failed per arm would be
	// strict-improvement failure under v1) with a strictly cheaper candidate
	// is admissible under v2 when bounds meet floor/required.
	calibration := fixtureEconomicsCalibration()
	// Make baseline training accepted as well: baseline 1/0, candidate 1/0.
	// Both arms share the same task key; outcomes equal.
	for i := range calibration.Training {
		calibration.Training[i].Outcome = OutcomeAccepted
		calibration.Training[i].FixerInvocations = 1
		calibration.Training[i].Repairs = 1
		calibration.Training[i].ObservedFixerProfiles = []Profile{calibration.Training[i].AssignedProfile}
	}
	selection, err := EvaluateCalibration(calibration)
	if err != nil || !selection.Selected {
		t.Fatalf("equal-quality cheaper candidate rejected: %+v %v", selection, err)
	}
	// The same equal-quality shape under v1 must still fall back.
	v1 := fixtureCalibration()
	v1.Training[0].Outcome = OutcomeAccepted
	if selection, err := EvaluateCalibration(v1); err != nil || selection.Selected || selection.Reason != "fallback_no_quality_improvement" {
		t.Fatalf("v1 strict improvement changed: %+v %v", selection, err)
	}
}

func TestEvaluateEconomicsFallbacks(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Calibration)
		reason string
	}{
		{"sparse", func(c *Calibration) { c.MinimumPerArm = 2 }, "fallback_insufficient_quality_evidence"},
		{"unknown", func(c *Calibration) { c.Training[0].Outcome = OutcomeUnknown }, "fallback_unknown_outcome"},
		{"not-exercised", func(c *Calibration) {
			row := &c.Training[0]
			row.ObservedFixerProfiles = nil
			row.FixerInvocations = 0
			row.Outcome = OutcomeNotExercised
		}, "fallback_fixer_not_exercised"},
		{"drift", func(c *Calibration) { c.Holdout[1].ObservedFixerProfiles[0] = c.Baseline }, "fallback_fixer_route_drift"},
		{"missing-cost", func(c *Calibration) { c.Training[0].Usage = nil }, "fallback_missing_cost"},
		{"cost-tie", func(c *Calibration) {
			for i := range c.Training {
				cost := int64(100)
				c.Training[i].Usage = &EmpiricalUsage{CostMicroUSD: &cost}
			}
			for i := range c.Holdout {
				cost := int64(100)
				c.Holdout[i].Usage = &EmpiricalUsage{CostMicroUSD: &cost}
			}
		}, "fallback_cost_not_lower"},
		{"cost-holdout-not-lower", func(c *Calibration) {
			cost := int64(500)
			c.Holdout[1].Usage = &EmpiricalUsage{CostMicroUSD: &cost}
		}, "fallback_cost_not_lower"},
		{"floor", func(c *Calibration) { *c.QualityFloorPPM = 999999 }, "fallback_quality_below_floor"},
		{"required", func(c *Calibration) {
			// Baseline clearly better than candidate; epsilon 0 forces fallback.
			c.Training[0].Outcome = OutcomeAccepted
			c.Holdout[0].Outcome = OutcomeAccepted
			c.Training[1].Outcome = OutcomeFailed
			c.Holdout[1].Outcome = OutcomeFailed
			*c.EpsilonPPM = 0
			*c.QualityFloorPPM = 0
		}, "fallback_quality_below_required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			calibration := fixtureEconomicsCalibration()
			test.mutate(&calibration)
			selection, err := EvaluateCalibration(calibration)
			if err != nil || selection.Selected || selection.Reason != test.reason || selection.Profile.Name != "luna" {
				t.Fatalf("%s: selection = %+v, err = %v", test.name, selection, err)
			}
			if selection.Economics == nil {
				t.Fatalf("%s: missing output-only diagnostics", test.name)
			}
		})
	}
}

func TestSumArmCostsOverflowCheckedDirectly(t *testing.T) {
	// Helper-level robustness for int64 limits, exercised without canonical
	// validation: checked addition must report overflow rather than wrap.
	const maxInt64 = int64(^uint64(0) >> 1)
	half, other := maxInt64/2+1, maxInt64/2+2
	baseline := Profile{Name: "luna", Runtime: "codex", Provider: "codex", Model: "gpt-6-luna", Effort: "high"}
	rows := []TaskPolicyOutcome{
		{SourceID: "source-main", TaskID: "task-a", RunID: "run-a", ObjectiveHash: hash('a'), NonFixerPolicyID: hash('e'), AssignedProfile: baseline, ObservedFixerProfiles: []Profile{baseline}, FixerInvocations: 1, Outcome: OutcomeAccepted, Repairs: 1, Usage: &EmpiricalUsage{CostMicroUSD: &half}},
		{SourceID: "source-main", TaskID: "task-b", RunID: "run-b", ObjectiveHash: hash('a'), NonFixerPolicyID: hash('e'), AssignedProfile: baseline, ObservedFixerProfiles: []Profile{baseline}, FixerInvocations: 1, Outcome: OutcomeAccepted, Repairs: 1, Usage: &EmpiricalUsage{CostMicroUSD: &other}},
	}
	_, _, complete, overflow := sumArmCosts(rows, baseline)
	if !overflow || !complete {
		t.Fatalf("int64 overflow not reported: complete=%v overflow=%v", complete, overflow)
	}
	// Exact boundary: maxInt64 itself does not overflow on a single row.
	single := []TaskPolicyOutcome{rows[0]}
	maxValue := maxInt64
	single[0].Usage = &EmpiricalUsage{CostMicroUSD: &maxValue}
	if _, _, _, overflow := sumArmCosts(single[:1], baseline); overflow {
		t.Fatal("single maxInt64 row incorrectly reported overflow")
	}
}

func TestEvaluateEconomicsRejectsNoncanonicalCostBeforeSum(t *testing.T) {
	// int64-max costs violate the canonical 53-bit integer domain, so Digest
	// (and therefore EvaluateCalibration) must reject before any cost sum.
	// Matched task-policy cohorts are preserved; rejection is by domain, not
	// by mismatch.
	const maxInt64 = int64(^uint64(0) >> 1)
	half := maxInt64/2 + 1
	calibration := fixtureEconomicsCalibration()
	calibration.Training[0].Usage = &EmpiricalUsage{CostMicroUSD: &half}
	cheap := int64(1)
	calibration.Training[1].Usage = &EmpiricalUsage{CostMicroUSD: &cheap}
	if !matchedTaskPolicies(calibration.Training, calibration) || !matchedTaskPolicies(calibration.Holdout, calibration) {
		t.Fatal("economics fixture cohorts must stay matched for the canonical-rejection case")
	}
	if _, err := calibration.Digest(); err == nil {
		t.Fatal("noncanonical int64-max cost accepted by canonical digest")
	}
	if _, err := EvaluateCalibration(calibration); err == nil {
		t.Fatal("noncanonical cost did not reject before cost sum")
	} else if got := err.Error(); got != "noncanonical integer domain" && !strings.Contains(got, "noncanonical") {
		t.Fatalf("rejection was not by canonical domain: %v", err)
	}
}

func TestEvaluateEconomicsAdmitsMaximumCanonicalCostsWithoutOverflow(t *testing.T) {
	// Admitted calibrations stay within the canonical 53-bit domain. Even 64
	// maximum-canonical rows cannot overflow int64, so overflow fallback is
	// helper-level robustness rather than an admitted-path outcome.
	const maxCanonical = int64(9007199254740991)
	const maxInt64 = int64(^uint64(0) >> 1)
	if int64(64)*maxCanonical >= maxInt64 {
		t.Fatal("test premise changed: 64 maximum-canonical rows would overflow int64")
	}
	calibration := fixtureEconomicsCalibration()
	baselineMax, candidateCheaper := maxCanonical, maxCanonical-1
	calibration.Training[0].Usage = &EmpiricalUsage{CostMicroUSD: &baselineMax}
	calibration.Holdout[0].Usage = &EmpiricalUsage{CostMicroUSD: &baselineMax}
	calibration.Training[1].Usage = &EmpiricalUsage{CostMicroUSD: &candidateCheaper}
	calibration.Holdout[1].Usage = &EmpiricalUsage{CostMicroUSD: &candidateCheaper}
	for _, cohort := range [][]TaskPolicyOutcome{calibration.Training, calibration.Holdout} {
		for _, profile := range []Profile{calibration.Baseline, calibration.Candidate} {
			if _, _, _, overflow := sumArmCosts(cohort, profile); overflow {
				t.Fatal("maximum-canonical admitted cost incorrectly reported overflow")
			}
		}
	}
	selection, err := EvaluateCalibration(calibration)
	if err != nil || !selection.Selected || selection.Reason != "candidate_cheaper_reliable_train_and_holdout" {
		t.Fatalf("maximum-canonical cheaper candidate rejected: %+v %v", selection, err)
	}
	if !selection.Economics.Training.CostsComplete || !selection.Economics.Holdout.CostsComplete {
		t.Fatalf("admitted maximum-canonical costs left incomplete: %+v", selection.Economics)
	}
}

func TestEvaluateEconomicsMissingCostDiagnosticsStayUnknown(t *testing.T) {
	// Missing-cost fallback keeps CostsComplete false; zero means there are
	// unknown, not a measured zero. A true measured zero stays 0 with
	// CostsComplete true.
	calibration := fixtureEconomicsCalibration()
	calibration.Training[0].Usage = nil
	selection, err := EvaluateCalibration(calibration)
	if err != nil || selection.Selected || selection.Reason != "fallback_missing_cost" || selection.Economics == nil {
		t.Fatalf("missing cost did not fall back as unknown: %+v %v", selection, err)
	}
	if selection.Economics.Training.CostsComplete {
		t.Fatalf("missing cost marked complete: %+v", selection.Economics.Training)
	}
	calibration = fixtureEconomicsCalibration()
	zero := int64(0)
	calibration.Training[0].Usage = &EmpiricalUsage{CostMicroUSD: &zero}
	calibration.Training[1].Usage = &EmpiricalUsage{CostMicroUSD: &zero}
	calibration.Holdout[0].Usage = &EmpiricalUsage{CostMicroUSD: &zero}
	calibration.Holdout[1].Usage = &EmpiricalUsage{CostMicroUSD: &zero}
	selection, err = EvaluateCalibration(calibration)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Economics == nil {
		t.Fatal("true-zero calibration lacks diagnostics")
	}
	// Equal true-zero means tie and keep the baseline, but costs are complete.
	if selection.Selected || selection.Reason != "fallback_cost_not_lower" {
		t.Fatalf("true-zero tie did not keep baseline: %+v", selection)
	}
	if !selection.Economics.Training.CostsComplete || selection.Economics.Training.BaselineMeanCostMicroUSD != 0 || selection.Economics.Training.CandidateMeanCostMicroUSD != 0 {
		t.Fatalf("true measured zero not preserved as 0 complete: %+v", selection.Economics.Training)
	}
}

func TestEvaluateEconomicsFloorBoundaryAndEpsilon(t *testing.T) {
	calibration := fixtureEconomicsCalibration()
	selection, err := EvaluateCalibration(calibration)
	if err != nil || !selection.Selected {
		t.Fatal(selection, err)
	}
	candidateLCB := selection.Economics.Training.CandidateLowerPPM
	if holdLCB := selection.Economics.Holdout.CandidateLowerPPM; holdLCB < candidateLCB {
		candidateLCB = holdLCB
	}
	// Floor exactly at the minimum candidate bound still passes.
	calibration.QualityFloorPPM = &candidateLCB
	if selection, err := EvaluateCalibration(calibration); err != nil || !selection.Selected {
		t.Fatalf("floor boundary rejected: %+v %v", selection, err)
	}
	// Floor one ppm above fails.
	above := candidateLCB + 1
	calibration.QualityFloorPPM = &above
	if selection, err := EvaluateCalibration(calibration); err != nil || selection.Selected || selection.Reason != "fallback_quality_below_floor" {
		t.Fatalf("floor boundary not enforced: %+v %v", selection, err)
	}
	// Epsilon 0 allowed; larger epsilon relaxes the requirement.
	calibration = fixtureEconomicsCalibration()
	calibration.Training[0].Outcome = OutcomeAccepted
	calibration.Holdout[0].Outcome = OutcomeAccepted
	calibration.Training[1].Outcome = OutcomeFailed
	calibration.Holdout[1].Outcome = OutcomeFailed
	zero := 0
	calibration.EpsilonPPM = &zero
	lowFloor := 0
	calibration.QualityFloorPPM = &lowFloor
	if selection, err := EvaluateCalibration(calibration); err != nil || selection.Selected {
		t.Fatalf("stronger baseline with epsilon 0 unexpectedly passed: %+v %v", selection, err)
	}
	// Baseline LCB minus candidate LCB is the minimum epsilon that admits.
	baselineLCB := lowerBoundPPM(1, 0)
	candidateLCB2 := lowerBoundPPM(0, 1)
	needed := baselineLCB - candidateLCB2
	calibration.EpsilonPPM = &needed
	if selection, err := EvaluateCalibration(calibration); err != nil || !selection.Selected {
		t.Fatalf("explicit epsilon did not admit within-bound degradation: %+v %v", selection, err)
	}
}

func TestEvaluateEconomicsTrainHoldoutGating(t *testing.T) {
	calibration := fixtureEconomicsCalibration()
	// Break only the holdout quality: baseline holdout accepts while the
	// candidate holdout fails, so training passes but holdout misses required.
	calibration.Holdout[0].Outcome = OutcomeAccepted
	calibration.Holdout[1].Outcome = OutcomeFailed
	selection, err := EvaluateCalibration(calibration)
	if err != nil || selection.Selected || selection.Reason != "fallback_quality_below_required" {
		t.Fatalf("holdout gating missed: %+v %v", selection, err)
	}
	calibration = fixtureEconomicsCalibration()
	calibration.Holdout[0].Outcome = OutcomeAccepted
	calibration.Holdout[1].Outcome = OutcomeFailed
	// Even with floor 0, the failed holdout candidate bound must miss required.
	zero := 0
	calibration.QualityFloorPPM = &zero
	selection, err = EvaluateCalibration(calibration)
	if err != nil || selection.Selected {
		t.Fatalf("failed holdout candidate selected: %+v %v", selection, err)
	}
}

func TestCalibrationV1ExactCompatibility(t *testing.T) {
	v1 := fixtureCalibration()
	if v1.Version != CalibrationVersion {
		t.Fatalf("v1 fixture version changed: %d", v1.Version)
	}
	if v1.EpsilonPPM != nil || v1.QualityFloorPPM != nil {
		t.Fatal("v1 fixture carries economics parameters")
	}
	selection, err := EvaluateCalibration(v1)
	if err != nil || !selection.Selected || selection.Economics != nil || selection.Reason != "candidate_quality_improved_train_and_holdout" {
		t.Fatalf("v1 selection changed: %+v %v", selection, err)
	}
	v1.EpsilonPPM = ptrInt(0)
	if err := v1.Validate(); err == nil {
		t.Fatal("v1 accepted economics parameters")
	}
	v2 := fixtureEconomicsCalibration()
	if err := v2.Validate(); err != nil {
		t.Fatalf("valid v2 rejected: %v", err)
	}
	bad := fixtureEconomicsCalibration()
	bad.EpsilonPPM = ptrInt(MaxQualityPPM + 1)
	if err := bad.Validate(); err == nil {
		t.Fatal("unbounded epsilon accepted")
	}
	bad = fixtureEconomicsCalibration()
	bad.QualityFloorPPM = nil
	if err := bad.Validate(); err == nil {
		t.Fatal("v2 accepted without floor")
	}
	// Digest binds economics parameters.
	first, err := v2.Digest()
	if err != nil {
		t.Fatal(err)
	}
	v2.EpsilonPPM = ptrInt(*v2.EpsilonPPM + 1)
	second, err := v2.Digest()
	if err != nil || first == second {
		t.Fatalf("epsilon mutation did not bind digest: %q %q %v", first, second, err)
	}
}

func TestMeanStrictlyLowerKeepsTies(t *testing.T) {
	if meanStrictlyLower(200, 2, 200, 2) {
		t.Fatal("equal means selected as strictly lower")
	}
	if !meanStrictlyLower(100, 1, 200, 1) {
		t.Fatal("strictly cheaper mean rejected")
	}
	// Quotient tie with remainder deciding: 3/2=1.5 vs 4/3≈1.33.
	if !meanStrictlyLower(4, 3, 3, 2) {
		t.Fatal("remainder comparison missed strictly cheaper mean")
	}
	if meanStrictlyLower(3, 2, 4, 3) {
		t.Fatal("remainder comparison overstated")
	}
}

func TestEconomicsFallbackDiagnosticsFilterDriftUnknownNoExercise(t *testing.T) {
	// Mixed artifact: drifted ACCEPTED, UNKNOWN with exercise, and
	// NOT_EXERCISED rows coexist. Canonical counts retain every row and the
	// fallback reason is preserved, while the displayed posterior must count
	// only exact-assigned, exercised, undrifted ACCEPTED/FAILED rows.
	baseline := Profile{Name: "luna", Runtime: "codex", Provider: "codex", Model: "gpt-6-luna", Effort: "high"}
	candidate := Profile{Name: "sol", Runtime: "codex", Provider: "codex", Model: "gpt-6-sol", Effort: "high"}
	expensive, cheap := int64(200), int64(100)
	row := func(taskID string, objective byte, profile Profile, outcome string, invocations int, observed []Profile, cost int64) TaskPolicyOutcome {
		value := cost
		return TaskPolicyOutcome{SourceID: "source-main", TaskID: taskID, RunID: taskID + "-" + profile.Name + "-run", ObjectiveHash: hash(objective), NonFixerPolicyID: hash('e'), AssignedProfile: profile, ObservedFixerProfiles: observed, FixerInvocations: invocations, Outcome: outcome, Repairs: 1, Usage: &EmpiricalUsage{CostMicroUSD: &value}}
	}
	clean := func(profile Profile) []Profile { return []Profile{profile} }
	calibration := Calibration{
		Version: CalibrationVersion2, FamilyID: "go-fixer", Role: "fixer",
		Objectives: []string{hash('a'), hash('b'), hash('c'), hash('d')}, Baseline: baseline, Candidate: candidate,
		MinimumPerArm: 1, EpsilonPPM: ptrInt(0), QualityFloorPPM: ptrInt(0),
		Training: []TaskPolicyOutcome{
			row("train-task-a", 'a', baseline, OutcomeFailed, 1, clean(baseline), expensive),
			row("train-task-a", 'a', candidate, OutcomeAccepted, 1, clean(candidate), cheap),
			row("train-task-b", 'b', baseline, OutcomeFailed, 1, clean(baseline), expensive),
			// Drifted ACCEPTED: assigned candidate but observed baseline.
			row("train-task-b", 'b', candidate, OutcomeAccepted, 1, clean(baseline), cheap),
		},
		Holdout: []TaskPolicyOutcome{
			// UNKNOWN with exercise: must not count as FAIL.
			row("holdout-task-c", 'c', baseline, OutcomeUnknown, 1, clean(baseline), expensive),
			row("holdout-task-c", 'c', candidate, OutcomeAccepted, 1, clean(candidate), cheap),
			row("holdout-task-d", 'd', baseline, OutcomeFailed, 1, clean(baseline), expensive),
			// NOT_EXERCISED: zero invocations, no observed profiles.
			row("holdout-task-d", 'd', candidate, OutcomeNotExercised, 0, nil, cheap),
		},
	}
	selection, err := EvaluateCalibration(calibration)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Selected || selection.Reason != "fallback_fixer_route_drift" || selection.Profile.Name != "luna" {
		t.Fatalf("mixed fallback not preserved: %+v", selection)
	}
	if selection.Economics == nil {
		t.Fatal("mixed fallback lacks diagnostics")
	}
	// Canonical evidence is retained, not dropped.
	if got := selection.Training[candidate.Name]; got.Accepted != 2 || got.Drifted != 1 {
		t.Fatalf("drifted ACCEPTED dropped from canonical counts: %+v", got)
	}
	if got := selection.Holdout[baseline.Name]; got.Unknown != 1 || got.Failed != 1 {
		t.Fatalf("unknown row dropped from canonical counts: %+v", got)
	}
	if got := selection.Holdout[candidate.Name]; got.NotExercised != 1 || got.Accepted != 1 {
		t.Fatalf("no-exercise row dropped from canonical counts: %+v", got)
	}
	// Diagnostic posterior excludes drifted ACCEPTED: training candidate is
	// one clean ACCEPTED, not two raw ACCEPTED.
	if want := lowerBoundPPM(1, 0); selection.Economics.Training.CandidateLowerPPM != want {
		t.Fatalf("training candidate diagnostics = %d, want filtered lowerBoundPPM(1,0) = %d (raw 2,0 = %d)", selection.Economics.Training.CandidateLowerPPM, want, lowerBoundPPM(2, 0))
	}
	if selection.Economics.Training.CandidateLowerPPM == lowerBoundPPM(2, 0) && lowerBoundPPM(2, 0) != lowerBoundPPM(1, 0) {
		t.Fatal("drifted ACCEPTED contaminated the displayed posterior")
	}
	// UNKNOWN with exercise must not count as FAIL: holdout baseline is
	// (0 accepted, 1 failed), not (0,2).
	if want := lowerBoundPPM(0, 1); selection.Economics.Holdout.BaselineLowerPPM != want {
		t.Fatalf("holdout baseline diagnostics = %d, want lowerBoundPPM(0,1) = %d (unknown-as-FAIL 0,2 = %d)", selection.Economics.Holdout.BaselineLowerPPM, want, lowerBoundPPM(0, 2))
	}
	// NOT_EXERCISED must not count as FAIL: holdout candidate is (1,0).
	if want := lowerBoundPPM(1, 0); selection.Economics.Holdout.CandidateLowerPPM != want {
		t.Fatalf("holdout candidate diagnostics = %d, want lowerBoundPPM(1,0) = %d", selection.Economics.Holdout.CandidateLowerPPM, want)
	}
	// Training baseline has two clean FAILED rows.
	if want := lowerBoundPPM(0, 2); selection.Economics.Training.BaselineLowerPPM != want {
		t.Fatalf("training baseline diagnostics = %d, want lowerBoundPPM(0,2) = %d", selection.Economics.Training.BaselineLowerPPM, want)
	}
}

func TestLowerBoundPPMExhaustiveExactBinomialProof(t *testing.T) {
	// Independent exact integer proof for every admitted (accepted, failed)
	// with accepted+failed <= 64 (2145 cases). No betaCDF oracle, no float
	// tolerance: with denominator 1e6, n = alpha+beta-1, and
	// S(p) = sum_{j=alpha..n} C(n,j) p^j (1e6-p)^(n-j), the returned ppm must
	// satisfy 20*S(ppm) <= 1e6^n (conservative, at most 5% tail) and, unless
	// ppm is the maximum, 20*S(ppm+1) > 1e6^n (tightest such ppm).
	const denom int64 = 1000000
	exactTail := func(n, alpha int, p int64) *big.Int {
		q := denom - p
		sum := big.NewInt(0)
		for j := alpha; j <= n; j++ {
			binom := new(big.Int).Binomial(int64(n), int64(j))
			powP := new(big.Int).Exp(big.NewInt(p), big.NewInt(int64(j)), nil)
			powQ := new(big.Int).Exp(big.NewInt(q), big.NewInt(int64(n-j)), nil)
			term := new(big.Int).Mul(binom, powP)
			term.Mul(term, powQ)
			sum.Add(sum, term)
		}
		return sum
	}
	denomPow := func(n int) *big.Int {
		return new(big.Int).Exp(big.NewInt(denom), big.NewInt(int64(n)), nil)
	}
	twenty := big.NewInt(20)
	cases := 0
	for accepted := 0; accepted <= 64; accepted++ {
		for failed := 0; failed <= 64-accepted; failed++ {
			cases++
			ppm := lowerBoundPPM(accepted, failed)
			alpha, beta := 1+accepted, 1+failed
			n := alpha + beta - 1
			denominator := denomPow(n)
			conservative := exactTail(n, alpha, int64(ppm))
			if new(big.Int).Mul(conservative, twenty).Cmp(denominator) > 0 {
				t.Fatalf("lowerBoundPPM(%d,%d) = %d overstates 5%% tail", accepted, failed, ppm)
			}
			if ppm < MaxQualityPPM {
				tight := exactTail(n, alpha, int64(ppm+1))
				if new(big.Int).Mul(tight, twenty).Cmp(denominator) <= 0 {
					t.Fatalf("lowerBoundPPM(%d,%d) = %d is not the tightest conservative ppm", accepted, failed, ppm)
				}
			}
		}
	}
	if cases != 2145 {
		t.Fatalf("exhaustive cases = %d, want 2145", cases)
	}
}

func fixtureEconomicsCalibration() Calibration {
	baseline := Profile{Name: "luna", Runtime: "codex", Provider: "codex", Model: "gpt-6-luna", Effort: "high"}
	candidate := Profile{Name: "sol", Runtime: "codex", Provider: "codex", Model: "gpt-6-sol", Effort: "high"}
	expensive, cheap := int64(200), int64(100)
	row := func(cohort string, objective byte, profile Profile, outcome string, cost int64) TaskPolicyOutcome {
		value := cost
		return TaskPolicyOutcome{SourceID: "source-main", TaskID: cohort + "-task", RunID: cohort + "-" + profile.Name + "-run", ObjectiveHash: hash(objective), NonFixerPolicyID: hash('e'), AssignedProfile: profile, ObservedFixerProfiles: []Profile{profile}, FixerInvocations: 1, Outcome: outcome, Repairs: 1, Usage: &EmpiricalUsage{CostMicroUSD: &value}}
	}
	epsilon, floor := 0, 0
	return Calibration{
		Version: CalibrationVersion2, FamilyID: "go-fixer", Role: "fixer",
		Objectives: []string{hash('a'), hash('c')}, Baseline: baseline, Candidate: candidate,
		MinimumPerArm: 1, EpsilonPPM: &epsilon, QualityFloorPPM: &floor,
		Training: []TaskPolicyOutcome{row("train", 'a', baseline, OutcomeFailed, expensive), row("train", 'a', candidate, OutcomeAccepted, cheap)},
		Holdout:  []TaskPolicyOutcome{row("holdout", 'c', baseline, OutcomeFailed, expensive), row("holdout", 'c', candidate, OutcomeAccepted, cheap)},
	}
}

func ptrInt(value int) *int {
	return &value
}
