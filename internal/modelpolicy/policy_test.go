package modelpolicy

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSelectBoundedPolicy(t *testing.T) {
	policy := FixturePolicy()
	for _, fixture := range FixtureBaseline() {
		decision, err := Select(policy, fixture.Request)
		if err != nil {
			t.Fatalf("%s: %v", fixture.Label, err)
		}
		if decision.Profile.Name != fixture.ExpectedProfile {
			t.Fatalf("%s: got %q, want %q", fixture.Label, decision.Profile.Name, fixture.ExpectedProfile)
		}
	}
}

func TestSelectEscalatesRiskContextUncertaintyAndFailures(t *testing.T) {
	policy := FixturePolicy()
	cases := []Request{
		{Role: "writer", Complexity: LevelLow, Risk: LevelHigh, Uncertainty: LevelLow, ContextTokens: 1},
		{Role: "writer", Complexity: LevelHigh, Risk: LevelLow, Uncertainty: LevelLow, ContextTokens: 1},
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelHigh, ContextTokens: 1},
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, ContextTokens: 1024},
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, ContextTokens: 1, Failures: 1},
	}
	for _, request := range cases {
		decision, err := Select(policy, request)
		if err != nil || decision.Profile.Name != "architect" {
			t.Fatalf("request %#v: decision %#v, err %v", request, decision, err)
		}
	}
}

func TestSelectKeepsMediumSignalsOnConfiguredDefault(t *testing.T) {
	decision, err := Select(FixturePolicy(), Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium, ContextTokens: 1023})
	if err != nil || decision.Profile.Name != "standard" {
		t.Fatalf("medium signals unexpectedly escalated: %#v, %v", decision, err)
	}
}

func TestSelectUsesExplicitContextBytesWithoutTreatingThemAsTokens(t *testing.T) {
	policy := FixturePolicy()
	policy.Rules["writer"] = Rule{DefaultProfile: "standard", EscalatedProfile: "architect", ContextEscalationTokens: 1024, ContextEscalationBytes: 256, FailureEscalationCount: 1}
	request := Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium, ContextBytes: 256}
	decision, err := Select(policy, request)
	if err != nil || decision.Profile.Name != "architect" || decision.Reason != "large-input-bytes" {
		t.Fatalf("byte-sized input did not use explicit byte threshold: %#v, %v", decision, err)
	}
	request.ContextBytes = 255
	decision, err = Select(policy, request)
	if err != nil || decision.Profile.Name != "standard" {
		t.Fatalf("byte threshold treated bytes as tokens: %#v, %v", decision, err)
	}
}

func TestSelectUsesBoundedReadOnlyEconomyRoute(t *testing.T) {
	policy := FixturePolicy()
	policy.Rules["explorer"] = Rule{DefaultProfile: "standard", CheapProfile: "economy", CheapContextBytes: 256, EscalatedProfile: "architect", ContextEscalationTokens: 1024, ContextEscalationBytes: 2048, FailureEscalationCount: 1}
	request := Request{Role: "explorer", ReadOnly: true, Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium, ContextBytes: 256}
	decision, err := Select(policy, request)
	if err != nil || decision.Profile.Name != "economy" || decision.Reason != "bounded-read-only-input" {
		t.Fatalf("bounded read-only input did not select economy profile: %#v, %v", decision, err)
	}
	request.ContextBytes = 257
	decision, err = Select(policy, request)
	if err != nil || decision.Profile.Name != "standard" {
		t.Fatalf("input above economy byte bound was cheap-routed: %#v, %v", decision, err)
	}
	request.ContextBytes = 256
	request.Failures = 1
	decision, err = Select(policy, request)
	if err != nil || decision.Profile.Name != "architect" {
		t.Fatalf("failure escalation did not take precedence over economy route: %#v, %v", decision, err)
	}
	request.Role, request.ReadOnly, request.Failures = "writer", true, 0
	if _, err := Select(policy, request); err == nil {
		t.Fatal("write role was allowed to claim read-only capability")
	}
}

func TestSelectRejectsUnboundedEconomyRoute(t *testing.T) {
	policy := FixturePolicy()
	policy.Rules["explorer"] = Rule{DefaultProfile: "standard", CheapProfile: "economy", EscalatedProfile: "architect", ContextEscalationTokens: 1024, FailureEscalationCount: 1}
	if _, err := Select(policy, Request{Role: "explorer", ReadOnly: true, Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium}); err == nil {
		t.Fatal("economy profile was admitted without a positive input-byte bound")
	}
}

