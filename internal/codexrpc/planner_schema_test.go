package codexrpc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/runtime"
)

func TestPlannerGraphSchemaBoundToWire(t *testing.T) {
	p := runtime.Profile{Runtime: "codex-app-server", Provider: "openai", Model: "exact", Effort: "high", Role: "planner"}
	for _, schema := range []json.RawMessage{nil, engineeringplan.PlannerJSONSchema()} {
		payload := map[string]any{}
		if schema != nil {
			payload["output_schema"] = schema
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		i, err := runtime.NewInvocation(p, string(raw))
		if err != nil {
			t.Fatal(err)
		}
		settings := ThreadSettings{ThreadID: "thread", Model: p.Model, Provider: p.Provider, Effort: &p.Effort, Directory: t.TempDir(), Approval: "never", Sandbox: "readOnly"}
		request := scriptedResponse(t, "", map[string]any{"turn": map[string]any{"id": "turn", "status": "completed", "items": []any{}}}, func(ctx context.Context, c *Client) error {
			_, _, err := c.StartTurn(ctx, settings, i)
			return err
		})
		var params map[string]json.RawMessage
		if err := json.Unmarshal(request.Params, &params); err != nil {
			t.Fatal(err)
		}
		wireSchema, present := params["outputSchema"]
		if present != (schema != nil) {
			t.Fatal("planner wire schema presence differs from invocation")
		}
		if present {
			got, _ := canonical.Hash("schema", wireSchema)
			want, _ := canonical.Hash("schema", schema)
			if got != want {
				t.Fatal("planner schema changed on wire")
			}
		}
	}
}

func TestPlannerSchemaSubstitutionFailsBeforeRPC(t *testing.T) {
	p := runtime.Profile{Runtime: "codex-app-server", Provider: "openai", Model: "exact", Effort: "high", Role: "planner"}
	i, err := runtime.NewInvocation(p, `{"output_schema":{"type":"string"}}`)
	if err != nil {
		t.Fatal(err)
	}
	settings := ThreadSettings{ThreadID: "thread", Model: p.Model, Provider: p.Provider, Effort: &p.Effort, Directory: t.TempDir(), Approval: "never", Sandbox: "readOnly"}
	_, _, err = (&Client{}).StartTurn(context.Background(), settings, i)
	if err == nil || !strings.Contains(err.Error(), "planner output schema substitution") {
		t.Fatal("invalid schema reached provider transport", err)
	}
}
