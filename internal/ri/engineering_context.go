package ri

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"unicode"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/safepath"
	"harness.local/engorch/internal/taskcontext"
)

// GoContextInput compiles a bounded prompt view from one graph and its exact
// source observations. It performs no filesystem reads or authorization.
type GoContextInput struct {
	SourceID     string
	Graph        GoEngineeringGraph
	Objective    string
	Files        []taskcontext.File
	ChangedPaths []string
	Limits       taskcontext.Limits
}

// GoContextManifest retains source excerpts and the partial graph observations
// behind their selection. Missing evidence never implies absence or safety.
type GoContextManifest struct {
	Schema         string               `json:"schema"`
	GraphDigest    string               `json:"graph_digest"`
	ProducerSHA256 string               `json:"producer_sha256"`
	Coverage       string               `json:"coverage"`
	QuerySHA256    string               `json:"query_sha256"`
	Selection      taskcontext.Manifest `json:"selection"`
	Symbols        []GoGraphNode        `json:"symbols"`
	Relations      []GoGraphEdge        `json:"relations"`
	FactsTruncated bool                 `json:"facts_truncated"`
	HintsTruncated bool                 `json:"hints_truncated"`
	Digest         string               `json:"digest"`
}

// CompileGoContext combines exact path/symbol seeds and bounded dependency
// hints with the existing lexical selector. Returned facts fit selected source
// spans; they are observations, not proof of semantic resolution or write scope.
func CompileGoContext(input GoContextInput) (GoContextManifest, error) {
	var empty GoContextManifest
	if err := ValidateGoEngineeringGraph(input.Graph); err != nil {
		return empty, err
	}
	if safepath.RequireDigest(input.SourceID) != nil || input.SourceID != input.Graph.SourceID || len(input.Files) != len(input.Graph.Files) {
		return empty, errors.New("Go context requires the complete bound graph source corpus")
	}
	expected := make(map[string]string, len(input.Graph.Files))
	for _, file := range input.Graph.Files {
		expected[file.Facts.Path] = file.Facts.SourceSHA256
	}
	seen := make(map[string]bool, len(input.Files))
	for _, file := range input.Files {
		hash := sha256.Sum256(file.Content)
		if seen[file.Path] || expected[file.Path] == "" || expected[file.Path] != file.Hash || file.Hash != hex.EncodeToString(hash[:]) {
			return empty, errors.New("Go context source bytes differ from graph")
		}
		seen[file.Path] = true
	}
	// Full identifiers rather than arbitrary substrings seed graph hints.
	terms := make(map[string]bool)
	for _, word := range strings.FieldsFunc(input.Objective, func(r rune) bool {
		return !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r))
	}) {
		terms[strings.ToLower(word)] = true
	}
	hints := make(map[string]bool)
	anchors := make(map[string]int)
	for _, node := range input.Graph.Nodes {
		if node.Kind == "declaration" && terms[strings.ToLower(node.Label)] && taskcontext.EligiblePath(node.Path) {
			hints[node.Path] = true
			if node.Range != nil {
				if prior, ok := anchors[node.Path]; !ok || node.Range.StartByte < prior {
					anchors[node.Path] = node.Range.StartByte
				}
			}
		}
	}
	for _, changed := range input.ChangedPaths {
		if !seen[changed] {
			return empty, errors.New("Go context changed path absent from bound corpus")
		}
	}
	seeds := append([]string(nil), input.ChangedPaths...)
	for path := range hints {
		seeds = append(seeds, path)
	}
	sort.Strings(seeds)
	seeds = compactContextPaths(seeds)
	hintsTruncated := len(seeds) > goEngineeringMaxSeeds
	if len(seeds) > goEngineeringMaxSeeds {
		seeds = seeds[:goEngineeringMaxSeeds]
	}
	if len(seeds) > 0 {
		impact, err := QueryGoImpact(input.Graph, GoImpactQuery{Paths: seeds, MaxDepth: 2, Limit: goEngineeringMaxSeeds})
		if err != nil {
			return empty, err
		}
		hintsTruncated = hintsTruncated || impact.Truncated
		for _, path := range impact.Paths {
			if taskcontext.EligiblePath(path) {
				hints[path] = true
			}
		}
	}
	orderedHints := make([]string, 0, len(hints))
	for path := range hints {
		orderedHints = append(orderedHints, path)
	}
	sort.Strings(orderedHints)
	if len(orderedHints) > 512 {
		orderedHints = orderedHints[:512]
		hintsTruncated = true
	}
	anchorPaths := make([]string, 0, len(anchors))
	for path := range anchors {
		anchorPaths = append(anchorPaths, path)
	}
	sort.Strings(anchorPaths)
	if len(anchorPaths) > goEngineeringMaxSeeds {
		hintsTruncated = true
		for _, path := range anchorPaths[goEngineeringMaxSeeds:] {
			delete(anchors, path)
		}
	}
	selection, err := taskcontext.Select(taskcontext.Input{Version: 1, Scope: taskcontext.Scope{SourceID: input.SourceID, CandidateID: input.Graph.CandidateID}, Objective: input.Objective, Files: input.Files, ChangedPaths: input.ChangedPaths, PathHints: orderedHints, Anchors: anchors, Limits: input.Limits})
	if err != nil {
		return empty, err
	}
	queryHash := sha256.Sum256([]byte(input.Objective))
	manifest := GoContextManifest{Schema: "engorch.ri.go-context.v1", GraphDigest: input.Graph.Digest, ProducerSHA256: input.Graph.ProducerSHA256, Coverage: "PARTIAL", QuerySHA256: hex.EncodeToString(queryHash[:]), Selection: selection, Symbols: []GoGraphNode{}, Relations: []GoGraphEdge{}, HintsTruncated: hintsTruncated}
	selected := make(map[string]taskcontext.SelectedFile, len(selection.Selected))
	for _, file := range selection.Selected {
		selected[file.Path] = file
	}
	visible := func(path string, span *GoRange) bool {
		file, ok := selected[path]
		if !ok {
			return false
		}
		return span == nil || int64(span.StartByte) >= file.Start && int64(span.EndByte) <= file.End
	}
	// Hard count ceilings bound structural metadata separately from source bytes.
	for _, node := range input.Graph.Nodes {
		if node.Kind == "declaration" && visible(node.Path, node.Range) {
			if len(manifest.Symbols) == 128 {
				manifest.FactsTruncated = true
				continue
			}
			manifest.Symbols = append(manifest.Symbols, cloneGoGraphNode(node))
		}
	}
	for _, edge := range input.Graph.Edges {
		if visible(edge.Path, edge.Range) {
			if len(manifest.Relations) == 128 {
				manifest.FactsTruncated = true
				continue
			}
			manifest.Relations = append(manifest.Relations, cloneGoGraphEdge(edge))
		}
	}
	manifest.Digest, err = canonical.Hash("harness.ri.go-context.v1", manifest)
	if err != nil {
		return empty, err
	}
	encoded, err := canonical.Bytes(manifest)
	if err != nil || len(encoded) > 128<<10 {
		return empty, errors.New("Go context manifest exceeds 128 KiB output budget")
	}
	return manifest, nil
}

func compactContextPaths(paths []string) []string {
	out := paths[:0]
	for _, path := range paths {
		if len(out) == 0 || out[len(out)-1] != path {
			out = append(out, path)
		}
	}
	return out
}
