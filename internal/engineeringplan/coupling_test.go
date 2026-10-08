package engineeringplan

import (
	"encoding/json"
	"strings"
	"testing"
)

func couplingFixtureGraph() Graph {
	return resourceFixtureGraph(
		resourceFixtureTask("impl-alpha", Implementation, []string{"research"}, []string{"alpha/api.go"}, 60),
		resourceFixtureTask("impl-beta", Implementation, []string{"research"}, []string{"beta/api.go"}, 50),
		resourceFixtureTask("impl-gamma", Implementation, []string{"research"}, []string{"gamma/api.go"}, 40),
	)
}

func validCoupling(from, to, level string) TaskCoupling {
	return TaskCoupling{From: from, To: to, Level: level, Reason: "shared_api", Provenance: CouplingProvenancePlannerAdvisory, Evidence: from + " observes " + to + " via shared API"}
}

func TestTaskCouplingValidationAcceptsBoundedSet(t *testing.T) {
	couplings := []TaskCoupling{
		{From: "impl-alpha", To: "impl-beta", Level: CouplingC3, Reason: "shared_api", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "alpha observes beta via shared API"},
		{From: "impl-alpha", To: "impl-gamma", Level: CouplingC1, Reason: "import_neighborhood", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "shared import neighborhood"},
		{From: "impl-beta", To: "impl-gamma", Level: CouplingC4, Reason: "observed_generator_family", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "gen.go directive //go:generate go run . binds beta and gamma outputs"},
	}
	if err := ValidateTaskCouplings(couplings); err != nil {
		t.Fatalf("valid couplings rejected: %v", err)
	}
	graph := couplingFixtureGraph()
	graph.Couplings = couplings
	// C4 without an implementation dependency is a shape-valid graph but an
	// unsafe staged split; generic validation passes, staged validation owns
	// the hard gate.
	if err := graph.Validate(); err != nil {
		t.Fatalf("graph with couplings rejected at shape level: %v", err)
	}
}

