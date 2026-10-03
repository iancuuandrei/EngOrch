package ri

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/taskcontext"
)

func contextFixture(t *testing.T) GoContextInput {
	t.Helper()
	producer := strings.Repeat("a", 64)
	source := []byte("package p\nfunc F() {}\n")
	hash := sha256.Sum256(source)
	facts := fixtureGoFacts("pkg/file.go", hex.EncodeToString(hash[:]), producer)
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: strings.Repeat("c", 64), CandidateID: strings.Repeat("b", 64), ProducerSHA256: producer, Files: []GoGraphFileInput{{Facts: facts, Source: source, Package: GoPackageBinding{ImportPath: "example/p", ModulePath: "example"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return GoContextInput{SourceID: strings.Repeat("c", 64), Graph: graph, Objective: "Fix F", Files: []taskcontext.File{{Path: facts.Path, Hash: facts.SourceSHA256, Content: source}}, Limits: taskcontext.DefaultLimits()}
}

func TestCompileGoContextBindsExactGraphSourceAndQuery(t *testing.T) {
	input := contextFixture(t)
	first, err := CompileGoContext(input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := CompileGoContext(input)
	if err != nil || first.Digest != second.Digest {
		t.Fatalf("nondeterministic manifest: %v", err)
	}
	if first.Coverage != "PARTIAL" || first.GraphDigest != input.Graph.Digest || len(first.Selection.Selected) != 1 || len(first.Symbols) != 1 || first.Symbols[0].Label != "F" {
		t.Fatalf("incorrect evidence: %+v", first)
	}
	before := input.Objective
	input.Objective = "Explain F"
	changed, err := CompileGoContext(input)
	if err != nil || first.QuerySHA256 == changed.QuerySHA256 || first.Digest == changed.Digest {
		t.Fatalf("query not bound: %v", err)
	}
	input.Objective = before
	input.Files[0].Content = []byte("dirty working tree")
	if _, err := CompileGoContext(input); err == nil {
		t.Fatal("unbound source accepted")
	}
}

func TestCompileGoContextRejectsMissingDuplicateCorruptAndForeignEvidence(t *testing.T) {
	for _, which := range []string{"source ID", "missing", "duplicate", "graph digest", "foreign changed path"} {
		t.Run(which, func(t *testing.T) {
			input := contextFixture(t)
			switch which {
			case "source ID":
				input.SourceID = "unbound"
			case "missing":
				input.Files = nil
			case "duplicate":
				input.Files = append(input.Files, input.Files[0])
			case "graph digest":
				input.Graph.Digest = strings.Repeat("d", 64)
			case "foreign changed path":
				input.ChangedPaths = []string{"foreign.go"}
			}
			if _, err := CompileGoContext(input); err == nil {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
}

func TestGoContextDigestAndReturnedValuesAreIndependent(t *testing.T) {
	input := contextFixture(t)
	manifest, err := CompileGoContext(input)
	if err != nil {
		t.Fatal(err)
	}
	digest := manifest.Digest
	manifest.Digest = ""
	calculated, err := canonical.Hash("harness.ri.go-context.v1", manifest)
	if err != nil || calculated != digest {
		t.Fatalf("manifest digest mismatch: %v", err)
	}
	manifest.Symbols[0].Range.StartByte++
	if err := ValidateGoEngineeringGraph(input.Graph); err != nil {
		t.Fatalf("returned ranges mutated graph: %v", err)
	}
}

func TestCompileGoContextRetrievesNamedDeclarationAheadOfLexicalNoise(t *testing.T) {
	producer := strings.Repeat("a", 64)
	wanted := "package p\nfunc Target() {}\n"
	noise := "package p\n// Target Target Target Target\nfunc Other() {}\n"
	inputs := []GoGraphFileInput{
		graphInput("src/target.go", wanted, GoPackageBinding{ImportPath: "example/p", ModulePath: "example"}, []GoSymbol{{Name: "Target", Kind: "function_declaration", Range: graphSpan(wanted, "Target", 0)}}, nil, nil, nil, producer),
		graphInput("src/noise.go", noise, GoPackageBinding{ImportPath: "example/p", ModulePath: "example"}, []GoSymbol{{Name: "Other", Kind: "function_declaration", Range: graphSpan(noise, "Other", 0)}}, nil, nil, nil, producer),
	}
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: strings.Repeat("c", 64), ProducerSHA256: producer, Files: inputs})
	if err != nil {
		t.Fatal(err)
	}
	files := []taskcontext.File{}
	for _, in := range inputs {
		files = append(files, taskcontext.File{Path: in.Facts.Path, Hash: in.Facts.SourceSHA256, Content: in.Source})
	}
	limits := taskcontext.DefaultLimits()
	limits.MaxFiles = 1
	input := GoContextInput{SourceID: graph.SourceID, Graph: graph, Objective: "Fix Target", Files: files, Limits: limits}
	manifest, err := CompileGoContext(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Selection.Selected) != 1 || manifest.Selection.Selected[0].Path != "src/target.go" || len(manifest.Symbols) != 1 || manifest.Symbols[0].Label != "Target" || manifest.Selection.OmittedCount != 1 {
		t.Fatalf("retrieval ground truth mismatch: %+v", manifest)
	}
	input.Files[0], input.Files[1] = input.Files[1], input.Files[0]
	reordered, err := CompileGoContext(input)
	if err != nil || reordered.Digest != manifest.Digest {
		t.Fatalf("corpus order changed selection: %v", err)
	}
	if manifest.Selection.Scope.CandidateID != "" {
		t.Fatal("base context falsely claims candidate identity")
	}
}

func TestCompileGoContextAnchorsLateUnicodeDeclaration(t *testing.T) {
	producer := strings.Repeat("a", 64)
	source := "package p\n" + strings.Repeat("// unrelated padding\n", 3000) + "func Cafés() {}\n"
	file := graphInput("src/late.go", source, GoPackageBinding{ImportPath: "example/p", ModulePath: "example"}, []GoSymbol{{Name: "Cafés", Kind: "function_declaration", Range: graphSpan(source, "Cafés", 0)}}, nil, nil, nil, producer)
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: strings.Repeat("c", 64), ProducerSHA256: producer, Files: []GoGraphFileInput{file}})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := CompileGoContext(GoContextInput{SourceID: graph.SourceID, Graph: graph, Objective: "Fix Cafés", Files: []taskcontext.File{{Path: file.Facts.Path, Hash: file.Facts.SourceSHA256, Content: file.Source}}, Limits: taskcontext.DefaultLimits()})
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Selection.Selected) != 1 || manifest.Selection.Selected[0].Start <= 32<<10 || !strings.Contains(manifest.Selection.Selected[0].Content, "func Cafés") || len(manifest.Symbols) != 1 {
		t.Fatalf("late declaration not visible in compiled context: %+v", manifest)
	}
}

func TestCompileGoContextTruncatesExcessDeclarationHints(t *testing.T) {
	producer := strings.Repeat("a", 64)
	source := []byte("package p\nfunc F() {}\n")
	sum := sha256.Sum256(source)
	files := make([]GoGraphFileInput, 0, 513)
	corpus := make([]taskcontext.File, 0, 513)
	for index := 0; index < 513; index++ {
		path := fmt.Sprintf("src/f%03d.go", index)
		fact := fixtureGoFacts(path, hex.EncodeToString(sum[:]), producer)
		files = append(files, GoGraphFileInput{Facts: fact, Source: source, Package: GoPackageBinding{ImportPath: "example/p", ModulePath: "example"}})
		corpus = append(corpus, taskcontext.File{Path: path, Hash: fact.SourceSHA256, Content: source})
	}
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: strings.Repeat("c", 64), ProducerSHA256: producer, Files: files})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := CompileGoContext(GoContextInput{SourceID: graph.SourceID, Graph: graph, Objective: "Fix F", Files: corpus, Limits: taskcontext.DefaultLimits()})
	if err != nil {
		t.Fatal(err)
	}
	if !manifest.HintsTruncated || len(manifest.Selection.Selected) != 12 {
		t.Fatal("hint overflow was not bounded and recorded")
	}
}
