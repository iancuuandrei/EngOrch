package ri

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestQueryGoSemanticGraphVocabularyIsBoundedAndPartial(t *testing.T) {
	graph := semanticGraphFixture(t)
	for name, query := range map[string]SemanticQuery{
		"symbol":     {Vocabulary: "symbol", NamePrefix: "", Path: "api/a.go", Limit: 8},
		"imports":    {Vocabulary: "imports", Path: "consumer/use.go", Limit: 8},
		"calls":      {Vocabulary: "calls", Path: "consumer/use.go", Limit: 8},
		"tests":      {Vocabulary: "tests", Path: "api/a_test.go", Limit: 8},
		"generators": {Vocabulary: "generators", Path: "api/generated.go", Limit: 8},
		"path":       {Vocabulary: "path", Path: "api/a.go", Limit: 8},
		"impact":     {Vocabulary: "impact", Paths: []string{"api/a.go"}, MaxDepth: 3, Limit: 8},
	} {
		t.Run(name, func(t *testing.T) {
			first, err := QueryGoSemantic(graph, query)
			if err != nil {
				t.Fatal(err)
			}
			second, err := QueryGoSemantic(graph, query)
			if err != nil {
				t.Fatal(err)
			}
			if first.Version != 1 || first.Coverage != "PARTIAL" || first.SourceID != graph.SourceID || first.CandidateID != graph.CandidateID || first.GraphDigest != graph.Digest || first.ProducerSHA256 != graph.ProducerSHA256 || !semanticItemsEqual(first.Items, second.Items) || first.Truncated != second.Truncated {
				t.Fatalf("semantic result lost graph binding or determinism: %#v", first)
			}
		})
	}
	calls, err := QueryGoSemantic(graph, SemanticQuery{Vocabulary: "calls", Path: "consumer/use.go", Limit: 1})
	if err != nil || len(calls.Items) != 1 || calls.Items[0].Resolution != "UNRESOLVED" || !calls.Truncated {
		t.Fatalf("call evidence must remain explicitly unresolved and bounded: %#v err=%v", calls, err)
	}
	generators, err := QueryGoSemantic(graph, SemanticQuery{Vocabulary: "generators", Path: "api/generated.go", Limit: 8})
	if err != nil || len(generators.Items) != 1 || generators.Items[0].Relation != "GENERATED_BY" || generators.Items[0].Resolution != "EXPLICIT_SOURCE_BOUND" {
		t.Fatalf("generator evidence lost explicit source binding: %#v err=%v", generators, err)
	}
	tests, err := QueryGoSemantic(graph, SemanticQuery{Vocabulary: "tests", Path: "api/a_test.go", Limit: 8})
	if err != nil || len(tests.Items) != 1 || tests.Items[0].Relation != "TESTS" {
		t.Fatalf("test relation evidence missing: %#v err=%v", tests, err)
	}
}

func TestQueryGoSemanticRejectsUnsupportedAndUnsafeSelectors(t *testing.T) {
	graph := semanticGraphFixture(t)
	for _, vocabulary := range []string{"references", "implementations"} {
		if _, err := QueryGoSemantic(graph, SemanticQuery{Vocabulary: vocabulary, Limit: 1}); !errors.Is(err, ErrSemanticVocabularyUnsupported) {
			t.Fatalf("%s was silently represented as empty evidence: %v", vocabulary, err)
		}
	}
	for _, query := range []SemanticQuery{
		{Vocabulary: "path", Path: "../escape.go", Limit: 1},
		{Vocabulary: "symbol", Path: "api/a.go", MaxDepth: 1, Limit: 1},
		{Vocabulary: "calls", Path: "consumer/use.go", MaxDepth: 1, Limit: 1},
		{Vocabulary: "imports", Limit: 1},
		{Vocabulary: "impact", Paths: []string{"../escape.go"}, Limit: 1},
		{Vocabulary: "unknown", Limit: 1},
	} {
		if _, err := QueryGoSemantic(graph, query); err == nil {
			t.Fatalf("invalid semantic selector was accepted: %#v", query)
		}
	}
}

