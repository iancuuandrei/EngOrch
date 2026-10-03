package control

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/ri"
	"harness.local/engorch/internal/safepath"
	"harness.local/engorch/internal/taskcontext"
)

const (
	plannerContextGoSourceV1 = "go-source-context-v1"
	plannerContextGoSourceV2 = "go-source-context-v2"
)

const (
	plannerGoContextVersion       = 2
	plannerGoContextMaxFiles      = 24
	plannerGoContextMaxFileBytes  = 1 << 20
	plannerGoContextMaxInputBytes = 8 << 20
	plannerGoContextMaxGraphBytes = 512 << 10
	plannerGoContextMaxRecord     = 768 << 10
)

// PlannerGoContextRecord is immutable, pre-planning engineering evidence for
// one committed Go corpus. It is partial by design: omitted paths, unresolved
// calls, and absent files never make an absence or semantic-correctness claim.
// The parser is a pinned local executable, not a model/provider invocation.
type PlannerGoContextRecord struct {
	Version            int                       `json:"version"`
	Source             ri.Source                 `json:"source"`
	Query              string                    `json:"query"`
	QueryTruncated     bool                      `json:"query_truncated,omitempty"`
	QueryHash          string                    `json:"query_hash"`
	QueryLen           int                       `json:"query_len"`
	RIExecutableSHA256 string                    `json:"ri_executable_sha256"`
	InventoryFiles     int                       `json:"inventory_files"`
	AttemptedFiles     int                       `json:"attempted_files"`
	ReadFiles          int                       `json:"read_files"`
	OmittedCount       int                       `json:"omitted_count"`
	Omissions          []taskcontext.Omission    `json:"omissions,omitempty"`
	OmissionsTrimmed   bool                      `json:"omissions_trimmed,omitempty"`
	Sources            []repository.SourceDigest `json:"sources,omitempty"`
	Graph              *ri.GoEngineeringGraph    `json:"graph,omitempty"`
	Context            *ri.GoContextManifest     `json:"context,omitempty"`
	Unavailable        string                    `json:"unavailable,omitempty"`
	RecordID           string                    `json:"record_id"`
	GenerationMetadata *ri.GoGenerationMetadata  `json:"generation_metadata,omitempty"`
	GenerationContext  *ri.GoGenerationContext   `json:"generation_context,omitempty"`
}

// plannerGoContextPrompt is the model-visible view of a larger durable
// record. It excludes the full graph/fact corpus while retaining the exact
// selected evidence, its graph digest, and the immutable record identity.
type plannerGoContextPrompt struct {
	RecordID           string                     `json:"record_id"`
	Coverage           string                     `json:"coverage"`
	RIExecutableSHA256 string                     `json:"ri_executable_sha256"`
	GraphDigest        string                     `json:"graph_digest,omitempty"`
	Context            *ri.GoContextManifest      `json:"context,omitempty"`
	Unavailable        string                     `json:"unavailable,omitempty"`
	Generation         *plannerGoGenerationPrompt `json:"generation,omitempty"`
}

// plannerGoGenerationPrompt is the compact planner view of source-bound
// generation evidence. The durable record retains full metadata; this view
// deliberately excludes source inventories and retained source bytes.
type plannerGoGenerationPrompt struct {
	MetadataDigest string                     `json:"metadata_digest"`
	SourceID       string                     `json:"source_id"`
	Coverage       string                     `json:"coverage"`
	Bindings       []plannerGoGenerationChain `json:"bindings,omitempty"`
	Context        *ri.GoGenerationContext    `json:"context,omitempty"`
	Truncated      bool                       `json:"truncated,omitempty"`
	OmittedCount   int                        `json:"omitted_count,omitempty"`
}

type plannerGoGenerationChain struct {
	OwnerPath  string   `json:"owner_path"`
	OutputPath string   `json:"output_path"`
	Directive  string   `json:"directive"`
	ToolPaths  []string `json:"tool_paths,omitempty"`
	ToolRoles  []string `json:"tool_roles,omitempty"`
	ToolSHA256 []string `json:"tool_sha256,omitempty"`
}

