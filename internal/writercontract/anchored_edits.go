package writercontract

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AnchoredEditsV1 is an edit proposal, not a complete-file replacement.
// Existing files carry exact original anchors; new files require explicit text.
const ContractAnchoredEditsV1 = "anchored-edits-v1"

// ContractAnchoredEditsV2 keeps the v1 proposal bytes and enables same-turn
// candidate anchor validation for writer/fixer model invocations.
const ContractAnchoredEditsV2 = "anchored-edits-v2"

// IsAnchoredEdits reports contracts that share the frozen anchored proposal
// schema and controller-side composition rules.
func IsAnchoredEdits(contract string) bool {
	return contract == ContractAnchoredEditsV1 || contract == ContractAnchoredEditsV2
}

const (
	// AnchoredEditsMaxChanges bounds changed paths in one proposal.
	AnchoredEditsMaxChanges = 64
	// AnchoredEditAnchorMax bounds each UTF-8 anchor and replacement string.
	AnchoredEditAnchorMax = 16 << 10
	// AnchoredEditsMaxBytes bounds all before and after strings combined.
	AnchoredEditsMaxBytes = 64 << 10
	// AnchoredNewFileMax bounds explicit contents for new files.
	AnchoredNewFileMax = 256 << 10
)

// AnchoredEdit replaces one exact before anchor with its after text.
type AnchoredEdit struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

// AnchoredChange describes either localized edits to an existing file or
// explicit content for a new file; null content is never deletion.
type AnchoredChange struct {
	Path           string         `json:"path"`
	BeforeHash     *string        `json:"before_hash"`
	Edits          []AnchoredEdit `json:"edits"`
	NewContentUTF8 *string        `json:"new_content_utf8"`
	Executable     bool           `json:"executable"`
}

// AnchoredProposal binds a list of file changes to one candidate ID.
type AnchoredProposal struct {
	CandidateID string           `json:"candidate_id"`
	Changes     []AnchoredChange `json:"changes"`
}

// AnchoredEditsSchema describes the explicit anchored-edit-v1 wire shape.
func AnchoredEditsSchema() json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"type":"object","additionalProperties":false,"required":["candidate_id","changes"],"properties":{"candidate_id":{"type":"string","pattern":"^[0-9a-f]{64}$"},"changes":{"type":"array","minItems":1,"maxItems":%d,"items":{"type":"object","additionalProperties":false,"required":["path","before_hash","edits","new_content_utf8","executable"],"properties":{"path":{"type":"string","minLength":1},"before_hash":{"type":["string","null"],"pattern":"^[0-9a-f]{64}$"},"edits":{"type":"array","maxItems":64,"items":{"type":"object","additionalProperties":false,"required":["before","after"],"properties":{"before":{"type":"string","minLength":1,"maxLength":%d},"after":{"type":"string","maxLength":%d}}}},"new_content_utf8":{"type":["string","null"],"maxLength":%d},"executable":{"type":"boolean"}}}}}}`, AnchoredEditsMaxChanges, AnchoredEditAnchorMax, AnchoredEditAnchorMax, AnchoredNewFileMax))
}

// AnchoredEditsSchemaForCandidate constrains output to the exact candidate ID.
func AnchoredEditsSchemaForCandidate(candidateID string) (json.RawMessage, error) {
	return schemaForCandidate(AnchoredEditsSchema(), candidateID)
}

func schemaForCandidate(schema json.RawMessage, candidateID string) (json.RawMessage, error) {
	if !validCandidateID(candidateID) {
		return nil, fmt.Errorf("candidate ID must be 64 lowercase hexadecimal characters")
	}
	const pattern = `"candidate_id":{"type":"string","pattern":"^[0-9a-f]{64}$"}`
	replacement := `"candidate_id":{"type":"string","enum":["` + candidateID + `"]}`
	updated := strings.Replace(string(schema), pattern, replacement, 1)
	if updated == string(schema) {
		return nil, fmt.Errorf("candidate-bound schema template unavailable")
	}
	return json.RawMessage(updated), nil
}