func TestTaskCouplingValidationRejectsMalformed(t *testing.T) {
	cases := []struct {
		name      string
		couplings []TaskCoupling
	}{
		{"self", []TaskCoupling{{From: "impl-a", To: "impl-a", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}}},
		{"reverse_order", []TaskCoupling{{From: "impl-b", To: "impl-a", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}}},
		{"duplicate", []TaskCoupling{validCoupling("impl-a", "impl-b", CouplingC1), validCoupling("impl-a", "impl-b", CouplingC2)}},
		{"bad_level", []TaskCoupling{{From: "impl-a", To: "impl-b", Level: "C9", Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}}},
		{"bad_provenance", []TaskCoupling{{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: "forged_source", Evidence: "e"}}},
		{"empty_evidence", []TaskCoupling{{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "  "}}},
		{"empty_reason", []TaskCoupling{{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}}},
		{"bad_id", []TaskCoupling{{From: "Impl-A", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateTaskCouplings(tc.couplings); err == nil {
				t.Fatalf("malformed coupling %s admitted", tc.name)
			}
		})
	}
	many := make([]TaskCoupling, 0, MaxTaskCouplings+1)
	for i := 0; i <= MaxTaskCouplings; i++ {
		to := "impl-b"
		if i < 10 {
			to = "impl-b0" + string(rune('0'+i))
		} else {
			to = "impl-b" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		}
		many = append(many, TaskCoupling{From: "impl-a", To: to, Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"})
	}
	// 29 distinct pairs; the bound is 28.
	if err := ValidateTaskCouplings(many); err == nil {
		t.Fatal("coupling set beyond 28 admitted")
	}
}

func TestCouplingDoesNotCreateReadinessOrOwnership(t *testing.T) {
	graph := couplingFixtureGraph()
	plainReady, err := Ready(graph)
	if err != nil {
		t.Fatal(err)
	}
	var plainImpls int
	for _, task := range plainReady {
		if task.Kind == Implementation {
			plainImpls++
		}
	}
	coupled := couplingFixtureGraph()
	coupled.Couplings = []TaskCoupling{validCoupling("impl-alpha", "impl-beta", CouplingC3)}
	if err := coupled.Validate(); err != nil {
		t.Fatalf("coupled graph rejected: %v", err)
	}
	coupledReady, err := Ready(coupled)
	if err != nil {
		t.Fatal(err)
	}
	var coupledImpls int
	for _, task := range coupledReady {
		if task.Kind == Implementation {
			coupledImpls++
		}
	}
	if plainImpls != coupledImpls || plainImpls != 3 {
		t.Fatalf("coupling changed readiness: plain=%d coupled=%d", plainImpls, coupledImpls)
	}
	// Ownership is unchanged: write paths are identical with and without coupling.
	plainByID := map[string]Task{}
	for _, task := range graph.Tasks {
		plainByID[task.ID] = task
	}
	for _, task := range coupled.Tasks {
		if len(task.WritePaths) != len(plainByID[task.ID].WritePaths) {
			t.Fatalf("coupling changed ownership for %q", task.ID)
		}
	}
}

func couplingWaveFixture() (Graph, []Task, []TaskResourceDemand, ResourceCapacity) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-alpha", Implementation, []string{"research"}, []string{"alpha/api.go"}, 100),
		resourceFixtureTask("impl-beta", Implementation, []string{"research"}, []string{"beta/api.go"}, 90),
		resourceFixtureTask("impl-gamma", Implementation, []string{"research"}, []string{"gamma/api.go"}, 10),
	)
	ready := []Task{}
	for _, task := range graph.Tasks {
		if task.Kind == Implementation {
			ready = append(ready, task)
		}
	}
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-alpha", 1, 1, 0, 1),
		resourceFixtureDemand("impl-beta", 1, 1, 0, 1),
		resourceFixtureDemand("impl-gamma", 1, 1, 0, 1),
	}
	capacity := resourceFixtureCapacity(10, 10, 3, 2)
	return graph, ready, demands, capacity
}

func TestCouplingAwareObjectiveAvoidsStrongCoupledSameWave(t *testing.T) {
	graph, ready, demands, capacity := couplingWaveFixture()
	couplings := []TaskCoupling{{From: "impl-alpha", To: "impl-beta", Level: CouplingC3, Reason: "shared_api", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "alpha and beta share critical API"}}
	cohort, err := SelectResourceCohortCouplingAware(graph, ready, demands, capacity, 2, couplings)
	if err != nil {
		t.Fatalf("coupling-aware cohort rejected: %v", err)
	}
	got := map[string]bool{}
	for _, task := range cohort.Tasks {
		got[task.ID] = true
	}
	// Greedy critical-path order would admit alpha+beta (100+90). The
	// coupling-aware optimum admits alpha+gamma instead, avoiding the strong
	// C3 pair despite disjoint paths with an independent third task available.
	if !(got["impl-alpha"] && got["impl-gamma"] && !got["impl-beta"]) {
		t.Fatalf("coupling-aware wave did not avoid C3 pair: %v", got)
	}
	greedy, err := SelectResourceCohort(graph, ready, demands, capacity, 2)
	if err != nil {
		t.Fatal(err)
	}
	greedyGot := map[string]bool{}
	for _, task := range greedy.Tasks {
		greedyGot[task.ID] = true
	}
	if !(greedyGot["impl-alpha"] && greedyGot["impl-beta"]) {
		t.Fatalf("greedy baseline did not prefer the critical pair: %v", greedyGot)
	}
}