func plannerGoGenerationPromptFor(metadata *ri.GoGenerationMetadata, generation *ri.GoGenerationContext) *plannerGoGenerationPrompt {
	if metadata == nil {
		return nil
	}
	view := &plannerGoGenerationPrompt{
		MetadataDigest: metadata.Digest,
		SourceID:       metadata.SourceID,
		Coverage:       metadata.Coverage,
		Context:        generation,
		Truncated:      metadata.Truncated,
		OmittedCount:   metadata.OmittedCount,
	}
	for _, binding := range metadata.Bindings {
		chain := plannerGoGenerationChain{OwnerPath: binding.GeneratorPath, OutputPath: binding.GeneratedPath, Directive: binding.Directive}
		for _, tool := range binding.ToolSources {
			chain.ToolPaths = append(chain.ToolPaths, tool.Path)
			chain.ToolRoles = append(chain.ToolRoles, tool.Role)
			chain.ToolSHA256 = append(chain.ToolSHA256, tool.Source.SHA256)
		}
		view.Bindings = append(view.Bindings, chain)
	}
	return view
}

func plannerGoContextPromptFor(record *PlannerGoContextRecord) *plannerGoContextPrompt {
	if record == nil {
		return nil
	}
	prompt := &plannerGoContextPrompt{RecordID: record.RecordID, Coverage: "PARTIAL", RIExecutableSHA256: record.RIExecutableSHA256, Unavailable: record.Unavailable}
	if record.Unavailable == "" && record.Graph != nil && record.Context != nil {
		prompt.GraphDigest = record.Graph.Digest
		context := *record.Context
		prompt.Context = &context
	}
	if record.GenerationMetadata != nil {
		prompt.Generation = plannerGoGenerationPromptFor(record.GenerationMetadata, record.GenerationContext)
	}
	return prompt
}

func plannerGoContextEnabled(s Snapshot) bool {
	return s.Creation.Execution != nil && (s.Creation.Execution.PlannerContext == plannerContextGoSourceV1 || s.Creation.Execution.PlannerContext == plannerContextGoSourceV2)
}

func plannerGoContextRecordID(rec PlannerGoContextRecord) (string, error) {
	rec.RecordID = ""
	domain := "harness.control.planner-go-context.v1"
	if rec.Version == 3 {
		domain = "harness.control.planner-go-context.v2"
	}
	return canonical.Hash(domain, rec)
}

