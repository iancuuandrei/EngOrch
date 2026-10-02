package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestTaskScriptReadsReadyRunWithoutDispatchAndReturnsOneObject(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell happy path")
	}
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("PowerShell 7 not installed")
	}
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "base"}} {
		if b, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(b))
		}
	}
	// A tracked change produces a nonempty diff summary, which must not leak
	// strings into the caller's result object pipeline.
	file := filepath.Join(root, "source.txt")
	if err := os.WriteFile(file, []byte("before\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "source.txt"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "source"}} {
		if b, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(b))
		}
	}
	if err := os.WriteFile(file, []byte("after\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runID := strings.Repeat("a", 64)
	state := map[string]any{
		"run_id": runID, "state": "READY", "lifecycle": map[string]string{"status": "ACTIVE"},
		"review":          map[string]any{"result": map[string]string{"output": `{"decision":"approve","findings":[]}`}},
		"workspace":       map[string]any{"request": map[string]string{"path": root}},
		"writer_proposal": map[string]any{"prepared": map[string]any{"proposal": map[string]any{"changes": []any{}}}},
	}
	b, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	calls := filepath.Join(root, "calls.txt")
	fixture := filepath.Join(root, "fabric-fixture.cmd")
	content := "@echo off\r\necho %*>>\"" + calls + "\"\r\necho " + string(b) + "\r\nexit /b 0\r\n"
	if err := os.WriteFile(fixture, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	probe := filepath.Join(root, "probe.ps1")
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "run-task.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(probe, []byte(`param($Script, $Fabric, $Repository, $RunId)
$ErrorActionPreference = 'Stop'
$result = & $Script -Fabric $Fabric -Repository $Repository -RunId $RunId -NonInteractive
if (@($result).Count -ne 1 -or $result.state -ne 'READY' -or $result.review -ne 'approve') { throw 'Invalid task result' }
`), 0600); err != nil {
		t.Fatal(err)
	}
	if b, err := exec.Command(pwsh, "-NoProfile", "-File", probe, script, fixture, root, runID).CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	observed, err := os.ReadFile(calls)
	if err != nil || strings.Count(strings.TrimSpace(string(observed)), "\n") != 0 || !strings.Contains(string(observed), "inspect "+runID) {
		t.Fatal("ready readback dispatched more than one inspect", err, string(observed))
	}
	for _, test := range []struct {
		name, workflow, outcome string
		writerPending           bool
		wantError               bool
	}{
		{"approval", "AWAITING_APPROVAL", "", false, false},
		{"pending-writer", "IMPLEMENTING", "", true, true},
		{"unknown-files", "IMPLEMENTING", "UNKNOWN", false, true},
		{"failed-checks", "REPAIRING", "CONFIRMED", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			state["state"] = test.workflow
			state["file_outcome"] = test.outcome
			state["plan"] = map[string]string{"output": "fixture plan"}
			state["writer_proposal"] = nil
			state["writer_host"] = nil
			if test.writerPending {
				state["writer_host"] = map[string]string{"status": "pending"}
			}
			b, _ := json.Marshal(state)
			content := "@echo off\r\necho %*>>\"" + calls + "\"\r\necho " + string(b) + "\r\nexit /b 0\r\n"
			if err := os.WriteFile(fixture, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(calls, nil, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(probe, []byte(`param($Script, $Fabric, $Repository, $RunId)
$ErrorActionPreference = 'Stop'
$result = & $Script -Fabric $Fabric -Repository $Repository -RunId $RunId -NonInteractive
if (@($result).Count -ne 1 -or $result.action_required -ne 'approve plan') { throw 'Approval bypassed' }
`), 0600); err != nil {
				t.Fatal(err)
			}
			b, err := exec.Command(pwsh, "-NoProfile", "-File", probe, script, fixture, root, runID).CombinedOutput()
			if (err != nil) != test.wantError {
				t.Fatal("unexpected stop behavior", err, string(b))
			}
			observed, err := os.ReadFile(calls)
			if err != nil || strings.Count(strings.TrimSpace(string(observed)), "\n") != 0 || !strings.Contains(string(observed), "inspect "+runID) {
				t.Fatal("stop path dispatched another operation", err, string(observed))
			}
		})
	}
}
