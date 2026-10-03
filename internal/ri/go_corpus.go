package ri

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/safepath"
	"harness.local/engorch/internal/taskcontext"
)

const (
	goCorpusMaxFiles       = 24
	goCorpusMaxFileBytes   = 1 << 20
	goCorpusMaxTotalBytes  = 8 << 20
	goCorpusMaxGraphBytes  = 512 << 10
	goCorpusMaxOmissions   = 64
	goCorpusSourceLocalV1  = "source_local_v1"
	goCorpusNoEligibleCode = "no_eligible_complete_go_source"
)

// GoCommittedCorpus is an immutable, bounded source-local Go corpus. Counts
// and omissions preserve incomplete coverage; successful parsing does not
// establish Go type resolution or semantic completeness.
type GoCommittedCorpus struct {
	Source         Source                    `json:"source"`
	Sources        []repository.SourceDigest `json:"sources"`
	GraphInputs    []GoGraphFileInput        `json:"graph_inputs"`
	ContextFiles   []taskcontext.File        `json:"context_files,omitempty"`
	Generators     []GoGeneratorBinding      `json:"generators"`
	InventoryFiles int                       `json:"inventory_files"`
	AttemptedFiles int                       `json:"attempted_files"`
	// ReadFiles counts complete files retained after parser-request and graph
	// admission bounds, not merely files whose committed bytes were observed.
	ReadFiles        int                    `json:"read_files"`
	OmittedCount     int                    `json:"omitted_count"`
	Omissions        []taskcontext.Omission `json:"omissions"`
	OmissionsTrimmed bool                   `json:"omissions_trimmed"`
	Unavailable      string                 `json:"unavailable,omitempty"`
	ModuleInventory  *GoModuleInventory     `json:"module_inventory,omitempty"`
}

// GoCorpusOptions opts into committed module declarations. Nil preserves the
// original source-local collection recipe and graph identities.
type GoCorpusOptions struct {
	ModuleInventory *GoModuleInventory
}

type goCorpusCandidate struct {
	entry repository.SourceEntry
	score int
}

type goCorpusCommittedFile struct {
	entry       repository.SourceEntry
	digest      repository.SourceDigest
	content     []byte
	packageInfo GoPackageBinding
}

type goCorpusBuffer struct {
	bytes.Buffer
	limit int64
}

// Write rejects any bytes beyond the exact object size admitted by the Git
// batch header.
func (b *goCorpusBuffer) Write(value []byte) (int, error) {
	if int64(len(value)) > b.limit-int64(b.Len()) {
		return 0, errors.New("committed Go source exceeds admitted corpus bound")
	}
	return b.Buffer.Write(value)
}

type goCorpusSink struct {
	buffer *goCorpusBuffer
}

// Write retains only bytes for a selected blob whose complete size was admitted.
func (s *goCorpusSink) Write(value []byte) (int, error) {
	return s.buffer.Write(value)
}

// Close satisfies io.WriteCloser without taking ownership of the shared buffer.
func (s *goCorpusSink) Close() error { return nil }

// CollectCommittedGoCorpus inventories the exact Git tree, chooses a bounded
// deterministic Go corpus, copies only selected objects through one cat-file
// batch process, and parses their facts through one pinned RI stream. cacheDir
// is intentionally required to be empty for the controller's first recipe.
func CollectCommittedGoCorpus(ctx context.Context, identity repository.Identity, client Client, cacheDir, objective string) (GoCommittedCorpus, error) {
	return CollectCommittedGoCorpusWithOptions(ctx, identity, client, cacheDir, objective, GoCorpusOptions{})
}

