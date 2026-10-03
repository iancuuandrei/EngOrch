package engineeringplan

import (
	"strings"
	"testing"
)

func parallelImplementationGraph() Graph {
	return Graph{Version: Version, Mode: ModeGraph, Summary: "split implementation", Tasks: []Task{
		{ID: "research", Kind: Research, Title: "Inspect API", ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "fact", Description: "API"}}, EstimatedSeconds: 10},
		{ID: "impl-api", Kind: Implementation, Title: "Implement API", Dependencies: []string{"research"}, ScopePaths: []string{"."}, WritePaths: []string{"bool.go"}, ExpectedEvidence: []Evidence{{Kind: "file", Description: "methods"}}, EstimatedSeconds: 10},
		{ID: "impl-tests", Kind: Implementation, Title: "Add tests", Dependencies: []string{"research"}, ScopePaths: []string{"."}, WritePaths: []string{"bool_test.go"}, ExpectedEvidence: []Evidence{{Kind: "test", Description: "cases"}}, EstimatedSeconds: 10},
		{ID: "verify", Kind: Verification, Title: "Run tests", Dependencies: []string{"impl-api", "impl-tests"}, ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "check", Description: "go test"}}, EstimatedSeconds: 10},
		{ID: "review", Kind: Review, Title: "Review changes", Dependencies: []string{"verify"}, ScopePaths: []string{"."}, ExpectedEvidence: []Evidence{{Kind: "review", Description: "findings"}}, EstimatedSeconds: 10},
	}}
}

func TestAutonomousGraphImplementationLimitKeepsLegacyAndAllowsIndependentPair(t *testing.T) {
	g := parallelImplementationGraph()
	if err := ValidateAutonomousGraph(g); err == nil {
		t.Fatal("legacy validator admitted multiple implementation tasks")
	}
	if err := ValidateAutonomousGraphWithImplementations(g, 2); err != nil {
		t.Fatal("opt-in validator rejected independent implementations:", err)
	}
	g.Tasks = append(g.Tasks[:2], g.Tasks[3:]...)
	g.Tasks[2].Dependencies = []string{"impl-api"}
	if err := ValidateAutonomousGraphWithImplementations(g, 2); err != nil {
		t.Fatal("one-writer serial baseline was not admitted:", err)
	}
	if err := ValidateAutonomousGraphWithImplementations(parallelImplementationGraph(), 1); err == nil {
		t.Fatal("one-writer validator admitted more than one initial implementation")
	}
	if err := ValidateAutonomousGraphWithImplementations(parallelImplementationGraph(), 3); err == nil {
		t.Fatal("unsupported implementation limit accepted")
	}
}

func TestAutonomousGraphParallelImplementationsRejectCollisionsAndMissingGate(t *testing.T) {
	for _, paths := range [][]string{
		{"bool.go", "bool.go"},
		{"internal", "internal/api.go"},
		{"Bool.go", "bool.go"},
		{"pkg/Foo.go", "pkg/foo.go"},
	} {
		g := parallelImplementationGraph()
		g.Tasks[1].WritePaths = []string{paths[0]}
		g.Tasks[2].WritePaths = []string{paths[1]}
		if err := ValidateAutonomousGraphWithImplementations(g, 2); err == nil {
			t.Fatalf("write collision accepted: %q and %q", paths[0], paths[1])
		}
	}
	g := parallelImplementationGraph()
	g.Tasks[3].Dependencies = []string{"impl-api"}
	g.Tasks[4].Dependencies = []string{"verify"}
	if err := ValidateAutonomousGraphWithImplementations(g, 2); err == nil || !strings.Contains(err.Error(), "impl-tests") {
		t.Fatal("native gate not required to depend on every implementation", err)
	}
	g = parallelImplementationGraph()
	g.Tasks[2].Dependencies = []string{"research", "impl-api"}
	if err := ValidateAutonomousGraphWithImplementations(g, 2); err == nil || !strings.Contains(err.Error(), "independent") {
		t.Fatal("dependent implementation siblings accepted", err)
	}
}

func TestAutonomousGraphIsolatedValidatorAllowsUpToEightWithoutWideningLegacy(t *testing.T) {
	g := parallelImplementationGraph()
	g.Tasks = append(g.Tasks, Task{ID: "impl-docs", Kind: Implementation, Title: "Update docs", Dependencies: []string{"research"}, ScopePaths: []string{"."}, WritePaths: []string{"README.md"}, ExpectedEvidence: []Evidence{{Kind: "file", Description: "docs"}}, EstimatedSeconds: 10})
	for i := range g.Tasks {
		if g.Tasks[i].ID == "verify" || g.Tasks[i].ID == "review" {
			g.Tasks[i].Dependencies = append(g.Tasks[i].Dependencies, "impl-docs")
		}
	}
	if err := ValidateAutonomousGraphWithIsolatedImplementations(g, 3); err != nil {
		t.Fatal("isolated validator rejected three independent implementations:", err)
	}
	if err := ValidateAutonomousGraphWithImplementations(g, 3); err == nil {
		t.Fatal("legacy parallel validator was widened beyond two implementations")
	}
	if err := ValidateAutonomousGraphWithIsolatedImplementations(g, 9); err == nil {
		t.Fatal("isolated validator accepted a limit above eight")
	}
}
