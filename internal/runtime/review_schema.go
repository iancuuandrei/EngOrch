package runtime

import "encoding/json"

// ReviewOutputSchema describes candidate-bound judgment. An empty finding path
// denotes a cross-cutting concern; commands are never file-path evidence.
func ReviewOutputSchema() json.RawMessage {
	return reviewOutputSchema(false)
}

// RepairReviewOutputSchema additionally requires explicit original concern IDs
// and bounded reviewer recheck judgments; it grants no native-check authority.
func RepairReviewOutputSchema() json.RawMessage {
	return reviewOutputSchema(true)
}

func reviewOutputSchema(rechecks bool) json.RawMessage {
	segment := `[^/\\:\x00-\x1f\x7f<>"|?*]*[^/\\:\x00-\x1f\x7f<>"|?* .]`
	schema := map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"candidate_id", "verification_plan_id", "decision", "findings"},
		"properties": map[string]any{
			"candidate_id":         map[string]any{"type": "string", "pattern": "^[a-f0-9]{64}$"},
			"verification_plan_id": map[string]any{"type": "string", "pattern": "^[a-f0-9]{64}$"},
			"decision":             map[string]any{"type": "string", "enum": []string{"approve", "changes_requested"}},
			"findings": map[string]any{"type": "array", "maxItems": 64, "items": map[string]any{
				"type": "object", "additionalProperties": false, "required": []string{"path", "message"},
				"properties": map[string]any{
					"path":    map[string]any{"type": "string", "maxLength": 1024, "pattern": "^(?:|" + segment + "(?:/" + segment + ")*)$"},
					"message": map[string]any{"type": "string", "minLength": 1, "maxLength": 4096},
				},
			}},
		},
	}
	if rechecks {
		schema["required"] = []string{"candidate_id", "verification_plan_id", "decision", "findings", "rechecks"}
		schema["properties"].(map[string]any)["rechecks"] = map[string]any{
			"type": "array", "maxItems": 64, "items": map[string]any{
				"type": "object", "additionalProperties": false, "required": []string{"finding_id", "decision", "rationale"},
				"properties": map[string]any{
					"finding_id": map[string]any{"type": "string", "pattern": "^[a-f0-9]{64}$"},
					"decision":   map[string]any{"type": "string", "enum": []string{"closed", "unresolved"}},
					"rationale":  map[string]any{"type": "string", "minLength": 1, "maxLength": 512},
				},
			},
		}
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	return raw
}