// CollectCommittedGoCorpusWithOptions admits an explicit source-bound module
// inventory for declared package ownership. It never establishes active Go
// build resolution, and malformed or omitted manifest coverage stays partial.
func CollectCommittedGoCorpusWithOptions(ctx context.Context, identity repository.Identity, client Client, cacheDir, objective string, options GoCorpusOptions) (GoCommittedCorpus, error) {
	corpus := GoCommittedCorpus{Sources: []repository.SourceDigest{}, GraphInputs: []GoGraphFileInput{}, Generators: []GoGeneratorBinding{}, Omissions: []taskcontext.Omission{}}
	if options.ModuleInventory != nil {
		if err := ValidateGoModuleInventory(*options.ModuleInventory, identity); err != nil {
			return corpus, err
		}
		corpus.ModuleInventory = options.ModuleInventory
	}
	if cacheDir != "" {
		return corpus, errors.New("committed Go corpus controller path does not permit an RI cache")
	}
	if len(objective) == 0 || len(objective) > 16<<10 || !utf8.ValidString(objective) {
		return corpus, errors.New("Go corpus objective is empty, oversized, or invalid UTF-8")
	}
	source, err := FromRepository(identity)
	if err != nil {
		return corpus, err
	}
	corpus.Source = source
	stream, err := client.OpenStream(ctx)
	if err != nil {
		return corpus, err
	}
	defer stream.Close()

	terms := goCorpusTerms(objective)
	candidates := make([]goCorpusCandidate, 0, 128)
	appendOmission := func(path, reason string) {
		corpus.OmittedCount++
		if len(corpus.Omissions) >= goCorpusMaxOmissions {
			corpus.OmissionsTrimmed = true
			return
		}
		if path != "" && (!taskcontext.EligiblePath(path) || safepath.Relative(path) != nil) {
			path = "[redacted]"
		}
		corpus.Omissions = append(corpus.Omissions, taskcontext.Omission{Path: path, Reason: reason})
	}
	err = repository.VisitSource(ctx, identity, func(entry repository.SourceEntry) error {
		if filepath.Ext(entry.Path) != ".go" {
			return nil
		}
		corpus.InventoryFiles++
		if entry.Kind != "file" {
			appendOmission(entry.Path, "unsupported_source_kind")
			return nil
		}
		if safepath.Relative(entry.Path) != nil {
			appendOmission(entry.Path, "unsafe_path")
			return nil
		}
		if !taskcontext.EligiblePath(entry.Path) {
			appendOmission(entry.Path, "sensitive_path")
			return nil
		}
		if err := safepath.Writable(entry.Path); err != nil {
			appendOmission(entry.Path, "protected_path")
			return nil
		}
		candidates = append(candidates, goCorpusCandidate{entry: entry, score: goCorpusPathScore(entry.Path, terms)})
		return nil
	})
	if err != nil {
		return GoCommittedCorpus{}, err
	}
	selected := selectGoCorpusCandidates(candidates)
	corpus.AttemptedFiles = len(selected)
	selectedSet := make(map[string]bool, len(selected))
	for _, candidate := range selected {
		selectedSet[candidate.entry.Path] = true
	}
	for _, candidate := range candidates {
		if !selectedSet[candidate.entry.Path] {
			appendOmission(candidate.entry.Path, "file_budget")
		}
	}
	if len(selected) == 0 {
		corpus.Unavailable = goCorpusNoEligibleCode
		return corpus, nil
	}

	buffers := make(map[string]*goCorpusBuffer, len(selected))
	discarded := make(map[string]string, len(selected))
	usedBytes := int64(0)
	graphInputs := make([]GoGraphFileInput, 0, len(selected))
	sources := make([]repository.SourceDigest, 0, len(selected))
	readContextFiles := make([]taskcontext.File, 0, len(selected))
	committed := make([]goCorpusCommittedFile, 0, len(selected))
	err = repository.CopySelectedSourceBatch(ctx, identity, goCorpusEntries(selected), func(entry repository.SourceEntry, size int64) (io.WriteCloser, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if size > goCorpusMaxFileBytes {
			discarded[entry.Path] = "file_byte_limit"
			return nil, nil
		}
		if size > int64(goCorpusMaxTotalBytes)-usedBytes {
			discarded[entry.Path] = "corpus_byte_limit"
			return nil, nil
		}
		buffer := &goCorpusBuffer{limit: size}
		buffers[entry.Path] = buffer
		usedBytes += size
		return &goCorpusSink{buffer: buffer}, nil
	}, func(entry repository.SourceEntry, digest *repository.SourceDigest) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if reason := discarded[entry.Path]; reason != "" {
			appendOmission(entry.Path, reason)
			return nil
		}
		buffer := buffers[entry.Path]
		if buffer == nil || digest == nil || digest.RepositoryID != source.RepositoryID || digest.Commit != source.Commit || digest.Path != entry.Path || digest.Blob != entry.Object || digest.Bytes != int64(buffer.Len()) {
			return fmt.Errorf("selected committed Go source binding mismatch: %q", entry.Path)
		}
		content := buffer.Bytes()
		delete(buffers, entry.Path)
		if !utf8.Valid(content) {
			appendOmission(entry.Path, "non_utf8")
			return nil
		}
		fileSet := token.NewFileSet()
		file, parseErr := parser.ParseFile(fileSet, entry.Path, content, parser.PackageClauseOnly)
		if parseErr != nil || file == nil || file.Name == nil || file.Name.Name == "" {
			appendOmission(entry.Path, "invalid_package_clause")
			return nil
		}
		binding, bindingErr := SourceLocalGoPackageBinding(source.RepositoryID, entry.Path, file.Name.Name)
		if options.ModuleInventory != nil {
			binding, bindingErr = DeclaredGoPackageBinding(*options.ModuleInventory, entry.Path, file.Name.Name)
		}
		if bindingErr != nil {
			return bindingErr
		}
		if binding.IdentityKind == goCorpusSourceLocalV1 && strings.HasSuffix(entry.Path, "_test.go") && strings.HasSuffix(file.Name.Name, "_test") {
			basePackageName := strings.TrimSuffix(file.Name.Name, "_test")
			baseBinding, err := SourceLocalGoPackageBinding(source.RepositoryID, entry.Path, basePackageName)
			if err != nil {
				return err
			}
			binding.TestOfPackageIdentity = baseBinding.PackageIdentity
		}
		committed = append(committed, goCorpusCommittedFile{entry: entry, digest: *digest, content: content, packageInfo: binding})
		return nil
	})
	if err != nil {
		return GoCommittedCorpus{}, err
	}
	selectedRank := make(map[string]int, len(selected))
	for index, candidate := range selected {
		selectedRank[candidate.entry.Path] = index
	}
	sort.Slice(committed, func(i, j int) bool {
		return selectedRank[committed[i].entry.Path] < selectedRank[committed[j].entry.Path]
	})
	for _, sourceFile := range committed {
		if err := ctx.Err(); err != nil {
			return GoCommittedCorpus{}, err
		}
		request, _, requestErr := goFileFactsRequest(sourceFile.entry.Path, sourceFile.content, "", stream.producerHash)
		if requestErr != nil {
			return GoCommittedCorpus{}, fmt.Errorf("committed Go facts request failed for %q: %w", sourceFile.entry.Path, requestErr)
		}
		if _, requestErr = canonical.Bytes(request); requestErr != nil {
			appendOmission(sourceFile.entry.Path, "facts_request_byte_limit")
			continue
		}
		facts, factsErr := stream.GoFileFacts(ctx, sourceFile.entry.Path, sourceFile.content, "")
		if factsErr != nil {
			return GoCommittedCorpus{}, fmt.Errorf("committed Go facts failed for %q: %w", sourceFile.entry.Path, factsErr)
		}
		input := GoGraphFileInput{Facts: facts, Source: sourceFile.content, Package: sourceFile.packageInfo}
		trial := make([]GoGraphFileInput, 0, len(graphInputs)+1)
		trial = append(trial, graphInputs...)
		trial = append(trial, input)
		fits, fitErr := goCorpusGraphFitsWithModules(source.RepositoryID, stream.producerHash, trial, options.ModuleInventory)
		if fitErr != nil {
			return GoCommittedCorpus{}, fitErr
		}
		if !fits {
			appendOmission(sourceFile.entry.Path, "graph_byte_budget")
			continue
		}
		graphInputs = append(graphInputs, input)
		sources = append(sources, sourceFile.digest)
		readContextFiles = append(readContextFiles, taskcontext.File{Path: sourceFile.entry.Path, Hash: sourceFile.digest.SHA256, Content: sourceFile.content})
		corpus.ReadFiles++
	}
	corpus.Sources = sources
	corpus.GraphInputs = graphInputs
	corpus.ContextFiles = readContextFiles
	if corpus.ReadFiles == 0 {
		corpus.Unavailable = goCorpusNoEligibleCode
	}
	return corpus, nil
}

