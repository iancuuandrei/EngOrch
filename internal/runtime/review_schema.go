package runtime

import "encoding/json"

// ReviewOutputSchema describes candidate-bound judgment. An empty finding path
// denotes a cross-cutting concern; commands are never file-path evidence.
func ReviewOutputSchema() json.RawMessage {
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
	raw, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	return raw
}
