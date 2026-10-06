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

func TestRunJournalSidecarNames(t *testing.T) {
	runID := strings.Repeat("a", 64)
	turnID := strings.Repeat("b", 64)
	invocationID := strings.Repeat("c", 64)
	valid := runID + ".jsonl.graph-schedule-0123456789abcdef.jsonl"
	writers := runID + ".jsonl.graph-writers-0123456789abcdef.jsonl"
	modelAccess := runID + ".jsonl.model-access.jsonl"
	ignored := []string{valid, writers, modelAccess}
	for _, role := range []string{"planner", "explorer", "writer", "fixer", "reviewer"} {
		for _, suffix := range []string{".opencode-runtime.jsonl", ".provider-gateway.jsonl", ".provider-runtime.jsonl"} {
			ignored = append(ignored, runID+".jsonl."+role+suffix)
		}
		// Repeated invocations only exist for opencode-runtime and its
		// provider-gateway; direct provider-runtime never takes an
		// invocation suffix.
		for _, suffix := range []string{".opencode-runtime.jsonl", ".provider-gateway.jsonl"} {
			ignored = append(ignored, runID+".jsonl."+role+".invocation-"+invocationID+suffix)
		}
		ignored = append(ignored,
			runID+".jsonl."+role+".opencode-runtime.jsonl.state-root.jsonl",
			runID+".jsonl."+role+".invocation-"+invocationID+".opencode-runtime.jsonl.state-root.jsonl",
		)
	}
	// Turn-scoped journals are only produced for explorer dynamic turns.
	for _, suffix := range []string{".opencode-runtime.jsonl", ".provider-gateway.jsonl", ".provider-runtime.jsonl"} {
		ignored = append(ignored, runID+".jsonl.explorer.turn-"+turnID+suffix)
	}
	ignored = append(ignored, runID+".jsonl.explorer.turn-"+turnID+".opencode-runtime.jsonl.state-root.jsonl")
	for _, name := range ignored {
		for _, candidate := range []string{name, filepath.Join("runs", name)} {
			if !runJournalSidecar(candidate) {
				t.Fatalf("canonical sidecar rejected: %s", candidate)
			}
		}
	}
	for _, name := range []string{
		runID + ".jsonl",
		"unknown.jsonl",
		runID + ".jsonl.model-access.jsonl.extra",
		runID + ".jsonl.model-access-extra.jsonl",
		strings.Repeat("A", 64) + ".jsonl.model-access.jsonl",
		runID + ".jsonl.planner.opencode-runtime.jsonl.extra",
		runID + ".jsonl.planner.provider-gateway.jsonl.extra",
		runID + ".jsonl.planner.provider-runtime.jsonl.extra",
		runID + ".jsonl.planner.opencode-runtime.jsonl.state-root.jsonl.extra",
		runID + ".jsonl.planner.provider-gateway.jsonl.state-root.jsonl",
		runID + ".jsonl.planner.provider-runtime.jsonl.state-root.jsonl",
		runID + ".jsonl.unknown.opencode-runtime.jsonl",
		runID + ".jsonl.Planner.opencode-runtime.jsonl",
		runID + ".jsonl.planner.opencode-Runtime.jsonl",
		runID + ".jsonl.planner.tool-receipts",
		runID + ".jsonl.planner.opencode-runtime.jsonl.state-root.jsonl-shm",
		runID + ".jsonl.explorer.turn-" + strings.Repeat("B", 64) + ".opencode-runtime.jsonl",
		runID + ".jsonl.explorer.turn-abc.opencode-runtime.jsonl",
		runID + ".jsonl.explorer.invocation-abc.provider-gateway.jsonl",
		runID + ".jsonl.planner.invocation-" + invocationID + ".provider-runtime.jsonl",
		runID + ".jsonl.explorer.invocation-" + invocationID + ".provider-runtime.jsonl",
		runID + ".jsonl.planner.turn-" + turnID + ".opencode-runtime.jsonl",
		runID + ".jsonl.writer.turn-" + turnID + ".provider-gateway.jsonl",
		runID + ".jsonl.reviewer.turn-" + turnID + ".provider-runtime.jsonl",
		runID + ".jsonl.planner.turn-" + turnID + ".opencode-runtime.jsonl.state-root.jsonl",
		runID + ".jsonl.planner.turn-" + turnID + ".invocation-" + invocationID + ".opencode-runtime.jsonl",
		runID + ".jsonl.planner.invocation-" + invocationID + ".turn-" + turnID + ".opencode-runtime.jsonl",
		runID + ".jsonl.planner.task-abc.opencode-runtime.jsonl",
		runID + ".jsonl..opencode-runtime.jsonl",
		".jsonl.planner.opencode-runtime.jsonl",
		strings.Repeat("A", 64) + ".jsonl.graph-schedule-0123456789abcdef.jsonl",
		runID + ".jsonl.graph-schedule-0123456789abcde.jsonl",
		runID + ".jsonl.graph-schedule-0123456789abcdeg.jsonl",
		runID + ".jsonl.graph-schedule-0123456789abcdef.jsonl.extra",
		runID + ".jsonl.graph-writers-0123456789abcdeg.jsonl",
		runID + ".jsonl.graph-writers-0123456789abcde.jsonl",
		strings.Repeat("A", 64) + ".jsonl.graph-writers-0123456789abcdef.jsonl",
		runID + ".jsonl.graph-writers-0123456789abcdef.jsonl.extra",
		runID + ".jsonl.graph-unknown-0123456789abcdef.jsonl",
	} {
		if runJournalSidecar(name) {
			t.Fatalf("noncanonical entry ignored: %s", name)
		}
	}
}

