package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphSchedulerSidecarNames(t *testing.T) {
	runID := strings.Repeat("a", 64)
	valid := runID + ".jsonl.graph-schedule-0123456789abcdef.jsonl"
	for _, name := range []string{valid, filepath.Join("runs", valid)} {
		if !graphSchedulerSidecar(name) {
			t.Fatalf("canonical scheduler sidecar rejected: %s", name)
		}
	}
	for _, name := range []string{
		runID + ".jsonl",
		"unknown.jsonl",
		strings.Repeat("A", 64) + ".jsonl.graph-schedule-0123456789abcdef.jsonl",
		runID + ".jsonl.graph-schedule-0123456789abcde.jsonl",
		runID + ".jsonl.graph-schedule-0123456789abcdeg.jsonl",
		runID + ".jsonl.graph-schedule-0123456789abcdef.jsonl.extra",
	} {
		if graphSchedulerSidecar(name) {
			t.Fatalf("noncanonical entry ignored: %s", name)
		}
	}
}

func TestRunListingIgnoresCanonicalGraphSchedulerSidecars(t *testing.T) {
	root, runID, _ := diffWorkspaceFixture(t)
	runs := filepath.Join(root, ".harness", "runs")
	sidecar := filepath.Join(runs, runID+".jsonl.graph-schedule-0123456789abcdef.jsonl")
	if err := os.WriteFile(sidecar, []byte("separate scheduler event contract"), 0600); err != nil {
		t.Fatal(err)
	}
	latest, err := latestRunID(root)
	if err != nil || latest != runID {
		t.Fatalf("latest selection treated scheduler journal as a run: %s %v", latest, err)
	}
	var listed []runStatus
	if err := json.Unmarshal(mustExecuteCLI(t, root, "status"), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].RunID != runID {
		t.Fatalf("status listed a scheduler sidecar: %#v", listed)
	}
	mustExecuteCLI(t, root, "inspect")
	mustExecuteCLI(t, root, "diff")
	// An unknown or malformed journal must retain the existing fail-closed
	// behavior; filtering scheduler names must not hide it.
	if err := os.WriteFile(filepath.Join(runs, "unknown.jsonl"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := latestRunID(root); err == nil {
		t.Fatal("latest selection hid an unknown corrupt journal")
	}
	var out bytes.Buffer
	if err := Execute(context.Background(), []string{"status"}, root, &out); err == nil {
		t.Fatal("status hid an unknown corrupt journal")
	}
}
