package canonical

import (
	"encoding/json"
	"strings"
	"testing"
)

type typedGeneratedChild struct {
	Name string `json:"name"`
	N    int    `json:"n,omitempty"`
}

type typedGeneratedValue struct {
	First  string                `json:"first"`
	Second string                `json:"second"`
	Values []typedGeneratedChild `json:"values,omitempty"`
	Map    map[string]string     `json:"map,omitempty"`
	Ptr    *typedGeneratedChild  `json:"ptr,omitempty"`
}

type typedGeneratedHook struct{}
type typedGeneratedNamedBytes []byte

func (typedGeneratedHook) MarshalJSON() ([]byte, error) { return []byte(`{"b":1,"b":2}`), nil }

func TestTypedGeneratedBytesParity(t *testing.T) {
	values := []any{
		nil,
		typedGeneratedValue{First: "<&>\\u2028\u2029", Second: "reversed", Values: []typedGeneratedChild{{Name: "x", N: 7}}, Map: map[string]string{"z": "last", "a": "first"}},
		typedGeneratedValue{First: "nil", Second: "shape", Values: []typedGeneratedChild{}, Map: map[string]string{}, Ptr: nil},
		typedGeneratedHook{},
		json.RawMessage(`{"x":1}`),
		[]byte{0, 1, 2},
		typedGeneratedNamedBytes{3, 4, 5},
		map[int]string{1: "invalid key"},
		struct {
			Number float64 `json:"number"`
		}{1},
	}
	for _, value := range values {
		want, wantErr := Bytes(value)
		got, gotErr := TypedGeneratedBytes(value)
		if (wantErr != nil) != (gotErr != nil) || wantErr == nil && string(want) != string(got) {
			t.Fatalf("typed parity mismatch for %T", value)
		}
		if wantErr == nil {
			wantHash, err := Hash("harness.typed.test.v1", value)
			if err != nil {
				t.Fatal(err)
			}
			gotHash, err := TypedGeneratedHash("harness.typed.test.v1", value)
			if err != nil || gotHash != wantHash {
				t.Fatalf("typed hash mismatch for %T", value)
			}
		}
	}
}

func TestTypedGeneratedBytesFallbackClosedShapeParity(t *testing.T) {
	values := []any{
		struct {
			Bytes []byte `json:"bytes"`
		}{Bytes: []byte{1, 2}},
		struct {
			Bytes typedGeneratedNamedBytes `json:"bytes"`
		}{Bytes: typedGeneratedNamedBytes{3, 4}},
		struct {
			Value any `json:"value"`
		}{Value: "interface payload"},
		struct {
			Value map[string]any `json:"value"`
		}{Value: map[string]any{"ok": int64(1)}},
	}
	for _, value := range values {
		want, wantErr := Bytes(value)
		got, gotErr := TypedGeneratedBytes(value)
		if (wantErr != nil) != (gotErr != nil) || wantErr == nil && string(want) != string(got) {
			t.Fatalf("closed-shape fallback parity mismatch for %T", value)
		}
	}
}

func TestTypedGeneratedBytesRetainsRawAndDepthBounds(t *testing.T) {
	tooLarge := typedGeneratedValue{First: strings.Repeat("x", MaxBytes), Second: "bound"}
	want, wantErr := Bytes(tooLarge)
	got, gotErr := TypedGeneratedBytes(tooLarge)
	if wantErr == nil || gotErr == nil || string(want) != string(got) {
		t.Fatal("raw byte bound diverged")
	}
	value := any("leaf")
	for i := 0; i < 66; i++ {
		value = []any{value}
	}
	want, wantErr = Bytes(value)
	got, gotErr = TypedGeneratedBytes(value)
	if wantErr == nil || gotErr == nil || string(want) != string(got) {
		t.Fatal("depth bound diverged")
	}
}

func TestTypedGeneratedBytesBoundedAdmitsOnlyExplicitLargerClosedValues(t *testing.T) {
	value := typedGeneratedValue{First: strings.Repeat("x", MaxBytes), Second: "bounded"}
	if _, err := TypedGeneratedBytes(value); err == nil {
		t.Fatal("default typed generated boundary accepted oversized value")
	}
	encoded, err := TypedGeneratedBytesBounded(value, 2<<20)
	if err != nil || len(encoded) <= MaxBytes {
		t.Fatalf("explicit bounded encoding = %d bytes, %v", len(encoded), err)
	}
	if _, err := TypedGeneratedHashBounded("harness.typed.bounded.v1", value, MaxBytes); err == nil {
		t.Fatal("default-size bounded hash accepted oversized value")
	}
	if _, err := TypedGeneratedHashBounded("harness.typed.bounded.v1", value, 2<<20); err != nil {
		t.Fatal(err)
	}
	overCap := typedGeneratedValue{First: strings.Repeat("x", MaxTypedGeneratedBytes), Second: "over-cap"}
	if _, err := TypedGeneratedBytesBounded(overCap, MaxTypedGeneratedBytes); err == nil {
		t.Fatal("explicit typed generated cap accepted an oversized value")
	}
}
