package engineeringplan

import (
	"encoding/json"
	"testing"
)

func TestRepositoryRootScopeKeepsConcreteWrites(t *testing.T) {
	impl := task("impl", 1)
	impl.ScopePaths = []string{"."}
	verify := task("verify", 1)
	verify.Kind, verify.WritePaths, verify.ScopePaths = Verification, nil, []string{"."}
	verify.Dependencies = []string{"impl"}
	review := task("review", 1)
	review.Kind, review.WritePaths, review.ScopePaths = Review, nil, []string{"."}
	review.Dependencies = []string{"verify"}
	g := graph(impl, verify, review)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParsePlannerGraph(data)
	if err != nil {
		t.Fatal("root scope rejected", err)
	}
	if err := ValidateAutonomousGraph(parsed); err != nil {
		t.Fatal(err)
	}
	g.Tasks[0].WritePaths = []string{"."}
	if err := g.Validate(); err == nil {
		t.Fatal("repository root admitted as a write path")
	}
}

func TestRootScopePreservesPathSafety(t *testing.T) {
	for _, unsafe := range []string{"..", "../escape", "/absolute", "C:/absolute", "a/../b", `a\b`} {
		for _, allowRoot := range []bool{true, false} {
			if err := validatePaths([]string{unsafe}, allowRoot); err == nil {
				t.Fatalf("unsafe path accepted: %q", unsafe)
			}
		}
	}
	if err := validatePaths([]string{".", "."}, true); err == nil {
		t.Fatal("duplicate root scope accepted")
	}
}
