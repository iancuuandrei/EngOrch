package engineeringplan

import (
	"encoding/json"
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

func TestSelectResourceWavesRejectsGlobalConflictUpfront(t *testing.T) {
	overlapping := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"shared/Api.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"shared/api.go"}, 10),
	)
	ready := readyImplementations(t, overlapping)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 1, 1, 1, 1),
		resourceFixtureDemand("impl-b", 1, 1, 1, 1),
	}
	if _, err := SelectResourceWaves(overlapping, ready, demands, resourceFixtureCapacity(2, 2, 2, 2), 1); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("maxTasks=1 overlapping writes slipped across waves: %v", err)
	}
	if _, err := SelectResourceWaves(overlapping, ready, demands, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("overlapping writes were deferred across waves: %v", err)
	}
	caseOverlapping := resourceFixtureGraph(
		resourceFixtureTask("impl-upper", Implementation, []string{"research"}, []string{"pkg/Foo.go"}, 20),
		resourceFixtureTask("impl-lower", Implementation, []string{"research"}, []string{"pkg/foo.go"}, 10),
	)
	caseReady := readyImplementations(t, caseOverlapping)
	caseDemands := []TaskResourceDemand{
		resourceFixtureDemand("impl-upper", 1, 1, 1, 1),
		resourceFixtureDemand("impl-lower", 1, 1, 1, 1),
	}
	if _, err := SelectResourceWaves(caseOverlapping, caseReady, caseDemands, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("casefolded prefix conflict was not rejected upfront: %v", err)
	}
	reversed := append([]Task(nil), ready...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	if _, err := SelectResourceWaves(overlapping, reversed, demands, resourceFixtureCapacity(2, 2, 2, 2), 1); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("conflict rejection changed with input order: %v", err)
	}
}

func TestSelectResourceCohortReportsConflictBeforeCohortLimit(t *testing.T) {
	overlapping := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"shared/Api.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"shared/api.go"}, 10),
	)
	ready := readyImplementations(t, overlapping)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 1, 1, 1, 1),
		resourceFixtureDemand("impl-b", 1, 1, 1, 1),
	}
	cohort, err := SelectResourceCohort(overlapping, ready, demands, resourceFixtureCapacity(2, 2, 2, 2), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(cohort.Tasks) != 1 || len(cohort.Blocked) != 1 || cohort.Blocked[0].Reason != "dependency_or_write_conflict" {
		t.Fatalf("conflict did not precede cohort_limit: %+v", cohort)
	}
}

func TestSelectResourceCohortSubsetRejectsNotReadyAndUnknown(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-blocked", Implementation, []string{"impl-a"}, []string{"b.go"}, 10),
	)
	byID := map[string]Task{}
	for _, task := range graph.Tasks {
		byID[task.ID] = task
	}
	blockedTask := byID["impl-blocked"]
	blockedDemands := []TaskResourceDemand{resourceFixtureDemand("impl-blocked", 1, 1, 1, 1)}
	if _, err := SelectResourceCohortSubset(graph, []Task{blockedTask}, blockedDemands, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "exact ready") {
		t.Fatalf("dependency-blocked candidate accepted: %v", err)
	}
	unknownGraph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
	)
	unknownTask := resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20)
	unknownTask.Attempts = []Attempt{{ID: "attempt-1", Outcome: AttemptUnknown}}
	unknownGraphWithAttempt := unknownGraph
	for i := range unknownGraphWithAttempt.Tasks {
		if unknownGraphWithAttempt.Tasks[i].ID == "impl-a" {
			unknownGraphWithAttempt.Tasks[i].Attempts = []Attempt{{ID: "attempt-1", Outcome: AttemptUnknown}}
		}
	}
	if _, err := SelectResourceCohortSubset(unknownGraphWithAttempt, []Task{unknownTask}, []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1)}, resourceFixtureCapacity(2, 2, 2, 2), 1); err == nil {
		t.Fatal("UNKNOWN candidate accepted")
	}
	readyGraph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
	)
	ready := readyImplementations(t, readyGraph)
	if _, err := SelectResourceWaves(readyGraph, ready, []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1)}, resourceFixtureCapacity(1, 1, 1, 1), 1); err != nil {
		t.Fatalf("exact ready waves rejected: %v", err)
	}
}

