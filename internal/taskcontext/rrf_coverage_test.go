package taskcontext

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func rrfFixtureInput() Input {
	return Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "Fix alpha beta handling",
		Files: []File{
			admitted("a.txt", "alpha first file content\n"),
			admitted("b.txt", "beta second file content alpha again\n"),
			admitted("c.txt", "unrelated zzz qqq content\n"),
		},
		ChangedPaths: []string{"b.txt"},
		PathHints:    []string{"a.txt", "b.txt"},
		Selector:     SelectorRRFCoverageV1,
		Limits:       DefaultLimits(),
	}
}

func TestRRFExactKnownRanks(t *testing.T) {
	files := []File{
		admitted("x.txt", "x"),
		admitted("y.txt", "y"),
		admitted("z.txt", "z"),
	}
	r1 := map[string]int{"x.txt": 1, "y.txt": 2, "z.txt": 3}
	r2 := map[string]int{"y.txt": 1, "z.txt": 2}
	ordered, scores := fuseRRF(files, []map[string]int{r1, r2})
	if len(ordered) != 3 || ordered[0].Path != "y.txt" || ordered[1].Path != "z.txt" || ordered[2].Path != "x.txt" {
		t.Fatalf("RRF order wrong: %+v", ordered)
	}
	// Exact: y = 1/62+1/61, z = 1/63+1/62, x = 1/61.
	if scores["y.txt"].Cmp(scores["z.txt"]) <= 0 || scores["z.txt"].Cmp(scores["x.txt"]) <= 0 {
		t.Fatal("RRF exact scores not ordered")
	}
	// Equal unweighted: swapping ranker order must not change fusion.
	again, _ := fuseRRF(files, []map[string]int{r2, r1})
	if !reflect.DeepEqual(ordered, again) {
		t.Fatal("RRF depends on ranker order")
	}
}