func TestCouplingAwareIsPermutationDeterministic(t *testing.T) {
	graph, ready, demands, capacity := couplingWaveFixture()
	couplings := []TaskCoupling{{From: "impl-alpha", To: "impl-beta", Level: CouplingC3, Reason: "shared_api", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "alpha and beta share critical API"}}
	want, err := SelectResourceCohortCouplingAware(graph, ready, demands, capacity, 2, couplings)
	if err != nil {
		t.Fatal(err)
	}
	// Reverse graph task order, ready order and demand order.
	reversedGraph := graph
	reversedGraph.Tasks = append([]Task(nil), graph.Tasks...)
	for i, j := 0, len(reversedGraph.Tasks)-1; i < j; i, j = i+1, j-1 {
		reversedGraph.Tasks[i], reversedGraph.Tasks[j] = reversedGraph.Tasks[j], reversedGraph.Tasks[i]
	}
	reversedReady := append([]Task(nil), ready...)
	for i, j := 0, len(reversedReady)-1; i < j; i, j = i+1, j-1 {
		reversedReady[i], reversedReady[j] = reversedReady[j], reversedReady[i]
	}
	reversedDemands := append([]TaskResourceDemand(nil), demands...)
	for i, j := 0, len(reversedDemands)-1; i < j; i, j = i+1, j-1 {
		reversedDemands[i], reversedDemands[j] = reversedDemands[j], reversedDemands[i]
	}
	got, err := SelectResourceCohortCouplingAware(reversedGraph, reversedReady, reversedDemands, capacity, 2, couplings)
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Tasks) != len(got.Tasks) {
		t.Fatalf("permutation changed cohort size: %+v vs %+v", want.Tasks, got.Tasks)
	}
	for i := range want.Tasks {
		if want.Tasks[i].ID != got.Tasks[i].ID {
			t.Fatalf("permutation changed cohort: %+v vs %+v", want.Tasks, got.Tasks)
		}
	}
	wavesWant, err := SelectResourceWavesCouplingAware(graph, ready, demands, capacity, 2, couplings)
	if err != nil {
		t.Fatal(err)
	}
	wavesGot, err := SelectResourceWavesCouplingAware(reversedGraph, reversedReady, reversedDemands, capacity, 2, couplings)
	if err != nil {
		t.Fatal(err)
	}
	if len(wavesWant) != len(wavesGot) {
		t.Fatalf("permutation changed wave count: %d vs %d", len(wavesWant), len(wavesGot))
	}
	for i := range wavesWant {
		if len(wavesWant[i].Tasks) != len(wavesGot[i].Tasks) {
			t.Fatalf("permutation changed wave %d", i)
		}
		for j := range wavesWant[i].Tasks {
			if wavesWant[i].Tasks[j].ID != wavesGot[i].Tasks[j].ID {
				t.Fatalf("permutation changed wave %d task %d", i, j)
			}
		}
	}
}

func TestCouplingAwareWavesRejectC4WithoutDependency(t *testing.T) {
	graph, ready, demands, capacity := couplingWaveFixture()
	couplings := []TaskCoupling{{From: "impl-alpha", To: "impl-beta", Level: CouplingC4, Reason: "observed_generator_family", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "gen.go directive binds alpha and beta outputs"}}
	if _, err := SelectResourceWavesCouplingAware(graph, ready, demands, capacity, 2, couplings); err == nil {
		t.Fatal("C4 generator split admitted to waves without dependency")
	}
	// Single-wave selection treats C4 as a hard conflict, not a silent fit.
	cohort, err := SelectResourceCohortCouplingAware(graph, ready, demands, capacity, 3, couplings)
	if err != nil {
		t.Fatalf("C4 cohort rejected outright: %v", err)
	}
	for _, a := range cohort.Tasks {
		for _, b := range cohort.Tasks {
			if a.ID != b.ID && ((a.ID == "impl-alpha" && b.ID == "impl-beta") || (a.ID == "impl-beta" && b.ID == "impl-alpha")) {
				t.Fatalf("C4 pair co-selected in one wave: %+v", cohort.Tasks)
			}
		}
	}
}

