package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCouplingAwareSelectorParsing(t *testing.T) {
	if version, err := validateAutonomousCohortSelector("coupling-aware-v1", true); err != nil || version != 2 {
		t.Fatalf("staged coupling-aware selector rejected: %d %v", version, err)
	}
	if _, err := validateAutonomousCohortSelector("coupling-aware-v1", false); err == nil {
		t.Fatal("non-staged coupling-aware selector admitted")
	}
	if got := cohortSelectorFlag(2); got != autonomousCohortSelectorCouplingAwareV1 {
		t.Fatalf("selector version 2 did not render its flag spelling: %q", got)
	}
	if got := cohortSelectorFlag(1); got != autonomousCohortSelectorLexicographicV1 {
		t.Fatalf("frozen selector 1 spelling changed: %q", got)
	}
	if got := cohortSelectorFlag(0); got != "" {
		t.Fatalf("greedy selector changed legacy spelling: %q", got)
	}
	if _, err := validateAutonomousCohortSelector("coupling-aware-v2", true); err == nil {
		t.Fatal("unknown coupling-aware selector admitted")
	}
	if _, err := validateAutonomousCohortSelector(cohortSelectorFlag(9), true); err == nil {
		t.Fatal("out-of-range selector version admitted")
	}
}

func TestCouplingAwareSelectorName(t *testing.T) {
	if got := cohortSelectorName(2); got != autonomousCohortSelectorCouplingAwareV1 {
		t.Fatalf("inspect-plan selector name wrong: %q", got)
	}
	if got := cohortSelectorName(1); got != autonomousCohortSelectorLexicographicV1 {
		t.Fatalf("frozen selector 1 display changed: %q", got)
	}
	if got := cohortSelectorName(0); got != "" {
		t.Fatalf("greedy display changed: %q", got)
	}
}

func TestCouplingAwareSelectorRequiresStagedAndFallbackClosed(t *testing.T) {
	root := autonomousCLIFixture(t)
	policyPath := filepath.Join(root, "isolation-policy.json")
	if err := os.WriteFile(policyPath, []byte(isolatedWriterPolicyJSON()), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"run", "--autonomous", "--cohort-selector", "coupling-aware-v1", "objective"},
		{"run", "--autonomous", "--isolated-writers", "--isolation-policy", policyPath, "--cohort-selector", "coupling-aware-v1", "objective"},
		{"run", "--autonomous", "--isolated-writer-waves", "--isolation-policy", policyPath, "--cohort-selector", "coupling-aware-v1", "objective"},
		{"run", "--autonomous", "--isolated-writer-staged", "--isolation-policy", policyPath, "--cohort-selector", "coupling-aware-v2", "objective"},
		{"run", "--autonomous", "--parallel-writers", "--cohort-selector", "coupling-aware-v1", "objective"},
	} {
		var out bytes.Buffer
		if err := Execute(context.Background(), args, root, &out); err == nil {
			t.Errorf("accepted invalid coupling-aware invocation %q", strings.Join(args, " "))
		}
	}
	if entries, err := filepath.Glob(filepath.Join(root, ".harness", "runs", "*.jsonl")); err != nil || len(entries) != 0 {
		t.Fatalf("invalid coupling-aware arguments created runs: %v, %v", entries, err)
	}
}
