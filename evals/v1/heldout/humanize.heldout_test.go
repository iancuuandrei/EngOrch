package humanize

import "testing"

func TestFabricV1Heldout(t *testing.T) {
	got, err := ParseBytes("1_024 B")
	if err != nil || got != 1024 {
		t.Fatalf("underscore integer: got %d, err %v", got, err)
	}
	for _, in := range []string{"1__024 B", "_1024 B", "1024_ B"} {
		if _, err := ParseBytes(in); err == nil {
			t.Fatalf("accepted malformed separators %q", in)
		}
	}
}
