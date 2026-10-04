package ri

import (
	"testing"

	"harness.local/engorch/internal/canonical"
)

func TestGraphTypedCanonicalSealMatchesGenericHash(t *testing.T) {
	producer := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sourceID := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	source := []byte("package p\nfunc Value() {}\n")
	facts := makeCorpusTestFacts(t, "pkg/value.go", source, producer, nil)
	binding, err := SourceLocalGoPackageBinding(sourceID, "pkg/value.go", "p")
	if err != nil {
		t.Fatal(err)
	}
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: sourceID, ProducerSHA256: producer, Files: []GoGraphFileInput{{Facts: facts, Source: source, Package: binding}}, Generators: []GoGeneratorBinding{}})
	if err != nil {
		t.Fatal(err)
	}
	want, err := canonical.Hash("harness.ri.go-engineering-graph.v1", graph.content())
	if err != nil {
		t.Fatal(err)
	}
	if graph.Digest != want || ValidateGoEngineeringGraph(graph) != nil {
		t.Fatal("typed graph seal diverged from generic canonical hash")
	}
}
