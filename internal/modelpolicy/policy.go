package modelpolicy

import (
	"errors"
	"sort"
	"strings"
)

const (
	// MaxContextBytes bounds exact UTF-8 input size used for routing. It is a
	// byte count and is never interpreted as provider-token usage.
	MaxContextBytes int64 = 256 << 10
	// MaxContextTokens bounds caller-provided context size for routing.
	// It is deliberately much larger than expected task inputs while
	// preventing unbounded or overflow-prone values from entering policy logic.
	MaxContextTokens int64 = 1_000_000_000
	// MaxFailures bounds caller-provided prior-failure counts for routing.
	// It is deliberately much larger than expected task inputs while
	// preventing unbounded or overflow-prone values from entering policy logic.
	MaxFailures = 1_000_000
	maxProfiles = 256
	maxRules    = 256
)

// Level is an ordered task signal. Unspecified is rejected so routing never
// quietly treats an unknown assessment as low risk.
type Level string

const (
	// LevelLow is a low task signal that permits the cheap profile path.
	LevelLow Level = "low"
	// LevelMedium is a moderate task signal that keeps the default profile.
	LevelMedium Level = "medium"
	// LevelHigh is an elevated task signal that forces escalation.
	LevelHigh Level = "high"
)

// Profile is one exact, configured routing target. Nil estimates mean the
// configuration has no usable estimate; they never mean zero.
type Profile struct {
	Name                  string `json:"name"`
	Runtime               string `json:"runtime"`
	Provider              string `json:"provider"`
	Model                 string `json:"model"`
	Effort                string `json:"effort"`
	ExpectedCostMicroUSD  *int64 `json:"expected_cost_micro_usd,omitempty"`
	ExpectedLatencyMillis *int64 `json:"expected_latency_millis,omitempty"`
}

// Rule names only profiles listed in Policy.Profiles. CheapProfile is optional
// and can only be selected for bounded read-only input under CheapContextBytes.
type Rule struct {
	DefaultProfile          string `json:"default_profile"`
	CheapProfile            string `json:"cheap_profile,omitempty"`
	CheapContextBytes       int64  `json:"cheap_context_bytes,omitempty"`
	EscalatedProfile        string `json:"escalated_profile"`
	ContextEscalationTokens int64  `json:"context_escalation_tokens"`
	ContextEscalationBytes  int64  `json:"context_escalation_bytes,omitempty"`
	FailureEscalationCount  int    `json:"failure_escalation_count"`
}

// Policy owns all allowable profile choices for this one routing decision.
// Rules are keyed by a role such as planner, explorer, writer, reviewer, or
// architect. Unknown roles reject rather than borrowing another role's route.
type Policy struct {
	Version                 int             `json:"version"`
	DecisionEvidenceVersion int             `json:"decision_evidence_version,omitempty" toml:"DecisionEvidenceVersion"`
	Profiles                []Profile       `json:"profiles"`
	Rules                   map[string]Rule `json:"rules"`
	// Calibration is optional, immutable task-level evidence for one explicitly
	// opted-in fixer route. Nil preserves historical policy serialization.
	Calibration *Calibration `json:"calibration,omitempty" toml:"Calibration"`
}

// Request contains only task facts already known to the caller. Failures is
// the count of prior accepted failed attempts for this task, not a provider
// retry count. Unknown provider outcomes must be resolved outside this policy.
type Request struct {
	Role          string `json:"role"`
	ReadOnly      bool   `json:"read_only,omitempty"`
	Complexity    Level  `json:"complexity"`
	Risk          Level  `json:"risk"`
	Uncertainty   Level  `json:"uncertainty"`
	ContextTokens int64  `json:"context_tokens"`
	ContextBytes  int64  `json:"context_bytes,omitempty"`
	Failures      int    `json:"failures"`
}

// Decision is an observable immutable selection. Reason contains the signal
// that chose the profile, allowing a caller to persist it with its invocation.
type Decision struct {
	Profile               Profile `json:"profile"`
	Reason                string  `json:"reason"`
	ExpectedCostMicroUSD  *int64  `json:"expected_cost_micro_usd,omitempty"`
	ExpectedLatencyMillis *int64  `json:"expected_latency_millis,omitempty"`
}

