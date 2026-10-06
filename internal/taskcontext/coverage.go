package taskcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"unicode/utf8"
)

// coverageConcepts derives finite, sparse, evidence-bound binary concepts:
// objective identifier terms actually admitted plus explicit path/anchor/task
// cues actually admitted. Path concepts use the exact admitted path string so
// coverage prefers complementary hinted files without inventing latent topics.
// Total concepts are capped to keep selection bounded.
//
// Concept admission mirrors the provenance bound enforced by
// validateProvenance: 2..64 UTF-8 bytes, valid UTF-8, non-blank. Overlong cues
// are skipped here, never byte-truncated and never fabricated; they remain
// valid RRF rank signals, they simply yield no coverage provenance.
func validCoverageConcept(c string) bool {
	if len(c) < 2 || len(c) > maxCoveredConceptLn || !utf8.ValidString(c) || strings.TrimSpace(c) == "" {
		return false
	}
	return true
}

func coverageConcepts(objective string, changed, hints []string, anchors map[string]int) []string {
	terms := words(objective)
	seen := map[string]bool{}
	out := []string{}
	for _, t := range terms {
		if !seen[t] && validCoverageConcept(t) {
			seen[t] = true
			out = append(out, t)
		}
	}
	paths := map[string]bool{}
	for _, p := range changed {
		if c := "path:" + p; validCoverageConcept(c) {
			paths[c] = true
		}
	}
	for _, p := range hints {
		if c := "path:" + p; validCoverageConcept(c) {
			paths[c] = true
		}
	}
	for p := range anchors {
		if c := "path:" + p; validCoverageConcept(c) {
			paths[c] = true
		}
	}
	sorted := make([]string, 0, len(paths))
	for p := range paths {
		sorted = append(sorted, p)
	}
	sort.Strings(sorted)
	// Keep total concepts bounded and sparse: objective terms first, then a
	// bounded slice of path cues.
	budget := 128 - len(out)
	if budget < 0 {
		return out[:128]
	}
	if len(sorted) > budget {
		sorted = sorted[:budget]
	}
	// Cap further to 64 path cues to keep per-excerpt matching sparse.
	if len(sorted) > 64 {
		sorted = sorted[:64]
	}
	out = append(out, sorted...)
	if len(out) > 128 {
		return out[:128]
	}
	return out
}

// excerptPivot reuses the safe existing excerpt pivot: rarest bounded content
// term occurrence, else bound anchor, else zero. It scans at most the capped
// 32KiB sample once per file; callers must not reread per selection.
func excerptPivot(content []byte, terms []string, anchor *int) int {
	if len(content) == 0 {
		return 0
	}
	sample := content
	if len(sample) > 32<<10 {
		cut := 32 << 10
		for cut > 0 && !utf8.RuneStart(sample[cut]) {
			cut--
		}
		sample = sample[:cut]
	}
	counts, firsts := tokenLexicalIndex(string(sample))
	pivot := -1
	pivotTerm := ""
	pivotCount := 0
	pivotLength := 0
	for _, term := range terms {
		n, ok := counts[term]
		if !ok {
			continue
		}
		off := firsts[term]
		termLength := utf8.RuneCountInString(term)
		if pivotTerm == "" || n < pivotCount || n == pivotCount && (termLength > pivotLength || termLength == pivotLength && (off < pivot || off == pivot && term < pivotTerm)) {
			pivot = off
			pivotTerm = term
			pivotCount = n
			pivotLength = termLength
		}
	}
	if pivot >= 0 {
		return pivot
	}
	if anchor != nil {
		return *anchor
	}
	return 0
}

