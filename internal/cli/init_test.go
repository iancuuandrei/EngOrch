package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"harness.local/engorch/internal/config"
)

func TestCodexInitCreatesCompleteRolesAndPreservesExistingConfig(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "base"}} {
		if b, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(b))
		}
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	auth := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(auth, []byte("not parsed during init"), 0600); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(t.TempDir(), "state")
	args := []string{"init", "--codex", binary, "--model", "fixture-model", "--auth-source", auth, "--state-root", state}
	var out bytes.Buffer
	if err := Execute(context.Background(), args, root, &out); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "harness.toml")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Parse(before)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Planner.Model != "fixture-model" || cfg.Writer == nil || cfg.Explorer == nil || cfg.Reviewer == nil || cfg.Codex.ExecutableHash == "" || cfg.BaseBranch != "HEAD" || cfg.ExplorerContract != "json-v2" || cfg.WriterContract != "anchored-edits-v1" {
		t.Fatal("incomplete real-role configuration")
	}
	if err := Execute(context.Background(), args, root, &out); err == nil {
		t.Fatal("configuration overwritten")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("existing configuration changed")
	}
}

func TestCodexInitRejectsInvalidSetupBeforeWriting(t *testing.T) {
	for _, args := range [][]string{{"init", "--model", "fixture"}, {"init", "--codex", "missing", "--model", "fixture"}, {"init", "unexpected"}} {
		root := t.TempDir()
		if err := Execute(context.Background(), args, root, &bytes.Buffer{}); err == nil {
			t.Fatal("invalid setup accepted", args)
		}
		if _, err := os.Stat(filepath.Join(root, "harness.toml")); !os.IsNotExist(err) {
			t.Fatal("configuration written on rejected setup")
		}
	}
}