// AdmitPlannerGoContext gathers one bounded, committed Go corpus and compiles
// it into read-only planning evidence before a planner runtime intent exists.
// Parser failures stop preparation; this function never falls back to a model
// or silently substitutes working-tree bytes.
func AdmitPlannerGoContext(ctx context.Context, path string) (PlannerGoContextRecord, error) {
	s, err := Inspect(path)
	if err != nil {
		return PlannerGoContextRecord{}, err
	}
	if s.State != "OBJECTIVE" || !plannerGoContextEnabled(s) {
		return PlannerGoContextRecord{}, errors.New("Go planner context is not enabled before planning")
	}
	if s.PlannerGoContext != nil {
		return *s.PlannerGoContext, nil
	}
	if s.PlannerContext != nil {
		return PlannerGoContextRecord{}, errors.New("conflicting planner context admission")
	}
	policy := s.Creation.Execution
	if policy == nil || policy.PlannerContextRIExecutable == "" || safepath.RequireDigest(policy.PlannerContextRIExecutableSHA256) != nil || !filepath.IsAbs(policy.PlannerContextRIExecutable) || filepath.Clean(policy.PlannerContextRIExecutable) != policy.PlannerContextRIExecutable {
		return PlannerGoContextRecord{}, errors.New("Go planner context requires an absolute pinned RI executable")
	}
	sourceID, err := s.Creation.Repository.ID()
	if err != nil {
		return PlannerGoContextRecord{}, err
	}
	source, err := ri.FromRepository(s.Creation.Repository)
	if err != nil {
		return PlannerGoContextRecord{}, err
	}
	query, truncated, queryHash, queryLen := truncateTaskQuery(s.Creation.Objective)
	client := ri.Client{Executable: policy.PlannerContextRIExecutable, ExecutableHash: policy.PlannerContextRIExecutableSHA256}
	corpus, err := ri.CollectCommittedGoCorpus(ctx, s.Creation.Repository, client, "", query)
	if err != nil {
		return PlannerGoContextRecord{}, err
	}
	record := PlannerGoContextRecord{
		Version: plannerGoContextVersion, Source: source, Query: query, QueryTruncated: truncated, QueryHash: queryHash, QueryLen: queryLen,
		RIExecutableSHA256: policy.PlannerContextRIExecutableSHA256, InventoryFiles: corpus.InventoryFiles, AttemptedFiles: corpus.AttemptedFiles,
		ReadFiles: corpus.ReadFiles, OmittedCount: corpus.OmittedCount, Omissions: append([]taskcontext.Omission(nil), corpus.Omissions...),
		OmissionsTrimmed: corpus.OmissionsTrimmed, Sources: append([]repository.SourceDigest(nil), corpus.Sources...),
	}
	// Corpus selection is relevance-ranked while durable source bindings are
	// path ordered. Discovery seeds this canonical admitted list so replay can
	// prove it was not pointed at a different generator search set.
	sort.Slice(record.Sources, func(i, j int) bool { return record.Sources[i].Path < record.Sources[j].Path })
	// An unavailable corpus has no admitted seed path. Preserve the legacy
	// unavailable representation rather than asking generation discovery to
	// manufacture an empty seed record.
	if policy.PlannerContext == plannerContextGoSourceV2 {
		record.Version = 3
		if record.ReadFiles > 0 {
			metadata, discoverErr := ri.DiscoverCommittedGoGenerators(ctx, s.Creation.Repository, sourcePaths(record.Sources))
			if discoverErr != nil {
				return PlannerGoContextRecord{}, discoverErr
			}
			if err := ri.ValidateGoGenerationMetadata(metadata); err != nil {
				return PlannerGoContextRecord{}, err
			}
			record.GenerationMetadata = &metadata
			if len(metadata.Bindings) > 0 {
				generation, compileErr := ri.CompileGoGenerationContext(metadata, query)
				if compileErr != nil {
					return PlannerGoContextRecord{}, compileErr
				}
				record.GenerationContext = &generation
			}
		}
	}
	if record.ReadFiles == 0 {
		record.Unavailable = corpus.Unavailable
		if record.Unavailable == "" {
			record.Unavailable = "no_eligible_complete_go_source"
		}
	} else {
		generators := append([]ri.GoGeneratorBinding(nil), corpus.Generators...)
		if record.GenerationMetadata != nil {
			generators = admittedGoGenerationBindings(*record.GenerationMetadata, record.Sources)
		}
		graph, buildErr := ri.BuildGoEngineeringGraph(ri.GoGraphSnapshotInput{SourceID: sourceID, ProducerSHA256: policy.PlannerContextRIExecutableSHA256, Files: corpus.GraphInputs, Generators: generators})
		if buildErr != nil {
			return PlannerGoContextRecord{}, buildErr
		}
		if err := boundedCanonical(graph, plannerGoContextMaxGraphBytes); err != nil {
			return PlannerGoContextRecord{}, errors.New("Go planner graph exceeds bounded admission size")
		}
		limits := taskcontext.DefaultLimits()
		if policy.PlannerContext == plannerContextGoSourceV2 {
			limits.MaxBytes = 24 << 10
		}
		manifest, compileErr := ri.CompileGoContext(ri.GoContextInput{ContextVersion: 2, SourceID: sourceID, Graph: graph, Objective: query, Files: corpus.ContextFiles, Limits: limits})
		if compileErr != nil {
			return PlannerGoContextRecord{}, compileErr
		}
		record.Graph, record.Context = &graph, &manifest
	}
	record.RecordID, err = plannerGoContextRecordID(record)
	if err != nil {
		return PlannerGoContextRecord{}, err
	}
	if err := validatePlannerGoContextRecord(s, record); err != nil {
		return PlannerGoContextRecord{}, err
	}
	if err := boundedCanonical(record, plannerGoContextMaxRecord); err != nil {
		return PlannerGoContextRecord{}, errors.New("Go planner context record exceeds bounded journal size")
	}
	if err := Append(path, "planner.go-context-admitted", record); err != nil {
		return PlannerGoContextRecord{}, err
	}
	return record, nil
}

func boundedCanonical(value any, maximum int) error {
	encoded, err := canonical.Bytes(value)
	if err != nil || len(encoded) > maximum {
		return errors.New("canonical value exceeds bound")
	}
	return nil
}

func sourcePaths(sources []repository.SourceDigest) []string {
	paths := make([]string, len(sources))
	for i, source := range sources {
		paths[i] = source.Path
	}
	return paths
}

