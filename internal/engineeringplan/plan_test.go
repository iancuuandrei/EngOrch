package engineeringplan

import (
	"encoding/json"
	"strings"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

func task(id string, seconds int) Task {
	return Task{ID: id, Kind: Implementation, Title: "Implement " + id, ScopePaths: []string{"src"}, WritePaths: []string{"src/" + id}, ExpectedEvidence: []Evidence{{Kind: "file", Description: "changed source"}}, EstimatedSeconds: seconds}
}

func graph(tasks ...Task) Graph {
	return Graph{Version: Version, Mode: ModeGraph, Summary: "bounded implementation", Tasks: tasks}
}

func TestParseJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	valid, err := json.Marshal(Graph{Version: Version, Mode: ModeDirect, Summary: "one task", Tasks: []Task{task("one", 3)}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseJSON(valid); err != nil {
		t.Fatalf("valid JSON rejected: %v", err)
	}
	for name, raw := range map[string]string{
		"unknown field": strings.TrimSuffix(string(valid), "}") + `,"extra":true}`,
		"second value":  string(valid) + ` {}`,
		"garbage":       string(valid) + ` trailing`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseJSON([]byte(raw)); err == nil {
				t.Fatal("expected strict parser rejection")
			}
		})
	}
}

func TestJSONSchemaBoundsAndStrictShape(t *testing.T) {
	var document any
	if err := json.Unmarshal(JSONSchema(), &document); err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	if err := c.AddResource("urn:engineering-plan", document); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile("urn:engineering-plan")
	if err != nil {
		t.Fatal(err)
	}
	valid := Graph{Version: Version, Mode: ModeDirect, Summary: "one task", Tasks: []Task{task("one", 3)}}
	var good any
	raw, _ := json.Marshal(valid)
	_ = json.Unmarshal(raw, &good)
	if err := schema.Validate(good); err != nil {
		t.Fatalf("valid plan violates schema: %v", err)
	}
	for name, mutate := range map[string]func(*Graph){
		"too many tasks": func(g *Graph) {
			g.Mode = ModeGraph
			for i := 0; i < 64; i++ {
				g.Tasks = append(g.Tasks, task("task"+strings.Repeat("x", i%2), 1))
			}
			g.Tasks = append(g.Tasks, task("last", 1))
		},
		"zero estimate": func(g *Graph) { g.Tasks[0].EstimatedSeconds = 0 },
		"unknown field": func(g *Graph) {},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			candidate.Tasks = append([]Task(nil), valid.Tasks...)
			mutate(&candidate)
			payload, _ := json.Marshal(candidate)
			if name == "unknown field" {
				payload = []byte(strings.TrimSuffix(string(payload), "}") + `,"unexpected":1}`)
			}
			var value any
			_ = json.Unmarshal(payload, &value)
			if schema.Validate(value) == nil {
				t.Fatal("expected schema rejection")
			}
		})
	}
}

func TestValidateRejectsCycleMissingDependencyAndWriteConflict(t *testing.T) {
	a, b := task("alpha", 1), task("beta", 1)
	a.Dependencies = []string{"beta"}
	b.Dependencies = []string{"alpha"}
	if err := graph(a, b).Validate(); err == nil {
		t.Fatal("cycle accepted")
	}
	a = task("alpha", 1)
	a.Dependencies = []string{"missing"}
	if err := graph(a).Validate(); err == nil {
		t.Fatal("missing dependency accepted")
	}
	a, b = task("alpha", 1), task("beta", 1)
	b.WritePaths = []string{"src/alpha/child"}
	if err := graph(a, b).Validate(); err == nil {
		t.Fatal("unordered overlapping writes accepted")
	}
	b.Dependencies = []string{"alpha"}
	if err := graph(a, b).Validate(); err != nil {
		t.Fatalf("ordered overlapping writes rejected: %v", err)
	}

	a = task("alpha", 1)
	a.WritePaths = []string{"other/file.go"}
	if err := graph(a).Validate(); err == nil {
		t.Fatal("write outside declared scope accepted")
	}
	a = task("alpha", 1)
	a.ScopePaths = []string{"C:/repo/src"}
	if err := graph(a).Validate(); err == nil {
		t.Fatal("Windows volume path accepted as repository-relative")
	}
}

func TestValidateRevisionPreservesCompletedTaskAndBlocksUnknownRetry(t *testing.T) {
	done := task("done", 10)
	done.Completed = true
	done.Attempts = []Attempt{{ID: "attempt-1", Outcome: AttemptCompleted}}
	old := graph(done, task("next", 2))
	next := graph(done, task("next", 2))
	next.Summary = "updated plan summary"
	if err := ValidateRevision(old, next); err != nil {
		t.Fatalf("preserving revision rejected: %v", err)
	}
	changed := graph(done, task("next", 2))
	changed.Tasks[0].Title = "rewritten completed task"
	if err := ValidateRevision(old, changed); err == nil {
		t.Fatal("modified completed task accepted")
	}

	uncertain := task("uncertain", 1)
	uncertain.Attempts = []Attempt{{ID: "attempt-1", Outcome: AttemptUnknown}}
	old = graph(uncertain)
	if ready, err := Ready(old); err != nil || len(ready) != 0 {
		t.Fatalf("unknown attempt became ready: %v %v", ready, err)
	}
	retry := task("uncertain", 1)
	retry.Attempts = []Attempt{{ID: "attempt-1", Outcome: AttemptUnknown}, {ID: "attempt-2", Outcome: AttemptFailed}}
	if err := ValidateRevision(old, graph(retry)); err == nil {
		t.Fatal("retry after unknown outcome accepted")
	}
}
