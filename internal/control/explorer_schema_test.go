package control

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"harness.local/engorch/internal/runtime"
)

func TestExplorerSchemaOptInPreservesLegacyInvocationAndRejectsObjectSummary(t *testing.T) {
	c := creation(t)
	c.Config.Explorer = &runtime.Profile{Runtime: "fake", Provider: "deterministic", Model: "explorer", Effort: "low", Role: "explorer"}
	path, _ := approvedRepositoryCreation(t, c)
	s, err := StartWorkspace(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := explorerInvocation(s, "Inspect source")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(legacy.Input, "output_schema") || strings.Contains(legacy.Input, "summary MUST") {
		t.Fatal("legacy invocation changed")
	}
	strict := s
	strict.Creation.Config.ExplorerContract = "json-v1"
	i, err := explorerInvocation(strict, "Inspect source")
	if err != nil || i.ID == legacy.ID || !strings.Contains(i.Input, "summary MUST") {
		t.Fatal("strict schema not bound into invocation", err)
	}
	var envelope struct {
		Schema json.RawMessage `json:"output_schema"`
	}
	if err := json.Unmarshal([]byte(i.Input), &envelope); err != nil || len(envelope.Schema) == 0 {
		t.Fatal("schema absent", err)
	}
	model := legacy.Profile.Model
	id, _ := s.Candidate.ID()
	r := ExplorerRecord{"Inspect source", legacy, runtime.Result{Version: 1, InvocationID: legacy.ID, Requested: legacy.Profile, ObservedModel: &model,
		Output: `{"candidate_id":"` + id + `","summary":{"observation":"source read"},"paths":[]}`}}
	if _, err := RecordExploration(path, r); err == nil {
		t.Fatal("object summary was silently admitted")
	}
	after, err := Inspect(path)
	if err != nil || len(after.Explorations) != 0 {
		t.Fatal("invalid result changed semantic history", err)
	}
}
