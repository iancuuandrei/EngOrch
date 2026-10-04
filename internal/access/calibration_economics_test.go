package access

import (
	"strings"
	"testing"
)

func TestCalibrationEconomicsV2ReasonsValidate(t *testing.T) {
	for _, reason := range []string{
		"candidate_cheaper_reliable_train_and_holdout",
		"fallback_missing_cost",
		"fallback_cost_overflow",
		"fallback_cost_not_lower",
		"fallback_quality_below_required",
		"fallback_quality_below_floor",
	} {
		selected := reason == "candidate_cheaper_reliable_train_and_holdout"
		decision := CalibrationDecision{
			Digest: strings.Repeat("e", 64), FamilyID: "go-fixer", BaselineProfile: "luna", CandidateProfile: "sol",
			SelectionReason: reason, ScopeMatched: true,
			CandidateSelected: selected, CandidateApplied: selected,
			Training: CalibrationArmCounts{
				Baseline:  CalibrationOutcomeCounts{Assigned: 1, Failed: 1},
				Candidate: CalibrationOutcomeCounts{Assigned: 1, Accepted: 1},
			},
			Holdout: CalibrationArmCounts{
				Baseline:  CalibrationOutcomeCounts{Assigned: 1, Failed: 1},
				Candidate: CalibrationOutcomeCounts{Assigned: 1, Accepted: 1},
			},
		}
		if err := decision.Validate(); err != nil {
			t.Fatalf("v2 reason %q rejected: %v", reason, err)
		}
	}
	// v2 success still requires selected flags; a mismatched flag rejects.
	mismatched := CalibrationDecision{
		Digest: strings.Repeat("e", 64), FamilyID: "go-fixer", BaselineProfile: "luna", CandidateProfile: "sol",
		SelectionReason: "candidate_cheaper_reliable_train_and_holdout", ScopeMatched: true,
		Training: CalibrationArmCounts{Baseline: CalibrationOutcomeCounts{Assigned: 1, Failed: 1}, Candidate: CalibrationOutcomeCounts{Assigned: 1, Accepted: 1}},
		Holdout:  CalibrationArmCounts{Baseline: CalibrationOutcomeCounts{Assigned: 1, Failed: 1}, Candidate: CalibrationOutcomeCounts{Assigned: 1, Accepted: 1}},
	}
	if err := mismatched.Validate(); err == nil {
		t.Fatal("v2 success without selected flags accepted")
	}
}