func TestRRFTiePermutationPreservesIdentity(t *testing.T) {
	in := rrfFixtureInput()
	first, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	in.Files[0], in.Files[2] = in.Files[2], in.Files[0]
	second, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("RRF selection depends on input order:\n%+v\n%+v", first, second)
	}
	if first.Selector != SelectorRRFCoverageV1 {
		t.Fatal("manifest selector not bound")
	}
	if _, err := first.ID(); err != nil {
		t.Fatal(err)
	}
	for _, sel := range first.Selected {
		if len(sel.Ranks) == 0 {
			t.Fatalf("selected item lacks ranker provenance: %+v", sel)
		}
		if err := validateProvenance(sel.Ranks, sel.Covered); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRRFSkipsAbsentRankers(t *testing.T) {
	files := []File{admitted("a.txt", "alpha here"), admitted("b.txt", "beta here")}
	empty := map[string]int(nil)
	r := map[string]int{"b.txt": 1}
	ordered, _ := fuseRRF(files, []map[string]int{empty, r, nil})
	if len(ordered) != 1 || ordered[0].Path != "b.txt" {
		t.Fatalf("absent rankers not skipped: %+v", ordered)
	}
}

func TestRRFNoSignalFallback(t *testing.T) {
	in := fixtureInput()
	in.Objective = "Fix qqqzzz wwwvv handling"
	in.Files = []File{admitted("one.txt", "unrelated"), admitted("two.txt", "also unrelated")}
	in.ChangedPaths, in.PathHints = nil, nil
	in.Selector = SelectorRRFCoverageV1
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 1 || got.Selected[0].Reason != "deterministic_fallback" || got.Selector != SelectorRRFCoverageV1 {
		t.Fatalf("no-signal fallback wrong: %+v", got)
	}
	if got.OmittedCount != 1 {
		t.Fatal("omitted count inexact", got)
	}
}

func TestCoveragePrefersComplementaryOverRedundant(t *testing.T) {
	in := Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "alpha beta",
		Files: []File{
			admitted("alpha1.txt", "alpha first\n"),
			admitted("alpha2.txt", "alpha second\n"),
			admitted("beta.txt", "beta only\n"),
		},
		ChangedPaths: nil,
		PathHints:    nil,
		Selector:     SelectorRRFCoverageV1,
		Limits:       Limits{MaxFiles: 3, MaxBytes: 4096, MaxBytesPerFile: 2048, MaxInputBytes: 16384, MaxOmissions: 8},
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 2 {
		t.Fatalf("expected 2 complementary excerpts: %+v", got)
	}
	paths := []string{got.Selected[0].Path, got.Selected[1].Path}
	hasAlpha := paths[0] == "alpha1.txt" || paths[1] == "alpha1.txt" || paths[0] == "alpha2.txt" || paths[1] == "alpha2.txt"
	hasBeta := paths[0] == "beta.txt" || paths[1] == "beta.txt"
	if !hasAlpha || !hasBeta {
		t.Fatalf("coverage did not prefer complementary evidence: %v", paths)
	}
	foundRedundant := false
	for _, o := range got.Omissions {
		if o.Reason == "redundant" {
			foundRedundant = true
		}
	}
	if !foundRedundant {
		t.Fatalf("redundant omission not reported: %+v", got.Omissions)
	}
}

func TestCoverageGainPerByteCost(t *testing.T) {
	// B covers one concept cheaply; A covers two concepts but is long.
	// Gain/cost must order B first under a matched byte cap.
	aContent := "alpha " + strings.Repeat("filler ", 200) + "beta\n"
	bContent := "alpha\n"
	in := Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "alpha beta",
		Files: []File{
			admitted("a_long.txt", aContent),
			admitted("b_short.txt", bContent),
		},
		Selector: SelectorRRFCoverageV1,
		Limits:   Limits{MaxFiles: 2, MaxBytes: 4096, MaxBytesPerFile: 4096, MaxInputBytes: 16384, MaxOmissions: 8},
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 2 || got.Selected[0].Path != "b_short.txt" {
		t.Fatalf("gain/cost priority wrong: %+v", got.Selected)
	}
}

func TestCoverageConceptsAreExcerptVisible(t *testing.T) {
	in := rrfFixtureInput()
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, sel := range got.Selected {
		for _, c := range sel.Covered {
			if strings.HasPrefix(c, "path:") {
				continue
			}
			counts, _ := tokenLexicalIndex(strings.ToLower(sel.Content))
			if _, ok := counts[c]; !ok {
				t.Fatalf("covered concept %q not visible in excerpt of %s", c, sel.Path)
			}
		}
		if !utf8.ValidString(sel.Content) || int64(len(sel.Content)) != sel.End-sel.Start {
			t.Fatal("invalid excerpt bounds", sel)
		}
	}
	if got.SelectedBytes > DefaultLimits().MaxBytes {
		t.Fatal("byte cap exceeded")
	}
	if len(got.Selected) > DefaultLimits().MaxFiles {
		t.Fatal("file count exceeded")
	}
}

func TestRRFSensitiveRedactedAndBounds(t *testing.T) {
	in := rrfFixtureInput()
	in.Files = append(in.Files, admitted(".env", "SECRET=live\n"))
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range got.Selected {
		if s.Path == ".env" || strings.Contains(s.Content, "SECRET") {
			t.Fatal("sensitive file selected")
		}
	}
	for _, o := range got.Omissions {
		if o.Reason == "sensitive_path" && o.Path != "[redacted]" {
			t.Fatalf("sensitive path disclosed: %+v", o)
		}
	}
}

func TestSelectorRejectsInvalidInputs(t *testing.T) {
	in := rrfFixtureInput()
	bad := in
	bad.Files[0].Hash = strings.Repeat("0", 64)
	if _, err := Select(bad); err == nil {
		t.Fatal("forged hash admitted")
	}
	alias := in
	alias.Files = []File{admitted("A.txt", "alpha"), admitted("a.txt", "beta")}
	if _, err := Select(alias); err == nil {
		t.Fatal("path alias admitted")
	}
	unknown := in
	unknown.Selector = "unknown-mode"
	if _, err := Select(unknown); err == nil {
		t.Fatal("unknown selector admitted")
	}
	if err := ValidateSelector("unknown-mode"); err == nil {
		t.Fatal("unknown selector validated")
	}
}

func TestCoverageEmptyRankedCorpusNoPanic(t *testing.T) {
	in := Input{
		Version:      1,
		Scope:        Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective:    "Fix empty handling",
		Files:        []File{admitted("empty.txt", "")},
		ChangedPaths: []string{"empty.txt"},
		Selector:     SelectorRRFCoverageV1,
		Limits:       DefaultLimits(),
	}
	first, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Selected) != 0 {
		t.Fatalf("empty corpus must select nothing: %+v", first.Selected)
	}
	if first.Selector != SelectorRRFCoverageV1 {
		t.Fatal("manifest selector not bound")
	}
	if first.OmittedCount != 1 || len(first.Omissions) != 1 || first.Omissions[0].Reason != "empty" || first.Omissions[0].Path != "empty.txt" {
		t.Fatalf("exact omission record wrong: %+v", first)
	}
	if first.OmissionsTrimmed {
		t.Fatal("unexpected omission trim")
	}
	if _, err := first.ID(); err != nil {
		t.Fatalf("empty ranked manifest ID invalid: %v", err)
	}
	second, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("empty ranked selection not deterministic")
	}
	if _, err := second.ID(); err != nil {
		t.Fatal(err)
	}
}

func TestCoverageMixedRankedEmptyAndUnranked(t *testing.T) {
	in := Input{
		Version:      1,
		Scope:        Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective:    "Fix qqqzzz wwwvv handling",
		Files:        []File{admitted("plain.txt", "unrelated text"), admitted("empty.txt", "")},
		ChangedPaths: []string{"empty.txt"},
		Selector:     SelectorRRFCoverageV1,
		Limits:       DefaultLimits(),
	}
	first, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Selected) != 0 {
		t.Fatalf("mixed corpus must select nothing: %+v", first.Selected)
	}
	if first.OmittedCount != 2 || len(first.Omissions) != 2 {
		t.Fatalf("exact omission count wrong: %+v", first)
	}
	seen := map[string]string{}
	for _, o := range first.Omissions {
		if _, dup := seen[o.Path]; dup {
			t.Fatalf("duplicate omission entry: %+v", first.Omissions)
		}
		seen[o.Path] = o.Reason
	}
	if seen["empty.txt"] != "empty" || seen["plain.txt"] != "unranked" {
		t.Fatalf("contradictory omission entries: %+v", first.Omissions)
	}
	if _, err := first.ID(); err != nil {
		t.Fatalf("mixed manifest ID invalid: %v", err)
	}
	second, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("mixed selection not deterministic under permutation")
	}
	if _, err := second.ID(); err != nil {
		t.Fatal(err)
	}
}