func TestSelectRejectsOutOfBoundsSignalsAndThresholds(t *testing.T) {
	policy := FixturePolicy()
	invalidRequests := []Request{
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, ContextTokens: -1},
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, ContextTokens: MaxContextTokens + 1},
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, Failures: -1},
		{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, Failures: MaxFailures + 1},
	}
	for _, request := range invalidRequests {
		if _, err := Select(policy, request); err == nil {
			t.Fatalf("out-of-bounds request accepted: %#v", request)
		}
	}

	policy.Rules["writer"] = Rule{DefaultProfile: "standard", EscalatedProfile: "architect", ContextEscalationTokens: MaxContextTokens + 1, FailureEscalationCount: 1}
	if _, err := Select(policy, Request{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow}); err == nil {
		t.Fatal("out-of-bounds context threshold accepted")
	}
}

func TestPolicyRejectsUnknownDecisionEvidenceVersion(t *testing.T) {
	policy := FixturePolicy()
	policy.DecisionEvidenceVersion = 3
	if _, err := Select(policy, Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium}); err == nil {
		t.Fatal("unknown decision evidence version was accepted")
	}
}

func TestDecisionEvidenceVersion2RequiresExactCalibrationProfiles(t *testing.T) {
	calibration := fixtureCalibration()
	policy := Policy{
		Version: 1, DecisionEvidenceVersion: 2, Calibration: &calibration,
		Profiles: []Profile{calibration.Baseline, calibration.Candidate, {Name: "strong", Runtime: "codex", Provider: "codex", Model: "gpt-6-astra", Effort: "high"}},
		Rules:    map[string]Rule{"fixer": {DefaultProfile: calibration.Baseline.Name, EscalatedProfile: "strong", ContextEscalationTokens: MaxContextTokens, ContextEscalationBytes: MaxContextBytes, FailureEscalationCount: 1}},
	}
	decision, err := Select(policy, Request{Role: "fixer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium, ContextTokens: 1})
	if err != nil || decision.Profile.Name != calibration.Baseline.Name {
		t.Fatalf("calibration artifact changed static baseline selection: %#v %v", decision, err)
	}
	policy.Calibration = nil
	if _, err := Select(policy, Request{Role: "fixer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium}); err == nil {
		t.Fatal("evidence version 2 accepted without calibration")
	}
	policy.DecisionEvidenceVersion = 1
	policy.Calibration = &calibration
	if _, err := Select(policy, Request{Role: "fixer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium}); err == nil {
		t.Fatal("calibration was accepted without evidence version 2")
	}
}

func TestDecisionEvidenceVersionZeroIsOmittedFromLegacyJSON(t *testing.T) {
	policy := FixturePolicy()
	legacy, err := json.Marshal(policy)
	if err != nil || strings.Contains(string(legacy), "decision_evidence_version") {
		t.Fatalf("default policy encoding changed legacy identity bytes: %s %v", legacy, err)
	}
	policy.DecisionEvidenceVersion = 1
	optedIn, err := json.Marshal(policy)
	if err != nil || !strings.Contains(string(optedIn), `"decision_evidence_version":1`) {
		t.Fatalf("opt-in evidence version was not encoded: %s %v", optedIn, err)
	}
}

func TestSelectRejectsMissingOrUnknownConfiguredProfile(t *testing.T) {
	policy := FixturePolicy()
	policy.Rules["writer"] = Rule{DefaultProfile: "standard", CheapProfile: "unavailable", EscalatedProfile: "architect", ContextEscalationTokens: 10, FailureEscalationCount: 1}
	if _, err := Select(policy, Request{Role: "writer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow}); err == nil {
		t.Fatal("missing cheap profile accepted")
	}
	if _, err := Select(FixturePolicy(), Request{Role: "reviewer", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow}); err == nil {
		t.Fatal("unknown role borrowed a route")
	}
}

func TestSelectKeepsUnknownEstimatesNilAndCopiesKnownEstimates(t *testing.T) {
	policy := FixturePolicy()
	decision, err := Select(policy, Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelLow})
	if err != nil || decision.ExpectedCostMicroUSD != nil || decision.ExpectedLatencyMillis != nil {
		t.Fatalf("unknown estimates were changed: %#v, %v", decision, err)
	}
	cost, latency := int64(4), int64(9)
	policy.Profiles[1].ExpectedCostMicroUSD, policy.Profiles[1].ExpectedLatencyMillis = &cost, &latency
	decision, err = Select(policy, Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelLow})
	if err != nil || decision.ExpectedCostMicroUSD == nil || decision.ExpectedLatencyMillis == nil {
		t.Fatalf("known estimates missing: %#v, %v", decision, err)
	}
	*decision.ExpectedCostMicroUSD = 99
	if *policy.Profiles[1].ExpectedCostMicroUSD != 4 {
		t.Fatal("decision aliases configuration")
	}
}

func TestRunFixtureBaseline(t *testing.T) {
	got, err := RunFixtureBaseline()
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range FixtureBaseline() {
		if got[fixture.Label] != fixture.ExpectedProfile {
			t.Fatalf("%s: got %q", fixture.Label, got[fixture.Label])
		}
	}
}