// excerptCovered matches concepts against the visible excerpt only, never
// hidden full-file text. Term concepts match whole tokens in the excerpt;
// path concepts match the excerpt's exact file path.
func excerptCovered(filePath, excerptContent string, concepts []string) []string {
	lower := strings.ToLower(excerptContent)
	counts, _ := tokenLexicalIndex(lower)
	out := []string{}
	for _, c := range concepts {
		if strings.HasPrefix(c, "path:") {
			if filePath == strings.TrimPrefix(c, "path:") {
				out = append(out, c)
			}
			continue
		}
		if _, ok := counts[c]; ok {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	if len(out) > maxCoveredConcepts {
		return out[:maxCoveredConcepts]
	}
	return out
}

type coverageCandidate struct {
	file    File
	pivot   int
	excerpt string
	start   int
	end     int
	covered []string
	rrfRank int
	ranks   map[string]int
}

// selectRRFCoverage implements the experimental rrf-coverage-v1 mode: fuse
// independent ordinal rankers with equal unweighted RRF, then greedily choose
// a bounded complementary excerpt set by marginal binary coverage per actual
// returned UTF-8 excerpt byte cost (gain/cost), with RRF as secondary ordering
// and path as final tie-break. This is the binary specialization of
// sum_c[1-product_(i in S)(1-a_ic)] with diminishing returns and no magic
// redundancy multiplier. Concept match applies to visible excerpts only.
func selectRRFCoverage(in Input, files []File) (Manifest, error) {
	manifest := Manifest{Version: version, Scope: in.Scope, Selected: []SelectedFile{}, Omissions: []Omission{}}
	for _, f := range files {
		manifest.InputBytes += len(f.Content)
	}
	inputHash, err := inputID(in, files)
	if err != nil {
		return Manifest{}, err
	}
	manifest.InputHash = inputHash
	manifest.Selector = SelectorRRFCoverageV1

	changed := pathSet(in.ChangedPaths)
	hints := pathSet(in.PathHints)
	terms := words(in.Objective)

	// Independent rankers with exact ordinals retained separately.
	rChanged := rankByPathSet(files, changed)
	rHint := rankByPathSet(files, hints)
	rPath := rankPathTokens(files, terms)
	rLex := rankContentLexical(files, terms)
	rAnchor := rankAnchors(files, in.Anchors)
	named := []struct {
		name    string
		ranking map[string]int
	}{
		{rankerChanged, rChanged},
		{rankerHint, rHint},
		{rankerPathToken, rPath},
		{rankerContentLex, rLex},
		{rankerAnchor, rAnchor},
	}
	rankings := []map[string]int{}
	for _, n := range named {
		if len(n.ranking) > 0 {
			rankings = append(rankings, n.ranking)
		}
	}
	concepts := coverageConcepts(in.Objective, in.ChangedPaths, in.PathHints, in.Anchors)

	// Eligible corpus: sensitive/non-UTF8 observations are omitted safely.
	eligible := []File{}
	for _, f := range files {
		if sensitivePath(f.Path) {
			appendSensitiveOmission(&manifest, in.Limits)
			continue
		}
		if !utf8.Valid(f.Content) {
			appendOmission(&manifest, in.Limits, f.Path, "non_utf8")
			continue
		}
		eligible = append(eligible, f)
	}
	if len(eligible) == 0 {
		manifest.OmittedCount = len(files) - len(manifest.Selected)
		manifest.OmissionsTrimmed = manifest.OmittedCount > len(manifest.Omissions)
		return manifest, nil
	}
	ordered, _ := fuseRRF(eligible, rankings)
	rrfOrder := map[string]int{}
	for i, f := range ordered {
		rrfOrder[f.Path] = i + 1
	}
	hasSignal := len(ordered) > 0
	if !hasSignal {
		// Explicit no-signal deterministic fallback: first path-sorted file.
		sorted := append([]File(nil), eligible...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
		return selectLegacyFallback(in, files, manifest, sorted[0])
	}

	// Precompute one visible excerpt per eligible file at MaxBytesPerFile.
	// No per-selection rereads of file bytes.
	cands := make([]*coverageCandidate, 0, len(eligible))
	for _, f := range eligible {
		rank, ok := rrfOrder[f.Path]
		if !ok {
			appendOmission(&manifest, in.Limits, f.Path, "unranked")
			continue
		}
		var anchor *int
		if v, ok := in.Anchors[f.Path]; ok {
			v := v
			anchor = &v
		}
		pivot := excerptPivot(f.Content, terms, anchor)
		want := in.Limits.MaxBytesPerFile
		start, end := excerpt(f.Content, pivot, want)
		if end-start < 1 {
			appendOmission(&manifest, in.Limits, f.Path, "empty")
			continue
		}
		content := string(f.Content[start:end])
		covered := excerptCovered(f.Path, content, concepts)
		ranks := provenanceRanks(f.Path, named)
		c := &coverageCandidate{file: f, pivot: pivot, excerpt: content, start: start, end: end, covered: covered, rrfRank: rank, ranks: ranks}
		cands = append(cands, c)
	}
	if len(cands) == 0 {
		// Every eligible file was ranked yet yielded no visible excerpt
		// (for example a single empty changed file). Return the bounded
		// omission record as-is: no duplicate entries, exact counts, and a
		// valid Manifest.ID. Admission replay stays fail-closed.
		manifest.OmittedCount = len(files) - len(manifest.Selected)
		manifest.OmissionsTrimmed = manifest.OmittedCount > len(manifest.Omissions)
		return manifest, nil
	}
	// Deterministic candidate order for evaluation: RRF order then path.
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].rrfRank != cands[j].rrfRank {
			return cands[i].rrfRank < cands[j].rrfRank
		}
		return cands[i].file.Path < cands[j].file.Path
	})

	coveredSet := map[string]bool{}
	selected := []*coverageCandidate{}
	selectedBytes := 0
	used := map[string]bool{}
	for len(selected) < in.Limits.MaxFiles {
		var best *coverageCandidate
		var bestGain int
		var bestCost int
		for _, c := range cands {
			if used[c.file.Path] {
				continue
			}
			gain := 0
			for _, concept := range c.covered {
				if !coveredSet[concept] {
					gain++
				}
			}
			if gain <= 0 {
				continue
			}
			cost := len(c.excerpt)
			if cost < 1 {
				continue
			}
			if selectedBytes+cost > in.Limits.MaxBytes {
				continue
			}
			if best == nil {
				best, bestGain, bestCost = c, gain, cost
				continue
			}
			// Primary: gain/cost; secondary: documented RRF order; final: path.
			lhs := int64(gain) * int64(bestCost)
			rhs := int64(bestGain) * int64(cost)
			if lhs != rhs {
				if lhs > rhs {
					best, bestGain, bestCost = c, gain, cost
				}
				continue
			}
			if c.rrfRank != best.rrfRank {
				if c.rrfRank < best.rrfRank {
					best, bestGain, bestCost = c, gain, cost
				}
				continue
			}
			if c.file.Path < best.file.Path {
				best, bestGain, bestCost = c, gain, cost
			}
		}
		if best == nil {
			break
		}
		remaining := in.Limits.MaxBytes - selectedBytes
		if remaining < 128 {
			break
		}
		// Actual returned byte cost: re-slice to remaining on UTF-8 boundary
		// without rescanning the corpus; recompute visible coverage exactly.
		want := min(in.Limits.MaxBytesPerFile, remaining)
		start, end := excerpt(best.file.Content, best.pivot, want)
		if end-start < 1 {
			appendOmission(&manifest, in.Limits, best.file.Path, "empty")
			used[best.file.Path] = true
			continue
		}
		content := best.file.Content[start:end]
		visible := string(content)
		actualCovered := excerptCovered(best.file.Path, visible, concepts)
		newCovered := []string{}
		for _, c := range actualCovered {
			if !coveredSet[c] {
				newCovered = append(newCovered, c)
			}
		}
		sort.Strings(newCovered)
		if len(newCovered) > maxCoveredConcepts {
			newCovered = newCovered[:maxCoveredConcepts]
		}
		if len(newCovered) == 0 {
			appendOmission(&manifest, in.Limits, best.file.Path, "redundant")
			used[best.file.Path] = true
			continue
		}
		hash := sha256.Sum256(content)
		manifest.Selected = append(manifest.Selected, SelectedFile{
			Path: best.file.Path, Hash: best.file.Hash,
			Start: int64(start), End: int64(end),
			ExcerptHash: hex.EncodeToString(hash[:]),
			Reason:      "rrf_coverage",
			Content:     visible,
			Ranks:       best.ranks,
			Covered:     newCovered,
		})
		selectedBytes += len(content)
		manifest.SelectedBytes += len(content)
		for _, c := range newCovered {
			coveredSet[c] = true
		}
		used[best.file.Path] = true
		selected = append(selected, best)
	}
	if len(manifest.Selected) == 0 {
		// No positive marginal coverage: explicit deterministic fallback on the
		// RRF-top file so the mode never returns an empty selection when
		// eligible bytes exist.
		top := cands[0]
		for _, c := range cands {
			if c.rrfRank < top.rrfRank || c.rrfRank == top.rrfRank && c.file.Path < top.file.Path {
				top = c
			}
		}
		return selectLegacyFallback(in, files, manifest, top.file)
	}
	// Omissions for unselected eligible files: redundant vs budget.
	for _, c := range cands {
		if used[c.file.Path] {
			continue
		}
		if len(manifest.Selected) >= in.Limits.MaxFiles {
			appendOmission(&manifest, in.Limits, c.file.Path, "file_limit")
			continue
		}
		remaining := in.Limits.MaxBytes - manifest.SelectedBytes
		if remaining < 128 || len(c.excerpt) > remaining {
			// Distinguish pure byte budget from already-covered redundancy.
			gain := 0
			for _, concept := range c.covered {
				if !coveredSet[concept] {
					gain++
				}
			}
			if gain <= 0 {
				appendOmission(&manifest, in.Limits, c.file.Path, "redundant")
			} else {
				appendOmission(&manifest, in.Limits, c.file.Path, "byte_limit")
			}
			continue
		}
		appendOmission(&manifest, in.Limits, c.file.Path, "redundant")
	}
	manifest.OmittedCount = len(files) - len(manifest.Selected)
	manifest.OmissionsTrimmed = manifest.OmittedCount > len(manifest.Omissions)
	return manifest, nil
}

