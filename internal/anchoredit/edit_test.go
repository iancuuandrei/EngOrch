package anchoredit

import (
	"strings"
	"testing"
)

func TestApplyPreservesUnmatchedBytesAndUsesOriginalOffsets(t *testing.T) {
	got, err := Apply([]byte("left A middle B right\n"), []Edit{
		{Before: "B", After: "second"},
		{Before: "A", After: "first"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "left first middle second right\n" {
		t.Fatalf("unexpected composition: %q", got)
	}
}

func TestApplyRejectsMissingRepeatedOverlappingAndOversized(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		edits  []Edit
	}{
		{name: "missing", source: "abc", edits: []Edit{{Before: "z", After: "x"}}},
		{name: "repeated", source: "abc abc", edits: []Edit{{Before: "abc", After: "x"}}},
		{name: "overlap", source: "abcdef", edits: []Edit{{Before: "abcd", After: "x"}, {Before: "cdef", After: "y"}}},
		{name: "overlapping repeated anchor", source: "aaa", edits: []Edit{{Before: "aa", After: "x"}}},
		{name: "empty anchor", source: "abc", edits: []Edit{{Before: "", After: "x"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Apply([]byte(tc.source), tc.edits); err == nil {
				t.Fatal("invalid anchored edit admitted")
			}
		})
	}
}

func TestApplyPreservesUnicodeAndEnforcesPayloadAndResultBounds(t *testing.T) {
	source := []byte("package p\n// păstrează această linie\nvar value = 1\n")
	got, err := Apply(source, []Edit{{Before: "var value = 1", After: "var value = 2"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "package p\n// păstrează această linie\nvar value = 2\n" {
		t.Fatalf("Unicode or untouched bytes changed: %q", got)
	}
	if _, err := Apply([]byte("anchor"), []Edit{{Before: "anchor", After: strings.Repeat("x", MaxEditBytes)}}); err == nil {
		t.Fatal("oversized edit payload admitted")
	}
	full := strings.Repeat("a", MaxSourceBytes-1) + "z"
	if _, err := Apply([]byte(full), []Edit{{Before: "z", After: "zz"}}); err == nil {
		t.Fatal("oversized composed result admitted")
	}
}
