package atomic

import (
	"encoding"
	"testing"
)

func TestFabricV1Heldout(t *testing.T) {
	var _ encoding.TextMarshaler = (*Bool)(nil)
	var _ encoding.TextUnmarshaler = (*Bool)(nil)
	b := NewBool(true)
	got, err := b.MarshalText()
	if err != nil || string(got) != "true" {
		t.Fatalf("MarshalText = %q, %v", got, err)
	}
	if err := b.UnmarshalText([]byte("false")); err != nil || b.Load() {
		t.Fatalf("UnmarshalText false: %v, value %v", err, b.Load())
	}
	if err := b.UnmarshalText([]byte("not-bool")); err == nil || b.Load() {
		t.Fatalf("invalid text must fail without changing value, err=%v value=%v", err, b.Load())
	}
}