func TestStagedC4RequiresOneOwnerOrDependency(t *testing.T) {
	base := func(tasks ...Task) Graph {
		all := []Task{{ID: "research", Kind: Research, Title: "Shared research", ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "fact", Description: "facts"}}, EstimatedSeconds: 5, Completed: true}}
		all = append(all, tasks...)
		all = append(all,
			Task{ID: "verify", Kind: Verification, Title: "Verify", Dependencies: []string{}, ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "check", Description: "checks"}}, EstimatedSeconds: 5},
			Task{ID: "review", Kind: Review, Title: "Review", Dependencies: []string{"verify"}, ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "review", Description: "review"}}, EstimatedSeconds: 5},
		)
		// Wire gates to every implementation.
		implIDs := []string{}
		for _, task := range tasks {
			if task.Kind == Implementation {
				implIDs = append(implIDs, task.ID)
			}
		}
		for i, task := range all {
			if task.ID == "verify" {
				all[i].Dependencies = append([]string(nil), implIDs...)
			}
		}
		return Graph{Version: Version, Mode: ModeGraph, Summary: "staged C4 fixture", Tasks: all}
	}
	impl := func(id string, deps []string, write string) Task {
		return Task{ID: id, Kind: Implementation, Title: "Task " + id, Dependencies: deps, ScopePaths: []string{"."}, WritePaths: []string{write}, ExpectedEvidence: []Evidence{{Kind: "file", Description: "output " + id}}, EstimatedSeconds: 10}
	}
	unsafe := base(impl("impl-a", []string{"research"}, "a/out.go"), impl("impl-b", []string{"research"}, "b/out.go"))
	unsafe.Couplings = []TaskCoupling{{From: "impl-a", To: "impl-b", Level: CouplingC4, Reason: "observed_generator_family", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "gen.go binds a/out.go and b/out.go"}}
	if err := ValidateAutonomousGraphWithStagedImplementations(unsafe, 8); err == nil {
		t.Fatal("unsafe C4 split without dependency admitted")
	}
	safe := base(impl("impl-a", []string{"research"}, "a/out.go"), impl("impl-b", []string{"impl-a", "research"}, "b/out.go"))
	safe.Couplings = []TaskCoupling{{From: "impl-a", To: "impl-b", Level: CouplingC4, Reason: "observed_generator_family", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "gen.go binds a/out.go and b/out.go"}}
	if err := ValidateAutonomousGraphWithStagedImplementations(safe, 8); err != nil {
		t.Fatalf("C4 with explicit dependency rejected: %v", err)
	}
	// One owner (no split) needs no coupling and is accepted.
	single := base(impl("impl-a", []string{"research"}, "a/out.go"))
	if err := ValidateAutonomousGraphWithStagedImplementations(single, 8); err != nil {
		t.Fatalf("single-owner staged graph rejected: %v", err)
	}
}

func TestLegacyGraphOmitsCouplingsField(t *testing.T) {
	legacy := Graph{Version: Version, Mode: ModeGraph, Summary: "legacy planner output", Tasks: []Task{
		resourceFixtureTask("impl-alpha", Implementation, []string{"research"}, []string{"alpha/api.go"}, 60),
		resourceFixtureTask("research", Research, nil, nil, 5),
		resourceFixtureTask("verify", Verification, []string{"impl-alpha"}, nil, 10),
		resourceFixtureTask("review", Review, []string{"verify"}, nil, 5),
	}}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "couplings") {
		t.Fatalf("legacy graph carries couplings bytes: %s", raw)
	}
	parsed, err := ParsePlannerGraph(raw)
	if err != nil {
		t.Fatalf("legacy planner JSON rejected: %v", err)
	}
	if len(parsed.Couplings) != 0 {
		t.Fatalf("legacy parse invented couplings: %+v", parsed.Couplings)
	}
}

