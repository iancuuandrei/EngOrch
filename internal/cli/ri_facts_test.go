package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRIGoFactsUsesCommittedSourceAndReturnsExactBinding(t *testing.T) {
	executable := os.Getenv("ENGORCH_RI_BINARY")
	if executable == "" {
		t.Skip("requires built Rust RI executable")
	}
	executable, err := filepath.Abs(executable)
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	exeHash := sha256.Sum256(binary)
	exeSHA := hex.EncodeToString(exeHash[:])
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatal(err, string(output))
		}
	}
	git("init", "-q")
	committed := []byte("package sample\nimport \"fmt\"\nfunc Committed() { fmt.Println(\"ok\") }\n")
	path := filepath.Join(root, "sample.go")
	if err := os.WriteFile(path, committed, 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "sample.go")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "facts fixture")
	var output bytes.Buffer
	if err := Execute(context.Background(), []string{"init"}, root, &output); err != nil {
		t.Fatal(err)
	}
	dirty := []byte("package wrong\nfunc Dirty() {}\n")
	if err := os.WriteFile(path, dirty, 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := Execute(context.Background(), []string{"ri", "facts", executable, exeSHA, "sample.go"}, root, &output); err != nil {
		t.Fatal(err)
	}
	var result goFactsResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	wantSHA := sha256.Sum256(committed)
	if result.Repository.Commit != result.Source.Commit || result.Repository.Tree == "" || result.Repository.RepositoryID != result.Source.RepositoryID || result.Source.Path != "sample.go" || result.Source.Bytes != int64(len(committed)) || result.Source.Commit == "" || result.Source.SHA256 != hex.EncodeToString(wantSHA[:]) {
		t.Fatalf("source result does not bind committed file: %+v", result.Source)
	}
	if result.Facts.Path != "sample.go" || result.Facts.SourceSHA256 != result.Source.SHA256 || result.Facts.ProducerSHA256 != exeSHA || result.Facts.Cache != "miss" || result.Facts.ParseCount != 1 || result.Facts.Coverage != "PARTIAL" {
		t.Fatalf("Go facts do not bind committed source and parser: %+v", result.Facts)
	}
	if len(result.Facts.Declarations) != 1 || result.Facts.Declarations[0].Name != "Committed" || len(result.Facts.Imports) != 1 || result.Facts.Imports[0].Path != "fmt" {
		t.Fatalf("facts were not extracted from committed source: %+v", result.Facts)
	}
	if bytes.Contains(output.Bytes(), dirty) {
		t.Fatal("output exposed working-tree contents")
	}
	cacheDir := filepath.Join(t.TempDir(), "go-facts-cache")
	for index, wantState := range []string{"miss", "hit"} {
		output.Reset()
		if err := Execute(context.Background(), []string{"ri", "facts", executable, exeSHA, "sample.go", cacheDir}, root, &output); err != nil {
			t.Fatal(err)
		}
		var cached goFactsResult
		if err := json.Unmarshal(output.Bytes(), &cached); err != nil {
			t.Fatal(err)
		}
		if cached.Facts.Cache != wantState || cached.Facts.ParseCount != 1-index || cached.Source.SHA256 != result.Source.SHA256 {
			t.Fatalf("unexpected cache metrics or binding: %+v", cached.Facts)
		}
	}
	for _, args := range [][]string{
		{"ri", "facts", executable, strings.Repeat("0", 64), "sample.go"},
		{"ri", "facts", executable, exeSHA, "../sample.go"},
		{"ri", "facts", executable, exeSHA, "sample.go", "relative-cache"},
		{"ri", "facts", executable, exeSHA, "missing.go"},
	} {
		output.Reset()
		if err := Execute(context.Background(), args, root, &output); err == nil || output.Len() != 0 {
			t.Fatalf("rejected request produced output or succeeded: args=%v err=%v", args, err)
		}
	}
	largePath := filepath.Join(root, "large.go")
	if err := os.WriteFile(largePath, bytes.Repeat([]byte("x"), (1<<20)+1), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "large.go")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "oversized source")
	output.Reset()
	if err := Execute(context.Background(), []string{"ri", "facts", executable, exeSHA, "large.go"}, root, &output); err == nil || output.Len() != 0 {
		t.Fatalf("oversized committed source succeeded or emitted output: %v", err)
	}
}

func TestRIGoFactsRejectsInvalidArgumentsAndUnboundedSource(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{
		{"ri", "facts"},
		{"ri", "facts", "exe", "hash"},
		{"ri", "facts", "exe", "hash", "file.go", "cache", "extra"},
	} {
		var output bytes.Buffer
		if err := Execute(context.Background(), args, root, &output); err == nil || output.Len() != 0 {
			t.Fatalf("invalid arguments accepted or output emitted: %v", args)
		}
	}
	var source goFactsSourceBuffer
	if n, err := source.Write(make([]byte, 1<<20)); err != nil || n != 1<<20 {
		t.Fatalf("exact source limit rejected: n=%d err=%v", n, err)
	}
	if _, err := source.Write([]byte("x")); err != errGoFactsSourceTooLarge {
		t.Fatalf("oversized source did not fail closed: %v", err)
	}
}
