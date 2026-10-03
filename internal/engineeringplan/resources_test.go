package engineeringplan

import (
	"reflect"
	"strings"
	"testing"
)

func resourceFixtureTask(id string, kind Kind, deps []string, writes []string, seconds int) Task {
	return Task{
		ID: id, Kind: kind, Title: "Task " + id, Dependencies: deps,
		ScopePaths: []string{"."}, WritePaths: writes,
		ExpectedEvidence: []Evidence{{Kind: "result", Description: "result for " + id}},
		EstimatedSeconds: seconds,
	}
}

func resourceFixtureGraph(implementations ...Task) Graph {
	tasks := []Task{
		{ID: "research", Kind: Research, Title: "Shared research", ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "fact", Description: "shared API facts"}}, EstimatedSeconds: 5, Completed: true},
	}
	tasks = append(tasks, implementations...)
	verifyDeps := make([]string, 0, len(implementations))
	for _, task := range implementations {
		verifyDeps = append(verifyDeps, task.ID)
	}
	tasks = append(tasks,
		resourceFixtureTask("verify", Verification, verifyDeps, nil, 10),
		resourceFixtureTask("review", Review, []string{"verify"}, nil, 5),
	)
	return Graph{Version: Version, Mode: ModeGraph, Summary: "resource selection fixture", Tasks: tasks}
}

var fixtureRuntime = RuntimeResourceKey{ProfileID: "profile-codex", Provider: "openai", Model: "gpt-6-luna"}

func resourceFixtureDemand(id string, cpu, memory int64, verify, runtime int) TaskResourceDemand {
	return TaskResourceDemand{TaskID: id, CPUMilli: cpu, MemoryMiB: memory, VerificationSlots: verify, Runtime: fixtureRuntime, RuntimeSlots: runtime}
}

func resourceFixtureCapacity(cpu, memory int64, verify, runtime int) ResourceCapacity {
	return ResourceCapacity{
		CPUMilli: cpu, MemoryMiB: memory, VerificationSlots: verify, TotalRuntimeSlots: maxResourceSlots,
		ProviderSlots: []ProviderSlotLimit{{Provider: fixtureRuntime.Provider, Slots: maxResourceSlots}},
		ModelSlots:    []ModelSlotLimit{{Model: providerModel(fixtureRuntime), Slots: maxResourceSlots}},
		RuntimeSlots:  []RuntimeSlotLimit{{Runtime: fixtureRuntime, Slots: runtime}},
	}
}

func readyImplementations(t *testing.T, graph Graph) []Task {
	t.Helper()
	ready, err := Ready(graph)
	if err != nil {
		t.Fatal(err)
	}
	var implementations []Task
	for _, task := range ready {
		if task.Kind == Implementation {
			implementations = append(implementations, task)
		}
	}
	return implementations
}