func TestForgedObservedProvenanceRejected(t *testing.T) {
	for _, provenance := range []string{CouplingProvenanceGeneratorObserved, CouplingProvenanceTopologyObserved, "forged_source"} {
		coupling := TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "shared_api", Provenance: provenance, Evidence: "planner claims observed binding"}
		if err := ValidateTaskCouplings([]TaskCoupling{coupling}); err == nil {
			t.Fatalf("forged provenance %q admitted", provenance)
		}
	}
	graph := couplingFixtureGraph()
	graph.Couplings = []TaskCoupling{{From: "impl-alpha", To: "impl-beta", Level: CouplingC1, Reason: "shared_api", Provenance: CouplingProvenanceGeneratorObserved, Evidence: "planner claims generator truth"}}
	if err := graph.Validate(); err == nil {
		t.Fatal("graph with forged observed provenance admitted")
	}
	raw := []byte(`{"version":1,"mode":"graph","summary":"forged","tasks":[{"id":"impl-a","parent_id":"","kind":"implementation","title":"Task impl-a","dependencies":[],"scope_paths":["."],"write_paths":["a/out.go"],"expected_evidence":[{"kind":"file","description":"output"}],"estimated_seconds":10}],"couplings":[{"from":"impl-a","to":"impl-b","level":"C1","reason":"shared_api","provenance":"observed_generator_binding","evidence":"planner claims observed truth"}]}`)
	if _, err := ParsePlannerGraph(raw); err == nil {
		t.Fatal("planner wire with forged observed provenance admitted")
	}
}

func TestCouplingWireSchemaAdmitsOnlyAdvisory(t *testing.T) {
	schema := string(PlannerJSONSchemaWithCouplings())
	if !strings.Contains(schema, "planner_declared_advisory") {
		t.Fatal("coupling wire schema omits planner_declared_advisory")
	}
	if strings.Contains(schema, "observed_generator_binding") || strings.Contains(schema, "observed_source_topology") {
		t.Fatal("coupling wire schema admits forged observed provenance")
	}
	base := string(PlannerJSONSchema())
	if strings.Contains(base, "couplings") {
		t.Fatal("frozen planner schema carries couplings bytes")
	}
	// Derived schema preserves every task property of the frozen schema.
	for _, token := range []string{`"expected_evidence"`, `"estimated_seconds"`, `"scope_paths"`, `"write_paths"`} {
		if !strings.Contains(schema, token) {
			t.Fatalf("derived coupling schema lost frozen task property %s", token)
		}
	}
}

func TestCouplingValidationRejectsEndpointsUTF8ControlSizeAndOrder(t *testing.T) {
	cases := []struct {
		name     string
		coupling TaskCoupling
	}{
		{"unknown_from", TaskCoupling{From: "Impl-A", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}},
		{"self", TaskCoupling{From: "impl-a", To: "impl-a", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}},
		{"reverse_order", TaskCoupling{From: "impl-b", To: "impl-a", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}},
		{"control_reason", TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "bad\x00reason", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}},
		{"control_evidence", TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "bad\x07evidence"}},
		{"invalid_utf8_reason", TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "bad\xffreason", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}},
		{"invalid_utf8_evidence", TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "bad\xffevidence"}},
		{"long_reason", TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: strings.Repeat("r", 129), Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}},
		{"long_evidence", TaskCoupling{From: "impl-a", To: "impl-b", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: strings.Repeat("e", 257)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateTaskCouplings([]TaskCoupling{tc.coupling}); err == nil {
				t.Fatalf("malformed coupling %s admitted", tc.name)
			}
		})
	}
	graph := couplingFixtureGraph()
	graph.Couplings = []TaskCoupling{{From: "impl-alpha", To: "research", Level: CouplingC1, Reason: "r", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "e"}}
	if err := graph.Validate(); err == nil {
		t.Fatal("coupling to non-implementation task admitted")
	}
}