func TestQueryGoSemanticConservativelyReportsUnderlyingQueryBound(t *testing.T) {
	symbolGraph := semanticLargeGraphFixture(t, goEngineeringMaxQuery+1, false)
	callGraph := semanticLargeGraphFixture(t, goEngineeringMaxQuery+1, true)
	for _, test := range []struct {
		graph GoEngineeringGraph
		query SemanticQuery
	}{
		{symbolGraph, SemanticQuery{Vocabulary: "symbol", NamePrefix: "F", Limit: goEngineeringMaxQuery}},
		{callGraph, SemanticQuery{Vocabulary: "calls", Path: "pkg/large.go", Limit: goEngineeringMaxQuery}},
		{symbolGraph, SemanticQuery{Vocabulary: "symbol", NamePrefix: "F", Limit: 3}},
		{callGraph, SemanticQuery{Vocabulary: "calls", Path: "pkg/large.go", Limit: 3}},
	} {
		result, err := QueryGoSemantic(test.graph, test.query)
		if err != nil || !result.Truncated || len(result.Items) != test.query.Limit || result.Coverage != "PARTIAL" {
			t.Fatalf("bounded semantic result hid an underlying or adapter truncation: result=%#v err=%v", result, err)
		}
	}
}

func TestQueryGoSemanticSourceLocalAndDeclaredModuleEvidence(t *testing.T) {
	producer := strings.Repeat("b", 64)
	sourceID := strings.Repeat("0", 64)
	local, err := SourceLocalGoPackageBinding(sourceID, "pkg/a.go", "pkg")
	if err != nil {
		t.Fatal(err)
	}
	localGraph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: sourceID, ProducerSHA256: producer, Files: []GoGraphFileInput{graphInput("pkg/a.go", "package pkg\n", local, nil, nil, nil, nil, producer)}})
	if err != nil {
		t.Fatal(err)
	}
	module, err := QueryGoSemantic(localGraph, SemanticQuery{Vocabulary: "module", Path: "pkg/a.go", Limit: 4})
	if err != nil || len(module.Items) != 0 || module.Coverage != "PARTIAL" {
		t.Fatalf("source-local package invented module ownership: %#v err=%v", module, err)
	}

	inventory := semanticModuleInventory(t, sourceID)
	declared, err := DeclaredGoPackageBinding(inventory, "pkg/a.go", "pkg")
	if err != nil {
		t.Fatal(err)
	}
	declaredGraph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: sourceID, ProducerSHA256: producer, Files: []GoGraphFileInput{graphInput("pkg/a.go", "package pkg\n", declared, nil, nil, nil, nil, producer)}, ModuleInventory: &inventory})
	if err != nil {
		t.Fatal(err)
	}
	module, err = QueryGoSemantic(declaredGraph, SemanticQuery{Vocabulary: "module", Path: "pkg/a.go", Limit: 4})
	if err != nil || len(module.Items) != 1 || module.Items[0].Resolution != "DECLARED" || module.Items[0].Relation != "DECLARED_OWNERSHIP" {
		t.Fatalf("declared module ownership was not inspectable: %#v err=%v", module, err)
	}

	first, err := QueryGoSemantic(declaredGraph, SemanticQuery{Vocabulary: "path", Path: "pkg/a.go", Limit: 1})
	if err != nil || len(first.Items) != 1 {
		t.Fatal(err)
	}
	first.Items[0].Label = "mutated"
	again, err := QueryGoSemantic(declaredGraph, SemanticQuery{Vocabulary: "path", Path: "pkg/a.go", Limit: 1})
	if err != nil || again.Items[0].Label == "mutated" {
		t.Fatal("semantic query result mutation altered immutable graph evidence")
	}
}

func semanticModuleInventory(t *testing.T, sourceID string) GoModuleInventory {
	t.Helper()
	inventory := GoModuleInventory{
		Version: 1, RepositoryID: sourceID, Commit: strings.Repeat("a", 64), Tree: strings.Repeat("b", 64), Coverage: "complete", ObservedManifestCount: 1,
		Files: []GoManifestObservation{{Kind: "go_mod", Path: "go.mod", Blob: strings.Repeat("c", 64), Bytes: 24, SHA256: strings.Repeat("d", 64), Status: "parsed", ModulePath: "example.com/m", Requires: []GoModuleRequirement{}, Replaces: []GoModuleReplacement{}}},
	}
	if err := finalizeGoModuleInventory(&inventory); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGoModuleInventoryRecord(inventory); err != nil {
		t.Fatal(err)
	}
	return inventory
}

