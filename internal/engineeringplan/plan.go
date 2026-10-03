// Package engineeringplan defines the bounded, inspectable task graph used by
// an autonomous engineering run. It is deliberately independent of execution
// and journaling: callers bind an accepted graph to their own durable run.
package engineeringplan

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Version is the exact planner wire version accepted by Validate and ParseJSON.
const Version = 1

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

// Mode selects the planner representation. Direct holds exactly one task;
// Graph holds an ordered dependency graph.
type Mode string

const (
	// ModeDirect is a single-task plan without planner graph overhead.
	ModeDirect Mode = "direct"
	// ModeGraph is a multi-task dependency graph plan.
	ModeGraph Mode = "graph"
)

// Kind classifies the engineering activity expected from a task.
type Kind string

const (
	// Research gathers facts before a change is designed.
	Research Kind = "research"
	// Design produces a change plan without modifying implementation.
	Design Kind = "design"
	// Implementation makes the scoped source change.
	Implementation Kind = "implementation"
	// Verification checks the change with tests or inspections.
	Verification Kind = "verification"
	// Review evaluates a completed change for correctness and safety.
	Review Kind = "review"
)

// AttemptOutcome records the durable result of one execution attempt. Unknown
// deliberately blocks automatic retry because external effects are unresolved.
type AttemptOutcome string

const (
	// AttemptCompleted marks an attempt whose effects are durably observed.
	AttemptCompleted AttemptOutcome = "completed"
	// AttemptFailed marks an attempt that finished without the expected effect.
	AttemptFailed AttemptOutcome = "failed"
	// AttemptUnknown marks an attempt whose external effect is unresolved.
	AttemptUnknown AttemptOutcome = "unknown"
)

// Evidence is a concrete result expected from a task, such as a source file,
// test receipt, or review finding. It is an expectation, never a claimed result.
type Evidence struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

// Attempt is execution history supplied by the durable runner. An unknown
// attempt deliberately blocks automatic retry because its external effect is
// unresolved.
type Attempt struct {
	ID      string         `json:"id"`
	Outcome AttemptOutcome `json:"outcome"`
}

// Task is one bounded unit of work. ParentID expresses hierarchy only;
// Dependencies control execution order.
type Task struct {
	ID               string     `json:"id"`
	ParentID         string     `json:"parent_id,omitempty"`
	Kind             Kind       `json:"kind"`
	Title            string     `json:"title"`
	Dependencies     []string   `json:"dependencies,omitempty"`
	ScopePaths       []string   `json:"scope_paths"`
	WritePaths       []string   `json:"write_paths,omitempty"`
	ExpectedEvidence []Evidence `json:"expected_evidence"`
	EstimatedSeconds int        `json:"estimated_seconds"`
	Completed        bool       `json:"completed,omitempty"`
	Attempts         []Attempt  `json:"attempts,omitempty"`
}

// Graph is a planner response. Direct plans retain the same representation but
// contain exactly one executable task, avoiding planner overhead for tiny work.
type Graph struct {
	Version int    `json:"version"`
	Mode    Mode   `json:"mode"`
	Summary string `json:"summary"`
	Tasks   []Task `json:"tasks"`
}

