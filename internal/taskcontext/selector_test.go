package taskcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func admitted(path, content string) File {
	b := []byte(content)
	h := sha256.Sum256(b)
	return File{Path: path, Hash: hex.EncodeToString(h[:]), Content: b}
}

func fixtureInput() Input {
	return Input{
		Version:   1,
		Scope:     Scope{SourceID: strings.Repeat("a", 64), CandidateID: strings.Repeat("b", 64)},
		Objective: "Fix greeting trimming and add regression coverage",
		Files: []File{
			admitted("README.md", "Fabric greeting usage\n"),
			admitted("internal/greeting/greeting.go", "package greeting\nfunc Hello(name string) string { return \"Hello, \" + name }\n"),
			admitted("internal/greeting/greeting_test.go", "package greeting\nfunc TestHelloTrimmed(t *testing.T) {}\n"),
			admitted(".env", "API_KEY=must-not-be-prompted\n"),
		},
		ChangedPaths: []string{"internal/greeting/greeting.go"},
		PathHints:    []string{"internal/greeting/greeting_test.go"},
		Limits:       DefaultLimits(),
	}
}

func TestSelectIsDeterministicRelevantAndBounded(t *testing.T) {
	in := fixtureInput()
	first, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	in.Files[0], in.Files[2] = in.Files[2], in.Files[0]
	in.ChangedPaths = []string{"internal/greeting/greeting.go"}
	second, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("selection depends on input order:\n%+v\n%+v", first, second)
	}
	if len(first.Selected) != 3 || first.Selected[0].Path != "internal/greeting/greeting.go" || first.Selected[0].Reason != "changed_path" || first.SelectedBytes > fixtureInput().Limits.MaxBytes {
		t.Fatal("unexpected relevance or bounds", first)
	}
	for _, selected := range first.Selected {
		if strings.Contains(selected.Content, "API_KEY") || selected.Start < 0 || selected.End <= selected.Start {
			t.Fatal("unsafe or invalid excerpt", selected)
		}
	}
	if first.OmittedCount != 1 || len(first.Omissions) != 1 || first.Omissions[0] != (Omission{Path: "[redacted]", Reason: "sensitive_path"}) {
		t.Fatal("sensitive omission was not observable", first)
	}
	if _, err := first.ID(); err != nil {
		t.Fatal(err)
	}
}

func TestSelectUsesExcerptAndReportsBudgetOmissions(t *testing.T) {
	in := fixtureInput()
	in.Files = []File{
		admitted("a.txt", strings.Repeat("prefix ", 80)+"greeting target"+strings.Repeat(" suffix", 80)),
		admitted("b.txt", "greeting second target"),
	}
	in.ChangedPaths = nil
	in.PathHints = nil
	in.Limits = Limits{MaxFiles: 1, MaxBytes: 256, MaxBytesPerFile: 256, MaxInputBytes: 4096, MaxOmissions: 8}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 1 || got.Selected[0].Path != "a.txt" || got.Selected[0].Start == 0 || got.SelectedBytes > 256 || got.OmittedCount != 1 {
		t.Fatal("excerpt or file budget failed", got)
	}
	if len(got.Omissions) != 1 || got.Omissions[0] != (Omission{Path: "b.txt", Reason: "file_limit"}) {
		t.Fatal("budget omission differs", got)
	}
}

func TestSelectRejectsUnboundOrForgedInputs(t *testing.T) {
	for name, mutate := range map[string]func(*Input){
		"missing source": func(in *Input) { in.Scope.SourceID = "" },
		"forged hash":    func(in *Input) { in.Files[0].Hash = strings.Repeat("0", 64) },
		"path escape":    func(in *Input) { in.Files[0].Path = "../secret" },
		"byte flood":     func(in *Input) { in.Limits.MaxInputBytes = 512 },
	} {
		t.Run(name, func(t *testing.T) {
			in := fixtureInput()
			mutate(&in)
			if _, err := Select(in); err == nil {
				t.Fatal("invalid input admitted")
			}
		})
	}
}

func TestSelectDoesNotClaimOmittedFilesAreAbsent(t *testing.T) {
	in := fixtureInput()
	in.Files = []File{admitted("one.txt", "unrelated"), admitted("two.txt", "also unrelated")}
	in.ChangedPaths, in.PathHints = nil, nil
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 1 || got.Selected[0].Reason != "deterministic_fallback" || got.OmittedCount != 1 || got.Omissions[0].Reason != "unranked" {
		t.Fatal("fallback semantics changed", got)
	}
}

