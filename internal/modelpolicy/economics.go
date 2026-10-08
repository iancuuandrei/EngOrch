package modelpolicy

import "math"

// Reliability-constrained cheaper-fixer selection (calibration version 2).
//
// The rule recommends a strictly cheaper candidate fixer only when its
// conditional Bayesian reliability bound meets an explicit quality requirement
// in both the training and holdout cohorts. It reuses the exact version 1 row
// domain: bounded disjoint matched whole-task training/holdout rows with
// shared source/task/objective/shared-policy bindings.
//
// Posterior: uniform Beta(1,1) prior. Each ACCEPTED whole-task outcome that
// exercised the fixer updates alpha by one; each exercised semantic FAILED
// updates beta by one. UNKNOWN, NOT_EXERCISED, and drifted rows never update
// the posterior: they force conservative fallback, are never treated as FAIL,
// and are never dropped from the counts.
//
// Bound: deterministic conservative lower quantile of the posterior at fixed
// 95% confidence (5% lower tail). This is a conditional Bayesian bound under
// the stated prior and exchangeability within each arm/cohort; it is not a
// frequentist proof and makes no generalization claim beyond the measured
// cohorts. Quantification is in integer parts-per-million with conservative
// (floor) rounding and deterministic replay: the same frozen artifact always
// yields the same integer bound.
//
// Requirement: requiredPPM = max(0, baselineLowerPPM - epsilonPPM) with
// epsilon in [0, 1000000] (zero allowed). The candidate lower bound must meet
// both requiredPPM and the explicit absolute quality floor, separately in
// training AND holdout, in addition to the operator minimum-per-arm evidence
// threshold. The candidate does NOT need strict observed quality improvement
// when it is admissibly cheaper; version 1 strict improvement is unchanged.
//
// Cost: the candidate mean whole-task CostMicroUSD must be strictly lower than
// the baseline mean in BOTH cohorts. Every row in both arms/cohorts must carry
// a complete measured CostMicroUSD; missing cost falls back as UNKNOWN and
// profile cost estimates are never used. Typed cached-input/reasoning subset
// accounting is preserved by row validation. All cost arithmetic is
// overflow-checked over int64 (usage values may be huge); overflow falls back
// conservatively. Exact ties keep the baseline.
//
// Precedence: static risk/prior-failure/context escalation still wins at
// routing time; this rule only substitutes the default-configured fixer when
// the objective is in scope and static escalation did not trigger. Diagnostics
// in CalibrationSelection.Economics are output only and never authority beyond
// the frozen policy decision.

const (
	// EconomicsConfidencePercent is the fixed lower-bound confidence.
	EconomicsConfidencePercent = 95
	// economicsTail is the lower-tail probability for the bound.
	economicsTail = 0.05
	// economicsQuantileIterations bounds the deterministic binary search.
	economicsQuantileIterations = 100
)

// EconomicsCohortDiagnostics reports the deterministic integer bounds and
// measured mean costs for one cohort. Means are exact integer-division floors
// with per-arm row counts, kept within the canonical integer domain for valid
// JSON-derived calibrations; the authoritative strict-cheaper comparison uses
// exact quotient/remainder arithmetic in meanStrictlyLower. CostsComplete is
// false when any row lacked measured cost or when an overflow was detected.
type EconomicsCohortDiagnostics struct {
	BaselineLowerPPM          int   `json:"baseline_lower_ppm"`
	CandidateLowerPPM         int   `json:"candidate_lower_ppm"`
	RequiredPPM               int   `json:"required_ppm"`
	BaselineMeanCostMicroUSD  int64 `json:"baseline_mean_cost_micro_usd"`
	CandidateMeanCostMicroUSD int64 `json:"candidate_mean_cost_micro_usd"`
	BaselineCostCount         int   `json:"baseline_cost_count"`
	CandidateCostCount        int   `json:"candidate_cost_count"`
	CostsComplete             bool  `json:"costs_complete"`
}