// Select returns the exact configured profile for a valid request. It never
// falls back to a different configured model if a named profile is absent.
func Select(policy Policy, request Request) (Decision, error) {
	profiles, err := policy.validate()
	if err != nil {
		return Decision{}, err
	}
	if err := request.validate(); err != nil {
		return Decision{}, err
	}
	rule, ok := policy.Rules[request.Role]
	if !ok {
		return Decision{}, errors.New("routing rule unavailable for role")
	}
	if err := rule.validate(profiles); err != nil {
		return Decision{}, err
	}

	name, reason := rule.DefaultProfile, "default-configured-profile"
	switch {
	case request.Role == "architect":
		name, reason = rule.EscalatedProfile, "architect-role"
	case request.Risk == LevelHigh:
		name, reason = rule.EscalatedProfile, "high-risk"
	case request.Complexity == LevelHigh:
		name, reason = rule.EscalatedProfile, "high-complexity"
	case request.Uncertainty == LevelHigh:
		name, reason = rule.EscalatedProfile, "high-uncertainty"
	case request.ContextTokens >= rule.ContextEscalationTokens:
		name, reason = rule.EscalatedProfile, "large-context"
	case rule.ContextEscalationBytes > 0 && request.ContextBytes >= rule.ContextEscalationBytes:
		name, reason = rule.EscalatedProfile, "large-input-bytes"
	case request.Failures >= rule.FailureEscalationCount:
		name, reason = rule.EscalatedProfile, "prior-failures"
	case rule.CheapProfile != "" && request.Role == "explorer" && request.ReadOnly && request.ContextBytes <= rule.CheapContextBytes && request.Failures == 0:
		name, reason = rule.CheapProfile, "bounded-read-only-input"
	}
	selected := profiles[name]
	return Decision{Profile: selected, Reason: reason, ExpectedCostMicroUSD: clone(selected.ExpectedCostMicroUSD), ExpectedLatencyMillis: clone(selected.ExpectedLatencyMillis)}, nil
}

func (p Policy) validate() (map[string]Profile, error) {
	if p.Version != 1 || (p.DecisionEvidenceVersion != 0 && p.DecisionEvidenceVersion != 1 && p.DecisionEvidenceVersion != 2) || len(p.Profiles) == 0 || len(p.Profiles) > maxProfiles || len(p.Rules) == 0 || len(p.Rules) > maxRules {
		return nil, errors.New("invalid model policy")
	}
	if (p.DecisionEvidenceVersion == 2) != (p.Calibration != nil) {
		return nil, errors.New("calibration requires decision evidence version 2")
	}
	profiles := make(map[string]Profile, len(p.Profiles))
	for _, profile := range p.Profiles {
		if err := profile.validate(); err != nil {
			return nil, err
		}
		if _, exists := profiles[profile.Name]; exists {
			return nil, errors.New("duplicate model profile")
		}
		profiles[profile.Name] = copyProfile(profile)
	}
	for role, rule := range p.Rules {
		if !validIdentifier(role) || rule.validate(profiles) != nil {
			return nil, errors.New("invalid model policy rule")
		}
	}
	if p.Calibration != nil {
		if err := p.Calibration.Validate(); err != nil {
			return nil, errors.New("invalid model policy calibration")
		}
		rule, ok := p.Rules["fixer"]
		baseline, baselineOK := profiles[p.Calibration.Baseline.Name]
		candidate, candidateOK := profiles[p.Calibration.Candidate.Name]
		if !ok || rule.DefaultProfile != p.Calibration.Baseline.Name || rule.EscalatedProfile == p.Calibration.Candidate.Name || !baselineOK || !candidateOK || !sameProfile(baseline, p.Calibration.Baseline) || !sameProfile(candidate, p.Calibration.Candidate) {
			return nil, errors.New("calibration profiles must be exact configured fixer profiles")
		}
	}
	return profiles, nil
}

func (r Rule) validate(profiles map[string]Profile) error {
	if r.ContextEscalationTokens < 1 || r.ContextEscalationTokens > MaxContextTokens || r.ContextEscalationBytes < 0 || r.ContextEscalationBytes > MaxContextBytes || r.FailureEscalationCount < 1 || r.FailureEscalationCount > MaxFailures || profiles[r.DefaultProfile].Name == "" || profiles[r.EscalatedProfile].Name == "" {
		return errors.New("invalid routing rule")
	}
	if r.CheapProfile != "" && (profiles[r.CheapProfile].Name == "" || r.CheapContextBytes < 1 || r.CheapContextBytes > MaxContextBytes) {
		return errors.New("cheap profile or byte threshold unavailable")
	}
	if r.CheapProfile == "" && r.CheapContextBytes != 0 {
		return errors.New("cheap byte threshold requires a cheap profile")
	}
	return nil
}