func semanticGraphFixture(t *testing.T) GoEngineeringGraph {
	t.Helper()
	producer := strings.Repeat("b", 64)
	sourceID := strings.Repeat("0", 64)
	apiSource := "package api\n//go:generate go run ./cmd/gen\nfunc New() {}\n"
	api := graphInput("api/a.go", apiSource, GoPackageBinding{ImportPath: "example.com/m/api", ModulePath: "example.com/m"},
		[]GoSymbol{{Name: "New", Kind: "function_declaration", Range: graphSpan(apiSource, "New", 0)}}, nil, nil,
		[]string{"//go:generate go run ./cmd/gen"}, producer)
	generatedSource := "// Code generated by command. DO NOT EDIT.\npackage api\n"
	generated := graphInput("api/generated.go", generatedSource, GoPackageBinding{ImportPath: "example.com/m/api", ModulePath: "example.com/m"}, nil, nil, nil,
		[]string{"// Code generated by command. DO NOT EDIT."}, producer)
	testSource := "package api_test\nfunc TestNew() {}\n"
	testFile := graphInput("api/a_test.go", testSource, GoPackageBinding{ImportPath: "example.com/m/api_test", ModulePath: "example.com/m", TestOfImportPath: "example.com/m/api"},
		[]GoSymbol{{Name: "TestNew", Kind: "function_declaration", Test: true, Range: graphSpan(testSource, "TestNew", 0)}}, nil, nil, nil, producer)
	consumerSource := "package consumer\nimport api \"example.com/m/api\"\nfunc Use() { api.New(); api.New() }\n"
	consumer := graphInput("consumer/use.go", consumerSource, GoPackageBinding{ImportPath: "example.com/m/consumer", ModulePath: "example.com/m"},
		[]GoSymbol{{Name: "Use", Kind: "function_declaration", Range: graphSpan(consumerSource, "Use", 0)}},
		[]GoImport{{Path: "example.com/m/api", Alias: ptrString("api"), Range: graphSpan(consumerSource, `api "example.com/m/api"`, 0)}},
		[]GoCall{{Spelling: "api.New", Resolution: "UNRESOLVED", Range: graphSpan(consumerSource, "api.New", 0)}, {Spelling: "api.New", Resolution: "UNRESOLVED", Range: graphSpan(consumerSource, "api.New", strings.Index(consumerSource, "api.New")+1)}}, nil, producer)
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{
		SourceID: sourceID, CandidateID: strings.Repeat("a", 64), ProducerSHA256: producer,
		Files:      []GoGraphFileInput{api, generated, testFile, consumer},
		Generators: []GoGeneratorBinding{{GeneratorPath: "api/a.go", GeneratedPath: "api/generated.go", Directive: "//go:generate go run ./cmd/gen"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func semanticLargeGraphFixture(t *testing.T, count int, callsOnly bool) GoEngineeringGraph {
	t.Helper()
	producer := strings.Repeat("b", 64)
	var source strings.Builder
	source.WriteString("package pkg\n")
	declarations := make([]GoSymbol, 0, count+1)
	if !callsOnly {
		for i := 0; i < count; i++ {
			name := fmt.Sprintf("F%04d", i)
			start := source.Len() + len("func ")
			source.WriteString("func " + name + "() {}\n")
			declarations = append(declarations, GoSymbol{Name: name, Kind: "function_declaration", Range: GoRange{StartByte: start, EndByte: start + len(name)}})
		}
	}
	useStart := source.Len()
	source.WriteString("func Use() { ")
	calls := make([]GoCall, 0, count)
	if callsOnly {
		for i := 0; i < count; i++ {
			name := fmt.Sprintf("F%04d", i)
			start := source.Len()
			source.WriteString(name + "();")
			calls = append(calls, GoCall{Spelling: name, Resolution: "UNRESOLVED", Range: GoRange{StartByte: start, EndByte: start + len(name)}})
		}
	}
	source.WriteString(" }\n")
	declarations = append(declarations, GoSymbol{Name: "Use", Kind: "function_declaration", Range: GoRange{StartByte: useStart + len("func "), EndByte: useStart + len("func Use")}})
	file := graphInput("pkg/large.go", source.String(), GoPackageBinding{ImportPath: "example.com/m/pkg", ModulePath: "example.com/m"}, declarations, nil, calls, nil, producer)
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: strings.Repeat("0", 64), ProducerSHA256: producer, Files: []GoGraphFileInput{file}})
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func semanticItemsEqual(left, right []SemanticItem) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].ID != right[i].ID || left[i].Kind != right[i].Kind || left[i].Path != right[i].Path || left[i].Label != right[i].Label || left[i].Relation != right[i].Relation || left[i].Target != right[i].Target || left[i].Resolution != right[i].Resolution {
			return false
		}
	}
	return true
}
