package control

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/fileeffects"
	"harness.local/engorch/internal/runtime"
	"harness.local/engorch/internal/worktree"
	"harness.local/engorch/internal/writercontract"
)

func anchoredWriterFixture(t *testing.T) (string, runtime.Invocation, string, string) {
	t.Helper()
	path, invocation := writerInvocationFixtureForContract(t, writercontract.ContractAnchoredEditsV1)
	s, err := Inspect(path)
	if err != nil || s.Workspace == nil || s.Candidate == nil {
		t.Fatal("confirmed candidate unavailable", err)
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	_, manifest, err := worktree.Capture(context.Background(), *s.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest {
		if file.Path == "file.txt" {
			return path, invocation, candidateID, file.Hash
		}
	}
	t.Fatal("fixture file.txt missing from candidate")
	return "", runtime.Invocation{}, "", ""
}

func anchoredWriterResult(t *testing.T, invocation runtime.Invocation, candidateID, beforeHash string) runtime.Result {
	t.Helper()
	proposal := writercontract.AnchoredProposal{
		CandidateID: candidateID,
		Changes: []writercontract.AnchoredChange{{
			Path: "file.txt", BeforeHash: &beforeHash,
			Edits:          []writercontract.AnchoredEdit{{Before: "base", After: "hello"}},
			NewContentUTF8: nil, Executable: false,
		}},
	}
	output, err := canonical.Bytes(proposal)
	if err != nil {
		t.Fatal(err)
	}
	return runtime.Result{Version: 1, InvocationID: invocation.ID, Requested: invocation.Profile, Output: string(output)}
}

func TestAnchoredWriterComposesAndReplaysExactCandidatePreimage(t *testing.T) {
	path, invocation, candidateID, beforeHash := anchoredWriterFixture(t)
	result := anchoredWriterResult(t, invocation, candidateID, beforeHash)
	record, err := RecordWriterProposal(context.Background(), path, invocation, result)
	if err != nil {
		t.Fatal(err)
	}
	if len(record.EditPreimages) != 1 || record.EditPreimages[0].Path != "file.txt" || record.EditPreimages[0].ContentUTF8 != "base\n" {
		t.Fatalf("unexpected replay evidence: %+v", record.EditPreimages)
	}
	if len(record.Prepared.Proposal.Changes) != 1 {
		t.Fatalf("unexpected prepared changes: %+v", record.Prepared.Proposal.Changes)
	}
	content, err := base64.StdEncoding.DecodeString(*record.Prepared.Proposal.Changes[0].ContentBase64)
	if err != nil || string(content) != "hello\n" {
		t.Fatalf("untouched bytes were not preserved: %q %v", content, err)
	}
	inspected, err := Inspect(path)
	if err != nil || inspected.WriterProposal == nil {
		t.Fatalf("anchored proposal did not replay: state=%+v err=%v", inspected, err)
	}
	if inspected.WriterProposal.Result.Output != result.Output || inspected.WriterProposal.EditPreimages[0].ContentUTF8 != "base\n" {
		t.Fatal("replay rewrote observed output or preimage evidence")
	}
}

func TestAnchoredWriterReplayRejectsCorruptedPreimage(t *testing.T) {
	path, invocation, candidateID, beforeHash := anchoredWriterFixture(t)
	result := anchoredWriterResult(t, invocation, candidateID, beforeHash)
	prepared, preimages, err := prepareWriterFiles(context.Background(), path, invocation, result)
	if err != nil {
		t.Fatal(err)
	}
	preimages[0].ContentUTF8 = "foreign\n"
	if err := Append(path, "writer.proposed", WriterRecord{Invocation: invocation, Result: result, Prepared: prepared, EditPreimages: preimages}); err == nil {
		t.Fatal("preimage not matching candidate hash was admitted")
	}
}

func TestAnchoredWriterReadsUnicodeAcrossCandidateReadPages(t *testing.T) {
	path, _, _, _ := anchoredWriterFixture(t)
	s, err := Inspect(path)
	if err != nil || s.Workspace == nil {
		t.Fatal("workspace unavailable", err)
	}
	_, manifest, err := worktree.Capture(context.Background(), *s.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	var beforeHash string
	for _, file := range manifest {
		if file.Path == "file.txt" {
			beforeHash = file.Hash
		}
	}
	large := strings.Repeat("a", 32767) + "ă" + strings.Repeat("b", 32760) + "TAIL\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(large))
	change := fileeffects.Change{Path: "file.txt", BeforeHash: &beforeHash, ContentBase64: &encoded}
	prepared, err := PrepareFiles(context.Background(), path, []fileeffects.Change{change})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyFiles(context.Background(), path, prepared, authorize(t, prepared)); err != nil {
		t.Fatal(err)
	}
	invocation, err := PrepareWriterInvocation(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err = Inspect(path)
	if err != nil || s.Candidate == nil {
		t.Fatal("updated candidate unavailable", err)
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	_, manifest, err = worktree.Capture(context.Background(), *s.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest {
		if file.Path == "file.txt" {
			beforeHash = file.Hash
		}
	}
	proposal := writercontract.AnchoredProposal{CandidateID: candidateID, Changes: []writercontract.AnchoredChange{{
		Path: "file.txt", BeforeHash: &beforeHash,
		Edits: []writercontract.AnchoredEdit{{Before: "TAIL", After: "DONE"}}, Executable: false,
	}}}
	output, err := canonical.Bytes(proposal)
	if err != nil {
		t.Fatal(err)
	}
	result := runtime.Result{Version: 1, InvocationID: invocation.ID, Requested: invocation.Profile, Output: string(output)}
	record, err := RecordWriterProposal(context.Background(), path, invocation, result)
	if err != nil {
		t.Fatal("split Unicode source page rejected", err)
	}
	if len(record.EditPreimages) != 1 || record.EditPreimages[0].ContentUTF8 != large {
		t.Fatal("multi-page preimage did not preserve exact original UTF-8 bytes")
	}
}

func TestAnchoredWriterRejectsNullAsExistingFileDeletion(t *testing.T) {
	path, invocation, candidateID, beforeHash := anchoredWriterFixture(t)
	proposal := writercontract.AnchoredProposal{CandidateID: candidateID, Changes: []writercontract.AnchoredChange{{
		Path: "file.txt", BeforeHash: &beforeHash, Edits: nil, NewContentUTF8: nil,
		Executable: false,
	}}}
	output, err := canonical.Bytes(proposal)
	if err != nil {
		t.Fatal(err)
	}
	result := runtime.Result{Version: 1, InvocationID: invocation.ID, Requested: invocation.Profile, Output: string(output)}
	if _, err := PrepareWriterFiles(context.Background(), path, invocation, result); err == nil {
		t.Fatal("null content was interpreted as deletion")
	}
	for _, output := range []string{
		`{"candidate_id":"` + candidateID + `","changes":[{"path":"file.txt","before_hash":"` + beforeHash + `","edits":[{"before":"base","after":"hello"}],"new_content_utf8":null}]}`,
		`{"candidate_id":"` + candidateID + `","changes":[{"path":"file.txt","edits":[],"new_content_utf8":"replacement","executable":false}]}`,
		`{"candidate_id":"` + candidateID + `","changes":[{"path":"new.go","before_hash":null,"edits":null,"new_content_utf8":"package new\n","executable":false}]}`,
	} {
		result.Output = output
		if _, err := PrepareWriterFiles(context.Background(), path, invocation, result); err == nil {
			t.Fatal("omitted required nullable/boolean field admitted")
		}
	}
}

func TestAnchoredWriterInstructionReplacesFullFileDirections(t *testing.T) {
	_, invocation := writerInvocationFixtureForContract(t, writercontract.ContractAnchoredEditsV1)
	var input struct {
		Instruction string          `json:"instruction"`
		Schema      json.RawMessage `json:"output_schema"`
	}
	if err := json.Unmarshal([]byte(invocation.Input), &input); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"localized edits", "controller composes", "Null never means delete", "before_hash"} {
		if !strings.Contains(input.Instruction, text) {
			t.Fatalf("anchored instruction omitted %q", text)
		}
	}
	for _, forbidden := range []string{"complete-file replacement", "content_base64", "entire current candidate file", "Complete source files"} {
		if strings.Contains(input.Instruction, forbidden) {
			t.Fatalf("anchored instruction retained conflicting direction %q", forbidden)
		}
	}
	if !strings.Contains(string(input.Schema), `"edits"`) || strings.Contains(string(input.Schema), `"content_base64"`) {
		t.Fatal("anchored invocation has the wrong wire schema")
	}
}