// admittedGoGenerationBindings exposes a generation edge only when both ends
// are in the already admitted Go corpus and retain the exact committed digest
// observed by discovery. Tool provenance remains prompt-only evidence: no
// inferred generation edge is added for a supporting source.
func admittedGoGenerationBindings(metadata ri.GoGenerationMetadata, admitted []repository.SourceDigest) []ri.GoGeneratorBinding {
	byPath := make(map[string]repository.SourceDigest, len(admitted))
	for _, source := range admitted {
		byPath[source.Path] = source
	}
	bindings := make([]ri.GoGeneratorBinding, 0, len(metadata.Bindings))
	for _, binding := range metadata.Bindings {
		owner, ownerOK := byPath[binding.GeneratorPath]
		output, outputOK := byPath[binding.GeneratedPath]
		if !ownerOK || !outputOK || owner != binding.DirectiveSource || output != binding.OutputSource {
			continue
		}
		bindings = append(bindings, ri.GoGeneratorBinding{GeneratorPath: binding.GeneratorPath, GeneratedPath: binding.GeneratedPath, Directive: binding.Directive})
	}
	return bindings
}

func replayPlannerGoContext(s *Snapshot, rec PlannerGoContextRecord) error {
	if err := validatePlannerGoContextRecord(*s, rec); err != nil {
		return err
	}
	copy := rec
	copy.Sources = append([]repository.SourceDigest(nil), rec.Sources...)
	copy.Omissions = append([]taskcontext.Omission(nil), rec.Omissions...)
	s.PlannerGoContext = &copy
	return nil
}

