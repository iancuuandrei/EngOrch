package opencoderuntime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"harness.local/engorch/internal/opencode"
)

func TestDispatchFailureSidecarPreservesFirstDiagnosticOnly(t *testing.T) {
	client, err := opencode.NewClient("http://127.0.0.1:1", "fixture", "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, cause := client.SubmitSynchronousToolTurn(ctx, "", "", opencode.SynchronousToolDispatchIntent{})
	if opencode.DispatchDiagnosticFromError(cause) == nil {
		t.Fatal("missing diagnostic", cause)
	}
	cfg := ExecuteConfig{RuntimePath: filepath.Join(t.TempDir(), "runtime.jsonl")}
	cfg.Intent.Invocation.ID = strings.Repeat("a", 64)
	cause = errors.Join(cause, errors.New("TOKEN=private"))
	if err := preserveDispatchDiagnostic(cfg, cause); err != nil {
		t.Fatal(err)
	}
	path := cfg.RuntimePath + ".failure.json"
	first, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(first), "private") || !strings.Contains(string(first), `"status":"UNKNOWN"`) || !strings.Contains(string(first), cfg.Intent.Invocation.ID) {
		t.Fatal("unsafe failure artifact", err, string(first))
	}
	if err := preserveDispatchDiagnostic(cfg, errors.New("plain failure")); err != nil {
		t.Fatal(err)
	}
	if err := preserveDispatchDiagnostic(cfg, cause); err != nil {
		t.Fatal(err)
	}
	again, _ := os.ReadFile(path)
	if string(first) != string(again) {
		t.Fatal("first diagnostic overwritten")
	}
	if _, err := os.Stat(cfg.RuntimePath); !os.IsNotExist(err) {
		t.Fatal("diagnostic fabricated a runtime journal")
	}
}
