package gitexec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCommandContextIgnoresHostilePATH(t *testing.T) {
	resolved, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("resolved executable is not absolute: %q", resolved)
	}
	hostile := t.TempDir()
	if runtime.GOOS == "windows" {
		if err := os.WriteFile(filepath.Join(hostile, "git.cmd"), []byte("@echo hostile-git\r\n"), 0600); err != nil {
			t.Fatal(err)
		}
	} else {
		fake := filepath.Join(hostile, "git")
		if err := os.WriteFile(fake, []byte("#!/bin/sh\necho hostile-git\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", hostile+string(os.PathListSeparator)+oldPath)

	command := CommandContext(context.Background(), "--version")
	if command.Path != resolved {
		t.Fatalf("command is not pinned to resolved executable: %q", command.Path)
	}
	if len(command.Args) == 0 || command.Args[0] != "git" || len(command.Args) != 2 || command.Args[1] != "--version" {
		t.Fatalf("Git argv changed: %#v", command.Args)
	}
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(output), "hostile-git") || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(string(output))), "git version ") {
		t.Fatalf("hostile PATH executable was used: %q", output)
	}
	pathGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal("fixture did not put a hostile git first on PATH", err)
	}
	if !strings.EqualFold(filepath.Clean(filepath.Dir(pathGit)), filepath.Clean(hostile)) {
		t.Fatalf("fixture PATH did not resolve the hostile executable first: %q", pathGit)
	}
}
