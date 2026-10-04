package taskcontext

import (
	"strings"
	"testing"
)

// pprTieFiles returns two files that tie on legacy scores: both are path
// hints with no lexical differentiator under the test objective.
func pprTieFiles() []File {
	return []File{
		admitted("aaa.go", "package aaa\n"),
		admitted("zzz.go", "package zzz\n"),
	}
}

// pprTieInput binds the tie fixture to one scope with both files hinted.
func pprTieInput(files []File, ppr []string) Input {
	limits := DefaultLimits()
	limits.MaxFiles = 1
	return Input{
		Version:   version,
		Scope:     Scope{SourceID: strings.Repeat("1", 64)},
		Objective: "update wiring",
		Files:     files,
		PathHints: []string{"aaa.go", "zzz.go"},
		PPR:       ppr,
		Limits:    limits,
	}
}

// TestSelectPPROrderBreaksScoreTies proves the advisory ordering decides
// score ties against the historical path order: without it the lexically
// first tied file wins, with a reversed ordering the other tied file wins
// under identical file/byte bounds.
func TestSelectPPROrderBreaksScoreTies(t *testing.T) {
	files := pprTieFiles()
	legacy, err := Select(pprTieInput(files, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy.Selected) != 1 || legacy.Selected[0].Path != "aaa.go" {
		t.Fatalf("legacy tie not broken by path: %+v", legacy.Selected)
	}
	ranked, err := Select(pprTieInput(files, []string{"zzz.go", "aaa.go"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranked.Selected) != 1 || ranked.Selected[0].Path != "zzz.go" {
		t.Fatalf("PPR ordering did not break the tie: %+v", ranked.Selected)
	}
	// The treatment changes only which tied file is visible: same excerpt
	// mechanics, same reason family, same byte budget.
	if ranked.Selected[0].Reason != legacy.Selected[0].Reason {
		t.Fatalf("PPR changed excerpt evidence: %q vs %q", ranked.Selected[0].Reason, legacy.Selected[0].Reason)
	}
	if ranked.SelectedBytes != legacy.SelectedBytes {
		t.Fatal("PPR changed the returned byte cost")
	}
	// Reordering the advisory input is a different input.
	flipped, err := Select(pprTieInput(files, []string{"aaa.go", "zzz.go"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(flipped.Selected) != 1 || flipped.Selected[0].Path != "aaa.go" {
		t.Fatalf("reordered PPR did not follow rank order: %+v", flipped.Selected)
	}
	if ranked.InputHash == legacy.InputHash || ranked.InputHash == flipped.InputHash {
		t.Fatal("PPR ordering is not bound into the input identity")
	}
}

// TestSelectPPRNeverPromotesZeroSignal proves a zero-score file stays
// omitted even when it carries an advisory ordinal: the ordering breaks
// ties, it never manufactures positive evidence.
func TestSelectPPRNeverPromotesZeroSignal(t *testing.T) {
	signalled := admitted("hit.go", "package hit\nfixit\n")
	silent := admitted("quiet.go", "package quiet\n")
	limits := DefaultLimits()
	limits.MaxFiles = 1
	input := Input{
		Version:   version,
		Scope:     Scope{SourceID: strings.Repeat("1", 64)},
		Objective: "repair fixit handling",
		Files:     []File{signalled, silent},
		PathHints: []string{"hit.go"},
		PPR:       []string{"quiet.go", "hit.go"},
		Limits:    limits,
	}
	manifest, err := Select(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Selected) != 1 || manifest.Selected[0].Path != "hit.go" {
		t.Fatalf("zero-signal file was promoted by PPR order: %+v", manifest.Selected)
	}
	found := false
	for _, omission := range manifest.Omissions {
		if omission.Path == "quiet.go" && omission.Reason == "unranked" {
			found = true
		}
	}
	if !found {
		t.Fatalf("zero-signal file lacks an explicit omission: %+v", manifest.Omissions)
	}
}

// TestSelectPPRValidation rejects orderings outside the admitted corpus,
// duplicates, oversized orderings and mixing with the experimental RRF mode.
func TestSelectPPRValidation(t *testing.T) {
	files := pprTieFiles()
	base := pprTieInput(files, nil)
	oversized := make([]string, 65)
	for i := range oversized {
		oversized[i] = "aaa.go"
	}
	for name, ppr := range map[string][]string{
		"foreign path":      {"missing.go"},
		"sensitive escape":  {"../escape.go"},
		"duplicate":         {"aaa.go", "aaa.go"},
		"case alias":        {"AAA.go"},
		"oversized":         oversized,
		"unadmitted member": {"aaa.go", "zzz.go", "extra.go"},
	} {
		t.Run(name, func(t *testing.T) {
			input := base
			input.PPR = ppr
			if _, err := Select(input); err == nil {
				t.Fatalf("invalid PPR ordering accepted: %q", ppr)
			}
		})
	}
	mixed := base
	mixed.Selector = SelectorRRFCoverageV1
	mixed.PPR = []string{"aaa.go"}
	if _, err := Select(mixed); err == nil {
		t.Fatal("PPR mixed with the experimental selector was accepted")
	}
}

// TestSelectPPREmptyPreservesLegacyIdentity proves an absent ordering keeps
// the exact historical selection bytes: the seam is opt-in only.
func TestSelectPPREmptyPreservesLegacyIdentity(t *testing.T) {
	files := pprTieFiles()
	first, err := Select(pprTieInput(files, nil))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Select(pprTieInput(files, []string{}))
	if err != nil {
		t.Fatal(err)
	}
	if first.InputHash != second.InputHash {
		t.Fatal("empty PPR changed the historical input hash")
	}
	if len(first.Selected) != 1 || len(second.Selected) != 1 || first.Selected[0].Path != second.Selected[0].Path {
		t.Fatal("empty PPR changed the historical selection")
	}
	for _, selected := range append(first.Selected, second.Selected...) {
		if len(selected.Ranks) != 0 || len(selected.Covered) != 0 {
			t.Fatal("legacy selection carries selector provenance")
		}
	}
	if _, err := first.ID(); err != nil {
		t.Fatalf("legacy manifest identity rejected: %v", err)
	}
}