// Task returns the task with the given ID, reporting whether it is present.
func (g Graph) Task(id string) (Task, bool) {
	for _, t := range g.Tasks {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// Validate checks version, shape, identifiers, path scoping, dependencies,
// acyclicity, and write ownership so only a bounded safe graph executes.
func (g Graph) Validate() error {
	if g.Version != Version || (g.Mode != ModeDirect && g.Mode != ModeGraph) || strings.TrimSpace(g.Summary) == "" || len(g.Tasks) == 0 || len(g.Tasks) > 64 {
		return errors.New("invalid engineering plan shape")
	}
	byID := make(map[string]Task, len(g.Tasks))
	for _, t := range g.Tasks {
		if !idPattern.MatchString(t.ID) || strings.TrimSpace(t.Title) == "" || !validKind(t.Kind) || len(t.Dependencies) > 64 || len(t.ScopePaths) == 0 || len(t.ScopePaths) > 32 || len(t.WritePaths) > 32 || len(t.ExpectedEvidence) == 0 || len(t.ExpectedEvidence) > 16 || len(t.Attempts) > 64 || t.EstimatedSeconds < 1 || t.EstimatedSeconds > 86400 {
			return fmt.Errorf("invalid task %q", t.ID)
		}
		if _, exists := byID[t.ID]; exists {
			return fmt.Errorf("duplicate task %q", t.ID)
		}
		if err := validatePaths(t.ScopePaths, true); err != nil {
			return fmt.Errorf("task %q scope: %w", t.ID, err)
		}
		if err := validatePaths(t.WritePaths, false); err != nil {
			return fmt.Errorf("task %q writes: %w", t.ID, err)
		}
		for _, p := range t.WritePaths {
			if !withinAny(p, t.ScopePaths) {
				return fmt.Errorf("task %q write path outside scope", t.ID)
			}
		}
		seenAttempt := map[string]bool{}
		for _, a := range t.Attempts {
			if !idPattern.MatchString(a.ID) || seenAttempt[a.ID] || (a.Outcome != AttemptCompleted && a.Outcome != AttemptFailed && a.Outcome != AttemptUnknown) {
				return fmt.Errorf("invalid task %q attempt", t.ID)
			}
			seenAttempt[a.ID] = true
		}
		for _, e := range t.ExpectedEvidence {
			if strings.TrimSpace(e.Kind) == "" || strings.TrimSpace(e.Description) == "" {
				return fmt.Errorf("invalid task %q evidence", t.ID)
			}
		}
		byID[t.ID] = t
	}
	if g.Mode == ModeDirect && len(g.Tasks) != 1 {
		return errors.New("direct plan must contain exactly one task")
	}
	for _, t := range g.Tasks {
		if t.ParentID != "" {
			if _, ok := byID[t.ParentID]; !ok || t.ParentID == t.ID {
				return fmt.Errorf("task %q parent missing", t.ID)
			}
		}
		seen := map[string]bool{}
		for _, dep := range t.Dependencies {
			if dep == t.ID || seen[dep] {
				return fmt.Errorf("invalid dependency for %q", t.ID)
			}
			if _, ok := byID[dep]; !ok {
				return fmt.Errorf("task %q dependency missing", t.ID)
			}
			seen[dep] = true
		}
	}
	if err := validateAcyclic(g, byID); err != nil {
		return err
	}
	if err := validateParentAcyclic(byID); err != nil {
		return err
	}
	for i, a := range g.Tasks {
		for _, b := range g.Tasks[i+1:] {
			if writesConflict(a, b) && !dependsOn(a.ID, b.ID, byID) && !dependsOn(b.ID, a.ID, byID) {
				return fmt.Errorf("write ownership conflict between %q and %q", a.ID, b.ID)
			}
		}
	}
	return nil
}

func validKind(k Kind) bool {
	return k == Research || k == Design || k == Implementation || k == Verification || k == Review
}

func validatePaths(paths []string, allowRoot bool) error {
	seen := map[string]bool{}
	for _, p := range paths {
		if p == "." && !allowRoot || p == ".." || strings.TrimSpace(p) == "" || path.IsAbs(p) || path.Clean(p) != p || strings.HasPrefix(p, "../") || strings.Contains(p, "\\") || windowsVolumePath(p) || seen[p] {
			return errors.New("path must be unique normalized repository-relative path")
		}
		seen[p] = true
	}
	return nil
}

func windowsVolumePath(p string) bool {
	return len(p) >= 2 && ((p[0] >= 'a' && p[0] <= 'z') || (p[0] >= 'A' && p[0] <= 'Z')) && p[1] == ':'
}

func withinAny(p string, scopes []string) bool {
	for _, s := range scopes {
		if s == "." || p == s || strings.HasPrefix(p, s+"/") {
			return true
		}
	}
	return false
}
func pathsOverlap(a, b string) bool {
	return a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
func writesConflict(a, b Task) bool {
	for _, x := range a.WritePaths {
		for _, y := range b.WritePaths {
			if pathsOverlap(x, y) {
				return true
			}
		}
	}
	return false
}

func dependsOn(id, target string, byID map[string]Task) bool {
	for _, d := range byID[id].Dependencies {
		if d == target || dependsOn(d, target, byID) {
			return true
		}
	}
	return false
}

func validateAcyclic(g Graph, byID map[string]Task) error {
	marks := map[string]uint8{}
	var visit func(string) error
	visit = func(id string) error {
		switch marks[id] {
		case 1:
			return fmt.Errorf("dependency cycle at %q", id)
		case 2:
			return nil
		}
		marks[id] = 1
		for _, dep := range byID[id].Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		marks[id] = 2
		return nil
	}
	for _, t := range g.Tasks {
		if err := visit(t.ID); err != nil {
			return err
		}
	}
	return nil
}

// ParseJSON accepts only the exact v1 planner wire shape. Unknown fields and
// trailing values are rejected before a graph becomes executable.
func ParseJSON(raw []byte) (Graph, error) {
	var g Graph
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&g); err != nil {
		return g, fmt.Errorf("invalid planner JSON: %w", err)
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		if err == nil {
			return g, errors.New("planner JSON contains trailing data")
		}
		return g, fmt.Errorf("invalid trailing planner JSON: %w", err)
	}
	if err := g.Validate(); err != nil {
		return g, err
	}
	return g, nil
}

func canonicalNormalize(raw []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func hashDomain(domain string, normalized []byte) (string, error) {
	if domain == "" {
		return "", errors.New("invalid hash domain")
	}
	h := sha256.New()
	h.Write([]byte(domain + "\n"))
	h.Write(normalized)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// JSONSchema is supplied to providers that support strict structured output.
func JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["version","mode","summary","tasks"],"properties":{"version":{"const":1},"mode":{"enum":["direct","graph"]},"summary":{"type":"string","minLength":1},"tasks":{"type":"array","minItems":1,"maxItems":64,"items":{"type":"object","additionalProperties":false,"required":["id","kind","title","scope_paths","expected_evidence","estimated_seconds"],"properties":{"id":{"type":"string","pattern":"^[a-z][a-z0-9_-]{0,63}$"},"parent_id":{"type":"string","pattern":"^[a-z][a-z0-9_-]{0,63}$"},"kind":{"enum":["research","design","implementation","verification","review"]},"title":{"type":"string","minLength":1},"dependencies":{"type":"array","maxItems":64,"items":{"type":"string","pattern":"^[a-z][a-z0-9_-]{0,63}$"}},"scope_paths":{"type":"array","minItems":1,"maxItems":32,"items":{"type":"string","minLength":1}},"write_paths":{"type":"array","maxItems":32,"items":{"type":"string","minLength":1}},"expected_evidence":{"type":"array","minItems":1,"maxItems":16,"items":{"type":"object","additionalProperties":false,"required":["kind","description"],"properties":{"kind":{"type":"string","minLength":1},"description":{"type":"string","minLength":1}}}},"estimated_seconds":{"type":"integer","minimum":1,"maximum":86400},"completed":{"type":"boolean"},"attempts":{"type":"array","maxItems":64,"items":{"type":"object","additionalProperties":false,"required":["id","outcome"],"properties":{"id":{"type":"string","pattern":"^[a-z][a-z0-9_-]{0,63}$"},"outcome":{"enum":["completed","failed","unknown"]}}}}}}}}}`)
}

// ValidateRevision retains all completed work byte-for-byte at task semantic
// level and forbids adding a retry after an unknown prior attempt.
func ValidateRevision(previous, next Graph) error {
	if err := previous.Validate(); err != nil {
		return fmt.Errorf("previous: %w", err)
	}
	if err := next.Validate(); err != nil {
		return fmt.Errorf("next: %w", err)
	}
	for _, old := range previous.Tasks {
		fresh, ok := next.Task(old.ID)
		if old.Completed && (!ok || !sameCompletedTask(old, fresh)) {
			return fmt.Errorf("completed task %q not retained", old.ID)
		}
		if !ok {
			continue
		}
		if len(fresh.Attempts) < len(old.Attempts) {
			return fmt.Errorf("task %q attempt history removed", old.ID)
		}
		for i := range old.Attempts {
			if fresh.Attempts[i] != old.Attempts[i] {
				return fmt.Errorf("task %q attempt history changed", old.ID)
			}
		}
		if len(fresh.Attempts) > len(old.Attempts) && len(old.Attempts) > 0 && old.Attempts[len(old.Attempts)-1].Outcome == AttemptUnknown {
			return fmt.Errorf("task %q has unresolved attempt; retry denied", old.ID)
		}
	}
	return nil
}

func sameCompletedTask(a, b Task) bool {
	return a.ID == b.ID && a.ParentID == b.ParentID && a.Kind == b.Kind && a.Title == b.Title &&
		stringListEqual(a.Dependencies, b.Dependencies) && stringListEqual(a.ScopePaths, b.ScopePaths) &&
		stringListEqual(a.WritePaths, b.WritePaths) && evidenceEqual(a.ExpectedEvidence, b.ExpectedEvidence) &&
		a.EstimatedSeconds == b.EstimatedSeconds && a.Completed == b.Completed
}
func stringListEqual(a, b []string) bool { return strings.Join(a, "\x00") == strings.Join(b, "\x00") }
func evidenceEqual(a, b []Evidence) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func validateParentAcyclic(byID map[string]Task) error {
	marks := map[string]uint8{}
	var visit func(string) error
	visit = func(id string) error {
		switch marks[id] {
		case 1:
			return fmt.Errorf("parent cycle at %q", id)
		case 2:
			return nil
		}
		marks[id] = 1
		if parent := byID[id].ParentID; parent != "" {
			if err := visit(parent); err != nil {
				return err
			}
		}
		marks[id] = 2
		return nil
	}
	for id := range byID {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

// ParsePlannerGraph accepts only strict planner output: valid v1 JSON with
// no planner-supplied Completed or Attempts. The runner owns all progress;
// any claimed completion is rejected as fictional.
func ParsePlannerGraph(raw []byte) (Graph, error) {
	g, err := ParseJSON(raw)
	if err != nil {
		return g, err
	}
	for _, t := range g.Tasks {
		if t.Completed || len(t.Attempts) != 0 {
			return Graph{}, fmt.Errorf("planner must not supply completion for %q", t.ID)
		}
	}
	return g, nil
}

// PlannerJSONSchema is the strict wire schema for autonomous graph planners:
// completed and attempts are absent because progress is runner-owned.
func PlannerJSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["version","mode","summary","tasks"],"properties":{"version":{"type":"integer","enum":[1]},"mode":{"type":"string","enum":["direct","graph"]},"summary":{"type":"string","minLength":1},"tasks":{"type":"array","minItems":1,"maxItems":64,"items":{"type":"object","additionalProperties":false,"required":["id","parent_id","kind","title","dependencies","scope_paths","write_paths","expected_evidence","estimated_seconds"],"properties":{"id":{"type":"string","pattern":"^[a-z][a-z0-9_-]{0,63}$"},"parent_id":{"type":"string","pattern":"^([a-z][a-z0-9_-]{0,63})?$"},"kind":{"type":"string","enum":["research","design","implementation","verification","review"]},"title":{"type":"string","minLength":1},"dependencies":{"type":"array","maxItems":64,"items":{"type":"string","pattern":"^[a-z][a-z0-9_-]{0,63}$"}},"scope_paths":{"type":"array","minItems":1,"maxItems":32,"items":{"type":"string","minLength":1}},"write_paths":{"type":"array","maxItems":32,"items":{"type":"string","minLength":1}},"expected_evidence":{"type":"array","minItems":1,"maxItems":16,"items":{"type":"object","additionalProperties":false,"required":["kind","description"],"properties":{"kind":{"type":"string","minLength":1},"description":{"type":"string","minLength":1}}}},"estimated_seconds":{"type":"integer","minimum":1,"maximum":86400}}}}}}`)
}

// Digest binds the exact graph bytes for durable progress and revision keys.
func Digest(g Graph) (string, error) {
	raw, err := json.Marshal(g)
	if err != nil {
		return "", err
	}
	normal, err := canonicalNormalize(raw)
	if err != nil {
		return "", err
	}
	return hashDomain("harness.engineering-graph.v1", normal)
}

// ValidateAutonomousGraph enforces the autonomous execution contract on top
// of Validate: parent/dependency cycles and write ownership plus exactly one
// initial implementation with concrete WritePaths, research/design
// dependencies, and native verification/review gates. ModeDirect avoids graph
// branches for simple tasks with a single implementation.
func ValidateAutonomousGraph(g Graph) error {
	return ValidateAutonomousGraphWithImplementations(g, 1)
}

// ValidateAutonomousGraphWithImplementations validates an autonomous graph
// whose immutable execution policy permits at most maxInitial initial
// implementation tasks. The compatibility validator above retains the
// historical one-implementation contract.
func ValidateAutonomousGraphWithImplementations(g Graph, maxInitial int) error {
	if maxInitial < 1 || maxInitial > 2 {
		return errors.New("initial implementation limit must be 1 or 2")
	}
	if err := g.Validate(); err != nil {
		return err
	}
	byID := map[string]Task{}
	for _, t := range g.Tasks {
		byID[t.ID] = t
	}
	if g.Mode == ModeDirect {
		if len(g.Tasks) != 1 {
			return errors.New("direct plan must contain exactly one task")
		}
		only := g.Tasks[0]
		if only.Kind != Implementation {
			return errors.New("direct plan must be a single implementation")
		}
		if len(only.WritePaths) == 0 {
			return errors.New("direct implementation requires concrete write paths")
		}
		if len(only.Dependencies) != 0 {
			return errors.New("direct implementation must not carry dependencies")
		}
		return nil
	}
	impls := []Task{}
	for _, t := range g.Tasks {
		if t.Kind == Implementation {
			impls = append(impls, t)
		}
	}
	if len(impls) == 0 || len(impls) > maxInitial {
		return fmt.Errorf("graph must contain one to %d initial implementations, found %d", maxInitial, len(impls))
	}
	for i, impl := range impls {
		if len(impl.WritePaths) == 0 {
			return fmt.Errorf("implementation %q requires concrete declared write paths", impl.ID)
		}
		if len(impl.Dependencies) == 0 {
			for _, t := range g.Tasks {
				if t.Kind == Research || t.Kind == Design {
					return errors.New("implementation requires research/design dependencies")
				}
			}
		}
		for _, dep := range impl.Dependencies {
			d := byID[dep]
			if d.Kind != Research && d.Kind != Design {
				return fmt.Errorf("implementation dependency %q must be research or design", dep)
			}
		}
		for _, other := range impls[i+1:] {
			if dependsOn(impl.ID, other.ID, byID) || dependsOn(other.ID, impl.ID, byID) {
				return errors.New("initial implementation tasks must be independent")
			}
			for _, a := range impl.WritePaths {
				for _, b := range other.WritePaths {
					if pathsOverlapFold(a, b) {
						return fmt.Errorf("initial implementation write paths overlap: %q and %q", a, b)
					}
				}
			}
		}
	}
	hasVerification, hasReview := false, false
	verificationCount, reviewCount := 0, 0
	for _, t := range g.Tasks {
		switch t.Kind {
		case Research, Design:
			if len(t.WritePaths) != 0 {
				return fmt.Errorf("research/design task %q must not declare writes", t.ID)
			}
		case Verification:
			verificationCount++
			hasVerification = true
			if len(t.WritePaths) != 0 {
				return fmt.Errorf("verification task %q must not declare writes", t.ID)
			}
			for _, impl := range impls {
				if !dependsOn(t.ID, impl.ID, byID) {
					return fmt.Errorf("verification task %q must depend on implementation %q", t.ID, impl.ID)
				}
			}
		case Review:
			reviewCount++
			hasReview = true
			if len(t.WritePaths) != 0 {
				return fmt.Errorf("review task %q must not declare writes", t.ID)
			}
			for _, impl := range impls {
				if !dependsOn(t.ID, impl.ID, byID) {
					return fmt.Errorf("review task %q must depend on implementation %q", t.ID, impl.ID)
				}
			}
		}
	}
	if !hasVerification {
		return errors.New("graph requires a native verification gate")
	}
	if !hasReview {
		return errors.New("graph requires a native review gate")
	}
	if verificationCount != 1 || reviewCount != 1 {
		return errors.New("initial autonomous graph requires exactly one verification and one review gate")
	}
	return nil
}

func pathsOverlapFold(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	return pathsOverlap(a, b)
}

// ValidateAutonomousRevision extends ValidateRevision for runner-owned
// progress: never-started nodes may be changed or removed, but started,
// completed or UNKNOWN nodes keep identity and attempt history. Completed
// results are preserved as history; UNKNOWN blocks any retry or revision of
// that node.
func ValidateAutonomousRevision(previous, next Graph) error {
	if err := ValidateRevision(previous, next); err != nil {
		return err
	}
	prevByID := map[string]Task{}
	for _, t := range previous.Tasks {
		prevByID[t.ID] = t
	}
	for _, fresh := range next.Tasks {
		old, exists := prevByID[fresh.ID]
		if (!exists || (!old.Completed && len(old.Attempts) == 0)) && (fresh.Completed || len(fresh.Attempts) != 0) {
			return fmt.Errorf("revision invents progress for task %q", fresh.ID)
		}
	}
	for _, old := range previous.Tasks {
		fresh, ok := next.Task(old.ID)
		if !ok {
			if len(old.Attempts) != 0 || old.Completed {
				return fmt.Errorf("started task %q cannot be removed", old.ID)
			}
			continue
		}
		started := old.Completed || len(old.Attempts) != 0
		if !started {
			continue
		}
		if fresh.Completed != old.Completed || len(fresh.Attempts) != len(old.Attempts) {
			return fmt.Errorf("revision changes progress of task %q", old.ID)
		}
		for i := range old.Attempts {
			if fresh.Attempts[i] != old.Attempts[i] {
				return fmt.Errorf("revision changes attempt history of task %q", old.ID)
			}
		}
		if fresh.Kind != old.Kind || fresh.ParentID != old.ParentID || fresh.Title != old.Title ||
			!stringListEqual(fresh.Dependencies, old.Dependencies) ||
			!stringListEqual(fresh.ScopePaths, old.ScopePaths) ||
			!stringListEqual(fresh.WritePaths, old.WritePaths) ||
			!evidenceEqual(fresh.ExpectedEvidence, old.ExpectedEvidence) ||
			fresh.EstimatedSeconds != old.EstimatedSeconds {
			return fmt.Errorf("started task %q is immutable", old.ID)
		}
	}
	if err := next.Validate(); err != nil {
		return err
	}
	return nil
}

// RepairExtension appends scoped repair implementation/verification/review
// nodes carrying the actual failed evidence. It never rewrites completed
// history; callers persist the result as a validated revision.
func RepairExtension(g Graph, failedTaskID, failedEvidence string, seq int) (Graph, error) {
	return repairExtension(g, failedTaskID, failedEvidence, seq, false)
}

// RepairDesignExtension inserts a bounded read-only design task before a
// repair implementation. The implementation starts with no write paths;
// callers may refine only those paths after recording the exact design result.
func RepairDesignExtension(g Graph, failedTaskID, failedEvidence string, seq int) (Graph, error) {
	return repairExtension(g, failedTaskID, failedEvidence, seq, true)
}

func repairExtension(g Graph, failedTaskID, failedEvidence string, seq int, withDesign bool) (Graph, error) {
	if err := g.Validate(); err != nil {
		return Graph{}, err
	}
	base, ok := g.Task(failedTaskID)
	if !ok {
		return Graph{}, fmt.Errorf("repair target %q missing", failedTaskID)
	}
	if len(base.Attempts) == 0 || base.Attempts[len(base.Attempts)-1].Outcome != AttemptFailed {
		return Graph{}, fmt.Errorf("repair target %q has no failed evidence", failedTaskID)
	}
	next := g
	mkID := func(prefix string) string {
		candidate := fmt.Sprintf("%s-repair-%d", prefix, seq)
		if len(candidate) > 64 {
			candidate = candidate[:64]
		}
		candidate = strings.ToLower(strings.ReplaceAll(candidate, "_", "-"))
		return candidate
	}
	implID := mkID("impl")
	designID := mkID("design")
	verifyID := mkID("verify")
	reviewID := mkID("review")
	// Planner IDs are not reserved. Pick a deterministic unused trio for
	// this repair attempt without changing its bounded slot identity.
	for suffix := 0; ; suffix++ {
		_, implExists := g.Task(implID)
		_, designExists := g.Task(designID)
		_, verifyExists := g.Task(verifyID)
		_, reviewExists := g.Task(reviewID)
		if !implExists && (!withDesign || !designExists) && !verifyExists && !reviewExists {
			break
		}
		implID = fmt.Sprintf("impl-repair-%d-%d", seq, suffix+1)
		designID = fmt.Sprintf("design-repair-%d-%d", seq, suffix+1)
		verifyID = fmt.Sprintf("verify-repair-%d-%d", seq, suffix+1)
		reviewID = fmt.Sprintf("review-repair-%d-%d", seq, suffix+1)
	}
	ids := []string{implID, verifyID, reviewID}
	if withDesign {
		ids = append(ids, designID)
	}
	for _, id := range ids {
		if _, exists := g.Task(id); exists {
			return Graph{}, fmt.Errorf("repair task %q already exists", id)
		}
		if !idPattern.MatchString(id) {
			return Graph{}, fmt.Errorf("repair task id %q invalid", id)
		}
	}
	var scope, writes []string
	for _, t := range g.Tasks {
		if t.Kind == Implementation && len(t.WritePaths) != 0 {
			scope = append([]string(nil), t.ScopePaths...)
			writes = append([]string(nil), t.WritePaths...)
			break
		}
	}
	if len(scope) == 0 {
		scope = []string{"src"}
		writes = []string{"src/repair"}
	}
	if strings.TrimSpace(failedEvidence) == "" || len(failedEvidence) > 256 {
		return Graph{}, errors.New("repair requires bounded recorded failure evidence")
	}
	// Repair must become ready after failure: depend on the failed task's
	// completed dependencies (which were satisfied when it was attempted),
	// not on the failed task itself (which is not Completed and would block
	// Ready forever).
	repairDeps := append([]string(nil), base.Dependencies...)
	if len(repairDeps) == 0 {
		// Fall back to the first completed implementation/research scope so
		// the repair remains ordered but runnable.
		for _, t := range g.Tasks {
			if t.Completed && (t.Kind == Implementation || t.Kind == Research || t.Kind == Design) {
				repairDeps = []string{t.ID}
				break
			}
		}
	}
	implDeps := repairDeps
	if withDesign {
		implDeps = []string{designID}
		design := Task{ID: designID, Kind: Design, ParentID: failedTaskID, Title: "Design repair after " + failedTaskID, Dependencies: repairDeps, ScopePaths: scope, ExpectedEvidence: []Evidence{{Kind: "failure", Description: failedEvidence}, {Kind: "repair-attempt", Description: fmt.Sprintf("%d", seq)}, {Kind: "paths", Description: "candidate-bound repair paths within the original implementation scope"}}, EstimatedSeconds: 180}
		next.Tasks = append(next.Tasks, design)
		writes = nil
	}
	next.Tasks = append(next.Tasks,
		Task{ID: implID, Kind: Implementation, ParentID: failedTaskID, Title: "Repair after " + failedTaskID, Dependencies: implDeps, ScopePaths: scope, WritePaths: writes, ExpectedEvidence: []Evidence{{Kind: "file", Description: "repaired source for " + failedTaskID}, {Kind: "failure", Description: failedEvidence}, {Kind: "repair-attempt", Description: fmt.Sprintf("%d", seq)}}, EstimatedSeconds: 600},
		Task{ID: verifyID, Kind: Verification, Title: "Verify repair " + implID, Dependencies: []string{implID}, ScopePaths: scope, ExpectedEvidence: []Evidence{{Kind: "test", Description: "fresh verification for " + implID}}, EstimatedSeconds: 300},
		Task{ID: reviewID, Kind: Review, Title: "Review repair " + implID, Dependencies: []string{verifyID}, ScopePaths: scope, ExpectedEvidence: []Evidence{{Kind: "review", Description: "fresh review for " + implID}}, EstimatedSeconds: 300},
	)
	if err := next.Validate(); err != nil {
		return Graph{}, err
	}
	return next, nil
}

// RefineRepairWritePaths fills the previously-empty write list of one
// never-started repair implementation. Paths must remain within the original
// implementation ceiling supplied by the runner.
func RefineRepairWritePaths(g Graph, implementationID string, writePaths, originalScope []string) (Graph, error) {
	if err := g.Validate(); err != nil {
		return Graph{}, err
	}
	for _, task := range g.Tasks {
		if len(task.Attempts) > 0 && task.Attempts[len(task.Attempts)-1].Outcome == AttemptUnknown {
			return Graph{}, errors.New("repair write refinement blocked by unresolved UNKNOWN task")
		}
	}
	if err := validatePaths(writePaths, false); err != nil || len(writePaths) == 0 {
		return Graph{}, errors.New("repair design requires concrete unique write paths")
	}
	if err := validatePaths(originalScope, true); err != nil || len(originalScope) == 0 {
		return Graph{}, errors.New("invalid original repair scope")
	}
	var target *Task
	for i := range g.Tasks {
		if g.Tasks[i].ID == implementationID {
			target = &g.Tasks[i]
			break
		}
	}
	if target == nil || target.Kind != Implementation || target.ParentID == "" || len(target.WritePaths) != 0 || target.Completed || len(target.Attempts) != 0 {
		return Graph{}, errors.New("repair implementation is not an unstarted design slot")
	}
	if !stringListEqual(target.ScopePaths, originalScope) {
		return Graph{}, errors.New("repair implementation changed its original scope ceiling")
	}
	designFound := false
	for _, dep := range target.Dependencies {
		design, ok := g.Task(dep)
		if !ok || design.Kind != Design || !design.Completed || len(design.Attempts) == 0 || design.Attempts[len(design.Attempts)-1].Outcome != AttemptCompleted {
			continue
		}
		for _, expected := range target.ExpectedEvidence {
			if expected.Kind == "repair-attempt" {
				for _, designExpected := range design.ExpectedEvidence {
					if designExpected.Kind == "repair-attempt" && designExpected.Description == expected.Description {
						designFound = true
					}
				}
			}
		}
	}
	if !designFound {
		return Graph{}, errors.New("repair implementation lacks completed matching design evidence")
	}
	for _, p := range writePaths {
		if !withinAny(p, originalScope) {
			return Graph{}, fmt.Errorf("repair path %q outside original implementation scope", p)
		}
	}
	next := g
	next.Tasks = append([]Task(nil), g.Tasks...)
	for i := range next.Tasks {
		if next.Tasks[i].ID == implementationID {
			next.Tasks[i].WritePaths = append([]string(nil), writePaths...)
		}
	}
	if err := next.Validate(); err != nil {
		return Graph{}, err
	}
	return next, nil
}

// Ready returns deterministic independent work whose dependencies completed.
func Ready(g Graph) ([]Task, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	byID := map[string]Task{}
	for _, t := range g.Tasks {
		byID[t.ID] = t
	}
	var out []Task
	for _, t := range g.Tasks {
		if t.Completed || hasUnknown(t) {
			continue
		}
		ok := true
		for _, d := range t.Dependencies {
			if !byID[d].Completed {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func hasUnknown(t Task) bool {
	return len(t.Attempts) > 0 && t.Attempts[len(t.Attempts)-1].Outcome == AttemptUnknown
}