func TestSelectResourceCohortUsesCriticalPathAndWeightedCapacities(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-alpha", Implementation, []string{"research"}, []string{"alpha/api.go"}, 60),
		resourceFixtureTask("impl-beta", Implementation, []string{"research"}, []string{"beta/api.go"}, 50),
		resourceFixtureTask("impl-gamma", Implementation, []string{"research"}, []string{"gamma/api.go"}, 40),
	)
	ready := readyImplementations(t, graph)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-alpha", 4, 8, 1, 1),
		resourceFixtureDemand("impl-beta", 2, 5, 1, 1),
		resourceFixtureDemand("impl-gamma", 1, 1, 1, 1),
	}
	cohort, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(10, 9, 2, 2), 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{cohort.Tasks[0].ID, cohort.Tasks[1].ID}; !reflect.DeepEqual(got, []string{"impl-alpha", "impl-gamma"}) {
		t.Fatalf("critical-path weighted cohort = %v", got)
	}
	if len(cohort.Blocked) != 1 || cohort.Blocked[0].TaskID != "impl-beta" || cohort.Blocked[0].Reason != "capacity" || cohort.Blocked[0].Resource != "memory_mib" {
		t.Fatalf("unexpected blocked estimate: %+v", cohort.Blocked)
	}
	wantTotals := ResourceTotals{
		EstimatedCPUMilli: 5, EstimatedMemoryMiB: 9, EstimatedVerificationSlots: 2,
		TotalRuntimeSlots: 2,
		ProviderSlots:     []ProviderSlotTotal{{Provider: fixtureRuntime.Provider, Slots: 2}},
		ModelSlots:        []ModelSlotTotal{{Model: providerModel(fixtureRuntime), Slots: 2}},
		RuntimeSlots:      []RuntimeSlotTotal{{Runtime: fixtureRuntime, Slots: 2}},
	}
	if !reflect.DeepEqual(cohort.Estimated, wantTotals) {
		t.Fatalf("estimated totals = %+v, want %+v", cohort.Estimated, wantTotals)
	}

	cpuBound, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(5, 20, 2, 2), 2)
	if err != nil || len(cpuBound.Tasks) != 2 || len(cpuBound.Blocked) != 1 || cpuBound.Blocked[0].TaskID != "impl-beta" || cpuBound.Blocked[0].Resource != "cpu_milli" {
		t.Fatalf("CPU estimate was not capacity checked: cohort=%+v err=%v", cpuBound, err)
	}
	runtimeBound, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(10, 20, 3, 1), 3)
	if err != nil || len(runtimeBound.Tasks) != 1 || len(runtimeBound.Blocked) != 2 || runtimeBound.Blocked[0].Resource != "runtime_slots" || runtimeBound.Blocked[0].Runtime == nil || *runtimeBound.Blocked[0].Runtime != fixtureRuntime {
		t.Fatalf("exact runtime route was not capacity checked: cohort=%+v err=%v", runtimeBound, err)
	}
	verificationBound, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(10, 20, 1, 3), 3)
	if err != nil || len(verificationBound.Tasks) != 1 || len(verificationBound.Blocked) != 2 || verificationBound.Blocked[0].Resource != "verification_slots" {
		t.Fatalf("verification estimate was not capacity checked: cohort=%+v err=%v", verificationBound, err)
	}
}