// EconomicsDiagnostics carries version 2 output-only diagnostics.
type EconomicsDiagnostics struct {
	ConfidencePercent int                        `json:"confidence_percent"`
	EpsilonPPM        int                        `json:"epsilon_ppm"`
	QualityFloorPPM   int                        `json:"quality_floor_ppm"`
	Training          EconomicsCohortDiagnostics `json:"training"`
	Holdout           EconomicsCohortDiagnostics `json:"holdout"`
}

// evaluateCheaperFixer implements version 2 selection. The caller supplies the
// digest-bound baseline result with Training/Holdout counts already computed.
func evaluateCheaperFixer(c Calibration, result CalibrationSelection) CalibrationSelection {
	if !matchedTaskPolicies(c.Training, c) || !matchedTaskPolicies(c.Holdout, c) {
		result.Reason = "fallback_unmatched_task_policy"
		result.Economics = economicsDiagnostics(c, result, false)
		return result
	}
	for _, cohort := range []map[string]CalibrationCounts{result.Training, result.Holdout} {
		for _, profile := range []Profile{c.Baseline, c.Candidate} {
			value := cohort[profile.Name]
			if value.Unknown > 0 {
				result.Reason = "fallback_unknown_outcome"
				result.Economics = economicsDiagnostics(c, result, false)
				return result
			}
			if value.NotExercised > 0 {
				result.Reason = "fallback_fixer_not_exercised"
				result.Economics = economicsDiagnostics(c, result, false)
				return result
			}
			if value.Drifted > 0 {
				result.Reason = "fallback_fixer_route_drift"
				result.Economics = economicsDiagnostics(c, result, false)
				return result
			}
			if value.Accepted+value.Failed < c.MinimumPerArm {
				result.Reason = "fallback_insufficient_quality_evidence"
				result.Economics = economicsDiagnostics(c, result, false)
				return result
			}
		}
	}
	trainBaseSum, trainBaseN, trainBaseOK, trainBaseOverflow := sumArmCosts(c.Training, c.Baseline)
	trainCandSum, trainCandN, trainCandOK, trainCandOverflow := sumArmCosts(c.Training, c.Candidate)
	holdBaseSum, holdBaseN, holdBaseOK, holdBaseOverflow := sumArmCosts(c.Holdout, c.Baseline)
	holdCandSum, holdCandN, holdCandOK, holdCandOverflow := sumArmCosts(c.Holdout, c.Candidate)
	if trainBaseOverflow || trainCandOverflow || holdBaseOverflow || holdCandOverflow {
		result.Reason = "fallback_cost_overflow"
		result.Economics = economicsDiagnostics(c, result, false)
		return result
	}
	if !trainBaseOK || !trainCandOK || !holdBaseOK || !holdCandOK {
		result.Reason = "fallback_missing_cost"
		result.Economics = economicsDiagnostics(c, result, false)
		return result
	}
	epsilon, floor := *c.EpsilonPPM, *c.QualityFloorPPM
	trainBaseLCB := lowerBoundPPM(result.Training[c.Baseline.Name].Accepted, result.Training[c.Baseline.Name].Failed)
	trainCandLCB := lowerBoundPPM(result.Training[c.Candidate.Name].Accepted, result.Training[c.Candidate.Name].Failed)
	holdBaseLCB := lowerBoundPPM(result.Holdout[c.Baseline.Name].Accepted, result.Holdout[c.Baseline.Name].Failed)
	holdCandLCB := lowerBoundPPM(result.Holdout[c.Candidate.Name].Accepted, result.Holdout[c.Candidate.Name].Failed)
	trainRequired := requiredQualityPPM(trainBaseLCB, epsilon)
	holdRequired := requiredQualityPPM(holdBaseLCB, epsilon)
	result.Economics = &EconomicsDiagnostics{
		ConfidencePercent: EconomicsConfidencePercent, EpsilonPPM: epsilon, QualityFloorPPM: floor,
		Training: EconomicsCohortDiagnostics{BaselineLowerPPM: trainBaseLCB, CandidateLowerPPM: trainCandLCB, RequiredPPM: trainRequired, BaselineMeanCostMicroUSD: trainBaseSum / int64(trainBaseN), CandidateMeanCostMicroUSD: trainCandSum / int64(trainCandN), BaselineCostCount: trainBaseN, CandidateCostCount: trainCandN, CostsComplete: true},
		Holdout:  EconomicsCohortDiagnostics{BaselineLowerPPM: holdBaseLCB, CandidateLowerPPM: holdCandLCB, RequiredPPM: holdRequired, BaselineMeanCostMicroUSD: holdBaseSum / int64(holdBaseN), CandidateMeanCostMicroUSD: holdCandSum / int64(holdCandN), BaselineCostCount: holdBaseN, CandidateCostCount: holdCandN, CostsComplete: true},
	}
	if trainCandLCB < floor || holdCandLCB < floor {
		result.Reason = "fallback_quality_below_floor"
		return result
	}
	if trainCandLCB < trainRequired || holdCandLCB < holdRequired {
		result.Reason = "fallback_quality_below_required"
		return result
	}
	if !meanStrictlyLower(trainCandSum, trainCandN, trainBaseSum, trainBaseN) || !meanStrictlyLower(holdCandSum, holdCandN, holdBaseSum, holdBaseN) {
		result.Reason = "fallback_cost_not_lower"
		return result
	}
	result.Profile, result.Selected, result.Reason = copyProfile(c.Candidate), true, "candidate_cheaper_reliable_train_and_holdout"
	return result
}