func TestSelectResourceCohortLexicographicBeatsGreedyPacking(t *testing.T) {
	// Adversarial declared estimates: the critical-path leader is heavy,
	// while two lighter tasks fit together but not alongside it. Greedy
	// takes the leader first and admits one task; the finite optimum
	// admits two. Estimates only; no wall-clock claim.
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 60),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
		resourceFixtureTask("impl-c", Implementation, []string{"research"}, []string{"c.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 6, 1, 1, 1),
		resourceFixtureDemand("impl-b", 5, 1, 1, 1),
		resourceFixtureDemand("impl-c", 5, 1, 1, 1),
	}
	capacity := resourceFixtureCapacity(10, 10, 3, 3)
	greedy, err := SelectResourceCohort(graph, ready, demands, capacity, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(greedy.Tasks) != 1 || greedy.Tasks[0].ID != "impl-a" {
		t.Fatalf("greedy did not take the heavy leader alone: %+v", greedy.Tasks)
	}
	optimal, err := SelectResourceCohortLexicographic(graph, ready, demands, capacity, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{optimal.Tasks[0].ID, optimal.Tasks[1].ID}; len(optimal.Tasks) != 2 || !reflect.DeepEqual(got, []string{"impl-b", "impl-c"}) {
		t.Fatalf("lexicographic optimum is not the lighter pair: %+v", optimal.Tasks)
	}
	if len(optimal.Blocked) != 1 || optimal.Blocked[0].TaskID != "impl-a" || optimal.Blocked[0].Reason != "capacity" || optimal.Blocked[0].Resource != "cpu_milli" {
		t.Fatalf("leader remainder misreported: %+v", optimal.Blocked)
	}
	if optimal.Estimated.EstimatedCPUMilli != 10 {
		t.Fatalf("optimal estimate is not the declared pair sum: %+v", optimal.Estimated)
	}
}

func TestSelectResourceCohortLexicographicIsPermutationDeterministic(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 60),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
		resourceFixtureTask("impl-c", Implementation, []string{"research"}, []string{"c.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 6, 1, 1, 1),
		resourceFixtureDemand("impl-b", 5, 1, 1, 1),
		resourceFixtureDemand("impl-c", 5, 1, 1, 1),
	}
	capacity := resourceFixtureCapacity(10, 10, 3, 3)
	want, err := SelectResourceCohortLexicographic(graph, ready, demands, capacity, 3)
	if err != nil {
		t.Fatal(err)
	}
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
	got, err := SelectResourceCohortLexicographic(reversedGraph, reversedReady, reversedDemands, capacity, 3)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("selection changed with input order: got=%+v want=%+v err=%v", got, want, err)
	}
	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	gotBytes, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(wantBytes) != string(gotBytes) {
		t.Fatalf("serialized selection differs with input order:\n%s\n%s", wantBytes, gotBytes)
	}
	var roundTrip ResourceCohort
	if err := json.Unmarshal(wantBytes, &roundTrip); err != nil || !reflect.DeepEqual(roundTrip, want) {
		t.Fatalf("serialized selection did not round-trip: %v", err)
	}
}

func TestSelectResourceCohortLexicographicEnforcesAllCeilings(t *testing.T) {
	fixture := func() (Graph, []Task) {
		graph := resourceFixtureGraph(
			resourceFixtureTask("impl-alpha", Implementation, []string{"research"}, []string{"alpha/api.go"}, 60),
			resourceFixtureTask("impl-beta", Implementation, []string{"research"}, []string{"beta/api.go"}, 50),
			resourceFixtureTask("impl-gamma", Implementation, []string{"research"}, []string{"gamma/api.go"}, 40),
		)
		ready := []Task{}
		for _, task := range graph.Tasks {
			if task.Kind == Implementation {
				ready = append(ready, task)
			}
		}
		return graph, ready
	}
	cases := []struct {
		name     string
		demands  []TaskResourceDemand
		capacity ResourceCapacity
		resource string
	}{
		{"cpu", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 2, 1, 1, 1), resourceFixtureDemand("impl-beta", 2, 1, 1, 1), resourceFixtureDemand("impl-gamma", 2, 1, 1, 1)}, resourceFixtureCapacity(3, 10, 3, 3), "cpu_milli"},
		{"memory", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 1, 2, 1, 1), resourceFixtureDemand("impl-beta", 1, 2, 1, 1), resourceFixtureDemand("impl-gamma", 1, 2, 1, 1)}, resourceFixtureCapacity(10, 3, 3, 3), "memory_mib"},
		{"verification", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 1, 1, 1, 1), resourceFixtureDemand("impl-beta", 1, 1, 1, 1), resourceFixtureDemand("impl-gamma", 1, 1, 1, 1)}, resourceFixtureCapacity(10, 10, 1, 3), "verification_slots"},
		{"total_runtime", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 1, 1, 1, 1), resourceFixtureDemand("impl-beta", 1, 1, 1, 1), resourceFixtureDemand("impl-gamma", 1, 1, 1, 1)}, func() ResourceCapacity {
			capacity := resourceFixtureCapacity(10, 10, 3, 3)
			capacity.TotalRuntimeSlots = 1
			return capacity
		}(), "total_runtime_slots"},
		{"provider", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 1, 1, 1, 1), resourceFixtureDemand("impl-beta", 1, 1, 1, 1), resourceFixtureDemand("impl-gamma", 1, 1, 1, 1)}, func() ResourceCapacity {
			capacity := resourceFixtureCapacity(10, 10, 3, 3)
			capacity.ProviderSlots = []ProviderSlotLimit{{Provider: fixtureRuntime.Provider, Slots: 1}}
			return capacity
		}(), "provider_slots"},
		{"model", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 1, 1, 1, 1), resourceFixtureDemand("impl-beta", 1, 1, 1, 1), resourceFixtureDemand("impl-gamma", 1, 1, 1, 1)}, func() ResourceCapacity {
			capacity := resourceFixtureCapacity(10, 10, 3, 3)
			capacity.ModelSlots = []ModelSlotLimit{{Model: providerModel(fixtureRuntime), Slots: 1}}
			return capacity
		}(), "model_slots"},
		{"runtime_route", []TaskResourceDemand{resourceFixtureDemand("impl-alpha", 1, 1, 1, 1), resourceFixtureDemand("impl-beta", 1, 1, 1, 1), resourceFixtureDemand("impl-gamma", 1, 1, 1, 1)}, resourceFixtureCapacity(10, 10, 3, 1), "runtime_slots"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			graph, ready := fixture()
			cohort, err := SelectResourceCohortLexicographic(graph, ready, tc.demands, tc.capacity, 3)
			if err != nil {
				t.Fatal(err)
			}
			if len(cohort.Tasks) != 1 || cohort.Tasks[0].ID != "impl-alpha" {
				t.Fatalf("ceiling %s did not bind to the critical-path singleton: %+v", tc.name, cohort.Tasks)
			}
			if len(cohort.Blocked) != 2 || cohort.Blocked[0].Resource != tc.resource || cohort.Blocked[1].Resource != tc.resource {
				t.Fatalf("ceiling %s misreported: %+v", tc.name, cohort.Blocked)
			}
			if tc.resource == "runtime_slots" || tc.resource == "provider_slots" || tc.resource == "model_slots" {
				if cohort.Blocked[0].Runtime == nil {
					t.Fatalf("ceiling %s omitted the route binding: %+v", tc.name, cohort.Blocked[0])
				}
			}
		})
	}
	t.Run("all_fit", func(t *testing.T) {
		graph, ready := fixture()
		demands := []TaskResourceDemand{
			resourceFixtureDemand("impl-alpha", 1, 1, 1, 1),
			resourceFixtureDemand("impl-beta", 1, 1, 1, 1),
			resourceFixtureDemand("impl-gamma", 1, 1, 1, 1),
		}
		cohort, err := SelectResourceCohortLexicographic(graph, ready, demands, resourceFixtureCapacity(10, 10, 3, 3), 3)
		if err != nil {
			t.Fatal(err)
		}
		if len(cohort.Tasks) != 3 || len(cohort.Blocked) != 0 {
			t.Fatalf("fitting cohort was not fully admitted: %+v", cohort)
		}
	})
}

