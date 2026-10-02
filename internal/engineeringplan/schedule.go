package engineeringplan

import (
	"fmt"
	"sort"
)

// Strategy determines ordering among independently-ready tasks. Both policies
// preserve dependency and write ownership safety; CriticalPath is a cheap
// deterministic heuristic, not a claim of optimal scheduling.
type Strategy string

const (
	// Lexical orders ready tasks by deterministic ID order.
	Lexical Strategy = "lexical"
	// CriticalPath orders ready tasks by longest downstream work first.
	CriticalPath Strategy = "critical-path"
)

// Select returns at most capacity conflict-free ready tasks. It is suitable for
// dispatchers that create one isolated candidate/workspace per selected task.
func Select(g Graph, capacity int, strategy Strategy) ([]Task, error) {
	if capacity < 1 {
		return nil, fmt.Errorf("capacity must be positive")
	}
	ready, err := Ready(g)
	if err != nil {
		return nil, err
	}
	if strategy != Lexical && strategy != CriticalPath {
		return nil, fmt.Errorf("unsupported scheduling strategy")
	}
	if strategy == CriticalPath {
		lengths := downstreamLengths(g)
		sort.SliceStable(ready, func(i, j int) bool {
			if lengths[ready[i].ID] != lengths[ready[j].ID] {
				return lengths[ready[i].ID] > lengths[ready[j].ID]
			}
			return ready[i].ID < ready[j].ID
		})
	}
	selected := make([]Task, 0, capacity)
	for _, t := range ready {
		if len(selected) == capacity {
			break
		}
		conflict := false
		for _, other := range selected {
			if writesConflict(t, other) {
				conflict = true
				break
			}
		}
		if !conflict {
			selected = append(selected, t)
		}
	}
	return selected, nil
}

func downstreamLengths(g Graph) map[string]int {
	children := map[string][]Task{}
	tasks := map[string]Task{}
	for _, t := range g.Tasks {
		tasks[t.ID] = t
		for _, d := range t.Dependencies {
			children[d] = append(children[d], t)
		}
	}
	memo := map[string]int{}
	var visit func(string) int
	visit = func(id string) int {
		if n, ok := memo[id]; ok {
			return n
		}
		n := tasks[id].EstimatedSeconds
		best := 0
		for _, child := range children[id] {
			if m := visit(child.ID); m > best {
				best = m
			}
		}
		memo[id] = n + best
		return memo[id]
	}
	for _, t := range g.Tasks {
		visit(t.ID)
	}
	return memo
}

// SimulateMakespan is a deterministic dependency-aware simulation. It is used
// for synthetic policy comparisons only; production runs must report observed
// wall-clock evidence separately.
func SimulateMakespan(g Graph, workers int, strategy Strategy) (int, error) {
	if err := g.Validate(); err != nil {
		return 0, err
	}
	if workers < 1 {
		return 0, fmt.Errorf("workers must be positive")
	}
	remaining := make(map[string]Task, len(g.Tasks))
	completed := map[string]bool{}
	for _, t := range g.Tasks {
		remaining[t.ID] = t
		if t.Completed {
			completed[t.ID] = true
		}
	}
	time := 0
	var running []simulationActive
	for len(remaining) > 0 || len(running) > 0 {
		for len(running) < workers {
			ready := simReady(remaining, completed, running, g, strategy)
			if len(ready) == 0 {
				break
			}
			t := ready[0]
			delete(remaining, t.ID)
			running = append(running, simulationActive{t, time + t.EstimatedSeconds})
		}
		if len(running) == 0 {
			return 0, fmt.Errorf("simulation stalled")
		}
		next := running[0].finish
		for _, r := range running {
			if r.finish < next {
				next = r.finish
			}
		}
		time = next
		still := running[:0]
		for _, r := range running {
			if r.finish == time {
				completed[r.task.ID] = true
			} else {
				still = append(still, r)
			}
		}
		running = still
	}
	return time, nil
}

type simulationActive struct {
	task   Task
	finish int
}

func simReady(remaining map[string]Task, completed map[string]bool, running []simulationActive, g Graph, strategy Strategy) []Task {
	var ready []Task
	for _, t := range remaining {
		ok := true
		for _, dep := range t.Dependencies {
			if !completed[dep] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		conflict := false
		for _, r := range running {
			if writesConflict(t, r.task) {
				conflict = true
				break
			}
		}
		if !conflict {
			ready = append(ready, t)
		}
	}
	if strategy == CriticalPath {
		lengths := downstreamLengths(g)
		sort.Slice(ready, func(i, j int) bool {
			if lengths[ready[i].ID] != lengths[ready[j].ID] {
				return lengths[ready[i].ID] > lengths[ready[j].ID]
			}
			return ready[i].ID < ready[j].ID
		})
	} else {
		sort.Slice(ready, func(i, j int) bool { return ready[i].ID < ready[j].ID })
	}
	return ready
}
