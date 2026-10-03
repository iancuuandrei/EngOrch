package control

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/runtime"
)

func TestPlannerGraphV2BindsSchemaWithoutChangingLegacyRecipes(t *testing.T) {
	c := config.Config{Planner: runtime.Profile{Runtime: "fake", Provider: "deterministic", Model: "fixture", Effort: "high", Role: "planner"}}
	for _, contract := range []string{plannerContractV1, plannerContractGraphV1} {
		c.PlannerContract = contract
		instruction := plannerAssignmentV1
		if contract == plannerContractGraphV1 {
			instruction = plannerGraphAssignmentV1 + fmt.Sprintf(" Use at most %d research/design tasks in total. Keep the initial plan to at most 40 tasks, leaving task slots for bounded repairs.", c.MaxExplorationRecords())
		}
		oldInput, err := canonical.Bytes(struct {
			Role        string `json:"role"`
			Instruction string `json:"instruction"`
			Objective   string `json:"objective"`
		}{"planner", instruction, "objective"})
		if err != nil {
			t.Fatal(err)
		}
		expected, err := runtime.NewInvocation(c.Planner, string(oldInput))
		if err != nil {
			t.Fatal(err)
		}
		actual, err := plannerInvocation(c, "objective")
		if err != nil || actual != expected {
			t.Fatal("legacy planner recipe changed", contract, err)
		}
	}
	c.PlannerContract = plannerContractGraphV2
	i, err := plannerInvocation(c, "objective")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Schema json.RawMessage `json:"output_schema"`
	}
	if err := json.Unmarshal([]byte(i.Input), &payload); err != nil {
		t.Fatal(err)
	}
	got, _ := canonical.Hash("schema", payload.Schema)
	want, _ := canonical.Hash("schema", engineeringplan.PlannerJSONSchema())
	if got != want {
		t.Fatal("strict graph schema not bound to planner invocation")
	}
	c.PlannerContract = plannerContractGraphV3
	i, err = plannerInvocation(c, "objective")
	if err != nil {
		t.Fatal(err)
	}
	var repairPayload struct {
		Instruction string          `json:"instruction"`
		Schema      json.RawMessage `json:"output_schema"`
	}
	if err := json.Unmarshal([]byte(i.Input), &repairPayload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(repairPayload.Instruction, "maximum repair area") || !strings.Contains(repairPayload.Instruction, "generated-source directives") {
		t.Fatal("repair planner contract omitted conservative-scope and generated-source guidance")
	}
	got, _ = canonical.Hash("schema", repairPayload.Schema)
	if got != want {
		t.Fatal("repair planner contract schema differs from strict graph schema")
	}
}

func TestPlannerGraphV4BindsParallelImplementationContract(t *testing.T) {
	c := config.Config{
		PlannerContract: plannerContractGraphV4,
		Planner:         runtime.Profile{Runtime: "fake", Provider: "deterministic", Model: "fixture", Effort: "high", Role: "planner"},
	}
	i, err := plannerInvocation(c, "objective")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Instruction string          `json:"instruction"`
		Schema      json.RawMessage `json:"output_schema"`
	}
	if err := json.Unmarshal([]byte(i.Input), &payload); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"one initial implementation task",
		"two initial implementation tasks only when the work can be split into genuinely independent paths",
		"concrete, disjoint write_paths",
		"Include native verification and review gates in graph mode; both gates must depend on both implementation tasks when there are two",
		"In direct mode include exactly one implementation",
		"Do not duplicate work to fill a slot",
		"maximum repair area",
		"generated-source directives",
	} {
		if !strings.Contains(payload.Instruction, required) {
			t.Errorf("parallel planner contract omitted %q", required)
		}
	}
	got, _ := canonical.Hash("schema", payload.Schema)
	want, _ := canonical.Hash("schema", engineeringplan.PlannerJSONSchema())
	if got != want {
		t.Fatal("parallel planner contract schema differs from strict graph schema")
	}
}
