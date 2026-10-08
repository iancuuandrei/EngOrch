package engineeringplan

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Typed coupling levels for staged scheduling (Phase E, brief section 10.3).
// Dependency determines legal readiness, coupling predicts concurrent
// interference risk, ownership determines write authority. Different files do
// not prove independence: C0 means no observed coupling, never guaranteed
// zero. No numeric weights, learned models, measured makespan, confidence or
// coupling-absence claims are encoded here.
const (
	CouplingC1 = "C1"
	CouplingC2 = "C2"
	CouplingC3 = "C3"
	CouplingC4 = "C4"
)

// Coupling provenance records where a typed relationship was observed or
// declared. This initial feature admits only planner-declared advisory risk:
// planner_declared_advisory never proves controller-observed truth, never
// grants write authority, readiness or ownership, and never proves absence
// (C0). The observed_generator_binding and observed_source_topology spellings
// are reserved for a future source-bound admission that validates the exact
// source/candidate/path binding against an existing admitted generator or
// topology record; they are rejected as forged until that Phase E admission
// exists. Graph source/candidate/journal hashing binds declared risk inputs,
// it does not prove a cited fact.
const (
	CouplingProvenancePlannerAdvisory   = "planner_declared_advisory"
	CouplingProvenanceGeneratorObserved = "observed_generator_binding"
	CouplingProvenanceTopologyObserved  = "observed_source_topology"
)

// MaxTaskCouplings bounds typed relationships to the complete graph on eight
// implementations (8*7/2 = 28). Larger sets are rejected rather than searched
// heuristically or truncated silently.
const MaxTaskCouplings = 28

// TaskCoupling is one typed, bounded advisory relationship between two
// implementation tasks. From is always lexicographically smaller than To so
// the canonical pair order is deterministic. Level is C1 (weak) through C4
// (hard). Reason names the observed or declared structural class, Provenance
// records its evidence source, Evidence carries a bounded human-auditable
// pointer such as a generator directive, file pair or topology reason. None
// of these fields grants readiness, ownership or write authority.
type TaskCoupling struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Level      string `json:"level"`
	Reason     string `json:"reason"`
	Provenance string `json:"provenance"`
	Evidence   string `json:"evidence"`
}

// validateTaskCoupling checks one relationship's shape. Task existence and
// implementation-kind checks stay with the graph validator so this helper
// remains reusable for subset contexts. Only planner-declared advisory
// provenance is admitted: observed generator/topology labels without an
// existing admitted source-bound record are rejected as forged rather than
// relabeled as controller-observed truth. Reason and evidence must be
// nonempty trimmed UTF-8 without control characters within their byte bounds;
// endpoints must match the task ID pattern in canonical from<to order.
func validateTaskCoupling(c TaskCoupling) error {
	if !idPattern.MatchString(c.From) || !idPattern.MatchString(c.To) {
		return fmt.Errorf("invalid coupling task %q-%q", c.From, c.To)
	}
	if c.From == c.To {
		return fmt.Errorf("coupling pair %q is self-coupled", c.From)
	}
	if c.From > c.To {
		return fmt.Errorf("coupling pair %q-%q is not in canonical order", c.From, c.To)
	}
	switch c.Level {
	case CouplingC1, CouplingC2, CouplingC3, CouplingC4:
	default:
		return fmt.Errorf("invalid coupling level %q", c.Level)
	}
	if strings.TrimSpace(c.Reason) == "" || len(c.Reason) > 128 || !utf8.ValidString(c.Reason) || strings.IndexFunc(c.Reason, unicode.IsControl) >= 0 {
		return fmt.Errorf("invalid coupling reason for %q-%q", c.From, c.To)
	}
	switch c.Provenance {
	case CouplingProvenancePlannerAdvisory:
	default:
		if c.Provenance == CouplingProvenanceGeneratorObserved || c.Provenance == CouplingProvenanceTopologyObserved {
			return fmt.Errorf("forged observed coupling provenance for %q-%q requires an admitted source-bound record", c.From, c.To)
		}
		return fmt.Errorf("invalid coupling provenance for %q-%q", c.From, c.To)
	}
	if strings.TrimSpace(c.Evidence) == "" || len(c.Evidence) > 256 || !utf8.ValidString(c.Evidence) || strings.IndexFunc(c.Evidence, unicode.IsControl) >= 0 {
		return fmt.Errorf("invalid coupling evidence for %q-%q", c.From, c.To)
	}
	return nil
}