func TestRunListingIgnoresCanonicalGraphSchedulerSidecars(t *testing.T) {
	root, runID, _ := diffWorkspaceFixture(t)
	runs := filepath.Join(root, ".harness", "runs")
	for _, kind := range []string{"schedule", "writers"} {
		sidecar := filepath.Join(runs, runID+".jsonl.graph-"+kind+"-0123456789abcdef.jsonl")
		if err := os.WriteFile(sidecar, []byte("separate scheduler event contract"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(runs, runID+".jsonl.model-access.jsonl"), []byte("separate model access event contract"), 0600); err != nil {
		t.Fatal(err)
	}
	// Canonical provider/runtime sidecars share the runs directory with the
	// controller journal. They must not break status or implicit latest
	// selection. Turn scoping is explorer-only and invocation scoping covers
	// opencode-runtime and provider-gateway; direct provider-runtime never
	// takes an invocation suffix.
	turnID := strings.Repeat("b", 64)
	invocationID := strings.Repeat("c", 64)
	for _, sidecar := range []string{
		runID + ".jsonl.planner.opencode-runtime.jsonl",
		runID + ".jsonl.planner.provider-gateway.jsonl",
		runID + ".jsonl.planner.opencode-runtime.jsonl.state-root.jsonl",
		runID + ".jsonl.explorer.turn-" + turnID + ".opencode-runtime.jsonl",
		runID + ".jsonl.explorer.turn-" + turnID + ".provider-gateway.jsonl",
		runID + ".jsonl.explorer.turn-" + turnID + ".provider-runtime.jsonl",
		runID + ".jsonl.writer.invocation-" + invocationID + ".opencode-runtime.jsonl",
		runID + ".jsonl.writer.invocation-" + invocationID + ".provider-gateway.jsonl",
		runID + ".jsonl.reviewer.provider-gateway.jsonl",
	} {
		if err := os.WriteFile(filepath.Join(runs, sidecar), []byte("separate provider runtime event contract"), 0600); err != nil {
			t.Fatal(err)
		}
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
	// A direct-provider invocation suffix is not a canonical producer
	// grammar, so it must stay fail-closed rather than being ignored.
	invalidDirect := filepath.Join(runs, runID+".jsonl.writer.invocation-"+invocationID+".provider-runtime.jsonl")
	if err := os.WriteFile(invalidDirect, []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := latestRunID(root); err == nil {
		t.Fatal("latest selection hid a noncanonical direct-provider sidecar")
	}
	var out bytes.Buffer
	if err := Execute(context.Background(), []string{"status"}, root, &out); err == nil {
		t.Fatal("status hid a noncanonical direct-provider sidecar")
	}
	if err := os.Remove(invalidDirect); err != nil {
		t.Fatal(err)
	}
	// An unknown or malformed journal must retain the existing fail-closed
	// behavior; filtering scheduler names must not hide it.
	if err := os.WriteFile(filepath.Join(runs, "unknown.jsonl"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := latestRunID(root); err == nil {
		t.Fatal("latest selection hid an unknown corrupt journal")
	}
	out.Reset()
	if err := Execute(context.Background(), []string{"status"}, root, &out); err == nil {
		t.Fatal("status hid an unknown corrupt journal")
	}
	if err := os.Remove(filepath.Join(runs, "unknown.jsonl")); err != nil {
		t.Fatal(err)
	}
	// An unknown sidecar grammar alongside a valid run must also stay
	// fail-closed rather than being silently ignored.
	if err := os.WriteFile(filepath.Join(runs, runID+".jsonl.unknown-sidecar.jsonl"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := latestRunID(root); err == nil {
		t.Fatal("latest selection hid an unknown sidecar journal")
	}
	out.Reset()
	if err := Execute(context.Background(), []string{"status"}, root, &out); err == nil {
		t.Fatal("status hid an unknown sidecar journal")
	}
	if err := os.Remove(filepath.Join(runs, runID+".jsonl.unknown-sidecar.jsonl")); err != nil {
		t.Fatal(err)
	}
	// A corrupt true run journal must stay fail-closed even when canonical
	// sidecars are present.
	corruptID := strings.Repeat("d", 64)
	if err := os.WriteFile(filepath.Join(runs, corruptID+".jsonl"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := latestRunID(root); err == nil {
		t.Fatal("latest selection hid a corrupt run journal")
	}
	out.Reset()
	if err := Execute(context.Background(), []string{"status"}, root, &out); err == nil {
		t.Fatal("status hid a corrupt run journal")
	}
}
