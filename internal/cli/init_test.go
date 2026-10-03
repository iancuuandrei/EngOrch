package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func codexInitRoleFixture(t *testing.T, extra []string) (config.Config, string, string, error) {
	t.Helper()
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
	args := append([]string{"init", "--codex", binary, "--model", "base-model", "--effort", "medium", "--auth-source", auth, "--state-root", state}, extra...)
	if err := Execute(context.Background(), args, root, &bytes.Buffer{}); err != nil {
		return config.Config{}, root, state, err
	}
	raw, err := os.ReadFile(filepath.Join(root, "harness.toml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Parse(raw)
	return cfg, root, state, err
}

func TestCodexInitSupportsRoleModelsAndEfforts(t *testing.T) {
	for _, tc := range []struct {
		name, writerModel, writerEffort, reviewerModel, reviewerEffort string
		extra                                                          []string
	}{
		{name: "defaults", writerModel: "base-model", writerEffort: "medium", reviewerModel: "base-model", reviewerEffort: "medium"},
		{name: "all overrides", writerModel: "coding-model", writerEffort: "high", reviewerModel: "review-model", reviewerEffort: "low", extra: []string{"--writer-model", "coding-model", "--writer-effort", "high", "--reviewer-model", "review-model", "--reviewer-effort", "low"}},
		{name: "writer effort", writerModel: "base-model", writerEffort: "high", reviewerModel: "base-model", reviewerEffort: "medium", extra: []string{"--writer-effort", "high"}},
		{name: "reviewer model", writerModel: "base-model", writerEffort: "medium", reviewerModel: "review-model", reviewerEffort: "medium", extra: []string{"--reviewer-model", "review-model"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _, _, err := codexInitRoleFixture(t, tc.extra)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Writer.Model != tc.writerModel || cfg.Writer.Effort != tc.writerEffort || cfg.Reviewer.Model != tc.reviewerModel || cfg.Reviewer.Effort != tc.reviewerEffort {
				t.Fatal("role allocation differs from explicit init flags")
			}
			if cfg.Planner.Model != "base-model" || cfg.Planner.Effort != "medium" || cfg.Explorer.Model != "base-model" || cfg.Explorer.Effort != "medium" {
				t.Fatal("writer/reviewer overrides changed read-only roles")
			}
		})
	}
}

func TestCodexInitRejectsInvalidRoleProfilesBeforeWriting(t *testing.T) {
	for _, extra := range [][]string{{"--writer-model", " "}, {"--reviewer-effort", " "}, {"--writer-model", strings.Repeat("x", 129)}} {
		_, root, state, err := codexInitRoleFixture(t, extra)
		if err == nil {
			t.Fatal("invalid role profile accepted")
		}
		for _, path := range []string{filepath.Join(root, "harness.toml"), state} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("rejected profile created configuration/runtime state", path, err)
			}
		}
	}
}
