package engineeringplan

import "testing"

func TestSelectCriticalPathChoosesDownstreamWorkFirst(t *testing.T) {
	shortA, shortB, long := task("alpha", 1), task("bravo", 1), task("zulu", 10)
	child := task("zulu-child", 10)
	child.Dependencies = []string{"zulu"}
	g := graph(shortA, shortB, long, child)
	lexical, err := Select(g, 1, Lexical)
	if err != nil {
		t.Fatal(err)
	}
	critical, err := Select(g, 1, CriticalPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(lexical) != 1 || lexical[0].ID != "alpha" {
		t.Fatalf("lexical selection: %+v", lexical)
	}
	if len(critical) != 1 || critical[0].ID != "zulu" {
		t.Fatalf("critical-path selection: %+v", critical)
	}
}

func TestSimulateMakespanCriticalPathBeatsDeclarationBaseline(t *testing.T) {
	shortA, shortB, long := task("alpha", 1), task("bravo", 1), task("zulu", 10)
	child := task("zulu-child", 10)
	child.Dependencies = []string{"zulu"}
	// This is a synthetic two-worker policy comparison: lexical declaration
	// order is the baseline; it is not observed execution evidence.
	g := graph(shortA, shortB, long, child)
	baseline, err := SimulateMakespan(g, 2, Lexical)
	if err != nil {
		t.Fatal(err)
	}
	critical, err := SimulateMakespan(g, 2, CriticalPath)
	if err != nil {
		t.Fatal(err)
	}
	if baseline != 21 || critical != 20 {
		t.Fatalf("synthetic makespans lexical=%d critical-path=%d, want 21 and 20", baseline, critical)
	}
}

func TestReadyDependencyAndUnknownSemantics(t *testing.T) {
	a := task("alpha", 1)
	b := task("beta", 1)
	b.Dependencies = []string{"alpha"}
	g := graph(a, b)
	ready, err := Ready(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 || ready[0].ID != "alpha" {
		t.Fatalf("initial ready set: %+v", ready)
	}
	g.Tasks[0].Completed = true
	ready, err = Ready(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 || ready[0].ID != "beta" {
		t.Fatalf("dependent ready set: %+v", ready)
	}
}
