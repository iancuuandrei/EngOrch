package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"harness.local/engorch/internal/control"
)

func cliGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, b)
	}
}

func autonomousCLIFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	cliGit(t, root, "init", "-q")
	if err := os.WriteFile(filepath.Join(root, "source.txt"), []byte("fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cliGit(t, root, "add", "source.txt")
	cliGit(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
	mustExecuteCLI(t, root, "init")
	return root
}

// mustExecuteCLI runs one CLI invocation against root and fails the test on
// error, returning the captured stdout. It factors the repeated
// Execute-plus-buffer boilerplate shared by CLI fixtures.
func mustExecuteCLI(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := Execute(context.Background(), args, root, &b); err != nil {
		t.Fatalf("execute %q: %v", strings.Join(args, " "), err)
	}
	return b.Bytes()
}

// planAutonomousFailureFixture plans one objective and returns the durable run
// path, snapshot and a reset buffer ready for reportAutonomousFailure.
func planAutonomousFailureFixture(t *testing.T, objective string) (string, control.Snapshot, bytes.Buffer) {
	t.Helper()
	root := autonomousCLIFixture(t)
	var planOut bytes.Buffer
	if err := Execute(context.Background(), []string{"plan", objective}, root, &planOut); err != nil {
		t.Fatal(err)
	}
	var s control.Snapshot
	if err := json.Unmarshal(planOut.Bytes(), &s); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".harness", "runs", s.RunID+".jsonl")
	var out bytes.Buffer
	return path, s, out
}

func TestRunAutonomousCreatesBoundPolicyAndReportsResumableBlocker(t *testing.T) {
	root := autonomousCLIFixture(t)
	var out bytes.Buffer
	err := Execute(context.Background(), []string{"run", "--autonomous", "--max-repairs", "3", "Make a bounded fixture change"}, root, &out)
	if err == nil {
		t.Fatal("fixture config has no implementation roles; autonomous run should report its blocker")
	}
	var result autonomousFailure
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("failure omitted sanitized run summary: %s (%v)", out.String(), err)
	}

	entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one durable run, entries=%v err=%v", entries, err)
	}
	s, err := control.Inspect(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if s.Creation.Objective != "Make a bounded fixture change" || s.Creation.Execution == nil || s.Creation.Execution.Mode != "autonomous-v1" || s.Creation.Execution.MaxRepairs != 3 {
		t.Fatalf("autonomous input was not durably bound: %#v", s.Creation)
	}
	if s.Creation.Execution.GraphVersion != 1 || s.Creation.Execution.MaxParallel != 3 || s.Creation.Execution.Context != "bounded-v1" {
		t.Fatalf("graph execution and bounded context not defaulted: %#v", s.Creation.Execution)
	}
	if s.State != "IMPLEMENTING" || result.RunID != s.RunID || result.State != s.State || result.Phase != "implementation" || result.BlockedReason == "" {
		t.Fatalf("failure summary or durable state mismatch: result=%#v snapshot=%#v", result, s)
	}

	// The CLI resolves an omitted RUN only from a validated journal bound to
	// this repository, then resumes the same blocked run without losing ID.
	var resumeOut bytes.Buffer
	err = Execute(context.Background(), []string{"resume", "--autonomous"}, root, &resumeOut)
	if err == nil {
		t.Fatal("resume should retain the missing-role blocker")
	}
	var resumed autonomousFailure
	if err := json.Unmarshal(resumeOut.Bytes(), &resumed); err != nil || resumed.RunID != s.RunID || resumed.State != s.State {
		t.Fatalf("resume did not report the same durable run: %#v err=%v", resumed, err)
	}
	s, err = control.Inspect(entries[0])
	if err != nil || s.State != "IMPLEMENTING" || s.Creation.Execution.MaxRepairs != 3 {
		t.Fatalf("resume changed the immutable policy or state: %#v, %v", s, err)
	}
}