func goCorpusGraphFits(sourceID, producer string, files []GoGraphFileInput) (bool, error) {
	return goCorpusGraphFitsWithModules(sourceID, producer, files, nil)
}

func goCorpusGraphFitsWithModules(sourceID, producer string, files []GoGraphFileInput, inventory *GoModuleInventory) (bool, error) {
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: sourceID, ProducerSHA256: producer, Files: files, Generators: []GoGeneratorBinding{}, ModuleInventory: inventory})
	if err != nil {
		if err.Error() == "JSON size or UTF-8 invalid" || strings.Contains(err.Error(), "exceeds") {
			return false, nil
		}
		return false, fmt.Errorf("committed Go graph admission failed: %w", err)
	}
	encoded, err := json.Marshal(graph)
	if err != nil {
		return false, err
	}
	canonicalBytes, err := canonical.Normalize(encoded)
	if err != nil {
		if err.Error() == "JSON size or UTF-8 invalid" {
			return false, nil
		}
		return false, err
	}
	return len(canonicalBytes) <= goCorpusMaxGraphBytes, nil
}

func goCorpusEntries(candidates []goCorpusCandidate) []repository.SourceEntry {
	entries := make([]repository.SourceEntry, len(candidates))
	for i, candidate := range candidates {
		entries[i] = candidate.entry
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries
}

func goCorpusTerms(objective string) []string {
	terms := strings.FieldsFunc(strings.ToLower(objective), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_')
	})
	sort.Strings(terms)
	result := terms[:0]
	for _, term := range terms {
		if len(term) > 1 && (len(result) == 0 || result[len(result)-1] != term) {
			result = append(result, term)
		}
	}
	return result
}

