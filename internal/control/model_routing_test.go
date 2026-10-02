package control

import (
	"strings"
	"testing"

	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/modelpolicy"
	"harness.local/engorch/internal/runtime"
)

func adaptivePlannerPolicy(base runtime.Profile, bytes int64) *modelpolicy.Policy {
	return &modelpolicy.Policy{
		Version: 1,
		Profiles: []modelpolicy.Profile{
			{Name: "standard", Runtime: base.Runtime, Provider: base.Provider, Model: base.Model, Effort: base.Effort},
			{Name: "escalated", Runtime: base.Runtime, Provider: base.Provider, Model: "planner-strong", Effort: "high"},
		},
		Rules: map[string]modelpolicy.Rule{
			"planner": {DefaultProfile: "standard", EscalatedProfile: "escalated", ContextEscalationTokens: modelpolicy.MaxContextTokens, ContextEscalationBytes: bytes, FailureEscalationCount: 1},
		},
	}
}

func TestAdaptivePlannerInvocationRoutesFromExactFinalInputBytes(t *testing.T) {
	c := creation(t)
	c.Config.PlannerContract = plannerContractV1
	c.Config.ModelPolicy = adaptivePlannerPolicy(c.Config.Planner, 100)
	objective := strings.Repeat("x", 150)
	i, err := plannerInvocation(c.Config, objective)
	if err != nil {
		t.Fatal(err)
	}
	if i.Profile.Model != "planner-strong" || i.Profile.Effort != "high" {
		t.Fatalf("large final planner input did not select configured escalation: %+v", i.Profile)
	}
	if i.Input == objective || len(i.Input) < len(objective) {
		t.Fatal("planner contract wrapper was not included in byte-based selection")
	}

	legacy := c.Config
	legacy.ModelPolicy = nil
	legacy.PlannerContract = ""
	legacyInvocation, err := plannerInvocation(legacy, objective)
	if err != nil {
		t.Fatal(err)
	}
	want, err := runtime.NewInvocation(legacy.Planner, objective)
	if err != nil || legacyInvocation != want {
		t.Fatalf("nil model policy changed legacy planner input or identity: %v", err)
	}
}

func TestAdaptiveWriterRoutingIgnoresUnknownGraphAttempt(t *testing.T) {
	c := creation(t)
	base := runtime.Profile{Runtime: "fake", Provider: "deterministic", Model: "writer-standard", Effort: "medium", Role: "writer"}
	c.Config.Writer = &base
	c.Config.ModelPolicy = &modelpolicy.Policy{
		Version: 1,
		Profiles: []modelpolicy.Profile{
			{Name: "standard", Runtime: base.Runtime, Provider: base.Provider, Model: base.Model, Effort: base.Effort},
			{Name: "escalated", Runtime: base.Runtime, Provider: base.Provider, Model: "writer-strong", Effort: "high"},
		},
		Rules: map[string]modelpolicy.Rule{
			"writer": {DefaultProfile: "standard", EscalatedProfile: "escalated", ContextEscalationTokens: modelpolicy.MaxContextTokens, ContextEscalationBytes: modelpolicy.MaxContextBytes, FailureEscalationCount: 1},
		},
	}
	graph := validGraphFixture()
	for i := range graph.Tasks {
		if graph.Tasks[i].ID == "impl" {
			graph.Tasks[i].Attempts = []engineeringplan.Attempt{{ID: "attempt-unknown", Outcome: engineeringplan.AttemptUnknown}}
		}
	}
	s := Snapshot{Creation: c, Graph: &GraphState{Graph: graph}}
	input := "final writer input"
	if got := modelFailureCount(s, "writer", input); got != 0 {
		t.Fatalf("unknown attempt counted as a completed failure: %d", got)
	}
	selected, err := modelProfileForSnapshot(s, "writer", input)
	if err != nil || selected != base {
		t.Fatalf("unknown attempt changed configured writer profile: %+v %v", selected, err)
	}

	for i := range s.Graph.Graph.Tasks {
		if s.Graph.Graph.Tasks[i].ID == "impl" {
			s.Graph.Graph.Tasks[i].Attempts = append(s.Graph.Graph.Tasks[i].Attempts, engineeringplan.Attempt{ID: "attempt-failed", Outcome: engineeringplan.AttemptFailed})
		}
	}
	if got := modelFailureCount(s, "writer", input); got != 1 {
		t.Fatalf("completed failed attempt was not counted: %d", got)
	}
	selected, err = modelProfileForSnapshot(s, "writer", input)
	if err != nil || selected.Model != "writer-strong" || selected.Effort != "high" {
		t.Fatalf("completed failure did not select configured escalation: %+v %v", selected, err)
	}
}

func TestConfigRejectsCheapProfileForWriteRole(t *testing.T) {
	c := creation(t)
	base := runtime.Profile{Runtime: "fake", Provider: "deterministic", Model: "writer-standard", Effort: "medium", Role: "writer"}
	c.Config.Writer = &base
	c.Config.ModelPolicy = &modelpolicy.Policy{
		Version: 1,
		Profiles: []modelpolicy.Profile{
			{Name: "standard", Runtime: base.Runtime, Provider: base.Provider, Model: base.Model, Effort: base.Effort},
			{Name: "economy", Runtime: base.Runtime, Provider: base.Provider, Model: "writer-economy", Effort: "low"},
			{Name: "strong", Runtime: base.Runtime, Provider: base.Provider, Model: "writer-strong", Effort: "high"},
		},
		Rules: map[string]modelpolicy.Rule{
			"writer": {DefaultProfile: "standard", CheapProfile: "economy", CheapContextBytes: 1024, EscalatedProfile: "strong", ContextEscalationTokens: modelpolicy.MaxContextTokens, ContextEscalationBytes: modelpolicy.MaxContextBytes, FailureEscalationCount: 1},
		},
	}
	if err := c.Config.Validate(); err == nil {
		t.Fatal("configuration admitted an economy route for a mutating writer role")
	}
}
