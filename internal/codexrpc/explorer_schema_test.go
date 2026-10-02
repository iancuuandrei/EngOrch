package codexrpc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"harness.local/engorch/internal/runtime"
)

func TestExplorerSchemaRequiresStringSummaryOnWire(t *testing.T) {
	p := runtime.Profile{Runtime: "codex-app-server", Provider: "openai", Model: "exact", Effort: "low", Role: "explorer"}
	raw, _ := json.Marshal(map[string]any{"output_schema": runtime.ExplorerOutputSchema()})
	i, err := runtime.NewInvocation(p, string(raw))
	if err != nil {
		t.Fatal(err)
	}
	settings := ThreadSettings{ThreadID: "thread", Model: p.Model, Provider: p.Provider, Effort: &p.Effort, Directory: t.TempDir(), Approval: "never", Sandbox: "readOnly"}
	request := scriptedResponse(t, "", map[string]any{"turn": map[string]any{"id": "turn", "status": "completed", "items": []any{}}}, func(ctx context.Context, c *Client) error {
		_, _, err := c.StartTurn(ctx, settings, i)
		return err
	})
	var params struct {
		Schema struct {
			Properties struct {
				Summary struct {
					Type string `json:"type"`
				} `json:"summary"`
			} `json:"properties"`
		} `json:"outputSchema"`
	}
	if err := json.Unmarshal(request.Params, &params); err != nil || params.Schema.Properties.Summary.Type != "string" {
		t.Fatal("explorer summary was not constrained to string", err)
	}
}

func TestExplorerSchemaSubstitutionFailsBeforeRPC(t *testing.T) {
	p := runtime.Profile{Runtime: "codex-app-server", Provider: "openai", Model: "exact", Effort: "low", Role: "explorer"}
	raw := strings.Replace(string(runtime.ExplorerOutputSchema()), `"summary":{"type":"string"}`, `"summary":{"type":"object"}`, 1)
	i, err := runtime.NewInvocation(p, `{"output_schema":`+raw+`}`)
	if err != nil {
		t.Fatal(err)
	}
	settings := ThreadSettings{ThreadID: "thread", Model: p.Model, Provider: p.Provider, Effort: &p.Effort, Directory: t.TempDir(), Approval: "never", Sandbox: "readOnly"}
	c := &Client{}
	_, _, err = c.StartTurn(context.Background(), settings, i)
	if err == nil || !strings.Contains(err.Error(), "schema substitution") {
		t.Fatal("invalid schema reached provider transport", err)
	}
}
