package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/control"
)

func TestPreviousRepairAnchorsBindRunAndBoundFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	raw, err := canonical.Bytes(control.RepairDiagnosis{Version: 1, RunID: "run", Findings: []control.RepairFinding{}, Specifications: []control.RepairSpecification{}, Anchors: []control.RepairAnchor{{ID: "anchor", FindingID: "finding"}}})
	if err != nil {
		t.Fatal(err)
	}
	var decoded control.RepairDiagnosis
	if err := canonical.Decode(raw, &decoded); err != nil {
		t.Fatal("fixture report decode", err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	anchors, err := previousRepairAnchors(path, "run")
	if err != nil || len(anchors) != 1 {
		t.Fatal("valid report not read", err)
	}
	if _, err := previousRepairAnchors(path, "different"); err == nil {
		t.Fatal("foreign run report accepted")
	}
	decoded.CandidateID = "different-candidate"
	mismatched, err := canonical.Bytes(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, mismatched, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := previousRepairAnchors(path, "run"); err == nil {
		t.Fatal("anchor/header candidate substitution accepted")
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", (1<<20)+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := previousRepairAnchors(path, "run"); err == nil {
		t.Fatal("oversized report accepted")
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"run_id":"run","unknown":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := previousRepairAnchors(path, "run"); err == nil {
		t.Fatal("unknown report fields accepted")
	}
}

func TestDiagnoseArgumentsRejectBeforeJournalAccess(t *testing.T) {
	for _, args := range [][]string{nil, {"run", "--previous", "report.json"}, {"run", "--unknown"}, {"run", "--anchor", "extra"}} {
		if err := diagnoseCommand(context.Background(), t.TempDir(), args, os.Stdout); err == nil {
			t.Fatalf("invalid diagnosis accepted %v", args)
		}
	}
}