func TestAutonomousCLIRejectsInvalidOptionsAndNonAutonomousResume(t *testing.T) {
	root := autonomousCLIFixture(t)
	for _, args := range [][]string{
		{"run", "--autonomous", "--max-repairs", "-1", "objective"},
		{"run", "--autonomous", "--max-repairs", "9", "objective"},
		{"run", "--autonomous", "--max-repairs", "not-a-number", "objective"},
		{"run", "--autonomous", "--max-parallel", "0", "objective"},
		{"run", "--autonomous", "--max-parallel", "9", "objective"},
		{"run", "--autonomous", "--max-parallel", "not-a-number", "objective"},
		{"run", "--autonomous"},
		{"resume", "--autonomous", "one", "two"},
	} {
		var out bytes.Buffer
		if err := Execute(context.Background(), args, root, &out); err == nil {
			t.Errorf("accepted invalid invocation %q", strings.Join(args, " "))
		}
	}
	if entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl")); err != nil || len(entries) != 0 {
		t.Fatalf("invalid autonomous options created a run: %v, %v", entries, err)
	}

	var planOut bytes.Buffer
	if err := Execute(context.Background(), []string{"plan", "ordinary fixture"}, root, &planOut); err != nil {
		t.Fatal(err)
	}
	var ordinary control.Snapshot
	if err := json.Unmarshal(planOut.Bytes(), &ordinary); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := Execute(context.Background(), []string{"resume", "--autonomous", ordinary.RunID}, root, &out)
	if err == nil || !strings.Contains(err.Error(), "is not an autonomous run") {
		t.Fatalf("autonomous resume accepted an interactive run: %v", err)
	}
}

func TestAutonomousUncertaintySummaryIsSanitized(t *testing.T) {
	path, s, out := planAutonomousFailureFixture(t, "uncertainty fixture")
	secretDetail := "file effect UNKNOWN credential=must-not-be-printed"
	err := reportAutonomousFailure(&out, path, s.RunID, errors.New(secretDetail))
	if err == nil {
		t.Fatal("sanitized boundary error missing")
	}
	if strings.Contains(err.Error(), "credential") || strings.Contains(err.Error(), secretDetail) {
		t.Fatalf("boundary error leaked provider detail: %v", err)
	}
	if !strings.Contains(err.Error(), s.RunID) || !strings.Contains(err.Error(), "effect_requires_reconciliation") {
		t.Fatalf("boundary error lost coarse identity: %v", err)
	}
	if errors.Is(err, context.Canceled) {
		t.Fatalf("non-cancellation cause reports as cancellation: %v", err)
	}
	var summary autonomousFailure
	if err := json.Unmarshal(out.Bytes(), &summary); err != nil {
		t.Fatalf("invalid uncertainty summary %q: %v", out.String(), err)
	}
	if summary.RunID != s.RunID || summary.State != s.State || summary.Phase != "approval" || summary.BlockedReason != "effect_requires_reconciliation" || strings.Contains(out.String(), "credential") {
		t.Fatalf("uncertainty summary leaked details or lost state: %#v (%s)", summary, out.String())
	}
}

func TestAutonomousFailurePreservesCancellationWithoutLeakingDetail(t *testing.T) {
	path, s, out := planAutonomousFailureFixture(t, "cancellation fixture")
	cause := errors.Join(errors.New("provider TOKEN=secret-must-not-print"), context.Canceled)
	err := reportAutonomousFailure(&out, path, s.RunID, cause)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation was not preserved through the sanitized boundary: %v", err)
	}
	if strings.Contains(err.Error(), "secret-must-not-print") {
		t.Fatalf("cancelled boundary error leaked provider detail: %v", err)
	}
	var summary autonomousFailure
	if err := json.Unmarshal(out.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.BlockedReason != "execution_interrupted" || strings.Contains(out.String(), "secret") {
		t.Fatalf("cancelled summary leaked details or lost reason: %#v (%s)", summary, out.String())
	}
}

func TestAutonomousObjectiveValidationRejectsBeforeDurableCreation(t *testing.T) {
	root := autonomousCLIFixture(t)
	invalid := []string{"   \n\t ", string([]byte{0xff, 0xfe})}
	for _, objective := range invalid {
		var out bytes.Buffer
		if err := Execute(context.Background(), []string{"run", "--autonomous", objective}, root, &out); err == nil {
			t.Fatalf("accepted invalid objective %q", objective)
		}
	}
	var out bytes.Buffer
	if err := Execute(context.Background(), []string{"run", "--autonomous", strings.Repeat("x", (256<<10)+1)}, root, &out); err == nil {
		t.Fatal("accepted oversized objective")
	}
	if entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl")); err != nil || len(entries) != 0 {
		t.Fatalf("invalid autonomous objective created a run: %v, %v", entries, err)
	}
	if err := validateAutonomousObjective("  "); err == nil {
		t.Fatal("whitespace-only objective admitted")
	}
	if err := validateAutonomousObjective(string([]byte{0xff})); err == nil {
		t.Fatal("invalid UTF-8 objective admitted")
	}
	if err := validateAutonomousObjective(strings.Repeat("x", (256<<10)+1)); err == nil {
		t.Fatal("oversized objective admitted")
	}
}