func validatePlannerGoContextRecord(s Snapshot, rec PlannerGoContextRecord) error {
	wantVersion := plannerGoContextVersion
	if s.Creation.Execution != nil && s.Creation.Execution.PlannerContext == plannerContextGoSourceV2 {
		wantVersion = 3
	}
	if s.State != "OBJECTIVE" || !plannerGoContextEnabled(s) || s.PlannerGoContext != nil || s.PlannerContext != nil || rec.Version != wantVersion || safepath.RequireDigest(rec.RIExecutableSHA256) != nil || s.Creation.Execution == nil || rec.RIExecutableSHA256 != s.Creation.Execution.PlannerContextRIExecutableSHA256 {
		return errors.New("Go planner context transition rejected")
	}
	sourceID, err := s.Creation.Repository.ID()
	if err != nil {
		return err
	}
	if rec.Source.RepositoryID != sourceID || rec.Source.ObjectFormat != s.Creation.Repository.ObjectFormat || rec.Source.Commit != s.Creation.Repository.Commit || rec.Source.Tree != s.Creation.Repository.Tree {
		return errors.New("Go planner context source substitution")
	}
	query, truncated, queryHash, queryLen := truncateTaskQuery(s.Creation.Objective)
	if rec.Query != query || rec.QueryTruncated != truncated || rec.QueryHash != queryHash || rec.QueryLen != queryLen {
		return errors.New("Go planner context query substitution")
	}
	if rec.InventoryFiles < 0 || rec.InventoryFiles > 500000 || rec.AttemptedFiles < 0 || rec.AttemptedFiles > plannerGoContextMaxFiles || rec.AttemptedFiles > rec.InventoryFiles || rec.ReadFiles < 0 || rec.ReadFiles > rec.AttemptedFiles || rec.ReadFiles != len(rec.Sources) || rec.OmittedCount < 0 || rec.OmittedCount < len(rec.Omissions) || len(rec.Omissions) > 64 || (rec.OmissionsTrimmed && rec.OmittedCount <= len(rec.Omissions)) || (!rec.OmissionsTrimmed && rec.OmittedCount != len(rec.Omissions)) {
		return errors.New("invalid Go planner context inventory")
	}
	for _, omission := range rec.Omissions {
		if (omission.Path != "[redacted]" && safepath.Relative(omission.Path) != nil) || strings.TrimSpace(omission.Reason) == "" || len(omission.Reason) > 64 {
			return errors.New("invalid Go planner context omission")
		}
	}
	if rec.Unavailable != "" {
		if rec.Unavailable != "no_eligible_complete_go_source" || rec.ReadFiles != 0 || len(rec.Sources) != 0 || rec.Graph != nil || rec.Context != nil {
			return errors.New("invalid unavailable Go planner context")
		}
	} else {
		if rec.ReadFiles == 0 {
			return errors.New("empty Go planner corpus")
		}
		if err := validatePlannerGoSources(rec, sourceID); err != nil {
			return err
		}
		if rec.Graph == nil || rec.Context == nil {
			return errors.New("Go planner graph or context is missing")
		}
		if err := ri.ValidateGoEngineeringGraph(*rec.Graph); err != nil || rec.Graph.SourceID != sourceID || rec.Graph.CandidateID != "" || rec.Graph.ProducerSHA256 != rec.RIExecutableSHA256 {
			return errors.New("invalid Go planner graph")
		}
		if rec.Version == 2 && len(rec.Graph.Generators) != 0 {
			return errors.New("legacy Go planner graph carries generation bindings")
		}
		if err := boundedCanonical(*rec.Graph, plannerGoContextMaxGraphBytes); err != nil {
			return errors.New("Go planner graph exceeds bound")
		}
		contextBudget := 48 << 10
		if rec.Version == 3 {
			contextBudget = 24 << 10
		}
		if err := validatePlannerGoManifest(*rec.Context, *rec.Graph, rec.Sources, rec.Query, contextBudget); err != nil {
			return err
		}
	}
	if rec.Version == 2 && (rec.GenerationMetadata != nil || rec.GenerationContext != nil) {
		return errors.New("legacy Go planner context carries generation data")
	}
	if rec.Version == 3 {
		if rec.Unavailable != "" {
			if rec.GenerationMetadata != nil || rec.GenerationContext != nil {
				return errors.New("unavailable Go planner context carries generation data")
			}
			id, err := plannerGoContextRecordID(rec)
			if err != nil || id != rec.RecordID || safepath.RequireDigest(rec.RecordID) != nil {
				return errors.New("Go planner context record substitution")
			}
			return boundedCanonical(rec, plannerGoContextMaxRecord)
		}
		if rec.GenerationMetadata == nil || rec.GenerationMetadata.SourceID != sourceID || ri.ValidateGoGenerationMetadata(*rec.GenerationMetadata) != nil {
			return errors.New("invalid Go generation metadata")
		}
		if !reflect.DeepEqual(rec.GenerationMetadata.SeedPaths, sourcePaths(rec.Sources)) {
			return errors.New("Go generation metadata seed substitution")
		}
		for _, source := range rec.GenerationMetadata.Sources {
			if source.RepositoryID != sourceID || source.Commit != rec.Source.Commit {
				return errors.New("Go generation metadata source substitution")
			}
		}
		if rec.Graph != nil && !reflect.DeepEqual(rec.Graph.Generators, admittedGoGenerationBindings(*rec.GenerationMetadata, rec.Sources)) {
			return errors.New("Go generation graph binding substitution")
		}
		if len(rec.GenerationMetadata.Bindings) > 0 && rec.GenerationContext == nil {
			return errors.New("Go generation context is missing")
		}
		if rec.GenerationContext != nil {
			if rec.GenerationContext.SourceID != sourceID || ri.ValidateGoGenerationContext(*rec.GenerationContext, *rec.GenerationMetadata) != nil {
				return errors.New("invalid Go generation context")
			}
		}
	}
	id, err := plannerGoContextRecordID(rec)
	if err != nil || id != rec.RecordID || safepath.RequireDigest(rec.RecordID) != nil {
		return errors.New("Go planner context record substitution")
	}
	return boundedCanonical(rec, plannerGoContextMaxRecord)
}

func validatePlannerGoSources(rec PlannerGoContextRecord, sourceID string) error {
	seen := map[string]bool{}
	total := int64(0)
	for index, source := range rec.Sources {
		if safepath.Relative(source.Path) != nil || filepath.Ext(source.Path) != ".go" || source.RepositoryID != sourceID || source.Commit != rec.Source.Commit || safepath.RequireDigest(source.SHA256) != nil || source.Bytes < 0 || source.Bytes > plannerGoContextMaxFileBytes || (index > 0 && rec.Sources[index-1].Path >= source.Path) || seen[strings.ToLower(source.Path)] {
			return errors.New("invalid Go planner source binding")
		}
		seen[strings.ToLower(source.Path)] = true
		total += source.Bytes
	}
	if total > plannerGoContextMaxInputBytes || rec.Graph == nil || len(rec.Graph.Files) != len(rec.Sources) {
		return errors.New("Go planner source corpus exceeds bound")
	}
	bound := make(map[string]string, len(rec.Sources))
	for _, source := range rec.Sources {
		bound[source.Path] = source.SHA256
	}
	for _, file := range rec.Graph.Files {
		if bound[file.Facts.Path] == "" || bound[file.Facts.Path] != file.Facts.SourceSHA256 {
			return errors.New("Go planner graph source mismatch")
		}
	}
	return nil
}