// economicsDiagnostics builds output-only diagnostics for early fallbacks.
// When withCosts is false, cost sums stay zero with CostsComplete false but
// the integer lower bounds remain deterministic. The displayed posterior
// counts only exact-assigned, exercised, undrifted ACCEPTED/FAILED rows via
// diagnosticPosteriorCounts, identical to the main selection drift exclusion:
// UNKNOWN, NOT_EXERCISED (FixerInvocations < 1), and drifted rows never
// update the posterior and are never treated as FAIL. Canonical
// CalibrationCounts (including Unknown/NotExercised/Drifted) and the
// fallback reason are retained unchanged; only the diagnostic bound input is
// filtered.
func economicsDiagnostics(c Calibration, result CalibrationSelection, withCosts bool) *EconomicsDiagnostics {
	epsilon, floor := 0, 0
	if c.EpsilonPPM != nil {
		epsilon = *c.EpsilonPPM
	}
	if c.QualityFloorPPM != nil {
		floor = *c.QualityFloorPPM
	}
	trainBaseAccepted, trainBaseFailed := diagnosticPosteriorCounts(c.Training, c.Baseline)
	trainCandAccepted, trainCandFailed := diagnosticPosteriorCounts(c.Training, c.Candidate)
	holdBaseAccepted, holdBaseFailed := diagnosticPosteriorCounts(c.Holdout, c.Baseline)
	holdCandAccepted, holdCandFailed := diagnosticPosteriorCounts(c.Holdout, c.Candidate)
	trainBaseLCB := lowerBoundPPM(trainBaseAccepted, trainBaseFailed)
	trainCandLCB := lowerBoundPPM(trainCandAccepted, trainCandFailed)
	holdBaseLCB := lowerBoundPPM(holdBaseAccepted, holdBaseFailed)
	holdCandLCB := lowerBoundPPM(holdCandAccepted, holdCandFailed)
	diagnostics := &EconomicsDiagnostics{
		ConfidencePercent: EconomicsConfidencePercent, EpsilonPPM: epsilon, QualityFloorPPM: floor,
		Training: EconomicsCohortDiagnostics{BaselineLowerPPM: trainBaseLCB, CandidateLowerPPM: trainCandLCB, RequiredPPM: requiredQualityPPM(trainBaseLCB, epsilon)},
		Holdout:  EconomicsCohortDiagnostics{BaselineLowerPPM: holdBaseLCB, CandidateLowerPPM: holdCandLCB, RequiredPPM: requiredQualityPPM(holdBaseLCB, epsilon)},
	}
	if withCosts {
		if sum, n, ok, overflow := sumArmCosts(c.Training, c.Baseline); ok && !overflow && n > 0 {
			diagnostics.Training.BaselineMeanCostMicroUSD, diagnostics.Training.BaselineCostCount = sum/int64(n), n
		}
		if sum, n, ok, overflow := sumArmCosts(c.Training, c.Candidate); ok && !overflow && n > 0 {
			diagnostics.Training.CandidateMeanCostMicroUSD, diagnostics.Training.CandidateCostCount = sum/int64(n), n
		}
		if sum, n, ok, overflow := sumArmCosts(c.Holdout, c.Baseline); ok && !overflow && n > 0 {
			diagnostics.Holdout.BaselineMeanCostMicroUSD, diagnostics.Holdout.BaselineCostCount = sum/int64(n), n
		}
		if sum, n, ok, overflow := sumArmCosts(c.Holdout, c.Candidate); ok && !overflow && n > 0 {
			diagnostics.Holdout.CandidateMeanCostMicroUSD, diagnostics.Holdout.CandidateCostCount = sum/int64(n), n
		}
	}
	return diagnostics
}