func TestCouplingRiskFirstSplitsC3PairWithCapacityForAll(t *testing.T) {
	graph, ready, demands, _ := couplingWaveFixture()
	capacity := resourceFixtureCapacity(10, 10, 3, 3)
	couplings := []TaskCoupling{{From: "impl-alpha", To: "impl-beta", Level: CouplingC3, Reason: "shared_api", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "alpha and beta share critical API"}}
	waves, err := SelectResourceWavesCouplingAware(graph, ready, demands, capacity, 3, couplings)
	if err != nil {
		t.Fatalf("coupling-aware waves rejected: %v", err)
	}
	if len(waves) != 2 {
		t.Fatalf("risk-first waves did not split the C3 pair with capacity for all 3: %d waves", len(waves))
	}
	first := map[string]bool{}
	for _, task := range waves[0].Tasks {
		first[task.ID] = true
	}
	if first["impl-alpha"] && first["impl-beta"] {
		t.Fatalf("C3 pair co-scheduled despite capacity to split: %v", first)
	}
	if !(first["impl-gamma"] && (first["impl-alpha"] || first["impl-beta"])) {
		t.Fatalf("independent third was not admitted with one C3 member: %v", first)
	}
	// The risk-excluded member carries the explicit bounded reason.
	foundRisk := false
	for _, blocked := range waves[0].Blocked {
		if blocked.Reason == "coupling_risk" && blocked.BlockingTaskID != "" {
			foundRisk = true
		}
	}
	if !foundRisk {
		t.Fatalf("risk-based exclusion lacks explicit coupling_risk reason: %+v", waves[0].Blocked)
	}
	// Determinism under reversed input order.
	reversedGraph := graph
	reversedGraph.Tasks = append([]Task(nil), graph.Tasks...)
	for i, j := 0, len(reversedGraph.Tasks)-1; i < j; i, j = i+1, j-1 {
		reversedGraph.Tasks[i], reversedGraph.Tasks[j] = reversedGraph.Tasks[j], reversedGraph.Tasks[i]
	}
	reversedReady := append([]Task(nil), ready...)
	for i, j := 0, len(reversedReady)-1; i < j; i, j = i+1, j-1 {
		reversedReady[i], reversedReady[j] = reversedReady[j], reversedReady[i]
	}
	reversedDemands := append([]TaskResourceDemand(nil), demands...)
	for i, j := 0, len(reversedDemands)-1; i < j; i, j = i+1, j-1 {
		reversedDemands[i], reversedDemands[j] = reversedDemands[j], reversedDemands[i]
	}
	again, err := SelectResourceWavesCouplingAware(reversedGraph, reversedReady, reversedDemands, capacity, 3, couplings)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != len(waves) || len(again[0].Tasks) != len(waves[0].Tasks) {
		t.Fatalf("permutation changed risk-first waves: %+v vs %+v", waves, again)
	}
	for i := range waves[0].Tasks {
		if again[0].Tasks[i].ID != waves[0].Tasks[i].ID {
			t.Fatalf("permutation changed risk-first cohort: %+v vs %+v", waves[0].Tasks, again[0].Tasks)
		}
	}
}

func TestFrozenV39SelectorStaysCountFirst(t *testing.T) {
	graph, ready, demands, _ := couplingWaveFixture()
	capacity := resourceFixtureCapacity(10, 10, 3, 3)
	couplings := []TaskCoupling{{From: "impl-alpha", To: "impl-beta", Level: CouplingC3, Reason: "shared_api", Provenance: CouplingProvenancePlannerAdvisory, Evidence: "alpha and beta share critical API"}}
	// The frozen v39 lexicographic selector ignores couplings and packs by
	// admitted count first: all three fit capacity 3 and are admitted together.
	lex, err := SelectResourceWavesLexicographic(graph, ready, demands, capacity, 3)
	if err != nil {
		t.Fatalf("frozen lexicographic waves rejected: %v", err)
	}
	if len(lex) != 1 || len(lex[0].Tasks) != 3 {
		t.Fatalf("frozen v39 policy changed count-first packing: %+v", lex)
	}
	// The coupling-aware selector differs only here: it splits the C3 pair.
	aware, err := SelectResourceWavesCouplingAware(graph, ready, demands, capacity, 3, couplings)
	if err != nil {
		t.Fatal(err)
	}
	if len(aware) == 1 && len(aware[0].Tasks) == 3 {
		t.Fatal("coupling-aware selector did not differ from frozen count-first packing")
	}
}