func validatePlannerGoManifest(manifest ri.GoContextManifest, graph ri.GoEngineeringGraph, sources []repository.SourceDigest, query string, contextBudget int) error {
	if manifest.Schema != "engorch.ri.go-context.v2" || manifest.GraphDigest != graph.Digest || manifest.ProducerSHA256 != graph.ProducerSHA256 || manifest.Coverage != "PARTIAL" || safepath.RequireDigest(manifest.Digest) != nil {
		return errors.New("invalid Go planner context manifest")
	}
	querySum := sha256.Sum256([]byte(query))
	selection := manifest.Selection
	if contextBudget < 256 || contextBudget > 48<<10 || manifest.QuerySHA256 != hex.EncodeToString(querySum[:]) || selection.Scope.SourceID != graph.SourceID || selection.Scope.CandidateID != "" || len(selection.Selected) == 0 || len(selection.Selected) > 12 || selection.SelectedBytes < 0 || selection.SelectedBytes > contextBudget || selection.InputBytes < 0 || selection.SelectedBytes > selection.InputBytes || selection.Version != 1 || safepath.RequireDigest(selection.InputHash) != nil {
		return errors.New("Go planner context selection mismatch")
	}
	if _, err := selection.ID(); err != nil {
		return errors.New("invalid Go planner context selection identity")
	}
	files := make(map[string]ri.GoGraphFile, len(graph.Files))
	sizes := make(map[string]int64, len(sources))
	for _, file := range graph.Files {
		files[file.Facts.Path] = file
	}
	inputBytes := int64(0)
	for _, source := range sources {
		inputBytes += source.Bytes
		sizes[source.Path] = source.Bytes
	}
	if inputBytes > plannerGoContextMaxInputBytes || int64(selection.InputBytes) != inputBytes {
		return errors.New("Go planner context input byte mismatch")
	}
	if selection.OmittedCount < 0 || selection.OmittedCount > 24 || len(selection.Omissions) > 64 || (selection.OmissionsTrimmed && selection.OmittedCount <= len(selection.Omissions)) || (!selection.OmissionsTrimmed && selection.OmittedCount != len(selection.Omissions)) {
		return errors.New("invalid Go planner context selection omissions")
	}
	for _, omission := range selection.Omissions {
		if (omission.Path != "[redacted]" && safepath.Relative(omission.Path) != nil) || strings.TrimSpace(omission.Reason) == "" || len(omission.Reason) > 64 {
			return errors.New("invalid Go planner context selection omission")
		}
	}
	allExcerpts := append(append([]taskcontext.SelectedFile(nil), selection.Selected...), manifest.ContractExcerpts...)
	if len(manifest.ContractExcerpts) > 32 {
		return errors.New("Go planner context contract excerpts exceed bounds")
	}
	ranges := make(map[string][]taskcontext.SelectedFile, len(allExcerpts))
	selectedBytes := 0
	for _, selected := range selection.Selected {
		file, ok := files[selected.Path]
		if !ok || safepath.RequireDigest(selected.Hash) != nil || selected.Hash != file.Facts.SourceSHA256 || safepath.RequireDigest(selected.ExcerptHash) != nil || selected.Start < 0 || selected.End <= selected.Start || selected.End > sizes[selected.Path] || int64(len(selected.Content)) != selected.End-selected.Start || !utf8.ValidString(selected.Content) || !plannerGoSelectedReason(selected.Reason) || plannerGoOverlaps(selected, ranges[selected.Path]) {
			return errors.New("invalid Go planner context selected excerpt")
		}
		ranges[selected.Path] = append(ranges[selected.Path], selected)
		sum := sha256.Sum256([]byte(selected.Content))
		if hex.EncodeToString(sum[:]) != selected.ExcerptHash {
			return errors.New("Go planner context excerpt substitution")
		}
		selectedBytes += len(selected.Content)
	}
	if selectedBytes != selection.SelectedBytes {
		return errors.New("Go planner context selected byte mismatch")
	}
	contractBytes := 0
	for _, selected := range manifest.ContractExcerpts {
		file, ok := files[selected.Path]
		if !ok || safepath.RequireDigest(selected.Hash) != nil || selected.Hash != file.Facts.SourceSHA256 || safepath.RequireDigest(selected.ExcerptHash) != nil || selected.Start < 0 || selected.End <= selected.Start || selected.End > sizes[selected.Path] || int64(len(selected.Content)) != selected.End-selected.Start || !utf8.ValidString(selected.Content) || !plannerGoSelectedReason(selected.Reason) || plannerGoOverlaps(selected, ranges[selected.Path]) {
			return errors.New("invalid Go planner context contract excerpt")
		}
		ranges[selected.Path] = append(ranges[selected.Path], selected)
		sum := sha256.Sum256([]byte(selected.Content))
		if hex.EncodeToString(sum[:]) != selected.ExcerptHash {
			return errors.New("Go planner context contract excerpt substitution")
		}
		contractBytes += len(selected.Content)
	}
	if selectedBytes+contractBytes > contextBudget {
		return errors.New("Go planner context combined excerpt budget exceeded")
	}
	if len(manifest.Symbols) > 128 || len(manifest.Relations) > 128 {
		return errors.New("Go planner context facts exceed bounds")
	}
	for _, symbol := range manifest.Symbols {
		if symbol.Kind != "declaration" || !plannerGoVisibleNode(symbol, allExcerpts) || !plannerGoGraphHasNode(graph, symbol) {
			return errors.New("Go planner context symbol is not visible graph evidence")
		}
	}
	for _, relation := range manifest.Relations {
		if !plannerGoVisibleEdge(relation, allExcerpts) || !plannerGoGraphHasEdge(graph, relation) {
			return errors.New("Go planner context relation is not visible graph evidence")
		}
	}
	copy := manifest
	copy.Digest = ""
	id, err := canonical.Hash("harness.ri.go-context.v2", copy)
	if err != nil || id != manifest.Digest {
		return errors.New("Go planner context digest mismatch")
	}
	if err := boundedCanonical(manifest, 128<<10); err != nil {
		return errors.New("Go planner context manifest exceeds bound")
	}
	return nil
}

