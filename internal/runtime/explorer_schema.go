package runtime

import "encoding/json"

// ExplorerOutputSchema declares the advisory explorer result's existing shape.
// Candidate binding and sorted paths are still checked by controller replay.
func ExplorerOutputSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["candidate_id","summary","paths"],"properties":{"candidate_id":{"type":"string"},"summary":{"type":"string"},"paths":{"type":"array","items":{"type":"string"}}}}`)
}
