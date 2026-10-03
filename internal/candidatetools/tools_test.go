package candidatetools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"harness.local/engorch/internal/anchoredit"
	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/worktree"
)

func TestCatalogPreservesCandidateSchemas(t *testing.T) {
	got, err := canonical.Bytes(Catalog())
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"Description":"List the exact admitted candidate's current regular files with hashes and modes. Use after empty initially, follow next_after until null. Unlike source_list this includes admitted modifications. Drift fails the request.","InputSchema":{"additionalProperties":false,"properties":{"after":{"type":"string"},"limit":{"maximum":128,"minimum":1,"type":"integer"}},"required":["after","limit"],"type":"object"},"Name":"candidate_list"},{"Description":"Read a byte page from the admitted candidate, including modifications. Returns exact full-file hash and binary/UTF-8 views. Follow next_offset until null. Any candidate drift fails the request.","InputSchema":{"additionalProperties":false,"properties":{"limit":{"maximum":32768,"minimum":1,"type":"integer"},"offset":{"minimum":0,"type":"integer"},"path":{"type":"string"}},"required":["path","offset","limit"],"type":"object"},"Name":"candidate_read"}]`
	if string(got) != want {
		t.Fatalf("catalog changed\n got: %s\nwant: %s", got, want)
	}
}

func TestExecuteReadsChangedBytesOmitsDeletionAndRejectsDrift(t *testing.T) {
	binding, lease := candidateFixture(t)
	defer lease.Close()
	root := binding.Workspace.Request.Path
	if err := os.WriteFile(filepath.Join(root, "changed.txt"), []byte("candidate bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "deleted.txt")); err != nil {
		t.Fatal(err)
	}
	candidate, err := worktree.Fingerprint(context.Background(), binding.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	binding.Candidate = candidate

	content, handled, err := Execute(context.Background(), binding, ListName, json.RawMessage(`{"after":"","limit":1}`))
	if err != nil || !handled {
		t.Fatal(handled, err)
	}
	first := content.(Page)
	if len(first.Files) != 1 || first.Files[0].Path != "changed.txt" || first.NextAfter == nil || *first.NextAfter != "changed.txt" {
		t.Fatal("candidate first page mismatch", first)
	}
	content, handled, err = Execute(context.Background(), binding, ListName, mustArguments(t, map[string]any{"after": *first.NextAfter, "limit": 128}))
	if err != nil || !handled {
		t.Fatal(handled, err)
	}
	second := content.(Page)
	if len(second.Files) != 1 || second.Files[0].Path != "unchanged.txt" || second.NextAfter != nil || second.CandidateID != first.CandidateID {
		t.Fatal("deleted path remained in candidate listing", second)
	}

	content, handled, err = Execute(context.Background(), binding, ReadName, json.RawMessage(`{"path":"changed.txt","offset":0,"limit":32768}`))
	if err != nil || !handled {
		t.Fatal(handled, err)
	}
	chunk := content.(worktree.SourceChunk)
	bytes, err := base64.StdEncoding.DecodeString(chunk.ContentBase64)
	if err != nil || string(bytes) != "candidate bytes" || chunk.CandidateID != first.CandidateID || chunk.SHA256 != first.Files[0].Hash {
		t.Fatal("candidate read did not bind changed bytes", chunk, err)
	}
	if _, handled, err = Execute(context.Background(), binding, ReadName, json.RawMessage(`{"path":"deleted.txt","offset":0,"limit":1}`)); err == nil || !handled {
		t.Fatal("deleted candidate path was readable", handled)
	}

	if err := os.WriteFile(filepath.Join(root, "drift.txt"), []byte("later"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, request := range []struct {
		name string
		args json.RawMessage
	}{
		{ListName, json.RawMessage(`{"after":"","limit":128}`)},
		{ReadName, json.RawMessage(`{"path":"changed.txt","offset":0,"limit":1}`)},
	} {
		if _, handled, err := Execute(context.Background(), binding, request.name, request.args); err == nil || !handled {
			t.Fatal("candidate drift admitted", request.name, handled)
		}
	}
}

func TestExecuteRejectsArgumentsAndLeavesUnknownUnhandled(t *testing.T) {
	for _, request := range []struct {
		name string
		args json.RawMessage
	}{
		{ListName, json.RawMessage(`{"after":"","limit":0}`)},
		{ListName, json.RawMessage(`{"after":"","limit":1,"extra":true}`)},
		{ReadName, json.RawMessage(`{"path":"changed.txt","offset":-1,"limit":1}`)},
		{ReadName, json.RawMessage(`{"path":"changed.txt","offset":0,"limit":32769}`)},
	} {
		if _, handled, err := Execute(context.Background(), Binding{}, request.name, request.args); err == nil || !handled {
			t.Fatal("invalid candidate arguments admitted", request.name, handled)
		}
	}
	if content, handled, err := Execute(context.Background(), Binding{}, "other", json.RawMessage(`{}`)); err != nil || handled || content != nil {
		t.Fatal("unknown candidate tool handled", content, handled, err)
	}
}

func TestAnchoredEditValidatorIsReadOnlyAndReportsAnchorDiagnostics(t *testing.T) {
	binding, lease := candidateFixture(t)
	defer lease.Close()
	binding.AnchorValidationVersion = AnchorValidationVersion
	candidateID, err := binding.Candidate.ID()
	if err != nil {
		t.Fatal(err)
	}
	_, files, err := worktree.Capture(context.Background(), binding.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	var beforeHash string
	for _, file := range files {
		if file.Path == "changed.txt" {
			beforeHash = file.Hash
		}
	}
	if beforeHash == "" {
		t.Fatal("fixture source absent")
	}

	valid := validateCandidateEdits(t, binding, AnchoredEditArgs{CandidateID: candidateID, Path: "changed.txt", BeforeHash: beforeHash, Edits: []anchoredit.Edit{{Before: "base", After: "candidate"}}})
	if !valid.Valid || valid.ResultHash == "" || valid.Diagnostic != "" {
		t.Fatal("unique anchor did not validate", valid)
	}
	encoded, err := canonical.Bytes(valid)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "base bytes") || strings.Contains(string(encoded), "candidate bytes") {
		t.Fatal("validator leaked file content", string(encoded))
	}

	for name, edits := range map[string][]anchoredit.Edit{
		"missing":           {{Before: "not present", After: "replacement"}},
		"ambiguous":         {{Before: "e", After: "replacement"}},
		"overlap":           {{Before: "base bytes", After: "x"}, {Before: "base", After: "y"}},
		"oversized-payload": {{Before: strings.Repeat("x", anchoredit.MaxEditBytes), After: "y"}},
	} {
		t.Run(name, func(t *testing.T) {
			result := validateCandidateEdits(t, binding, AnchoredEditArgs{CandidateID: candidateID, Path: "changed.txt", BeforeHash: beforeHash, Edits: edits})
			if result.Valid || result.Diagnostic == "" {
				t.Fatal("invalid anchors were not diagnosed", result)
			}
		})
	}
}

func TestAnchoredEditValidatorRejectsWrongBindingDriftAndOversizedSource(t *testing.T) {
	binding, lease := candidateFixture(t)
	defer lease.Close()
	binding.AnchorValidationVersion = AnchorValidationVersion
	candidateID, _ := binding.Candidate.ID()
	_, files, err := worktree.Capture(context.Background(), binding.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	var beforeHash string
	for _, file := range files {
		if file.Path == "changed.txt" {
			beforeHash = file.Hash
		}
	}
	args := AnchoredEditArgs{CandidateID: candidateID, Path: "changed.txt", BeforeHash: beforeHash, Edits: []anchoredit.Edit{{Before: "base", After: "candidate"}}}
	for _, mutate := range []func(*AnchoredEditArgs){
		func(a *AnchoredEditArgs) { a.CandidateID = strings.Repeat("b", 64) },
		func(a *AnchoredEditArgs) { a.BeforeHash = strings.Repeat("b", 64) },
	} {
		copy := args
		mutate(&copy)
		if _, handled, err := Execute(context.Background(), binding, ValidateAnchoredEditsName, mustArguments(t, copy)); err == nil || !handled {
			t.Fatal("wrong candidate/hash binding admitted", handled, err)
		}
	}
	if err := os.WriteFile(filepath.Join(binding.Workspace.Request.Path, "drift.txt"), []byte("drift"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, handled, err := Execute(context.Background(), binding, ValidateAnchoredEditsName, mustArguments(t, args)); err == nil || !handled {
		t.Fatal("candidate drift admitted", handled, err)
	}

	oversized, oversizedLease := candidateFixture(t)
	defer oversizedLease.Close()
	oversized.AnchorValidationVersion = AnchorValidationVersion
	large := strings.Repeat("x", anchoredit.MaxSourceBytes+1)
	if err := os.WriteFile(filepath.Join(oversized.Workspace.Request.Path, "changed.txt"), []byte(large), 0600); err != nil {
		t.Fatal(err)
	}
	oversized.Candidate, err = worktree.Fingerprint(context.Background(), oversized.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	overID, _ := oversized.Candidate.ID()
	_, largeFiles, err := worktree.Capture(context.Background(), oversized.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range largeFiles {
		if file.Path == "changed.txt" {
			args.CandidateID, args.BeforeHash = overID, file.Hash
		}
	}
	if _, handled, err := Execute(context.Background(), oversized, ValidateAnchoredEditsName, mustArguments(t, args)); err == nil || !handled {
		t.Fatal("oversized source admitted", handled, err)
	}
}

func TestAnchoredEditValidatorHandlesUTF8SplitAcrossReadPages(t *testing.T) {
	binding, lease := candidateFixture(t)
	defer lease.Close()
	binding.AnchorValidationVersion = AnchorValidationVersion
	source := strings.Repeat("x", 32767) + "🦊" + strings.Repeat("y", 40000)
	if err := os.WriteFile(filepath.Join(binding.Workspace.Request.Path, "changed.txt"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var err error
	binding.Candidate, err = worktree.Fingerprint(context.Background(), binding.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	candidateID, _ := binding.Candidate.ID()
	_, files, err := worktree.Capture(context.Background(), binding.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	var beforeHash string
	for _, file := range files {
		if file.Path == "changed.txt" {
			beforeHash = file.Hash
		}
	}
	result := validateCandidateEdits(t, binding, AnchoredEditArgs{CandidateID: candidateID, Path: "changed.txt", BeforeHash: beforeHash, Edits: []anchoredit.Edit{{Before: "🦊", After: "fox"}}})
	if !result.Valid {
		t.Fatal("valid UTF-8 split across pages rejected", result)
	}
}

func validateCandidateEdits(t *testing.T, binding Binding, args AnchoredEditArgs) AnchoredEditValidation {
	t.Helper()
	content, handled, err := Execute(context.Background(), binding, ValidateAnchoredEditsName, mustArguments(t, args))
	if err != nil || !handled {
		t.Fatal("anchored validation failed", handled, err)
	}
	result, ok := content.(AnchoredEditValidation)
	if !ok {
		t.Fatalf("unexpected validation result %T", content)
	}
	return result
}

func candidateFixture(t *testing.T) (Binding, *worktree.Lease) {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	git("init", "-q")
	for name, body := range map[string]string{"changed.txt": "base bytes", "deleted.txt": "remove me", "unchanged.txt": "retained"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("add", "changed.txt", "deleted.txt", "unchanged.txt")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "base")
	source, err := repository.Discover(context.Background(), root, "candidate-tools-fixture")
	if err != nil {
		t.Fatal(err)
	}
	request, err := worktree.Prepare(strings.Repeat("a", 64), source)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := worktree.Acquire(request)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := worktree.Create(context.Background(), request)
	if err != nil {
		lease.Close()
		t.Fatal(err)
	}
	candidate, err := worktree.Fingerprint(context.Background(), workspace)
	if err != nil {
		lease.Close()
		t.Fatal(err)
	}
	return Binding{Workspace: workspace, Candidate: candidate}, lease
}

func mustArguments(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := canonical.Bytes(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