// ValidateTaskCouplings checks a complete coupling set's shape: bound,
// per-item shape, canonical order, no self pairs and no duplicate unordered
// pairs. It does not grant readiness or ownership and does not check
// dependency satisfaction; staged validation owns the C4 hard gate.
func ValidateTaskCouplings(couplings []TaskCoupling) error {
	if len(couplings) > MaxTaskCouplings {
		return errors.New("task coupling set exceeds 28 relationships")
	}
	seen := make(map[string]bool, len(couplings))
	for _, c := range couplings {
		if err := validateTaskCoupling(c); err != nil {
			return err
		}
		key := c.From + "\x00" + c.To
		if seen[key] {
			return fmt.Errorf("duplicate coupling pair %q-%q", c.From, c.To)
		}
		seen[key] = true
	}
	return nil
}

// validateGraphCouplings binds couplings to exact implementation tasks in the
// graph. Every endpoint must be an existing implementation task; research,
// design, verification and review tasks cannot carry implementation coupling.
// Couplings never create readiness or ownership: this function only checks
// membership and shape.
func validateGraphCouplings(g Graph) error {
	if len(g.Couplings) == 0 {
		return nil
	}
	if err := ValidateTaskCouplings(g.Couplings); err != nil {
		return err
	}
	implByID := make(map[string]bool, len(g.Tasks))
	for _, t := range g.Tasks {
		if t.Kind == Implementation {
			implByID[t.ID] = true
		}
	}
	for _, c := range g.Couplings {
		if !implByID[c.From] {
			return fmt.Errorf("coupling task %q is not an implementation", c.From)
		}
		if !implByID[c.To] {
			return fmt.Errorf("coupling task %q is not an implementation", c.To)
		}
	}
	return nil
}

// couplingPairKey returns the canonical unordered key for two task IDs.
// Callers must supply IDs in any order; the key is order-independent.
func couplingPairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "\x00" + b
}

// CouplingLevelByPair indexes couplings by unordered task pair. Later
// duplicates are impossible after validation; the first entry wins.
func CouplingLevelByPair(couplings []TaskCoupling) map[string]string {
	byPair := make(map[string]string, len(couplings))
	for _, c := range couplings {
		key := couplingPairKey(c.From, c.To)
		if _, exists := byPair[key]; !exists {
			byPair[key] = c.Level
		}
	}
	return byPair
}