func goCorpusPathScore(filePath string, terms []string) int {
	folded := strings.ToLower(filePath)
	score := 0
	for _, term := range terms {
		if strings.Contains(folded, term) {
			score++
		}
		if (term == "generate" || term == "generated" || term == "generator") && (strings.Contains(folded, "gen") || strings.Contains(folded, "template")) {
			score++
		}
	}
	return score
}

func selectGoCorpusCandidates(candidates []goCorpusCandidate) []goCorpusCandidate {
	ordered := append([]goCorpusCandidate(nil), candidates...)
	better := func(left, right goCorpusCandidate) bool {
		if left.score != right.score {
			return left.score > right.score
		}
		return left.entry.Path < right.entry.Path
	}
	sort.Slice(ordered, func(i, j int) bool { return better(ordered[i], ordered[j]) })
	if len(ordered) <= goCorpusMaxFiles {
		return ordered
	}
	seedCount := 16
	if len(ordered) < seedCount {
		seedCount = len(ordered)
	}
	seeds := append([]goCorpusCandidate(nil), ordered[:seedCount]...)
	selected := append([]goCorpusCandidate(nil), seeds...)
	seen := make(map[string]bool, goCorpusMaxFiles)
	for _, seed := range seeds {
		seen[seed.entry.Path] = true
	}
	directories := make(map[string]bool)
	for _, seed := range seeds {
		directories[path.Dir(seed.entry.Path)] = true
	}
	related := make([]goCorpusCandidate, 0)
	for _, candidate := range ordered[seedCount:] {
		if !directories[path.Dir(candidate.entry.Path)] {
			continue
		}
		related = append(related, candidate)
	}
	sort.Slice(related, func(i, j int) bool {
		iTest := strings.HasSuffix(related[i].entry.Path, "_test.go")
		jTest := strings.HasSuffix(related[j].entry.Path, "_test.go")
		if iTest != jTest {
			return iTest
		}
		return better(related[i], related[j])
	})
	for _, candidate := range related {
		if len(selected) == goCorpusMaxFiles {
			break
		}
		if seen[candidate.entry.Path] {
			continue
		}
		seen[candidate.entry.Path] = true
		selected = append(selected, candidate)
	}
	for _, candidate := range ordered[seedCount:] {
		if len(selected) == goCorpusMaxFiles {
			break
		}
		if seen[candidate.entry.Path] {
			continue
		}
		seen[candidate.entry.Path] = true
		selected = append(selected, candidate)
	}
	return selected
}