func (p Profile) validate() error {
	if !validIdentifier(p.Name) || !validIdentifier(p.Runtime) || !validIdentifier(p.Provider) || strings.TrimSpace(p.Model) == "" || len(p.Model) > 256 || !validIdentifier(p.Effort) {
		return errors.New("invalid model profile")
	}
	for _, value := range []*int64{p.ExpectedCostMicroUSD, p.ExpectedLatencyMillis} {
		if value != nil && *value < 0 {
			return errors.New("invalid profile estimate")
		}
	}
	return nil
}

func (r Request) validate() error {
	if !validIdentifier(r.Role) || (r.ReadOnly && r.Role != "explorer") || !validLevel(r.Complexity) || !validLevel(r.Risk) || !validLevel(r.Uncertainty) || r.ContextTokens < 0 || r.ContextTokens > MaxContextTokens || r.ContextBytes < 0 || r.ContextBytes > MaxContextBytes || r.Failures < 0 || r.Failures > MaxFailures {
		return errors.New("invalid routing request")
	}
	return nil
}

func validLevel(value Level) bool {
	return value == LevelLow || value == LevelMedium || value == LevelHigh
}

func validIdentifier(value string) bool {
	if len(value) == 0 || len(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

func clone(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyProfile(value Profile) Profile {
	value.ExpectedCostMicroUSD = clone(value.ExpectedCostMicroUSD)
	value.ExpectedLatencyMillis = clone(value.ExpectedLatencyMillis)
	return value
}

// FixtureCase is a labeled policy input for the executable baseline. It is not
// an evaluation of model output quality.
type FixtureCase struct {
	Label           string  `json:"label"`
	Request         Request `json:"request"`
	ExpectedProfile string  `json:"expected_profile"`
}

// FixtureBaseline returns deterministic routing cases used to guard policy
// behavior before live quality and intervention measurements exist.
func FixtureBaseline() []FixtureCase {
	return []FixtureCase{
		{Label: "cheap-read", Request: Request{Role: "explorer", ReadOnly: true, Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelMedium, ContextBytes: 120, ContextTokens: 120, Failures: 0}, ExpectedProfile: "economy"},
		{Label: "ordinary-write", Request: Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelLow, ContextTokens: 120, Failures: 0}, ExpectedProfile: "standard"},
		{Label: "failed-write", Request: Request{Role: "writer", Complexity: LevelMedium, Risk: LevelMedium, Uncertainty: LevelLow, ContextTokens: 120, Failures: 1}, ExpectedProfile: "architect"},
		{Label: "architect-plan", Request: Request{Role: "architect", Complexity: LevelLow, Risk: LevelLow, Uncertainty: LevelLow, ContextTokens: 120, Failures: 0}, ExpectedProfile: "architect"},
	}
}

// FixturePolicy is the explicit configuration for FixtureBaseline. It exists
// for tests and executable comparisons only, never as a production default.
func FixturePolicy() Policy {
	profiles := []Profile{{Name: "economy", Runtime: "fixture", Provider: "fixture", Model: "economy", Effort: "low"}, {Name: "standard", Runtime: "fixture", Provider: "fixture", Model: "standard", Effort: "medium"}, {Name: "architect", Runtime: "fixture", Provider: "fixture", Model: "architect", Effort: "high"}}
	rule := Rule{DefaultProfile: "standard", CheapProfile: "economy", CheapContextBytes: 1024, EscalatedProfile: "architect", ContextEscalationTokens: 1024, FailureEscalationCount: 1}
	return Policy{Version: 1, Profiles: profiles, Rules: map[string]Rule{"explorer": rule, "writer": rule, "architect": rule}}
}

// RunFixtureBaseline executes the labeled fixture cases and reports the names
// selected by this policy. Results are sorted by label for stable comparison.
func RunFixtureBaseline() (map[string]string, error) {
	policy := FixturePolicy()
	results := make(map[string]string)
	cases := FixtureBaseline()
	sort.Slice(cases, func(i, j int) bool { return cases[i].Label < cases[j].Label })
	for _, fixture := range cases {
		decision, err := Select(policy, fixture.Request)
		if err != nil {
			return nil, err
		}
		results[fixture.Label] = decision.Profile.Name
		if decision.Profile.Name != fixture.ExpectedProfile {
			return nil, errors.New("fixture baseline selection changed")
		}
	}
	return results, nil
}