func plannerGoOverlaps(candidate taskcontext.SelectedFile, existing []taskcontext.SelectedFile) bool {
	for _, prior := range existing {
		if candidate.Start < prior.End && prior.Start < candidate.End {
			return true
		}
	}
	return false
}

func plannerGoSelectedReason(reason string) bool {
	switch reason {
	case "changed_path", "path_hint", "path_match", "lexical_match", "deterministic_fallback", "graph_anchor", "graph_caller":
		return true
	default:
		return false
	}
}

func plannerGoVisibleNode(node ri.GoGraphNode, selected []taskcontext.SelectedFile) bool {
	if node.Path == "" || node.Range == nil {
		return false
	}
	for _, file := range selected {
		if file.Path == node.Path && int64(node.Range.StartByte) >= file.Start && int64(node.Range.EndByte) <= file.End {
			return true
		}
	}
	return false
}

func plannerGoVisibleEdge(edge ri.GoGraphEdge, selected []taskcontext.SelectedFile) bool {
	if edge.Path == "" || edge.Range == nil {
		return false
	}
	for _, file := range selected {
		if file.Path == edge.Path && int64(edge.Range.StartByte) >= file.Start && int64(edge.Range.EndByte) <= file.End {
			return true
		}
	}
	return false
}

func plannerGoGraphHasNode(graph ri.GoEngineeringGraph, node ri.GoGraphNode) bool {
	for _, candidate := range graph.Nodes {
		if reflect.DeepEqual(candidate, node) {
			return true
		}
	}
	return false
}

func plannerGoGraphHasEdge(graph ri.GoEngineeringGraph, edge ri.GoGraphEdge) bool {
	for _, candidate := range graph.Edges {
		if reflect.DeepEqual(candidate, edge) {
			return true
		}
	}
	return false
}