// diagnosticPosteriorCounts returns the posterior ACCEPTED/FAILED evidence for
// one arm within a cohort. It counts only rows with the exact assigned
// profile, at least one fixer invocation, every observed fixer profile exactly
// matching the assignment, and a terminal ACCEPTED/FAILED outcome. UNKNOWN,
// NOT_EXERCISED, and drifted rows are excluded; missing exercise and unknown
// outcomes never count as FAIL. The drift exclusion (any observed profile
// differing from the assignment) is identical to the main selection guard in
// evaluateCheaperFixer and to countsFor drift detection, so the success-path
// raw Accepted/Failed already equal these filtered counts once the guards
// pass; the fallback path must use this filtered form for the displayed bound.
func diagnosticPosteriorCounts(rows []TaskPolicyOutcome, profile Profile) (accepted, failed int) {
	for _, row := range rows {
		if !sameProfile(row.AssignedProfile, profile) {
			continue
		}
		if row.FixerInvocations < 1 {
			continue
		}
		drifted := false
		for _, observed := range row.ObservedFixerProfiles {
			if !sameProfile(observed, row.AssignedProfile) {
				drifted = true
				break
			}
		}
		if drifted {
			continue
		}
		switch row.Outcome {
		case OutcomeAccepted:
			accepted++
		case OutcomeFailed:
			failed++
		}
	}
	return accepted, failed
}

// requiredQualityPPM derives the per-cohort requirement from the baseline
// lower bound minus the explicit epsilon, clamped at zero.
func requiredQualityPPM(baselineLowerPPM, epsilonPPM int) int {
	required := baselineLowerPPM - epsilonPPM
	if required < 0 {
		return 0
	}
	return required
}

// lowerBoundPPM returns the conservative integer parts-per-million lower bound
// for accepted successes and failed outcomes under a uniform Beta(1,1) prior.
func lowerBoundPPM(accepted, failed int) int {
	return betaLowerQuantilePPM(1+accepted, 1+failed)
}

// betaLowerQuantile returns the 5% lower quantile of Beta(alpha,beta) via a
// bounded deterministic binary search on the regularized incomplete beta
// function. Alpha and beta are positive integers bounded by the calibration
// row limits.
func betaLowerQuantile(alpha, beta int) float64 {
	low, high := 0.0, 1.0
	for i := 0; i < economicsQuantileIterations; i++ {
		mid := (low + high) / 2
		if betaCDF(mid, alpha, beta) < economicsTail {
			low = mid
		} else {
			high = mid
		}
	}
	return high
}

// quantileToPPM converts a quantile in [0,1] to integer ppm with floor
// rounding. Callers seeking the tightest conservative bound use
// betaLowerQuantilePPM, which corrects the floor against the computed CDF.
func quantileToPPM(quantile float64) int {
	if math.IsNaN(quantile) || quantile <= 0 {
		return 0
	}
	if quantile >= 1 {
		return MaxQualityPPM
	}
	ppm := int(quantile * MaxQualityPPM)
	if ppm < 0 {
		return 0
	}
	if ppm > MaxQualityPPM {
		return MaxQualityPPM
	}
	return ppm
}

