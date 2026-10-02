package runtime

import (
	"encoding/json"
	"errors"
	"strings"

	"harness.local/engorch/internal/canonical"
)

// ExplorerOutputSchema declares the advisory explorer result's existing shape.
// Candidate binding and sorted paths are still checked by controller replay.
func ExplorerOutputSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["candidate_id","summary","paths"],"properties":{"candidate_id":{"type":"string"},"summary":{"type":"string"},"paths":{"type":"array","items":{"type":"string"}}}}`)
}

// ExplorerOutputSchemaForCandidate binds json-v2's candidate_id field to the
// exact candidate already selected by the invocation. The controller still
// validates the returned value independently during replay.
func ExplorerOutputSchemaForCandidate(candidateID string) (json.RawMessage, error) {
	if len(candidateID) != 64 {
		return nil, errors.New("candidate ID must be 64 lowercase hexadecimal characters")
	}
	for _, c := range candidateID {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return nil, errors.New("candidate ID must be 64 lowercase hexadecimal characters")
		}
	}
	raw, err := canonical.Bytes(map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"candidate_id", "summary", "paths"},
		"properties": map[string]any{
			"candidate_id": map[string]any{"type": "string", "enum": []string{candidateID}},
			"summary":      map[string]any{"type": "string"},
			"paths":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}