func TestSelectResourceCohortLexicographicRejectsOverlapDependencyUnknownAndBounds(t *testing.T) {
	overlapping := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"shared/Api.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"shared/api.go"}, 10),
	)
	ready := readyImplementations(t, overlapping)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 1, 1, 1, 1),
		resourceFixtureDemand("impl-b", 1, 1, 1, 1),
	}
	cohort, err := SelectResourceCohortLexicographic(overlapping, ready, demands, resourceFixtureCapacity(2, 2, 2, 2), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(cohort.Tasks) != 1 || cohort.Tasks[0].ID != "impl-a" || len(cohort.Blocked) != 1 || cohort.Blocked[0].Reason != "dependency_or_write_conflict" || cohort.Blocked[0].BlockingTaskID != "impl-a" {
		t.Fatalf("casefolded overlap was not the blocking reason: %+v", cohort)
	}

	chained := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-blocked", Implementation, []string{"impl-a"}, []string{"b.go"}, 10),
	)
	complete := readyImplementations(t, chained)
	byID := map[string]Task{}
	for _, task := range chained.Tasks {
		byID[task.ID] = task
	}
	withBlocked := append(append([]Task(nil), complete...), byID["impl-blocked"])
	withDemands := []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1), resourceFixtureDemand("impl-blocked", 1, 1, 1, 1)}
	if _, err := SelectResourceCohortLexicographic(chained, withBlocked, withDemands, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "exact ready implementation set") {
		t.Fatalf("dependency-blocked input accepted: %v", err)
	}

	unknownGraph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
	)
	for i := range unknownGraph.Tasks {
		if unknownGraph.Tasks[i].ID == "impl-b" {
			unknownGraph.Tasks[i].Attempts = []Attempt{{ID: "attempt-1", Outcome: AttemptUnknown}}
		}
	}
	unknownByID := map[string]Task{}
	for _, task := range unknownGraph.Tasks {
		unknownByID[task.ID] = task
	}
	unknownReady := []Task{unknownByID["impl-a"], unknownByID["impl-b"]}
	unknownDemands := []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1), resourceFixtureDemand("impl-b", 1, 1, 1, 1)}
	if _, err := SelectResourceCohortLexicographic(unknownGraph, unknownReady, unknownDemands, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "exact ready implementation set") {
		t.Fatalf("UNKNOWN attempt input accepted: %v", err)
	}
	exactReady := []Task{unknownByID["impl-a"]}
	exact, err := SelectResourceCohortLexicographic(unknownGraph, exactReady, []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1)}, resourceFixtureCapacity(2, 2, 2, 2), 1)
	if err != nil || len(exact.Tasks) != 1 || exact.Tasks[0].ID != "impl-a" {
		t.Fatalf("resolved subset around UNKNOWN was not selected: %+v %v", exact, err)
	}

	graph := resourceFixtureGraph(resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 10))
	single := readyImplementations(t, graph)
	singleDemand := []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1)}
	singleCapacity := resourceFixtureCapacity(1, 1, 1, 1)
	if _, err := SelectResourceCohortLexicographic(graph, single, singleDemand, singleCapacity, 0); err == nil || !strings.Contains(err.Error(), "resource cohort bound must be 1..8") {
		t.Fatalf("zero cohort bound accepted: %v", err)
	}
	if _, err := SelectResourceCohortLexicographic(graph, single, singleDemand, singleCapacity, 9); err == nil || !strings.Contains(err.Error(), "resource cohort bound must be 1..8") {
		t.Fatalf("oversized cohort bound accepted: %v", err)
	}
	missingRuntime := resourceFixtureCapacity(1, 1, 1, 1)
	missingRuntime.RuntimeSlots = []RuntimeSlotLimit{{Runtime: RuntimeResourceKey{ProfileID: "other", Provider: "openai", Model: "gpt-6-luna"}, Slots: 1}}
	if _, err := SelectResourceCohortLexicographic(graph, single, singleDemand, missingRuntime, 1); err == nil || !strings.Contains(err.Error(), "missing exact runtime capacity") {
		t.Fatalf("missing exact runtime capacity accepted: %v", err)
	}
	missingProvider := resourceFixtureCapacity(1, 1, 1, 1)
	missingProvider.ProviderSlots = []ProviderSlotLimit{{Provider: "other", Slots: 1}}
	if _, err := SelectResourceCohortLexicographic(graph, single, singleDemand, missingProvider, 1); err == nil || !strings.Contains(err.Error(), "missing provider capacity") {
		t.Fatalf("missing provider capacity accepted: %v", err)
	}
	zeroMemory := resourceFixtureCapacity(1, 1, 1, 1)
	zeroMemory.MemoryMiB = 0
	if _, err := SelectResourceCohortLexicographic(graph, single, singleDemand, zeroMemory, 1); err == nil {
		t.Fatal("zero memory capacity treated as unlimited")
	}
	manyTasks := []Task{{ID: "research", Kind: Research, Title: "Shared research", ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "fact", Description: "shared"}}, EstimatedSeconds: 5, Completed: true}}
	manyDemands := []TaskResourceDemand{}
	for i := 0; i < 9; i++ {
		id := strings.Repeat("a", i+1)
		manyTasks = append(manyTasks, resourceFixtureTask(id, Implementation, []string{"research"}, []string{id + ".go"}, 10))
		manyDemands = append(manyDemands, resourceFixtureDemand(id, 1, 1, 1, 1))
	}
	manyTasks = append(manyTasks,
		resourceFixtureTask("verify", Verification, []string{"a", "aa", "aaa", "aaaa", "aaaaa", "aaaaaa", "aaaaaaa", "aaaaaaaa", "aaaaaaaaa"}, nil, 10),
		resourceFixtureTask("review", Review, []string{"verify"}, nil, 5),
	)
	many := Graph{Version: Version, Mode: ModeGraph, Summary: "nine ready implementations", Tasks: manyTasks}
	manyReady := []Task{}
	for _, task := range many.Tasks {
		if task.Kind == Implementation {
			manyReady = append(manyReady, task)
		}
	}
	if len(manyReady) != 9 {
		t.Fatalf("nine-ready fixture did not stay ready: %d", len(manyReady))
	}
	if _, err := SelectResourceCohortLexicographic(many, manyReady, manyDemands, resourceFixtureCapacity(64, 64, 64, 64), 8); err == nil || !strings.Contains(err.Error(), "one to eight ready implementations") {
		t.Fatalf("nine-ready finite bound bypassed: %v", err)
	}
}

