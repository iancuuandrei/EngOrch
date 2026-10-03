package repository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCopySelectedSourceBatchReadsOnlyBoundSelectedLeaves(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if output, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(output))
		}
	}
	run("init", "-q")
	for path, content := range map[string]string{"a.go": "package a\n", "b.go": "package b\n", "notes.txt": "unselected"} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	run("add", ".")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "selected fixture")
	identity, err := Discover(context.Background(), root, "selected-fixture")
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]SourceEntry{}
	if err := VisitSource(context.Background(), identity, func(entry SourceEntry) error {
		entries[entry.Path] = entry
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// A dirty checkout must not change bytes from the selected committed object.
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package dirty\n"), 0600); err != nil {
		t.Fatal(err)
	}
	writer := &selectedSourceWriter{}
	opened := []string{}
	visited := []string{}
	if err := CopySelectedSourceBatch(context.Background(), identity, []SourceEntry{entries["a.go"]}, func(entry SourceEntry, size int64) (io.WriteCloser, error) {
		opened = append(opened, entry.Path)
		if size != int64(len("package a\n")) {
			return nil, errors.New("wrong committed size")
		}
		return writer, nil
	}, func(entry SourceEntry, digest *SourceDigest) error {
		if digest == nil || !writer.closed || writer.String() != "package a\n" {
			return errors.New("selected bytes were not delivered as committed content")
		}
		want := sha256.Sum256([]byte("package a\n"))
		if digest.RepositoryID == "" || digest.Commit != identity.Commit || digest.Path != entry.Path || digest.Blob != entry.Object || digest.Bytes != int64(len("package a\n")) || digest.SHA256 != hex.EncodeToString(want[:]) {
			return errors.New("selected digest binding mismatch")
		}
		visited = append(visited, entry.Path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(opened) != 1 || opened[0] != "a.go" || len(visited) != 1 || visited[0] != "a.go" {
		t.Fatalf("unselected leaves were delivered: opened=%v visited=%v", opened, visited)
	}
}

func TestCopySelectedSourceBatchRejectsInvalidOrSubstitutedSelection(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if output, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(output))
		}
	}
	run("init", "-q")
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n"), 0600); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "selection fixture")
	identity, err := Discover(context.Background(), root, "selected-fixture")
	if err != nil {
		t.Fatal(err)
	}
	var observed SourceEntry
	if err := VisitSource(context.Background(), identity, func(entry SourceEntry) error { observed = entry; return nil }); err != nil {
		t.Fatal(err)
	}
	noop := func(SourceEntry, int64) (io.WriteCloser, error) { return discardWriteCloser{}, nil }
	visit := func(SourceEntry, *SourceDigest) error { return nil }
	if err := CopySelectedSourceBatch(context.Background(), identity, []SourceEntry{observed, observed}, noop, visit); err == nil {
		t.Fatal("duplicate selection accepted")
	}
	substituted := observed
	substituted.Object = string(bytes.Repeat([]byte{'0'}, len(observed.Object)))
	if err := CopySelectedSourceBatch(context.Background(), identity, []SourceEntry{substituted}, noop, visit); err == nil {
		t.Fatal("object substitution accepted")
	}
	missing := observed
	missing.Path = "missing.go"
	if err := CopySelectedSourceBatch(context.Background(), identity, []SourceEntry{missing}, noop, visit); err == nil {
		t.Fatal("path absent from exact tree accepted")
	}
}

func TestCopySelectedSourceBatchCanOmitLargeBlobWithoutOpeningContents(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if output, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(output))
		}
	}
	run("init", "-q")
	largeSize := int64(64<<20) + 1
	largeFile, err := os.Create(filepath.Join(root, "a-large.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := largeFile.Truncate(largeSize); err != nil {
		largeFile.Close()
		t.Fatal(err)
	}
	if err := largeFile.Close(); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string][]byte{"b-small.go": []byte("package small\n")} {
		if err := os.WriteFile(filepath.Join(root, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	run("add", ".")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "large blob skip fixture")
	identity, err := Discover(context.Background(), root, "selected-fixture")
	if err != nil {
		t.Fatal(err)
	}
	selected := []SourceEntry{}
	if err := VisitSource(context.Background(), identity, func(entry SourceEntry) error { selected = append(selected, entry); return nil }); err != nil {
		t.Fatal(err)
	}
	opened := []string{}
	visited := map[string]*SourceDigest{}
	if err := CopySelectedSourceBatch(context.Background(), identity, selected, func(entry SourceEntry, size int64) (io.WriteCloser, error) {
		if entry.Path == "a-large.go" {
			if size != largeSize {
				return nil, errors.New("wrong large blob size")
			}
			return nil, nil
		}
		opened = append(opened, entry.Path)
		return discardWriteCloser{}, nil
	}, func(entry SourceEntry, digest *SourceDigest) error {
		visited[entry.Path] = digest
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(opened) != 1 || opened[0] != "b-small.go" || visited["a-large.go"] != nil || visited["b-small.go"] == nil {
		t.Fatalf("large object was not skipped before contents: opened=%v visited=%v", opened, visited)
	}
}

type selectedSourceWriter struct {
	bytes.Buffer
	closed bool
}

func (w *selectedSourceWriter) Close() error { w.closed = true; return nil }

type discardWriteCloser struct{}

func (discardWriteCloser) Write(value []byte) (int, error) { return io.Discard.Write(value) }
func (discardWriteCloser) Close() error                    { return nil }