func TestSelectResourceCohortSharesProviderAndModelCapacityAcrossProfiles(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	routeA := RuntimeResourceKey{ProfileID: "profile-a", Provider: "openai", Model: "gpt-6-luna"}
	routeB := RuntimeResourceKey{ProfileID: "profile-b", Provider: "openai", Model: "gpt-6-luna"}
	demands := []TaskResourceDemand{
		{TaskID: "impl-a", CPUMilli: 1, MemoryMiB: 1, VerificationSlots: 1, Runtime: routeA, RuntimeSlots: 1},
		{TaskID: "impl-b", CPUMilli: 1, MemoryMiB: 1, VerificationSlots: 1, Runtime: routeB, RuntimeSlots: 1},
	}
	capacity := ResourceCapacity{
		CPUMilli: 2, MemoryMiB: 2, VerificationSlots: 2, TotalRuntimeSlots: 2,
		ProviderSlots: []ProviderSlotLimit{{Provider: "openai", Slots: 2}},
		ModelSlots:    []ModelSlotLimit{{Model: ProviderModelKey{Provider: "openai", Model: "gpt-6-luna"}, Slots: 1}},
		RuntimeSlots:  []RuntimeSlotLimit{{Runtime: routeA, Slots: 2}, {Runtime: routeB, Slots: 2}},
	}
	cohort, err := SelectResourceCohort(graph, ready, demands, capacity, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(cohort.Tasks) != 1 || cohort.Tasks[0].ID != "impl-a" || len(cohort.Blocked) != 1 || cohort.Blocked[0].TaskID != "impl-b" || cohort.Blocked[0].Resource != "model_slots" || cohort.Blocked[0].Runtime == nil || *cohort.Blocked[0].Runtime != routeB {
		t.Fatalf("shared provider/model capacity was split by profile: %+v", cohort)
	}
	if cohort.Estimated.TotalRuntimeSlots != 1 || !reflect.DeepEqual(cohort.Estimated.ModelSlots, []ModelSlotTotal{{Model: ProviderModelKey{Provider: "openai", Model: "gpt-6-luna"}, Slots: 1}}) {
		t.Fatalf("shared model totals are incorrect: %+v", cohort.Estimated)
	}
}

func TestSelectResourceCohortEnforcesProviderAndGlobalRuntimeLimits(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	routeA := RuntimeResourceKey{ProfileID: "profile-a", Provider: "openai", Model: "model-a"}
	routeB := RuntimeResourceKey{ProfileID: "profile-b", Provider: "openai", Model: "model-b"}
	demands := []TaskResourceDemand{
		{TaskID: "impl-a", CPUMilli: 1, MemoryMiB: 1, VerificationSlots: 1, Runtime: routeA, RuntimeSlots: 1},
		{TaskID: "impl-b", CPUMilli: 1, MemoryMiB: 1, VerificationSlots: 1, Runtime: routeB, RuntimeSlots: 1},
	}
	baseCapacity := ResourceCapacity{
		CPUMilli: 2, MemoryMiB: 2, VerificationSlots: 2, TotalRuntimeSlots: 2,
		ProviderSlots: []ProviderSlotLimit{{Provider: "openai", Slots: 2}},
		ModelSlots: []ModelSlotLimit{
			{Model: ProviderModelKey{Provider: "openai", Model: "model-a"}, Slots: 2},
			{Model: ProviderModelKey{Provider: "openai", Model: "model-b"}, Slots: 2},
		},
		RuntimeSlots: []RuntimeSlotLimit{{Runtime: routeA, Slots: 2}, {Runtime: routeB, Slots: 2}},
	}
	providerCapacity := baseCapacity
	providerCapacity.ProviderSlots = []ProviderSlotLimit{{Provider: "openai", Slots: 1}}
	providerBound, err := SelectResourceCohort(graph, ready, demands, providerCapacity, 2)
	if err != nil || len(providerBound.Tasks) != 1 || providerBound.Blocked[0].Resource != "provider_slots" {
		t.Fatalf("provider-wide limit was not enforced: cohort=%+v err=%v", providerBound, err)
	}
	globalCapacity := baseCapacity
	globalCapacity.TotalRuntimeSlots = 1
	globalBound, err := SelectResourceCohort(graph, ready, demands, globalCapacity, 2)
	if err != nil || len(globalBound.Tasks) != 1 || globalBound.Blocked[0].Resource != "total_runtime_slots" {
		t.Fatalf("global runtime limit was not enforced: cohort=%+v err=%v", globalBound, err)
	}
}

func TestSelectResourceCohortAllowsZeroVerificationDemandForParallelWriters(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 30),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 20),
		resourceFixtureTask("impl-c", Implementation, []string{"research"}, []string{"c.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 1, 1, 0, 1),
		resourceFixtureDemand("impl-b", 1, 1, 0, 1),
		resourceFixtureDemand("impl-c", 1, 1, 0, 1),
	}
	cohort, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(3, 3, 1, 3), 3)
	if err != nil || len(cohort.Tasks) != 3 || len(cohort.Blocked) != 0 || cohort.Estimated.EstimatedVerificationSlots != 0 {
		t.Fatalf("zero verification demand consumed a slot: cohort=%+v err=%v", cohort, err)
	}

	// A task that does request a verification slot still consumes the strict
	// positive capacity and cannot be duplicated beyond that limit.
	verifyDemands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 1, 1, 1, 1),
		resourceFixtureDemand("impl-b", 1, 1, 1, 1),
		resourceFixtureDemand("impl-c", 1, 1, 0, 1),
	}
	verifyBound, err := SelectResourceCohort(graph, ready, verifyDemands, resourceFixtureCapacity(3, 3, 1, 3), 3)
	if err != nil || len(verifyBound.Tasks) != 2 || len(verifyBound.Blocked) != 1 || verifyBound.Blocked[0].TaskID != "impl-b" || verifyBound.Blocked[0].Resource != "verification_slots" || verifyBound.Estimated.EstimatedVerificationSlots != 1 {
		t.Fatalf("positive verification demand escaped capacity: cohort=%+v err=%v", verifyBound, err)
	}
}