func TestSelectRedactsSensitivePathsWithoutLeaking(t *testing.T) {
	in := fixtureInput()
	in.Objective = "Fix greeting trimming"
	in.ChangedPaths, in.PathHints = nil, nil
	in.Files = []File{
		admitted("internal/greeting/greeting.go", "package greeting\nfunc Hello() string { return \"hi\" }\n"),
		admitted("production.env", "API_KEY=live-secret\n"),
		admitted("config/.env.local", "TOKEN=abc\n"),
		admitted("secrets/deploy.txt", "token material\n"),
		admitted("config/credentials.json", "{\"token\":\"x\"}\n"),
		admitted("deploy/id_rsa", "private key material\n"),
		admitted("notes/secretary_minutes.md", "meeting notes about greeting\n"),
		admitted("docs/keyboard_shortcuts.md", "ordinary editor shortcuts\n"),
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range got.Omissions {
		if o.Reason == "sensitive_path" {
			if o.Path != "[redacted]" {
				t.Fatalf("sensitive path disclosed: %+v", o)
			}
		} else {
			if o.Path == "[redacted]" {
				t.Fatalf("ordinary omission redacted: %+v", o)
			}
		}
		if strings.Contains(o.Path, "production.env") || strings.Contains(o.Path, "API_KEY") || strings.Contains(o.Path, "deploy.txt") {
			t.Fatalf("sensitive material leaked in omission: %+v", o)
		}
	}
	if got.OmittedCount != len(in.Files)-len(got.Selected) {
		t.Fatal("omitted count inexact", got)
	}
	sensitiveCount := 0
	for _, f := range in.Files {
		if f.Path == "production.env" || f.Path == "config/.env.local" || f.Path == "secrets/deploy.txt" || f.Path == "config/credentials.json" || f.Path == "deploy/id_rsa" {
			sensitiveCount++
			for _, s := range got.Selected {
				if s.Path == f.Path {
					t.Fatalf("sensitive file selected: %s", f.Path)
				}
			}
		}
	}
	if sensitiveCount != 5 {
		t.Fatal("fixture miscount")
	}
	foundOrdinary := false
	for _, s := range got.Selected {
		if s.Path == "notes/secretary_minutes.md" || s.Path == "docs/keyboard_shortcuts.md" {
			foundOrdinary = true
		}
		if strings.Contains(s.Content, "API_KEY") {
			t.Fatal("sensitive content selected")
		}
	}
	if !foundOrdinary {
		t.Fatal("ordinary sources with key substrings were excluded", got)
	}
}

func TestSelectFallbackOnlyWhenNoPositiveMatches(t *testing.T) {
	in := fixtureInput()
	in.ChangedPaths, in.PathHints = nil, nil
	in.Files = []File{
		admitted("a.txt", "greeting target present here"),
		admitted("b.txt", "totally unrelated zzz"),
		admitted("c.txt", "another unrelated qqq"),
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range got.Selected {
		if s.Reason == "deterministic_fallback" {
			t.Fatalf("fallback assigned despite positive matches: %+v", got)
		}
	}
	if len(got.Selected) != 1 || got.Selected[0].Path != "a.txt" {
		t.Fatal("positive match not preferred", got)
	}
	for _, o := range got.Omissions {
		if o.Reason != "unranked" {
			t.Fatalf("expected unranked omissions, got %+v", o)
		}
	}
}

func TestSelectUnicodeExcerptKeepsByteBounds(t *testing.T) {
	prefix := strings.Repeat("préfixe ☃ ", 40)
	suffix := strings.Repeat(" suffixe ☃", 40)
	content := prefix + "GRÜSSE Ziel" + suffix
	in := fixtureInput()
	in.Objective = "find grüsse ziel"
	in.ChangedPaths, in.PathHints = nil, nil
	in.Files = []File{admitted("unicode/notes.txt", content)}
	in.Limits = Limits{MaxFiles: 1, MaxBytes: 512, MaxBytesPerFile: 256, MaxInputBytes: 4096, MaxOmissions: 8}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 1 {
		t.Fatal("unicode file not selected", got)
	}
	sel := got.Selected[0]
	raw := []byte(content)
	if int(sel.Start) < 0 || int(sel.End) > len(raw) || sel.Start >= sel.End {
		t.Fatal("invalid excerpt bounds", sel)
	}
	if !utf8.ValidString(sel.Content) {
		t.Fatal("excerpt splits UTF-8")
	}
	if string(raw[sel.Start:sel.End]) != sel.Content {
		t.Fatal("excerpt bounds do not match original bytes")
	}
	if !utf8.RuneStart(raw[sel.Start]) {
		t.Fatal("excerpt start splits rune", sel)
	}
	if int(sel.End) < len(raw) && !utf8.RuneStart(raw[sel.End]) {
		t.Fatal("excerpt end splits rune", sel)
	}
	// Window should be centered on the lexical hit, not clamped to zero.
	if sel.Start == 0 {
		t.Fatal("lexical window ignored pivot", sel)
	}
	if !strings.Contains(strings.ToLower(sel.Content), "grüsse") && !strings.Contains(strings.ToLower(sel.Content), "ziel") {
		t.Fatal("excerpt misses lexical hit", sel)
	}
}

func TestSelectRequiresCandidateForChangedPaths(t *testing.T) {
	in := fixtureInput()
	in.Scope.CandidateID = ""
	in.ChangedPaths = []string{"internal/greeting/greeting.go"}
	if _, err := Select(in); err == nil {
		t.Fatal("changed paths without candidate admitted")
	}
	in.Scope.CandidateID = strings.Repeat("b", 64)
	in.ChangedPaths = nil
	if _, err := Select(in); err != nil {
		t.Fatal(err)
	}
}

func TestSelectFixTermDoesNotMatchInsidePrefix(t *testing.T) {
	in := fixtureInput()
	in.Objective = "fix it"
	in.ChangedPaths, in.PathHints = nil, nil
	in.Files = []File{
		admitted("a.txt", strings.Repeat("prefix ", 50)),
		admitted("b.txt", "totally unrelated zzz qqq"),
	}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range got.Selected {
		if s.Reason == "lexical_match" {
			t.Fatalf("short term matched inside larger word: %+v", got)
		}
	}
	if len(got.Selected) != 1 || got.Selected[0].Reason != "deterministic_fallback" {
		t.Fatal("expected fallback when only substring hits exist", got)
	}
}

func TestSelectMultipleTermsPreferMeaningfulExcerpt(t *testing.T) {
	in := fixtureInput()
	in.Objective = "Fix greeting trimming and add regression coverage"
	in.ChangedPaths, in.PathHints = nil, nil
	in.Files = []File{
		admitted("a.txt", strings.Repeat("prefix ", 80)+"greeting target"+strings.Repeat(" suffix", 80)),
		admitted("b.txt", strings.Repeat("prefix ", 80)+"greeting regression coverage"+strings.Repeat(" suffix", 80)),
	}
	in.Limits = Limits{MaxFiles: 1, MaxBytes: 512, MaxBytesPerFile: 256, MaxInputBytes: 8192, MaxOmissions: 8}
	got, err := Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Selected) != 1 || got.Selected[0].Path != "b.txt" {
		t.Fatal("multi-term ranking failed", got)
	}
	sel := got.Selected[0]
	if sel.Start == 0 {
		t.Fatal("meaningful excerpt pivoted at zero", sel)
	}
	if !utf8.ValidString(sel.Content) || string([]byte(sel.Content)) != sel.Content {
		t.Fatal("invalid excerpt", sel)
	}
	lower := strings.ToLower(sel.Content)
	if !strings.Contains(lower, "greeting") {
		t.Fatal("excerpt misses meaningful hit", sel)
	}
}

func TestSelectPivotPrefersRareLongTargetTermIndependentOfObjectiveOrder(t *testing.T) {
	content := strings.Repeat("and common filler words\n", 500) + "Compare nearby context\n" + strings.Repeat("padding elsewhere\n", 900) + "func SplitLines(s string) []string { return split(s) }\n"
	if at := strings.Index(content, "func SplitLines"); at <= 8<<10 {
		t.Fatal("fixture target must be beyond the first excerpt window", at)
	}
	base := fixtureInput()
	base.ChangedPaths, base.PathHints = nil, nil
	base.Files = []File{admitted("difflib/difflib.go", content)}
	base.Limits = Limits{MaxFiles: 1, MaxBytes: 8 << 10, MaxBytesPerFile: 8 << 10, MaxInputBytes: 1 << 20, MaxOmissions: 8}

	base.Objective = "and Compare SplitLines"
	first, err := Select(base)
	if err != nil {
		t.Fatal(err)
	}
	base.Objective = "SplitLines and Compare"
	second, err := Select(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Selected) != 1 || len(second.Selected) != 1 {
		t.Fatal("target file was not selected", first, second)
	}
	a, b := first.Selected[0], second.Selected[0]
	if a.Start != b.Start || a.End != b.End || a.Content != b.Content {
		t.Fatal("objective term order changed the selected excerpt", a, b)
	}
	if a.Start <= 0 || !strings.Contains(a.Content, "func SplitLines") {
		t.Fatal("excerpt pivot missed the distinctive target near end of file", a)
	}
	if len(a.Content) > base.Limits.MaxBytesPerFile || first.SelectedBytes > base.Limits.MaxBytes || !utf8.ValidString(a.Content) {
		t.Fatal("pivot exceeded excerpt bounds or produced invalid UTF-8", a, first.SelectedBytes)
	}
}

func TestEligiblePathReusesSensitiveRules(t *testing.T) {
	for _, p := range []string{".env", "config/.env.local", "secrets/deploy.txt", "config/credentials.json", "deploy/id_rsa", "deploy/key.pem"} {
		if EligiblePath(p) {
			t.Fatalf("sensitive path eligible: %s", p)
		}
	}
	for _, p := range []string{"internal/greeting/greeting.go", "docs/keyboard_shortcuts.md", "notes/secretary_minutes.md", "file.txt"} {
		if !EligiblePath(p) {
			t.Fatalf("ordinary path ineligible: %s", p)
		}
	}
}