func selectLegacyFallback(in Input, files []File, manifest Manifest, top File) (Manifest, error) {
	// Reuse safe excerpt logic pivoted at zero for the deterministic fallback.
	want := min(in.Limits.MaxBytesPerFile, in.Limits.MaxBytes)
	start, end := excerpt(top.Content, 0, want)
	if end-start < 1 {
		appendOmission(&manifest, in.Limits, top.Path, "empty")
		manifest.OmittedCount = len(files)
		manifest.OmissionsTrimmed = manifest.OmittedCount > len(manifest.Omissions)
		return manifest, nil
	}
	content := top.Content[start:end]
	hash := sha256.Sum256(content)
	manifest.Selected = append(manifest.Selected, SelectedFile{
		Path: top.Path, Hash: top.Hash,
		Start: int64(start), End: int64(end),
		ExcerptHash: hex.EncodeToString(hash[:]),
		Reason:      "deterministic_fallback",
		Content:     string(content),
	})
	manifest.SelectedBytes += len(content)
	for _, f := range files {
		if f.Path == top.Path {
			continue
		}
		if sensitivePath(f.Path) {
			continue // already recorded as redacted
		}
		appendOmission(&manifest, in.Limits, f.Path, "unranked")
	}
	manifest.OmittedCount = len(files) - len(manifest.Selected)
	manifest.OmissionsTrimmed = manifest.OmittedCount > len(manifest.Omissions)
	return manifest, nil
}
