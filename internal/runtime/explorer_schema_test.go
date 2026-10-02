package runtime

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestExplorerOutputSchemaForCandidateBindsExactCandidate(t *testing.T) {
	const legacySchema = `{"type":"object","additionalProperties":false,"required":["candidate_id","summary","paths"],"properties":{"candidate_id":{"type":"string"},"summary":{"type":"string"},"paths":{"type":"array","items":{"type":"string"}}}}`
	if string(ExplorerOutputSchema()) != legacySchema {
		t.Fatal("json-v1 explorer schema bytes changed")
	}
	candidateID := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	raw, err := ExplorerOutputSchemaForCandidate(candidateID)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Type                 string                     `json:"type"`
		AdditionalProperties bool                       `json:"additionalProperties"`
		Required             []string                   `json:"required"`
		Properties           map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	var candidate struct {
		Type string   `json:"type"`
		Enum []string `json:"enum"`
	}
	if err := json.Unmarshal(schema.Properties["candidate_id"], &candidate); err != nil {
		t.Fatal(err)
	}
	if schema.Type != "object" || schema.AdditionalProperties || !reflect.DeepEqual(schema.Required, []string{"candidate_id", "summary", "paths"}) || candidate.Type != "string" || !reflect.DeepEqual(candidate.Enum, []string{candidateID}) {
		t.Fatalf("candidate-bound explorer schema malformed: %s", raw)
	}
	if !json.Valid(ExplorerOutputSchema()) {
		t.Fatal("legacy explorer schema changed into invalid JSON")
	}
	for _, invalid := range []string{"", "0123", candidateID[:63], candidateID[:63] + "G", "0123456789ABCDEF0123456789abcdef0123456789abcdef0123456789abcdef"} {
		if _, err := ExplorerOutputSchemaForCandidate(invalid); err == nil {
			t.Fatalf("invalid candidate ID admitted: %q", invalid)
		}
	}
}