// betaLowerQuantilePPM returns the tightest conservative integer ppm bound:
// the largest ppm with CDF(ppm/1e6) <= 0.05, corrected from the binary-search
// quantile by a bounded deterministic unit-step adjustment for integer-exact
// replay.
func betaLowerQuantilePPM(alpha, beta int) int {
	quantile := betaLowerQuantile(alpha, beta)
	ppm := quantileToPPM(quantile)
	for i := 0; i < 8 && ppm > 0 && betaCDF(float64(ppm)/MaxQualityPPM, alpha, beta) > economicsTail; i++ {
		ppm--
	}
	for i := 0; i < 8 && ppm < MaxQualityPPM && betaCDF(float64(ppm+1)/MaxQualityPPM, alpha, beta) <= economicsTail; i++ {
		ppm++
	}
	return ppm
}

// betaCDF returns the regularized incomplete beta I_x(alpha,beta) for integer
// alpha,beta >= 1 via the binomial sum. Inputs are bounded by calibration
// row limits (alpha+beta <= 130), so the loop is bounded and deterministic.
func betaCDF(x float64, alpha, beta int) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	n := alpha + beta - 1
	// Initial term j=alpha computed in log space to avoid intermediate
	// binomial overflow: log C(n,alpha) + alpha*log(x) + (n-alpha)*log(1-x).
	lagN, _ := math.Lgamma(float64(n + 1))
	lagA, _ := math.Lgamma(float64(alpha + 1))
	lagNA, _ := math.Lgamma(float64(n - alpha + 1))
	term := math.Exp(lagN - lagA - lagNA + float64(alpha)*math.Log(x) + float64(n-alpha)*math.Log(1-x))
	sum := term
	for j := alpha + 1; j <= n; j++ {
		term *= float64(n-j+1) / float64(j) * x / (1 - x)
		sum += term
		if math.IsNaN(sum) || math.IsInf(sum, 1) {
			return 1
		}
	}
	if sum < 0 {
		return 0
	}
	if sum > 1 {
		return 1
	}
	return sum
}

// sumArmCosts totals measured whole-task CostMicroUSD for one arm within a
// cohort. Complete is false when any row lacks measured cost. Overflow reports
// a checked int64 addition overflow; sums stay bounded by the row limit but
// individual usage values may be huge.
func sumArmCosts(rows []TaskPolicyOutcome, profile Profile) (sum int64, count int, complete bool, overflow bool) {
	const maxInt64 = int64(^uint64(0) >> 1)
	complete = true
	for _, row := range rows {
		if !sameProfile(row.AssignedProfile, profile) {
			continue
		}
		if row.Usage == nil || row.Usage.CostMicroUSD == nil {
			return sum, count, false, false
		}
		value := *row.Usage.CostMicroUSD
		if value < 0 {
			return sum, count, false, false
		}
		if sum > maxInt64-value {
			return sum, count, true, true
		}
		sum += value
		count++
	}
	if count == 0 {
		return 0, 0, false, false
	}
	return sum, count, true, false
}

// meanStrictlyLower reports whether candidate mean cost is strictly lower than
// baseline mean cost using exact integer comparison without overflow-prone
// cross-multiplication of sums. Equal means keep the baseline.
func meanStrictlyLower(candidateSum int64, candidateN int, baselineSum int64, baselineN int) bool {
	if candidateN <= 0 || baselineN <= 0 {
		return false
	}
	quotientCandidate := candidateSum / int64(candidateN)
	remainderCandidate := candidateSum % int64(candidateN)
	quotientBaseline := baselineSum / int64(baselineN)
	remainderBaseline := baselineSum % int64(baselineN)
	if quotientCandidate != quotientBaseline {
		return quotientCandidate < quotientBaseline
	}
	return remainderCandidate*int64(baselineN) < remainderBaseline*int64(candidateN)
}