func TestSumResourceDemandsMatchesCohortEstimated(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 2, 3, 1, 1),
		resourceFixtureDemand("impl-b", 1, 1, 0, 1),
	}
	cohort, err := SelectResourceCohort(graph, ready, demands, resourceFixtureCapacity(5, 5, 2, 2), 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := SumResourceDemands(demands); !reflect.DeepEqual(got, cohort.Estimated) {
		t.Fatalf("shared totals differ: sum=%+v cohort=%+v", got, cohort.Estimated)
	}
}

func TestSelectResourceWavesLexicographicReordersAdversarialPacking(t *testing.T) {
	// Same adversarial declared estimates as the single-cohort test: the
	// critical-path leader is heavy while two lighter tasks fit together
	// but not alongside it. Greedy waves open with the leader alone; the
	// lexicographic optimum opens with the lighter pair. Estimates only;
	// no wall-clock claim.
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 60),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
		resourceFixtureTask("impl-c", Implementation, []string{"research"}, []string{"c.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-a", 6, 1, 1, 1),
		resourceFixtureDemand("impl-b", 5, 1, 1, 1),
		resourceFixtureDemand("impl-c", 5, 1, 1, 1),
	}
	capacity := resourceFixtureCapacity(10, 10, 3, 3)
	greedy, err := SelectResourceWaves(graph, ready, demands, capacity, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(greedy) != 2 || len(greedy[0].Tasks) != 1 || greedy[0].Tasks[0].ID != "impl-a" {
		t.Fatalf("greedy waves did not open with the heavy leader: %+v", greedy)
	}
	optimal, err := SelectResourceWavesLexicographic(graph, ready, demands, capacity, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(optimal) != 2 || len(optimal[0].Tasks) != 2 || optimal[0].Tasks[0].ID != "impl-b" || optimal[0].Tasks[1].ID != "impl-c" {
		t.Fatalf("lexicographic waves did not open with the lighter pair: %+v", optimal)
	}
	if len(optimal[1].Tasks) != 1 || optimal[1].Tasks[0].ID != "impl-a" {
		t.Fatalf("lexicographic remainder is not the leader alone: %+v", optimal[1])
	}
	if len(optimal[1].Blocked) != 0 {
		t.Fatalf("final wave retains blocked tasks: %+v", optimal[1].Blocked)
	}
	// Coverage is identical; only the partition order differs.
	greedyIDs, optimalIDs := map[string]bool{}, map[string]bool{}
	for _, wave := range greedy {
		for _, task := range wave.Tasks {
			greedyIDs[task.ID] = true
		}
	}
	for _, wave := range optimal {
		for _, task := range wave.Tasks {
			optimalIDs[task.ID] = true
		}
	}
	if !reflect.DeepEqual(greedyIDs, optimalIDs) {
		t.Fatalf("wave coverage differs: greedy=%v lexicographic=%v", greedyIDs, optimalIDs)
	}
}

func TestSelectResourceCohortSubsetLexicographicUsesSubsetContract(t *testing.T) {
	graph := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 60),
		resourceFixtureTask("impl-b", Implementation, []string{"research"}, []string{"b.go"}, 10),
		resourceFixtureTask("impl-c", Implementation, []string{"research"}, []string{"c.go"}, 10),
	)
	ready := readyImplementations(t, graph)
	byID := map[string]Task{}
	for _, task := range ready {
		byID[task.ID] = task
	}
	// An explicit remainder subset (after admitting impl-a elsewhere) still
	// selects the exact optimum over that subset.
	candidates := []Task{byID["impl-b"], byID["impl-c"]}
	demands := []TaskResourceDemand{
		resourceFixtureDemand("impl-b", 5, 1, 1, 1),
		resourceFixtureDemand("impl-c", 5, 1, 1, 1),
	}
	cohort, err := SelectResourceCohortSubsetLexicographic(graph, candidates, demands, resourceFixtureCapacity(10, 10, 3, 3), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(cohort.Tasks) != 2 || len(cohort.Blocked) != 0 {
		t.Fatalf("subset optimum did not admit the fitting pair: %+v", cohort)
	}
	// A dependency-blocked candidate is rejected, never silently dropped.
	chained := resourceFixtureGraph(
		resourceFixtureTask("impl-a", Implementation, []string{"research"}, []string{"a.go"}, 20),
		resourceFixtureTask("impl-blocked", Implementation, []string{"impl-a"}, []string{"b.go"}, 10),
	)
	complete := readyImplementations(t, chained)
	chainedByID := map[string]Task{}
	for _, task := range chained.Tasks {
		chainedByID[task.ID] = task
	}
	withBlocked := append(append([]Task(nil), complete...), chainedByID["impl-blocked"])
	withDemands := []TaskResourceDemand{resourceFixtureDemand("impl-a", 1, 1, 1, 1), resourceFixtureDemand("impl-blocked", 1, 1, 1, 1)}
	if _, err := SelectResourceCohortSubsetLexicographic(chained, withBlocked, withDemands, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "not an exact ready implementation") {
		t.Fatalf("dependency-blocked subset input accepted: %v", err)
	}
	if _, err := SelectResourceCohortSubsetLexicographic(graph, nil, nil, resourceFixtureCapacity(2, 2, 2, 2), 2); err == nil || !strings.Contains(err.Error(), "subset is empty") {
		t.Fatalf("empty subset accepted: %v", err)
	}
}