func TestAutonomousFileGoalUsesPlanObjectiveConventions(t *testing.T) {
	root := autonomousCLIFixture(t)
	want := "# Obiectiv\r\nPăstrează textul exact.\n"
	goalPath := filepath.Join(root, "goal.md")
	if err := os.WriteFile(goalPath, []byte(want), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := Execute(context.Background(), []string{"run", "--autonomous", "--max-repairs", "1", "--file", "goal.md"}, root, &out)
	if err == nil {
		t.Fatal("fixture config has no implementation roles; file-goal run should report its blocker")
	}
	if strings.Contains(err.Error(), want) {
		t.Fatalf("sanitized file-goal error echoed objective detail: %v", err)
	}
	entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one durable file-goal run, entries=%v err=%v", entries, err)
	}
	s, err := control.Inspect(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if s.Creation.Objective != want || s.Creation.Execution == nil || s.Creation.Execution.Mode != "autonomous-v1" {
		t.Fatalf("file-goal bytes were not preserved durably: %#v", s.Creation)
	}
	if err := os.WriteFile(goalPath, []byte(" \n"), 0600); err != nil {
		t.Fatal(err)
	}
	var rejected bytes.Buffer
	if err := Execute(context.Background(), []string{"run", "--autonomous", "--file", "goal.md"}, root, &rejected); err == nil {
		t.Fatal("whitespace-only goal file admitted")
	}
	var both bytes.Buffer
	if err := Execute(context.Background(), []string{"run", "--autonomous", "--file", "goal.md", "extra"}, root, &both); err == nil {
		t.Fatal("positional objective alongside --file admitted")
	}
	if entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl")); err != nil || len(entries) != 1 {
		t.Fatalf("invalid file-goal input created a run: %v, %v", entries, err)
	}
}

func TestAutonomousParallelFlagDefaultsSequentialOverride(t *testing.T) {
	root := autonomousCLIFixture(t)
	var seqOut bytes.Buffer
	if err := Execute(context.Background(), []string{"run", "--autonomous", "--max-parallel", "1", "Sequential fixture"}, root, &seqOut); err == nil {
		t.Fatal("sequential fixture should still report its blocker")
	}
	entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one sequential run, entries=%v err=%v", entries, err)
	}
	s, err := control.Inspect(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if s.Creation.Execution == nil || s.Creation.Execution.GraphVersion != 1 || s.Creation.Execution.MaxParallel != 1 {
		t.Fatalf("sequential override not bound: %#v", s.Creation.Execution)
	}
}

func TestAutonomousResumeWithoutIDSelectsLatestAutonomousRun(t *testing.T) {
	root := autonomousCLIFixture(t)
	var autoOut bytes.Buffer
	if err := Execute(context.Background(), []string{"run", "--autonomous", "Make a bounded fixture change"}, root, &autoOut); err == nil {
		t.Fatal("expected autonomous blocker for fixture config")
	}
	var autoResult autonomousFailure
	if err := json.Unmarshal(autoOut.Bytes(), &autoResult); err != nil {
		t.Fatalf("missing autonomous failure summary: %v", err)
	}
	// A newer interactive run must not divert an omitted autonomous resume.
	var planOut bytes.Buffer
	if err := Execute(context.Background(), []string{"plan", "newer ordinary fixture"}, root, &planOut); err != nil {
		t.Fatal(err)
	}
	var ordinary control.Snapshot
	if err := json.Unmarshal(planOut.Bytes(), &ordinary); err != nil {
		t.Fatal(err)
	}
	if ordinary.RunID == autoResult.RunID {
		t.Fatal("fixture reused the autonomous run identity")
	}
	var resumeOut bytes.Buffer
	err := Execute(context.Background(), []string{"resume", "--autonomous"}, root, &resumeOut)
	if err == nil {
		t.Fatal("resume should retain the missing-role blocker")
	}
	var resumed autonomousFailure
	if err := json.Unmarshal(resumeOut.Bytes(), &resumed); err != nil || resumed.RunID != autoResult.RunID {
		t.Fatalf("resume did not select the latest autonomous run: %#v err=%v", resumed, err)
	}
}
