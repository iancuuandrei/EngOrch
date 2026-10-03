package control

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/ri"
	"harness.local/engorch/internal/taskcontext"
)

func plannerGoContextUnavailableFixture(t *testing.T) (string, Creation, PlannerGoContextRecord) {
	t.Helper()
	c := autonomousCreation(t, 0)
	c.Execution.PlannerContext = plannerContextGoSourceV1
	c.Execution.PlannerContextRIExecutable = filepath.Join(t.TempDir(), "engorch-ri.exe")
	c.Execution.PlannerContextRIExecutableSHA256 = strings.Repeat("a", 64)
	root := c.Repository.Root
	autonomousGitInit(t, root)
	var err error
	c.Repository, err = repository.Discover(context.Background(), root, c.Config.Repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := ri.FromRepository(c.Repository)
	if err != nil {
		t.Fatal(err)
	}
	query, truncated, hash, length := truncateTaskQuery(c.Objective)
	record := PlannerGoContextRecord{
		Version: plannerGoContextVersion, Source: source, Query: query, QueryTruncated: truncated, QueryHash: hash, QueryLen: length,
		RIExecutableSHA256: c.Execution.PlannerContextRIExecutableSHA256, Unavailable: "no_eligible_complete_go_source",
	}
	record.RecordID, err = plannerGoContextRecordID(record)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "run.jsonl")
	if err := Append(path, "run.created", c); err != nil {
		t.Fatal(err)
	}
	return path, c, record
}

func TestGoPlannerContextUnavailableAdmissionBindsPlannerInputAndReplays(t *testing.T) {
	path, _, record := plannerGoContextUnavailableFixture(t)
	if err := Append(path, "planner.go-context-admitted", record); err != nil {
		t.Fatal(err)
	}
	s, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.PlannerGoContext == nil || s.PlannerGoContext.RecordID != record.RecordID || s.PlannerGoContext.Unavailable == "" {
		t.Fatalf("unavailable Go context was not replayed: %+v", s.PlannerGoContext)
	}
	invocation, err := plannerInvocationForSnapshot(s)
	if err != nil || !strings.Contains(invocation.Input, record.RecordID) {
		t.Fatalf("planner invocation lacks durable Go context: %v", err)
	}
}

func TestGoPlannerContextReplayRejectsRecordAndSourceSubstitution(t *testing.T) {
	_, c, record := plannerGoContextUnavailableFixture(t)
	for name, change := range map[string]func(*PlannerGoContextRecord){
		"record": func(r *PlannerGoContextRecord) { r.Unavailable = "forged" },
		"source": func(r *PlannerGoContextRecord) { r.Source.Tree = strings.Repeat("b", 64) },
		"parser": func(r *PlannerGoContextRecord) { r.RIExecutableSHA256 = strings.Repeat("b", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "run.jsonl")
			if err := Append(path, "run.created", c); err != nil {
				t.Fatal(err)
			}
			forged := record
			change(&forged)
			if err := Append(path, "planner.go-context-admitted", forged); err == nil {
				t.Fatal("forged Go planner context was admitted")
			}
		})
	}
}

func TestGoPlannerContextPolicyRequiresOnlyItsPinnedParser(t *testing.T) {
	base := ExecutionPolicy{Mode: "autonomous-v1", MaxRepairs: 0}
	if err := base.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := base
	bad.PlannerContext = plannerContextGoSourceV1
	if err := bad.Validate(); err == nil {
		t.Fatal("Go planner context accepted without parser pin")
	}
	good := bad
	good.PlannerContextRIExecutable = filepath.Join(t.TempDir(), "engorch-ri.exe")
	good.PlannerContextRIExecutableSHA256 = strings.Repeat("a", 64)
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
	legacyPlanner := base
	legacyPlanner.PlannerContextRIExecutable = good.PlannerContextRIExecutable
	legacyPlanner.PlannerContextRIExecutableSHA256 = good.PlannerContextRIExecutableSHA256
	if err := legacyPlanner.Validate(); err == nil {
		t.Fatal("legacy planner context accepted Go parser fields")
	}
}

// TestGoPlannerContextNonemptyAdmissionAndReplay uses the pinned local RI
// executable when supplied by the test environment. It creates a committed
// fixture and exercises only admission/replay: no planner, writer, or model
// runtime is started.
func TestGoPlannerContextNonemptyAdmissionAndReplay(t *testing.T) {
	executable := os.Getenv("ENGORCH_RI_BINARY")
	if executable == "" {
		t.Skip("ENGORCH_RI_BINARY is required for local RI fixture")
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(binary)
	c := autonomousCreation(t, 0)
	c.Objective = "Correct SplitLines semantics and document the example contract"
	c.Execution.PlannerContext = plannerContextGoSourceV1
	c.Execution.PlannerContextRIExecutable = filepath.Clean(executable)
	c.Execution.PlannerContextRIExecutableSHA256 = hex.EncodeToString(digest[:])
	root := c.Repository.Root
	autonomousGitInit(t, root)
	for path, content := range map[string]string{
		"difflib/difflib.go":      "package difflib\n\n// SplitLines preserves existing terminators.\nfunc SplitLines(input string) []string { return []string{input} }\n\nfunc GetContextDiffString(input string) string { return SplitLines(input)[0] }\n",
		"difflib/difflib_test.go": "package difflib\n\nimport \"fmt\"\n\nfunc ExampleGetContextDiffString() {\n\tfmt.Println(GetContextDiffString(\"line\\n\"))\n\t// Output: line\n}\n",
	} {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "difflib"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Go fixture"}} {
		if out, err := autonomousGitCmd(t, root, args); err != nil {
			t.Fatal(err, string(out))
		}
	}
	c.Repository, err = repository.Discover(context.Background(), root, c.Config.Repository)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "run.jsonl")
	if err := Append(path, "run.created", c); err != nil {
		t.Fatal(err)
	}
	record, err := AdmitPlannerGoContext(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if record.Unavailable != "" || record.Graph == nil || record.Context == nil || record.Context.Schema != "engorch.ri.go-context.v2" || len(record.Sources) == 0 {
		t.Fatalf("expected nonempty v2 evidence: %+v", record)
	}
	if err := boundedCanonical(*record.Graph, plannerGoContextMaxGraphBytes); err != nil {
		t.Fatal(err)
	}
	if err := boundedCanonical(record, plannerGoContextMaxRecord); err != nil {
		t.Fatal(err)
	}
	s, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.PlannerGoContext == nil || s.PlannerAccess != nil || len(s.ModelAccess) != 0 || s.WriterProposal != nil || s.PlannerGoContext.RecordID != record.RecordID {
		t.Fatalf("admission created an unexpected effect or failed replay: %+v", s)
	}
	all := append(append([]taskcontext.SelectedFile(nil), record.Context.Selection.Selected...), record.Context.ContractExcerpts...)
	joined := ""
	for _, excerpt := range all {
		joined += string(excerpt.Content)
	}
	if !strings.Contains(joined, "SplitLines") || !strings.Contains(joined, "ExampleGetContextDiffString") {
		t.Fatalf("v2 context lacks target contract evidence: %q", joined)
	}
	forged := record
	forged.Source.Tree = strings.Repeat("f", len(forged.Source.Tree))
	forged.RecordID, err = plannerGoContextRecordID(forged)
	if err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(t.TempDir(), "forged.jsonl")
	if err := Append(badPath, "run.created", c); err != nil {
		t.Fatal(err)
	}
	if err := Append(badPath, "planner.go-context-admitted", forged); err == nil {
		t.Fatal("self-rehashed v2 source substitution was admitted")
	}
}
