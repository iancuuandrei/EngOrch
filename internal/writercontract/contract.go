// Package writercontract defines the finite mutation-proposal output contract.
package writercontract

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// MaxChanges bounds the number of changes in one proposal.
const MaxChanges = 64

// ContractChangesJSONV1 is the R59 writer/fixer wire contract: the provider
// constructs one flat string (changes_json) carrying the JSON array, and the
// deterministic controller reconstructs the semantic proposal. The semantic
// contract stays WriterProposal{CandidateID, Changes}; only the wire
// representation differs from utf8-v2. There is deliberately no dual
// representation: v1 outputs are invalid under v2 and vice versa.
const ContractChangesJSONV1 = "changes-json-v1"

// ContractUTF8ReplaceV3 adds explicit complete-file replacement instructions
// while retaining the strict UTF-8 wire schema and decoder.
const ContractUTF8ReplaceV3 = "utf8-replace-v3"

// ContractUTF8ScopedV4 adds the current ready graph implementation task and
// exact declared paths to the writer input while retaining v3's wire shape.
const ContractUTF8ScopedV4 = "utf8-scoped-v4"

// ChangesJSONMaxLength bounds the outer string payload. The structured-output
// value bound (256 KiB) still governs the complete terminal value.
const ChangesJSONMaxLength = 500000

// UTF8Schema leaves encoding to the controller, not the language model.
func UTF8Schema() json.RawMessage {
	return json.RawMessage(strings.ReplaceAll(string(Schema()), "content_base64", "content_utf8"))
}

// UTF8SchemaForCandidate additionally constrains candidate_id to the exact
// invocation candidate. Controller validation remains authoritative.
func UTF8SchemaForCandidate(candidateID string) (json.RawMessage, error) {
	return schemaForCandidate(UTF8Schema(), candidateID)
}

func validCandidateID(candidateID string) bool {
	decoded, err := hex.DecodeString(candidateID)
	return err == nil && len(decoded) == 32 && strings.ToLower(candidateID) == candidateID
}

// ErrEmptyChangeset reports a proposal with no changes.
var ErrEmptyChangeset = errors.New("WRITER_PROPOSAL_INVALID: EMPTY_CHANGESET")

// ErrTooManyChanges reports a proposal exceeding MaxChanges.
var ErrTooManyChanges = errors.New("WRITER_PROPOSAL_INVALID: TOO_MANY_CHANGES")

// ValidateCount validates that n is within one to MaxChanges.
func ValidateCount(n int) error {
	if n == 0 {
		return fmt.Errorf("%w: one to 64 changes required", ErrEmptyChangeset)
	}
	if n < 0 || n > MaxChanges {
		return fmt.Errorf("%w: one to 64 changes required", ErrTooManyChanges)
	}
	return nil
}

// Schema returns fresh bytes so callers cannot mutate a shared contract.
// Filesystem authority, candidate identity and byte validation remain independent.
func Schema() json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"type":"object","additionalProperties":false,"required":["candidate_id","changes"],"properties":{"candidate_id":{"type":"string","pattern":"^[0-9a-f]{64}$"},"changes":{"type":"array","minItems":1,"maxItems":%d,"items":{"type":"object","additionalProperties":false,"required":["path","before_hash","content_base64","executable"],"properties":{"path":{"type":"string","minLength":1},"before_hash":{"type":["string","null"],"pattern":"^[0-9a-f]{64}$"},"content_base64":{"type":["string","null"]},"executable":{"type":"boolean"}}}}}}`, MaxChanges))
}

// ChangesJSONSchema returns the R59 v2 outer wire schema: candidate_id stays a
// directly verifiable binding and the complete change array travels as one
// JSON-encoded string. Semantic strictness (array shape, item schema, count,
// duplicates, unknown fields) is enforced deterministically after decoding,
// not by asking the model for canonical bytes.
func ChangesJSONSchema() json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"type":"object","additionalProperties":false,"required":["candidate_id","changes_json"],"properties":{"candidate_id":{"type":"string","pattern":"^[0-9a-f]{64}$"},"changes_json":{"type":"string","minLength":2,"maxLength":%d}}}`, ChangesJSONMaxLength))
}