func TestSelectResourceCohortPreservesExplicitGeneratorOwnership(t *testing.T) {
	// A generator owns its generated output path. Its dependent consumer is
	// not ready until the generator is complete, while unrelated work can join
	// the current cohort.
	dependent := resourceFixtureGraph(
		resourceFixtureTask("impl-generator", Implementation, []string{"research"}, []string{"cmd/generate.go", "internal/generated"}, 20),
		resourceFixtureTask("impl-consumer", Implementation, []string{"impl-generator"}, []string{"internal/generated/api.go"}, 30),
		resourceFixtureTask("impl-other", Implementation, []string{"research"}, []string{"pkg/other.go"}, 10),
	)
	ready := readyImplementations(t, dependent)
	if len(ready) != 2 || ready[0].ID != "impl-generator" || ready[1].ID != "impl-other" {
		t.Fatalf("dependency did not hold generated consumer: %+v", ready)
	}
	cohort, err := SelectResourceCohort(dependent, ready, []TaskResourceDemand{
		resourceFixtureDemand("impl-generator", 1, 1, 1, 1),
		resourceFixtureDemand("impl-other", 1, 1, 1, 1),
	}, resourceFixtureCapacity(2, 2, 2, 2), 2)
	if err != nil || len(cohort.Tasks) != 2 || cohort.Tasks[0].ID != "impl-generator" || cohort.Tasks[1].ID != "impl-other" {
		t.Fatalf("generator cohort = %+v, err=%v", cohort, err)
	}

	// Even an explicit shared-prefix ownership conflict is rejected before it
	// can be considered for scheduling; the selector never guesses at missing
	// generator relationships or partial ownership.
	caseConflict := resourceFixtureGraph(
		resourceFixtureTask("impl-upper", Implementation, []string{"research"}, []string{"pkg/Foo.go"}, 20),
		resourceFixtureTask("impl-lower", Implementation, []string{"research"}, []string{"pkg/foo.go"}, 10),
	)
	caseCohort, err := SelectResourceCohort(caseConflict, readyImplementations(t, caseConflict), []TaskResourceDemand{
		resourceFixtureDemand("impl-upper", 1, 1, 1, 1),
		resourceFixtureDemand("impl-lower", 1, 1, 1, 1),
	}, resourceFixtureCapacity(2, 2, 2, 2), 2)
	if err != nil || len(caseCohort.Tasks) != 1 || caseCohort.Tasks[0].ID != "impl-upper" || len(caseCohort.Blocked) != 1 || caseCohort.Blocked[0].Reason != "dependency_or_write_conflict" || caseCohort.Blocked[0].BlockingTaskID != "impl-upper" {
		t.Fatalf("case-folded shared ownership was not blocked: cohort=%+v err=%v", caseCohort, err)
	}

	prefixConflict := resourceFixtureGraph(
		resourceFixtureTask("impl-parent", Implementation, []string{"research"}, []string{"internal/generated"}, 20),
		resourceFixtureTask("impl-child", Implementation, []string{"research"}, []string{"internal/generated/api.go"}, 10),
	)
	if err := prefixConflict.Validate(); err == nil || !strings.Contains(err.Error(), "write ownership conflict") {
		t.Fatalf("shared-prefix ownership was not rejected by graph validation: %v", err)
	}
}

