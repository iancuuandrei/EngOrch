package engineeringplan

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// RuntimeResourceKey identifies one exact configured runtime route. Profile,
// provider, and model remain separate so model aliases cannot collide.
type RuntimeResourceKey struct {
	ProfileID string `json:"profile_id"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
}

// ProviderModelKey represents capacity shared across every configured route
// that dispatches the same provider/model pair, regardless of access profile
// or effort variant.
type ProviderModelKey struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// TaskResourceDemand contains declared scheduling estimates. These values are
// not runtime measurements or execution authority. RuntimeSlots are charged
// against exact route, shared provider/model, provider-wide, and global runtime
// ceilings.
type TaskResourceDemand struct {
	TaskID            string             `json:"task_id"`
	CPUMilli          int64              `json:"estimated_cpu_milli"`
	MemoryMiB         int64              `json:"estimated_memory_mib"`
	VerificationSlots int                `json:"estimated_verification_slots"`
	Runtime           RuntimeResourceKey `json:"runtime"`
	RuntimeSlots      int                `json:"runtime_slots"`
}

// RuntimeSlotLimit gives one exact profile/provider/model route a positive
// capacity. A slice keeps the serialized representation deterministic.
type RuntimeSlotLimit struct {
	Runtime RuntimeResourceKey `json:"runtime"`
	Slots   int                `json:"slots"`
}

// ProviderSlotLimit bounds all selected tasks routed through one provider.
type ProviderSlotLimit struct {
	Provider string `json:"provider"`
	Slots    int    `json:"slots"`
}

// ModelSlotLimit bounds all selected tasks for one provider/model pair,
// including different configured profiles or effort variants.
type ModelSlotLimit struct {
	Model ProviderModelKey `json:"model"`
	Slots int              `json:"slots"`
}

// RuntimeSlotTotal reports the selected cohort's estimated slots for one exact
// runtime route.
type RuntimeSlotTotal struct {
	Runtime RuntimeResourceKey `json:"runtime"`
	Slots   int                `json:"slots"`
}

// ProviderSlotTotal reports estimated selected runtime slots for one provider.
type ProviderSlotTotal struct {
	Provider string `json:"provider"`
	Slots    int    `json:"slots"`
}

// ModelSlotTotal reports estimated slots shared by one provider/model pair.
type ModelSlotTotal struct {
	Model ProviderModelKey `json:"model"`
	Slots int              `json:"slots"`
}

// ResourceCapacity fixes explicit positive ceilings at every runtime layer.
// Missing keys are errors; zero never means unlimited.
type ResourceCapacity struct {
	CPUMilli          int64               `json:"cpu_milli"`
	MemoryMiB         int64               `json:"memory_mib"`
	VerificationSlots int                 `json:"verification_slots"`
	TotalRuntimeSlots int                 `json:"total_runtime_slots"`
	ProviderSlots     []ProviderSlotLimit `json:"provider_slots"`
	ModelSlots        []ModelSlotLimit    `json:"model_slots"`
	RuntimeSlots      []RuntimeSlotLimit  `json:"runtime_slots"`
}

// ResourceTotals sums estimates for a selected cohort. It must not be reported
// as measured CPU, memory, or runtime usage.
type ResourceTotals struct {
	EstimatedCPUMilli          int64               `json:"estimated_cpu_milli"`
	EstimatedMemoryMiB         int64               `json:"estimated_memory_mib"`
	EstimatedVerificationSlots int                 `json:"estimated_verification_slots"`
	TotalRuntimeSlots          int                 `json:"total_runtime_slots"`
	ProviderSlots              []ProviderSlotTotal `json:"provider_slots"`
	ModelSlots                 []ModelSlotTotal    `json:"model_slots"`
	RuntimeSlots               []RuntimeSlotTotal  `json:"runtime_slots"`
}

// ResourceBlockReason explains why a ready implementation was left for a
// later cohort.
type ResourceBlockReason struct {
	TaskID         string              `json:"task_id"`
	Reason         string              `json:"reason"`
	BlockingTaskID string              `json:"blocking_task_id,omitempty"`
	Resource       string              `json:"resource,omitempty"`
	Runtime        *RuntimeResourceKey `json:"runtime,omitempty"`
}

// ResourceCohort is a deterministic, bounded subset of the exact ready
// implementation set. Selection is a greedy critical-path heuristic, not an
// optimal packing claim.
type ResourceCohort struct {
	Tasks     []Task                `json:"tasks"`
	Blocked   []ResourceBlockReason `json:"blocked"`
	Estimated ResourceTotals        `json:"estimated_totals"`
}

const (
	maxResourceCPUMilli  int64 = 1 << 20 // millicores
	maxResourceMemoryMiB int64 = 1 << 30
	maxResourceSlots           = 64
)

// SelectResourceCohort selects ready implementation tasks by critical-path
// priority, exact declared runtime slots, CPU/memory estimates, and verification
// estimates. The supplied tasks and demands must exactly match the graph's
// ready implementation set so callers cannot hide dependencies or shared
// ownership by passing a partial graph.
func SelectResourceCohort(g Graph, ready []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int) (ResourceCohort, error) {
	if maxTasks < 1 || maxTasks > 8 {
		return ResourceCohort{}, errors.New("resource cohort bound must be 1..8")
	}
	if err := g.Validate(); err != nil {
		return ResourceCohort{}, err
	}
	if err := validateResourceCapacity(capacity); err != nil {
		return ResourceCohort{}, err
	}
	graphReady, err := Ready(g)
	if err != nil {
		return ResourceCohort{}, err
	}
	wantReady := map[string]Task{}
	for _, task := range graphReady {
		if task.Kind == Implementation {
			wantReady[task.ID] = task
		}
	}
	provided := map[string]Task{}
	for _, task := range ready {
		if task.Kind != Implementation {
			return ResourceCohort{}, fmt.Errorf("resource cohort task %q is not an implementation", task.ID)
		}
		if _, duplicate := provided[task.ID]; duplicate {
			return ResourceCohort{}, fmt.Errorf("duplicate resource cohort task %q", task.ID)
		}
		provided[task.ID] = task
	}
	if len(provided) != len(wantReady) {
		return ResourceCohort{}, errors.New("resource cohort input differs from exact ready implementation set")
	}
	for _, task := range graphReady {
		if task.Kind != Implementation {
			continue
		}
		if got, ok := provided[task.ID]; !ok || !reflect.DeepEqual(got, task) {
			return ResourceCohort{}, fmt.Errorf("resource cohort task %q is stale or substituted", task.ID)
		}
	}
	if len(demands) != len(provided) {
		return ResourceCohort{}, errors.New("resource demand set differs from ready implementations")
	}
	demandByID := make(map[string]TaskResourceDemand, len(demands))
	for _, demand := range demands {
		if _, ok := provided[demand.TaskID]; !ok {
			return ResourceCohort{}, fmt.Errorf("resource demand for non-ready task %q", demand.TaskID)
		}
		if _, duplicate := demandByID[demand.TaskID]; duplicate {
			return ResourceCohort{}, fmt.Errorf("duplicate resource demand for task %q", demand.TaskID)
		}
		if err := validateResourceDemand(demand); err != nil {
			return ResourceCohort{}, fmt.Errorf("task %q: %w", demand.TaskID, err)
		}
		if _, ok := runtimeCapacity(capacity.RuntimeSlots, demand.Runtime); !ok {
			return ResourceCohort{}, fmt.Errorf("missing exact runtime capacity for task %q", demand.TaskID)
		}
		if _, ok := providerCapacity(capacity.ProviderSlots, demand.Runtime.Provider); !ok {
			return ResourceCohort{}, fmt.Errorf("missing provider capacity for task %q", demand.TaskID)
		}
		if _, ok := modelCapacity(capacity.ModelSlots, providerModel(demand.Runtime)); !ok {
			return ResourceCohort{}, fmt.Errorf("missing provider/model capacity for task %q", demand.TaskID)
		}
		demandByID[demand.TaskID] = demand
	}

	ordered := orderResourceCandidates(provided, downstreamLengths(g))
	byID := make(map[string]Task, len(g.Tasks))
	for _, task := range g.Tasks {
		byID[task.ID] = task
	}
	return selectResourceCohortGreedy(ordered, demandByID, capacity, maxTasks, byID), nil
}

// SelectResourceCohortSubset selects from an explicit candidate subset of the
// graph's exact ready implementation set. Every candidate must be ready in the
// graph (dependencies completed, no terminal UNKNOWN attempt) with identical
// bytes; dependency-blocked or UNKNOWN candidates are rejected. It never
// fabricates Completed mutations to trick Ready. Callers deriving
// deterministic waves use this for each remainder after the first exact-ready
// SelectResourceCohort.
func SelectResourceCohortSubset(g Graph, candidates []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int) (ResourceCohort, error) {
	if maxTasks < 1 || maxTasks > 8 {
		return ResourceCohort{}, errors.New("resource cohort bound must be 1..8")
	}
	if err := g.Validate(); err != nil {
		return ResourceCohort{}, err
	}
	if err := validateResourceCapacity(capacity); err != nil {
		return ResourceCohort{}, err
	}
	if len(candidates) == 0 {
		return ResourceCohort{}, errors.New("resource cohort subset is empty")
	}
	graphReady, err := Ready(g)
	if err != nil {
		return ResourceCohort{}, err
	}
	readyByID := make(map[string]Task, len(graphReady))
	for _, task := range graphReady {
		if task.Kind == Implementation {
			readyByID[task.ID] = task
		}
	}
	byID := make(map[string]Task, len(g.Tasks))
	for _, task := range g.Tasks {
		byID[task.ID] = task
	}
	provided := make(map[string]Task, len(candidates))
	for _, task := range candidates {
		if task.Kind != Implementation {
			return ResourceCohort{}, fmt.Errorf("resource cohort task %q is not an implementation", task.ID)
		}
		if _, duplicate := provided[task.ID]; duplicate {
			return ResourceCohort{}, fmt.Errorf("duplicate resource cohort task %q", task.ID)
		}
		readyTask, ok := readyByID[task.ID]
		if !ok || !reflect.DeepEqual(readyTask, task) {
			return ResourceCohort{}, fmt.Errorf("resource cohort task %q is not an exact ready implementation", task.ID)
		}
		graphTask, ok := byID[task.ID]
		if !ok || !reflect.DeepEqual(graphTask, task) {
			return ResourceCohort{}, fmt.Errorf("resource cohort task %q is stale or substituted", task.ID)
		}
		if graphTask.Completed {
			return ResourceCohort{}, fmt.Errorf("resource cohort task %q is already completed", task.ID)
		}
		provided[task.ID] = task
	}
	if len(demands) != len(provided) {
		return ResourceCohort{}, errors.New("resource demand set differs from candidate implementations")
	}
	demandByID := make(map[string]TaskResourceDemand, len(demands))
	for _, demand := range demands {
		if _, ok := provided[demand.TaskID]; !ok {
			return ResourceCohort{}, fmt.Errorf("resource demand for non-candidate task %q", demand.TaskID)
		}
		if _, duplicate := demandByID[demand.TaskID]; duplicate {
			return ResourceCohort{}, fmt.Errorf("duplicate resource demand for task %q", demand.TaskID)
		}
		if err := validateResourceDemand(demand); err != nil {
			return ResourceCohort{}, fmt.Errorf("task %q: %w", demand.TaskID, err)
		}
		if _, ok := runtimeCapacity(capacity.RuntimeSlots, demand.Runtime); !ok {
			return ResourceCohort{}, fmt.Errorf("missing exact runtime capacity for task %q", demand.TaskID)
		}
		if _, ok := providerCapacity(capacity.ProviderSlots, demand.Runtime.Provider); !ok {
			return ResourceCohort{}, fmt.Errorf("missing provider capacity for task %q", demand.TaskID)
		}
		if _, ok := modelCapacity(capacity.ModelSlots, providerModel(demand.Runtime)); !ok {
			return ResourceCohort{}, fmt.Errorf("missing provider/model capacity for task %q", demand.TaskID)
		}
		demandByID[demand.TaskID] = demand
	}
	ordered := orderResourceCandidates(provided, downstreamLengths(g))
	return selectResourceCohortGreedy(ordered, demandByID, capacity, maxTasks, byID), nil
}

// SelectResourceWaves deterministically partitions the exact ready
// implementation set into nonempty resource-bounded waves. The ready set must
// already be globally independent: every pair is dependency-free with disjoint
// casefolded write prefixes, rejected upfront under the existing independent
// cohort contract rather than deferred across waves. The first wave uses the
// exact-ready SelectResourceCohort contract; each remainder uses the subset
// contract without fabricating Completed mutations. Every selected task
// identity is exact against the graph. A task that fits no empty wave yields
// a no-fit capacity diagnostic. Blocked reasons and per-wave estimates are
// retained in each returned in-memory cohort (not durable preparation state);
// the final cohort's Blocked must be empty when all tasks are covered.
func SelectResourceWaves(g Graph, ready []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int) ([]ResourceCohort, error) {
	if maxTasks < 1 || maxTasks > 8 {
		return nil, errors.New("resource cohort bound must be 1..8")
	}
	if len(ready) == 0 || len(ready) > 8 {
		return nil, errors.New("resource waves require one to eight ready implementations")
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	byID := make(map[string]Task, len(g.Tasks))
	for _, task := range g.Tasks {
		byID[task.ID] = task
	}
	sorted := append([]Task(nil), ready...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if dependsOn(sorted[i].ID, sorted[j].ID, byID) || dependsOn(sorted[j].ID, sorted[i].ID, byID) {
				return nil, fmt.Errorf("resource waves reject conflicting task %q blocked by %q", sorted[j].ID, sorted[i].ID)
			}
			if pathsOverlapFoldAny(sorted[i].WritePaths, sorted[j].WritePaths) {
				return nil, fmt.Errorf("resource waves reject conflicting task %q blocked by %q", sorted[j].ID, sorted[i].ID)
			}
		}
	}
	first, err := SelectResourceCohort(g, ready, demands, capacity, maxTasks)
	if err != nil {
		return nil, err
	}
	for _, blocked := range first.Blocked {
		if blocked.Reason == "dependency_or_write_conflict" {
			return nil, fmt.Errorf("resource waves reject conflicting task %q blocked by %q", blocked.TaskID, blocked.BlockingTaskID)
		}
	}
	if len(first.Tasks) == 0 {
		if len(first.Blocked) == 0 {
			return nil, errors.New("resource waves selected no implementation tasks")
		}
		leading := first.Blocked[0]
		if leading.Reason == "capacity" {
			resource := leading.Resource
			if resource == "" {
				resource = "capacity"
			}
			return nil, fmt.Errorf("resource waves admit no implementation task: task %q exceeds %s capacity", leading.TaskID, resource)
		}
		return nil, fmt.Errorf("resource waves admit no implementation task: task %q blocked (%s)", leading.TaskID, leading.Reason)
	}
	waves := []ResourceCohort{first}
	if len(first.Blocked) == 0 && len(first.Tasks) == len(ready) {
		return waves, nil
	}
	selected := map[string]bool{}
	for _, task := range first.Tasks {
		selected[task.ID] = true
	}
	demandByID := make(map[string]TaskResourceDemand, len(demands))
	for _, demand := range demands {
		demandByID[demand.TaskID] = demand
	}
	for len(selected) < len(ready) {
		var remainder []Task
		var remainderDemands []TaskResourceDemand
		for _, task := range ready {
			if !selected[task.ID] {
				remainder = append(remainder, task)
				remainderDemands = append(remainderDemands, demandByID[task.ID])
			}
		}
		next, err := SelectResourceCohortSubset(g, remainder, remainderDemands, capacity, maxTasks)
		if err != nil {
			return nil, err
		}
		for _, blocked := range next.Blocked {
			if blocked.Reason == "dependency_or_write_conflict" {
				return nil, fmt.Errorf("resource waves reject conflicting task %q blocked by %q", blocked.TaskID, blocked.BlockingTaskID)
			}
		}
		if len(next.Tasks) == 0 {
			if len(next.Blocked) == 0 {
				return nil, errors.New("resource waves selected no implementation tasks")
			}
			leading := next.Blocked[0]
			if leading.Reason == "capacity" {
				resource := leading.Resource
				if resource == "" {
					resource = "capacity"
				}
				return nil, fmt.Errorf("resource waves admit no implementation task: task %q exceeds %s capacity", leading.TaskID, resource)
			}
			return nil, fmt.Errorf("resource waves admit no implementation task: task %q blocked (%s)", leading.TaskID, leading.Reason)
		}
		waves = append(waves, next)
		for _, task := range next.Tasks {
			if selected[task.ID] {
				return nil, fmt.Errorf("resource waves duplicated task %q", task.ID)
			}
			selected[task.ID] = true
		}
		if len(waves) > 8 {
			return nil, errors.New("resource waves exceed eight-task bound")
		}
	}
	last := waves[len(waves)-1]
	if len(last.Blocked) != 0 {
		return nil, errors.New("resource waves retain unexpected blocked tasks after covering all implementations")
	}
	return waves, nil
}

func orderResourceCandidates(provided map[string]Task, criticalPath map[string]int) []Task {
	ordered := make([]Task, 0, len(provided))
	for _, task := range provided {
		ordered = append(ordered, task)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if criticalPath[ordered[i].ID] != criticalPath[ordered[j].ID] {
			return criticalPath[ordered[i].ID] > criticalPath[ordered[j].ID]
		}
		return ordered[i].ID < ordered[j].ID
	})
	return ordered
}

func selectResourceCohortGreedy(ordered []Task, demandByID map[string]TaskResourceDemand, capacity ResourceCapacity, maxTasks int, byID map[string]Task) ResourceCohort {
	out := ResourceCohort{Tasks: make([]Task, 0, maxTasks), Blocked: []ResourceBlockReason{}, Estimated: ResourceTotals{ProviderSlots: []ProviderSlotTotal{}, ModelSlots: []ModelSlotTotal{}, RuntimeSlots: []RuntimeSlotTotal{}}}
	for _, task := range ordered {
		conflict := ""
		for _, chosen := range out.Tasks {
			if dependsOn(task.ID, chosen.ID, byID) || dependsOn(chosen.ID, task.ID, byID) {
				conflict = chosen.ID
				break
			}
			if pathsOverlapFoldAny(task.WritePaths, chosen.WritePaths) {
				conflict = chosen.ID
				break
			}
		}
		if conflict != "" {
			out.Blocked = append(out.Blocked, ResourceBlockReason{TaskID: task.ID, Reason: "dependency_or_write_conflict", BlockingTaskID: conflict})
			continue
		}
		if len(out.Tasks) >= maxTasks {
			out.Blocked = append(out.Blocked, ResourceBlockReason{TaskID: task.ID, Reason: "cohort_limit"})
			continue
		}
		demand := demandByID[task.ID]
		if resource := resourceOverCapacity(out.Estimated, demand, capacity); resource != "" {
			reason := ResourceBlockReason{TaskID: task.ID, Reason: "capacity", Resource: resource}
			if resource == "runtime_slots" || resource == "provider_slots" || resource == "model_slots" {
				runtime := demand.Runtime
				reason.Runtime = &runtime
			}
			out.Blocked = append(out.Blocked, reason)
			continue
		}
		out.Tasks = append(out.Tasks, task)
		addResourceTotals(&out.Estimated, demand)
	}
	return out
}

func validateResourceCapacity(capacity ResourceCapacity) error {
	if capacity.CPUMilli < 1 || capacity.CPUMilli > maxResourceCPUMilli || capacity.MemoryMiB < 1 || capacity.MemoryMiB > maxResourceMemoryMiB || capacity.VerificationSlots < 1 || capacity.VerificationSlots > maxResourceSlots || capacity.TotalRuntimeSlots < 1 || capacity.TotalRuntimeSlots > maxResourceSlots || len(capacity.ProviderSlots) == 0 || len(capacity.ProviderSlots) > maxResourceSlots || len(capacity.ModelSlots) == 0 || len(capacity.ModelSlots) > maxResourceSlots || len(capacity.RuntimeSlots) == 0 || len(capacity.RuntimeSlots) > maxResourceSlots {
		return errors.New("resource capacity requires positive bounded CPU, memory, verification, global runtime, provider, model, and exact route limits")
	}
	seenRuntime := map[RuntimeResourceKey]bool{}
	for _, limit := range capacity.RuntimeSlots {
		if !validRuntimeKey(limit.Runtime) || limit.Slots < 1 || limit.Slots > maxResourceSlots || seenRuntime[limit.Runtime] {
			return errors.New("invalid exact runtime capacity")
		}
		seenRuntime[limit.Runtime] = true
	}
	seenProviders := map[string]bool{}
	for _, limit := range capacity.ProviderSlots {
		if !validResourceIdentifier(limit.Provider) || limit.Slots < 1 || limit.Slots > maxResourceSlots || seenProviders[limit.Provider] {
			return errors.New("invalid provider capacity")
		}
		seenProviders[limit.Provider] = true
	}
	seenModels := map[ProviderModelKey]bool{}
	for _, limit := range capacity.ModelSlots {
		if !validResourceIdentifier(limit.Model.Provider) || !validResourceIdentifier(limit.Model.Model) || limit.Slots < 1 || limit.Slots > maxResourceSlots || seenModels[limit.Model] {
			return errors.New("invalid provider/model capacity")
		}
		seenModels[limit.Model] = true
	}
	return nil
}

func validateResourceDemand(demand TaskResourceDemand) error {
	if !idPattern.MatchString(demand.TaskID) || demand.CPUMilli < 1 || demand.CPUMilli > maxResourceCPUMilli || demand.MemoryMiB < 1 || demand.MemoryMiB > maxResourceMemoryMiB || demand.VerificationSlots < 0 || demand.VerificationSlots > maxResourceSlots || demand.RuntimeSlots < 1 || demand.RuntimeSlots > maxResourceSlots || !validRuntimeKey(demand.Runtime) {
		return errors.New("resource demand must be positive, bounded, and bound to an exact runtime")
	}
	return nil
}

func validRuntimeKey(key RuntimeResourceKey) bool {
	for _, value := range []string{key.ProfileID, key.Provider, key.Model} {
		if value == "" || len(value) > 512 || strings.TrimSpace(value) != value || !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return false
		}
	}
	return true
}

func validResourceIdentifier(value string) bool {
	return value != "" && len(value) <= 512 && strings.TrimSpace(value) == value && utf8.ValidString(value) && strings.IndexFunc(value, unicode.IsControl) < 0
}

func pathsOverlapFoldAny(first, second []string) bool {
	for _, a := range first {
		for _, b := range second {
			a, b = strings.ToLower(strings.TrimSuffix(a, "/")), strings.ToLower(strings.TrimSuffix(b, "/"))
			if a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/") {
				return true
			}
		}
	}
	return false
}

func resourceOverCapacity(used ResourceTotals, demand TaskResourceDemand, capacity ResourceCapacity) string {
	if demand.CPUMilli > capacity.CPUMilli-used.EstimatedCPUMilli {
		return "cpu_milli"
	}
	if demand.MemoryMiB > capacity.MemoryMiB-used.EstimatedMemoryMiB {
		return "memory_mib"
	}
	if demand.VerificationSlots > capacity.VerificationSlots-used.EstimatedVerificationSlots {
		return "verification_slots"
	}
	if demand.RuntimeSlots > capacity.TotalRuntimeSlots-used.TotalRuntimeSlots {
		return "total_runtime_slots"
	}
	providerSlots, _ := providerCapacity(capacity.ProviderSlots, demand.Runtime.Provider)
	if demand.RuntimeSlots > providerSlots-providerTotal(used.ProviderSlots, demand.Runtime.Provider) {
		return "provider_slots"
	}
	modelKey := providerModel(demand.Runtime)
	modelSlots, _ := modelCapacity(capacity.ModelSlots, modelKey)
	if demand.RuntimeSlots > modelSlots-modelTotal(used.ModelSlots, modelKey) {
		return "model_slots"
	}
	capacitySlots, _ := runtimeCapacity(capacity.RuntimeSlots, demand.Runtime)
	if demand.RuntimeSlots > capacitySlots-runtimeTotal(used.RuntimeSlots, demand.Runtime) {
		return "runtime_slots"
	}
	return ""
}

func addResourceTotals(total *ResourceTotals, demand TaskResourceDemand) {
	total.EstimatedCPUMilli += demand.CPUMilli
	total.EstimatedMemoryMiB += demand.MemoryMiB
	total.EstimatedVerificationSlots += demand.VerificationSlots
	total.TotalRuntimeSlots += demand.RuntimeSlots
	addProviderTotal(total, demand.Runtime.Provider, demand.RuntimeSlots)
	addModelTotal(total, providerModel(demand.Runtime), demand.RuntimeSlots)
	for i := range total.RuntimeSlots {
		if total.RuntimeSlots[i].Runtime == demand.Runtime {
			total.RuntimeSlots[i].Slots += demand.RuntimeSlots
			return
		}
	}
	total.RuntimeSlots = append(total.RuntimeSlots, RuntimeSlotTotal{Runtime: demand.Runtime, Slots: demand.RuntimeSlots})
	sort.Slice(total.RuntimeSlots, func(i, j int) bool {
		return runtimeKeyLess(total.RuntimeSlots[i].Runtime, total.RuntimeSlots[j].Runtime)
	})
}

// SumResourceDemands derives the exact overall work total for the supplied
// demands in deterministic serialized order. It is the shared derivation for
// durable preparation totals; per-wave Estimated values remain the separate
// per-wave peak for that wave's members.
func SumResourceDemands(demands []TaskResourceDemand) ResourceTotals {
	total := ResourceTotals{ProviderSlots: []ProviderSlotTotal{}, ModelSlots: []ModelSlotTotal{}, RuntimeSlots: []RuntimeSlotTotal{}}
	for _, demand := range demands {
		addResourceTotals(&total, demand)
	}
	return total
}

func providerModel(runtime RuntimeResourceKey) ProviderModelKey {
	return ProviderModelKey{Provider: runtime.Provider, Model: runtime.Model}
}

func providerCapacity(limits []ProviderSlotLimit, provider string) (int, bool) {
	for _, limit := range limits {
		if limit.Provider == provider {
			return limit.Slots, true
		}
	}
	return 0, false
}

func modelCapacity(limits []ModelSlotLimit, model ProviderModelKey) (int, bool) {
	for _, limit := range limits {
		if limit.Model == model {
			return limit.Slots, true
		}
	}
	return 0, false
}

func providerTotal(totals []ProviderSlotTotal, provider string) int {
	for _, total := range totals {
		if total.Provider == provider {
			return total.Slots
		}
	}
	return 0
}

func modelTotal(totals []ModelSlotTotal, model ProviderModelKey) int {
	for _, total := range totals {
		if total.Model == model {
			return total.Slots
		}
	}
	return 0
}

func addProviderTotal(total *ResourceTotals, provider string, slots int) {
	for i := range total.ProviderSlots {
		if total.ProviderSlots[i].Provider == provider {
			total.ProviderSlots[i].Slots += slots
			return
		}
	}
	total.ProviderSlots = append(total.ProviderSlots, ProviderSlotTotal{Provider: provider, Slots: slots})
	sort.Slice(total.ProviderSlots, func(i, j int) bool { return total.ProviderSlots[i].Provider < total.ProviderSlots[j].Provider })
}

func addModelTotal(total *ResourceTotals, model ProviderModelKey, slots int) {
	for i := range total.ModelSlots {
		if total.ModelSlots[i].Model == model {
			total.ModelSlots[i].Slots += slots
			return
		}
	}
	total.ModelSlots = append(total.ModelSlots, ModelSlotTotal{Model: model, Slots: slots})
	sort.Slice(total.ModelSlots, func(i, j int) bool {
		if total.ModelSlots[i].Model.Provider != total.ModelSlots[j].Model.Provider {
			return total.ModelSlots[i].Model.Provider < total.ModelSlots[j].Model.Provider
		}
		return total.ModelSlots[i].Model.Model < total.ModelSlots[j].Model.Model
	})
}

func runtimeCapacity(limits []RuntimeSlotLimit, key RuntimeResourceKey) (int, bool) {
	for _, limit := range limits {
		if limit.Runtime == key {
			return limit.Slots, true
		}
	}
	return 0, false
}

func runtimeTotal(totals []RuntimeSlotTotal, key RuntimeResourceKey) int {
	for _, total := range totals {
		if total.Runtime == key {
			return total.Slots
		}
	}
	return 0
}

func runtimeKeyLess(a, b RuntimeResourceKey) bool {
	if a.ProfileID != b.ProfileID {
		return a.ProfileID < b.ProfileID
	}
	if a.Provider != b.Provider {
		return a.Provider < b.Provider
	}
	return a.Model < b.Model
}
