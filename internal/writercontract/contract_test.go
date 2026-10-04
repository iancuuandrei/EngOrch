package writercontract

import (
	"encoding/json"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"strings"
	"testing"
)

func TestSchemaAndDomainCardinality(t *testing.T) {
	var doc any
	if err := json.Unmarshal(Schema(), &doc); err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	if err := c.AddResource("urn:writer", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile("urn:writer")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 1, 64, 65} {
		changes := []any{}
		for k := 0; k < n; k++ {
			changes = append(changes, map[string]any{"path": "new.py", "before_hash": nil, "content_base64": "eA==", "executable": false})
		}
		payload := map[string]any{"candidate_id": strings.Repeat("a", 64), "changes": changes}
		valid := n >= 1 && n <= 64
		if (schema.Validate(payload) == nil) != valid {
			t.Fatalf("schema count %d", n)
		}
		if (ValidateCount(n) == nil) != valid {
			t.Fatalf("domain count %d", n)
		}
	}
}

func TestUTF8SchemaAndDomainCardinality(t *testing.T) {
	var doc any
	if err := json.Unmarshal(UTF8Schema(), &doc); err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	if err := c.AddResource("urn:writer", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile("urn:writer")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 1, 64, 65} {
		changes := []any{}
		for k := 0; k < n; k++ {
			changes = append(changes, map[string]any{"path": "new.py", "before_hash": nil, "content_utf8": "eA==", "executable": false})
		}
		payload := map[string]any{"candidate_id": strings.Repeat("a", 64), "changes": changes}
		valid := n >= 1 && n <= 64
		if (schema.Validate(payload) == nil) != valid {
			t.Fatalf("schema count %d", n)
		}
		if (ValidateCount(n) == nil) != valid {
			t.Fatalf("domain count %d", n)
		}
	}
}

func TestUTF8SchemaForCandidateBindsOnlyExactLowercaseDigest(t *testing.T) {
	candidate := strings.Repeat("a", 64)
	raw, err := UTF8SchemaForCandidate(candidate)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	if got := schema.Properties["candidate_id"].Enum; len(got) != 1 || got[0] != candidate {
		t.Fatalf("candidate enum mismatch: %v", got)
	}
	if !strings.Contains(string(UTF8Schema()), `"pattern":"^[0-9a-f]{64}$"`) {
		t.Fatal("legacy UTF-8 schema bytes/contract changed")
	}
	for _, invalid := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 63) + "G", strings.Repeat("A", 64)} {
		if _, err := UTF8SchemaForCandidate(invalid); err == nil {
			t.Fatalf("invalid candidate admitted: %q", invalid)
		}
	}
}

func TestAnchoredEditSchemasAreClosedAndCandidateBound(t *testing.T) {
	candidate := strings.Repeat("a", 64)
	schema, err := AnchoredEditsSchemaForCandidate(candidate)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		AdditionalProperties bool `json:"additionalProperties"`
		Properties           map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(schema, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.AdditionalProperties || len(decoded.Properties["candidate_id"].Enum) != 1 || decoded.Properties["candidate_id"].Enum[0] != candidate {
		t.Fatalf("candidate-bound anchored schema malformed: %s", schema)
	}
	if _, err := AnchoredEditsSchemaForCandidate(strings.Repeat("B", 64)); err == nil {
		t.Fatal("invalid candidate ID admitted")
	}
}

func TestAnchoredEditSchemaEncodesBothExplicitFileForms(t *testing.T) {
	schema := compileWriterSchema(t, "anchored-writer", AnchoredEditsSchema())
	digest := strings.Repeat("a", 64)
	for _, change := range anchoredExplicitChanges(digest) {
		payload := map[string]any{"candidate_id": digest, "changes": []any{change}}
		if err := schema.Validate(payload); err != nil {
			t.Fatalf("valid explicit change rejected: %v", err)
		}
	}
}

func TestStrictAnchoredEditSchemaRejectsNoOpExistingFiles(t *testing.T) {
	schema := compileWriterSchema(t, "strict-anchored-writer", StrictAnchoredEditsSchema())
	digest := strings.Repeat("a", 64)
	for _, change := range anchoredExplicitChanges(digest) {
		if err := schema.Validate(map[string]any{"candidate_id": digest, "changes": []any{change}}); err != nil {
			t.Fatalf("valid strict change rejected: %v", err)
		}
	}
	for _, change := range []map[string]any{
		{"path": "noop.go", "before_hash": digest, "edits": []any{}, "new_content_utf8": nil, "executable": false},
		{"path": "replacement.go", "before_hash": digest, "edits": []any{map[string]any{"before": "x", "after": "y"}}, "new_content_utf8": "unexpected", "executable": false},
		{"path": "new.go", "before_hash": nil, "edits": []any{map[string]any{"before": "x", "after": "y"}}, "new_content_utf8": "package new\n", "executable": false},
	} {
		if err := schema.Validate(map[string]any{"candidate_id": digest, "changes": []any{change}}); err == nil {
			t.Fatalf("invalid strict change admitted: %#v", change)
		}
	}
	bound, err := StrictAnchoredEditsSchemaForCandidate(digest)
	if err != nil || !strings.Contains(string(bound), `"enum":["`+digest+`"]`) {
		t.Fatalf("strict candidate schema was not bound: %v", err)
	}
}

func compileWriterSchema(t *testing.T, name string, raw []byte) *jsonschema.Schema {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	uri := "urn:" + name
	if err := compiler.AddResource(uri, doc); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(uri)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func anchoredExplicitChanges(digest string) []map[string]any {
	return []map[string]any{
		{"path": "existing.go", "before_hash": digest, "edits": []any{map[string]any{"before": "x", "after": "y"}}, "new_content_utf8": nil, "executable": false},
		{"path": "new.go", "before_hash": nil, "edits": []any{}, "new_content_utf8": "package new\n", "executable": false},
	}
}
