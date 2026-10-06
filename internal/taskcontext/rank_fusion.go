package taskcontext

import (
	"errors"
	"math/big"
	"sort"
	"strings"
	"unicode/utf8"
)

// SelectorRRFCoverageV1 is the single experimental context-selector mode.
// Empty preserves exact historical/default behavior; unknown modes are
// rejected pre-run/pre-provider with no default promotion.
const SelectorRRFCoverageV1 = "rrf-coverage-v1"

// fusionK is the fixed RRF denominator offset. It is provenance from
// Cormack/Clarke/Buettcher SIGIR 2009 primary paper
// https://cormack.uwaterloo.ca/cormacksigir09-rrf.pdf (paper lines 39-40),
// not tuned on any Fabric benchmark. All rankers fuse equally unweighted as
// RRF(d) = sum_r 1/(k+rank_r(d)) with 1-based ranks.
const fusionK = 60

// Independent ranker identities. Each retains exact provenance/ordinals
// separately; no SCIP/PPR/SBFL/type relations are claimed.
const (
	rankerChanged      = "changed"
	rankerHint         = "hint"
	rankerPathToken    = "path_token"
	rankerContentLex   = "content_lexical"
	rankerAnchor       = "anchor"
	maxRankProvenance  = 5
	maxCoveredConcepts = 16
	// maxCoveredConceptLn bounds one covered concept in UTF-8 bytes
	// (2..64), matching validateProvenance and coverage admission.
	maxCoveredConceptLn = 64
)

// ValidateSelector admits only empty legacy and the single experimental mode.
func ValidateSelector(s string) error {
	if s == "" || s == SelectorRRFCoverageV1 {
		return nil
	}
	return errors.New("invalid task context selector")
}

func rankByPathSet(files []File, set map[string]bool) map[string]int {
	matched := []string{}
	for _, f := range files {
		if set[f.Path] {
			matched = append(matched, f.Path)
		}
	}
	if len(matched) == 0 {
		return nil
	}
	sort.Strings(matched)
	out := make(map[string]int, len(matched))
	for i, p := range matched {
		out[p] = i + 1
	}
	return out
}

func rankPathTokens(files []File, terms []string) map[string]int {
	type hit struct {
		path  string
		count int
	}
	hits := []hit{}
	for _, f := range files {
		lower := strings.ToLower(f.Path)
		n := 0
		for _, term := range terms {
			if strings.Contains(lower, term) {
				n++
			}
		}
		if n > 0 {
			hits = append(hits, hit{f.Path, n})
		}
	}
	if len(hits) == 0 {
		return nil
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].count != hits[j].count {
			return hits[i].count > hits[j].count
		}
		return hits[i].path < hits[j].path
	})
	out := make(map[string]int, len(hits))
	for i, h := range hits {
		out[h.path] = i + 1
	}
	return out
}

func rankContentLexical(files []File, terms []string) map[string]int {
	type hit struct {
		path     string
		distinct int
		total    int
	}
	termSet := map[string]bool{}
	for _, t := range terms {
		termSet[t] = true
	}
	hits := []hit{}
	for _, f := range files {
		if !utf8.Valid(f.Content) {
			continue
		}
		sample := f.Content
		if len(sample) > 32<<10 {
			cut := 32 << 10
			for cut > 0 && !utf8.RuneStart(sample[cut]) {
				cut--
			}
			sample = sample[:cut]
		}
		counts, _ := tokenLexicalIndex(string(sample))
		distinct, total := 0, 0
		for term := range termSet {
			if n, ok := counts[term]; ok {
				distinct++
				total += n
			}
		}
		if distinct > 0 {
			hits = append(hits, hit{f.Path, distinct, total})
		}
	}
	if len(hits) == 0 {
		return nil
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].distinct != hits[j].distinct {
			return hits[i].distinct > hits[j].distinct
		}
		if hits[i].total != hits[j].total {
			return hits[i].total > hits[j].total
		}
		return hits[i].path < hits[j].path
	})
	out := make(map[string]int, len(hits))
	for i, h := range hits {
		out[h.path] = i + 1
	}
	return out
}

func rankAnchors(files []File, anchors map[string]int) map[string]int {
	if len(anchors) == 0 {
		return nil
	}
	present := map[string]bool{}
	for _, f := range files {
		if _, ok := anchors[f.Path]; ok {
			present[f.Path] = true
		}
	}
	return rankByPathSet(files, present)
}

// fuseRRF combines independent ordinal rankings with equal unweighted RRF.
// Each ranking maps path to 1-based rank; absent means no signal from that
// ranker for that file. Empty/nil rankings (no signal) must be skipped by the
// caller. Scores are exact rationals (big.Rat); no floats enter canonical
// JSON. Ties break deterministically by path so equivalent input permutations
// preserve identity.
func fuseRRF(files []File, rankings []map[string]int) (ordered []File, scores map[string]*big.Rat) {
	byPath := make(map[string]File, len(files))
	for _, f := range files {
		byPath[f.Path] = f
	}
	scores = make(map[string]*big.Rat, len(files))
	for _, f := range files {
		sum := new(big.Rat)
		hit := false
		for _, r := range rankings {
			if r == nil {
				continue
			}
			if rank, ok := r[f.Path]; ok && rank >= 1 {
				term := new(big.Rat).SetFrac(big.NewInt(1), big.NewInt(int64(fusionK+rank)))
				sum.Add(sum, term)
				hit = true
			}
		}
		if hit {
			scores[f.Path] = sum
		}
	}
	ordered = make([]File, 0, len(scores))
	for path := range scores {
		ordered = append(ordered, byPath[path])
	}
	sort.Slice(ordered, func(i, j int) bool {
		ci := scores[ordered[i].Path].Cmp(scores[ordered[j].Path])
		if ci != 0 {
			return ci > 0
		}
		return ordered[i].Path < ordered[j].Path
	})
	return ordered, scores
}

func provenanceRanks(path string, named []struct {
	name    string
	ranking map[string]int
}) map[string]int {
	out := map[string]int{}
	for _, n := range named {
		if n.ranking == nil {
			continue
		}
		if r, ok := n.ranking[path]; ok {
			out[n.name] = r
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func validateProvenance(ranks map[string]int, covered []string) error {
	if len(ranks) > maxRankProvenance {
		return errors.New("invalid task context rank provenance")
	}
	for k, v := range ranks {
		switch k {
		case rankerChanged, rankerHint, rankerPathToken, rankerContentLex, rankerAnchor:
		default:
			return errors.New("invalid task context ranker")
		}
		if v < 1 || v > 4096 {
			return errors.New("invalid task context rank ordinal")
		}
	}
	if len(covered) > maxCoveredConcepts {
		return errors.New("invalid task context coverage provenance")
	}
	seen := map[string]bool{}
	for _, c := range covered {
		if len(c) < 2 || len(c) > maxCoveredConceptLn || !utf8.ValidString(c) || strings.TrimSpace(c) == "" {
			return errors.New("invalid task context covered concept")
		}
		if seen[c] {
			return errors.New("duplicate covered concept")
		}
		seen[c] = true
	}
	return nil
}