func TestCoverageSkipsOverlongTermProvenance(t *testing.T) {
	long := strings.Repeat("a", 100)
	in := Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "Fix " + long + " alpha",
		Files:     []File{admitted("a.txt", long+" alpha\n")},
		Selector:  SelectorRRFCoverageV1,
		Limits:    Limits{MaxFiles: 2, MaxBytes: 4096, MaxBytesPerFile: 4096, MaxInputBytes: 16384, MaxOmissions: 8},
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := got.ID(); err != nil {
		t.Fatalf("overlong-term manifest ID invalid: %v", err)
	}
	for _, sel := range got.Selected {
		for _, c := range sel.Covered {
			if len(c) < 2 || len(c) > 64 || !utf8.ValidString(c) {
				t.Fatalf("invalid covered concept bytes: %q", c)
			}
			if c == long {
				t.Fatal("overlong term fabricated into provenance")
			}
		}
	}
}

func TestCoverageSkipsOverlongPathProvenance(t *testing.T) {
	longPath := strings.Repeat("x", 60) + ".txt"
	if len("path:"+longPath) <= 64 {
		t.Fatal("fixture path concept not overlong")
	}
	in := Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "alpha handling",
		Files: []File{
			admitted(longPath, "alpha content\n"),
			admitted("b.txt", "alpha other\n"),
		},
		PathHints: []string{longPath},
		Selector:  SelectorRRFCoverageV1,
		Limits:    Limits{MaxFiles: 4, MaxBytes: 8192, MaxBytesPerFile: 4096, MaxInputBytes: 32768, MaxOmissions: 8},
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := got.ID(); err != nil {
		t.Fatalf("overlong-path manifest ID invalid: %v", err)
	}
	for _, sel := range got.Selected {
		for _, c := range sel.Covered {
			if len(c) > 64 {
				t.Fatalf("overlong path concept in provenance: %q", c)
			}
			if c == "path:"+longPath {
				t.Fatal("overlong path identity fabricated into provenance")
			}
		}
	}
}

func TestCoverageMultibyteConceptByteBound(t *testing.T) {
	term64 := strings.Repeat("é", 32)
	if len(term64) != 64 {
		t.Fatalf("fixture not 64 bytes: %d", len(term64))
	}
	in := Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "Fix " + term64,
		Files:     []File{admitted("m.txt", term64+"\n")},
		Selector:  SelectorRRFCoverageV1,
		Limits:    Limits{MaxFiles: 1, MaxBytes: 4096, MaxBytesPerFile: 4096, MaxInputBytes: 16384, MaxOmissions: 8},
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := got.ID(); err != nil {
		t.Fatalf("64-byte concept manifest ID invalid: %v", err)
	}
	found := false
	for _, sel := range got.Selected {
		for _, c := range sel.Covered {
			if c == term64 {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("exact 64-byte concept missing from provenance: %+v", got.Selected)
	}
	term66 := strings.Repeat("é", 33)
	if len(term66) <= 64 {
		t.Fatal("fixture not overlong")
	}
	over := Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "Fix " + term66 + " alpha",
		Files:     []File{admitted("m.txt", term66+" alpha\n")},
		Selector:  SelectorRRFCoverageV1,
		Limits:    Limits{MaxFiles: 1, MaxBytes: 4096, MaxBytesPerFile: 4096, MaxInputBytes: 16384, MaxOmissions: 8},
	}
	gotOver, err := Select(over)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gotOver.ID(); err != nil {
		t.Fatalf("overlong multibyte manifest ID invalid: %v", err)
	}
	for _, sel := range gotOver.Selected {
		for _, c := range sel.Covered {
			if c == term66 {
				t.Fatal("overlong multibyte concept fabricated into provenance")
			}
			if len(c) > 64 {
				t.Fatalf("covered concept exceeds byte bound: %q (%d bytes)", c, len(c))
			}
		}
	}
}

func TestLegacyEmptySelectorPreservesExactIdentity(t *testing.T) {
	in := fixtureInput()
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Selector != "" {
		t.Fatal("legacy manifest carries selector")
	}
	if got.InputHash != "69beccd08a7a02e5754064bf355c0c24cc90460e418e72357f8b1cc799e21158" {
		t.Fatalf("legacy input identity changed: %s", got.InputHash)
	}
	for _, sel := range got.Selected {
		if len(sel.Ranks) != 0 || len(sel.Covered) != 0 {
			t.Fatalf("legacy selection carries provenance: %+v", sel)
		}
	}
	if _, err := got.ID(); err != nil {
		t.Fatal(err)
	}
}