// countCouplingLevels counts C1, C2 and C3 pairs fully contained in the
// selected set. C4 pairs are hard gates handled by feasibility, not counted
// here. C0 (absent) contributes zero and never proves independence.
func countCouplingLevels(selected map[string]bool, byPair map[string]string) (c3, c2, c1 int) {
	ids := make([]string, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for i := range ids {
		for _, other := range ids[i+1:] {
			switch byPair[couplingPairKey(ids[i], other)] {
			case CouplingC3:
				c3++
			case CouplingC2:
				c2++
			case CouplingC1:
				c1++
			}
		}
	}
	return c3, c2, c1
}

// couplingSubsetFeasible reports whether a masked subset is pairwise
// independent under hard gates plus the C4 hard-coupling gate. Any C4 pair
// co-selected in the same wave is infeasible: same-generated-family work
// requires one owner or an explicit dependency, and serial waves forked from
// the same parent base do not make an unsafe split safe. Dependency-staged
// completion with parent advancement is required instead. The shared v39
// feasibility (cohort bound, dependency/write-overlap independence, every
// explicit resource ceiling in sorted ID order) is reused rather than
// duplicated; only the C4 gate is added here.
func couplingSubsetFeasible(ids []string, mask int, provided map[string]Task, demandByID map[string]TaskResourceDemand, capacity ResourceCapacity, maxTasks int, byID map[string]Task, byPair map[string]string) bool {
	if !lexicographicSubsetFeasible(ids, mask, provided, demandByID, capacity, maxTasks, byID) {
		return false
	}
	for i := range ids {
		if mask&(1<<i) == 0 {
			continue
		}
		for j := i + 1; j < len(ids); j++ {
			if mask&(1<<j) == 0 {
				continue
			}
			if byPair[couplingPairKey(ids[i], ids[j])] == CouplingC4 {
				return false
			}
		}
	}
	return true
}

// couplingSubsetBetter applies the coupling-aware lexicographic objective
// order between two feasible masks over declared estimates only:
//
//  1. minimize C3 pairs co-scheduled in the same wave (strong coupling risk
//     before utilization: a smaller wave that avoids a known strong pair
//     beats a fuller wave that co-schedules it);
//  2. maximize admitted task count;
//  3. minimize C2 pairs co-scheduled in the same wave;
//  4. minimize C1 pairs co-scheduled in the same wave;
//  5. maximize summed downstream critical-path length (declared
//     EstimatedSeconds, not a measurement);
//  6. minimize summed estimated CPU, then memory, then verification slots,
//     then runtime slots, each as a separate level;
//  7. prefer the lexicographically smallest sorted task-ID list.
//
// Levels 1, 3 and 4 optimize ordinal concurrency/integration risk before and
// around utilization, as justified by actual source-declared coupling data.
// Only C4 is a hard serial mandate; C3/C2/C1 never force serialization by
// themselves, and no numeric weights, measured makespan, confidence or
// coupling-absence claims are used. Sorted ID order supplies the final
// tie-break, so the verdict is independent of enumeration and input order.
// The frozen v1.1.39 selector 1 keeps its count-first order unchanged; only
// this coupling-aware selector 2 differs.
func couplingSubsetBetter(ids []string, mask, best int, demandByID map[string]TaskResourceDemand, critical map[string]int, byPair map[string]string) bool {
	maskSelected := map[string]bool{}
	bestSelected := map[string]bool{}
	for i, id := range ids {
		if mask&(1<<i) != 0 {
			maskSelected[id] = true
		}
		if best&(1<<i) != 0 {
			bestSelected[id] = true
		}
	}
	maskC3, maskC2, maskC1 := countCouplingLevels(maskSelected, byPair)
	bestC3, bestC2, bestC1 := countCouplingLevels(bestSelected, byPair)
	if maskC3 != bestC3 {
		return maskC3 < bestC3
	}
	count, bestCount := len(maskSelected), len(bestSelected)
	if count != bestCount {
		return count > bestCount
	}
	if maskC2 != bestC2 {
		return maskC2 < bestC2
	}
	if maskC1 != bestC1 {
		return maskC1 < bestC1
	}
	criticalSum, bestCriticalSum := 0, 0
	var cpu, mem int64
	var verify, runtime int
	var bestCPU, bestMem int64
	var bestVerify, bestRuntime int
	for i, id := range ids {
		if mask&(1<<i) != 0 {
			criticalSum += critical[id]
			demand := demandByID[id]
			cpu += demand.CPUMilli
			mem += demand.MemoryMiB
			verify += demand.VerificationSlots
			runtime += demand.RuntimeSlots
		}
		if best&(1<<i) != 0 {
			bestCriticalSum += critical[id]
			demand := demandByID[id]
			bestCPU += demand.CPUMilli
			bestMem += demand.MemoryMiB
			bestVerify += demand.VerificationSlots
			bestRuntime += demand.RuntimeSlots
		}
	}
	if criticalSum != bestCriticalSum {
		return criticalSum > bestCriticalSum
	}
	if cpu != bestCPU {
		return cpu < bestCPU
	}
	if mem != bestMem {
		return mem < bestMem
	}
	if verify != bestVerify {
		return verify < bestVerify
	}
	if runtime != bestRuntime {
		return runtime < bestRuntime
	}
	for i := range ids {
		inMask, inBest := mask&(1<<i) != 0, best&(1<<i) != 0
		if inMask != inBest {
			return inMask
		}
	}
	return false
}

// smallestCoupledPartner returns the smallest selected task ID that shares a
// C1, C2 or C3 coupling with the excluded task, or empty when the exclusion
// is not explained by coupling risk. C4 pairs never reach here: they are
// infeasible and render as dependency_or_write_conflict above.
func smallestCoupledPartner(excluded string, selected map[string]bool, byPair map[string]string) string {
	partner := ""
	for id := range selected {
		switch byPair[couplingPairKey(excluded, id)] {
		case CouplingC1, CouplingC2, CouplingC3:
			if partner == "" || id < partner {
				partner = id
			}
		}
	}
	return partner
}

// couplingOptimum enumerates at most 255 nonempty subsets of at most eight
// candidates and renders the coupling-aware optimum with greedy-compatible
// order and blocked provenance. C4 pairs are infeasible in the same wave;
// C3 risk is minimized before admitted count, then C2/C1 pairs are minimized
// before critical/resource packing. Larger candidate sets are rejected by callers.
func couplingOptimum(g Graph, provided map[string]Task, demandByID map[string]TaskResourceDemand, capacity ResourceCapacity, maxTasks int, couplings []TaskCoupling) (ResourceCohort, error) {
	ids := make([]string, 0, len(provided))
	for id := range provided {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	critical := downstreamLengths(g)
	byID := graphTaskIndex(g)
	byPair := CouplingLevelByPair(couplings)
	best := -1
	for mask := 1; mask < (1 << len(ids)); mask++ {
		if !couplingSubsetFeasible(ids, mask, provided, demandByID, capacity, maxTasks, byID, byPair) {
			continue
		}
		if best == -1 || couplingSubsetBetter(ids, mask, best, demandByID, critical, byPair) {
			best = mask
		}
	}
	selected := map[string]bool{}
	if best != -1 {
		for i, id := range ids {
			if best&(1<<i) != 0 {
				selected[id] = true
			}
		}
	}
	return couplingCohortResult(provided, selected, demandByID, capacity, maxTasks, critical, byID, byPair)
}

// couplingCohortResult renders the optimal set with greedy-compatible display
// order and blocked provenance. C4 conflicts render as
// dependency_or_write_conflict so wave drivers reject them with the existing
// vocabulary; risk-based exclusions render as the explicit bounded reason
// coupling_risk with the smallest coupled partner as BlockingTaskID, so a
// remainder that fits capacity but was excluded to avoid a C3/C2/C1 pair is
// explained instead of failing closed with the v39 remainder error. Every
// other non-selected candidate must be a conflict, cohort-limit or capacity
// remainder of the optimal set. Display order, totals derivation
// (addResourceTotals), capacity checks (resourceOverCapacity,
// capacityBlockedReason) and candidate ordering (orderResourceCandidates) are
// shared with the frozen selectors rather than duplicated.
func couplingCohortResult(provided map[string]Task, selected map[string]bool, demandByID map[string]TaskResourceDemand, capacity ResourceCapacity, maxTasks int, critical map[string]int, byID map[string]Task, byPair map[string]string) (ResourceCohort, error) {
	ordered := orderResourceCandidates(provided, critical)
	out := ResourceCohort{Tasks: make([]Task, 0, maxTasks), Blocked: []ResourceBlockReason{}, Estimated: ResourceTotals{ProviderSlots: []ProviderSlotTotal{}, ModelSlots: []ModelSlotTotal{}, RuntimeSlots: []RuntimeSlotTotal{}}}
	for _, task := range ordered {
		if selected[task.ID] {
			out.Tasks = append(out.Tasks, task)
			addResourceTotals(&out.Estimated, demandByID[task.ID])
		}
	}
	for _, task := range ordered {
		if selected[task.ID] {
			continue
		}
		conflict := ""
		for _, chosen := range ordered {
			if !selected[chosen.ID] {
				continue
			}
			if dependsOn(task.ID, chosen.ID, byID) || dependsOn(chosen.ID, task.ID, byID) || pathsOverlapFoldAny(task.WritePaths, chosen.WritePaths) || byPair[couplingPairKey(task.ID, chosen.ID)] == CouplingC4 {
				if conflict == "" || chosen.ID < conflict {
					conflict = chosen.ID
				}
			}
		}
		if conflict != "" {
			out.Blocked = append(out.Blocked, ResourceBlockReason{TaskID: task.ID, Reason: "dependency_or_write_conflict", BlockingTaskID: conflict})
			continue
		}
		if len(selected) >= maxTasks {
			out.Blocked = append(out.Blocked, ResourceBlockReason{TaskID: task.ID, Reason: "cohort_limit"})
			continue
		}
		demand := demandByID[task.ID]
		if resource := resourceOverCapacity(out.Estimated, demand, capacity); resource != "" {
			out.Blocked = append(out.Blocked, capacityBlockedReason(task.ID, resource, demand))
			continue
		}
		if partner := smallestCoupledPartner(task.ID, selected, byPair); partner != "" {
			out.Blocked = append(out.Blocked, ResourceBlockReason{TaskID: task.ID, Reason: "coupling_risk", BlockingTaskID: partner})
			continue
		}
		return ResourceCohort{}, fmt.Errorf("coupling-aware remainder %q fits the optimal cohort", task.ID)
	}
	return out, nil
}

// SelectResourceCohortCouplingAware selects the exact finite coupling-aware
// optimum over the ready implementation set. Hard gates match the frozen
// selectors (exact-ready identity, pairwise dependency/write-overlap
// independence, every explicit resource ceiling, 1..8 bound) plus the C4
// hard-coupling gate. Objectives minimize C3 risk before admitted count, then
// minimize C2, then C1 co-scheduling before existing critical/resource
// packing. At most eight ready
// implementations are admitted; unknown attempts are rejected via Ready.
func SelectResourceCohortCouplingAware(g Graph, ready []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int, couplings []TaskCoupling) (ResourceCohort, error) {
	if err := checkResourceCohortBound(maxTasks); err != nil {
		return ResourceCohort{}, err
	}
	if err := checkResourceGraphCapacity(g, capacity); err != nil {
		return ResourceCohort{}, err
	}
	if len(ready) == 0 || len(ready) > 8 {
		return ResourceCohort{}, errors.New("coupling-aware cohort requires one to eight ready implementations")
	}
	_, provided, err := exactReadyCohortMaps(g, ready)
	if err != nil {
		return ResourceCohort{}, err
	}
	demandByID, err := bindExactCohortDemands(provided, demands, capacity)
	if err != nil {
		return ResourceCohort{}, err
	}
	if err := ValidateTaskCouplings(couplings); err != nil {
		return ResourceCohort{}, err
	}
	return couplingOptimum(g, provided, demandByID, capacity, maxTasks, couplings)
}

// SelectResourceCohortSubsetCouplingAware selects the exact finite
// coupling-aware optimum over an explicit candidate subset. It never
// fabricates Completed mutations to trick Ready. Staged wave derivation uses
// this for each remainder after the first exact-ready optimum.
func SelectResourceCohortSubsetCouplingAware(g Graph, candidates []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int, couplings []TaskCoupling) (ResourceCohort, error) {
	if err := checkResourceCohortBound(maxTasks); err != nil {
		return ResourceCohort{}, err
	}
	if err := checkResourceGraphCapacity(g, capacity); err != nil {
		return ResourceCohort{}, err
	}
	if len(candidates) == 0 {
		return ResourceCohort{}, errors.New("resource cohort subset is empty")
	}
	provided, _, err := subsetCohortMaps(g, candidates)
	if err != nil {
		return ResourceCohort{}, err
	}
	demandByID, err := bindSubsetCohortDemands(provided, demands, capacity)
	if err != nil {
		return ResourceCohort{}, err
	}
	if err := ValidateTaskCouplings(couplings); err != nil {
		return ResourceCohort{}, err
	}
	return couplingOptimum(g, provided, demandByID, capacity, maxTasks, couplings)
}

// SelectResourceWavesCouplingAware deterministically partitions the exact
// ready implementation set into nonempty resource-bounded waves using the
// coupling-aware optimum for the first wave and each remainder. Hard gates,
// error identities and coverage diagnostics match SelectResourceWaves; only
// the per-wave optimum differs. C4 pairs never share a wave; C3 risk is
// minimized before admitted count, then C2/C1 pairs are minimized before
// critical/resource packing. A C4 pair among currently
// ready tasks is rejected as an unsafe generator split (one owner or an
// explicit dependency is required); serial waves from the same parent base
// do not make it safe.
func SelectResourceWavesCouplingAware(g Graph, ready []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int, couplings []TaskCoupling) ([]ResourceCohort, error) {
	if err := ValidateTaskCouplings(couplings); err != nil {
		return nil, err
	}
	readyIDs := make(map[string]bool, len(ready))
	for _, task := range ready {
		readyIDs[task.ID] = true
	}
	for _, c := range couplings {
		if c.Level != CouplingC4 {
			continue
		}
		if readyIDs[c.From] && readyIDs[c.To] {
			return nil, fmt.Errorf("coupling-aware waves reject C4 generator split between %q and %q without one owner or an explicit dependency", c.From, c.To)
		}
	}
	return partitionResourceWaves(g, ready, demands, capacity, maxTasks,
		func(g Graph, ready []Task, demands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int) (ResourceCohort, error) {
			return SelectResourceCohortCouplingAware(g, ready, demands, capacity, maxTasks, couplings)
		},
		func(g Graph, remainder []Task, remainderDemands []TaskResourceDemand, capacity ResourceCapacity, maxTasks int) (ResourceCohort, error) {
			return SelectResourceCohortSubsetCouplingAware(g, remainder, remainderDemands, capacity, maxTasks, couplings)
		})
}