func TestSelectResourceCohortIsDeterministicAndSkipsUnreadyDependencyChain(t *testing.T) {
	first := resourceFixtureTask("impl-first", Implementation, []string{"research"}, []string{"first.go"}, 25)
	second := resourceFixtureTask("impl-second", Implementation, []string{"impl-first"}, []string{"second.go"}, 50)
	third := resourceFixtureTask("impl-third", Implementation, []string{"research"}, []string{"third.go"}, 25)
	graph := resourceFixtureGraph(first, second, third)
	ready := readyImplementations(t, graph)
	if len(ready) != 2 {
		t.Fatalf("dependency chain was not held: %+v", ready)
	}
	demands := []TaskResourceDemand{resourceFixtureDemand("impl-first", 1, 1, 1, 1), resourceFixtureDemand("impl-third", 1, 1, 1, 1)}
	want, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(2, 2, 2, 2), 2)
	if err != nil {
		t.Fatal(err)
	}
	// Reversing task, ready, and demand order must not change the result.
	reversedGraph := graph
	reversedGraph.Tasks = append([]Task(nil), graph.Tasks...)
	for i, j := 0, len(reversedGraph.Tasks)-1; i < j; i, j = i+1, j-1 {
		reversedGraph.Tasks[i], reversedGraph.Tasks[j] = reversedGraph.Tasks[j], reversedGraph.Tasks[i]
	}
	for i, j := 0, len(ready)-1; i < j; i, j = i+1, j-1 {
		ready[i], ready[j] = ready[j], ready[i]
	}
	demands[0], demands[1] = demands[1], demands[0]
	got, err := SelectResourceCohort(reversedGraph, ready, demands, resourceFixtureCapacity(2, 2, 2, 2), 2)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("selection changed with input order: got=%+v want=%+v err=%v", got, want, err)
	}
}

func TestSelectResourceCohortRejectsMissingCapacityBadEstimateAndPartialReadyInput(t *testing.T) {
	graph := resourceFixtureGraph(resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 10))
	ready := readyImplementations(t, graph)
	demand := resourceFixtureDemand("impl-a", 1, 1, 1, 1)
	capacity := resourceFixtureCapacity(1, 1, 1, 1)

	missingRuntime := resourceFixtureCapacity(1, 1, 1, 1)
	missingRuntime.RuntimeSlots = []RuntimeSlotLimit{{Runtime: RuntimeResourceKey{ProfileID: "other", Provider: "openai", Model: "gpt-6-luna"}, Slots: 1}}
	if _, err := SelectResourceCohort(graph, ready, []TaskResourceDemand{demand}, missingRuntime, 1); err == nil || !strings.Contains(err.Error(), "missing exact runtime capacity") {
		t.Fatalf("missing exact runtime capacity accepted: %v", err)
	}
	if _, err := SelectResourceCohort(graph, ready, []TaskResourceDemand{demand}, capacity, 0); err == nil {
		t.Fatal("zero cohort limit accepted")
	}
	bad := demand
	bad.CPUMilli = maxResourceCPUMilli + 1
	if _, err := SelectResourceCohort(graph, ready, []TaskResourceDemand{bad}, capacity, 1); err == nil {
		t.Fatal("out-of-range resource estimate accepted")
	}
	if _, err := SelectResourceCohort(graph, nil, []TaskResourceDemand{demand}, capacity, 1); err == nil || !strings.Contains(err.Error(), "exact ready implementation set") {
		t.Fatalf("partial ready input accepted: %v", err)
	}
	capacity.MemoryMiB = 0
	if _, err := SelectResourceCohort(graph, ready, []TaskResourceDemand{demand}, capacity, 1); err == nil {
		t.Fatal("zero memory capacity treated as unlimited")
	}
	capacity = resourceFixtureCapacity(1, 1, 1, 1)
	capacity.TotalRuntimeSlots = 0
	if _, err := SelectResourceCohort(graph, ready, []TaskResourceDemand{demand}, capacity, 1); err == nil {
		t.Fatal("zero global runtime capacity treated as unlimited")
	}
	capacity = resourceFixtureCapacity(maxResourceCPUMilli+1, 1, 1, 1)
	if _, err := SelectResourceCohort(graph, ready, []TaskResourceDemand{demand}, capacity, 1); err == nil {
		t.Fatal("overflow-prone CPU capacity accepted")
	}
}
