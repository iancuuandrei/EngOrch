package control

import (
	"encoding/json"
	"errors"
	"fmt"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/runtime"
)

const plannerContractV1 = "plan-v1"

const plannerContractGraphV1 = "plan-graph-v1"
const plannerContractGraphV2 = "plan-graph-v2"
const plannerContractGraphV3 = "plan-graph-v3"
const plannerContractGraphV4 = "plan-graph-v4"

const plannerAssignmentV1 = "As planner, produce an implementation plan using read-only source tools. Editing and testing belong to later roles. Read-only access is expected and is not a blocker. Return the plan without requesting additional capability."

const plannerGraphAssignmentV1 = "As planner, produce a bounded engineering task graph as strict JSON matching the engineeringplan v1 schema (version 1, mode direct|graph, summary, tasks with id/kind/title/scope_paths/write_paths/expected_evidence/estimated_seconds and optional parent_id/dependencies). Use kinds research, design, implementation, verification, review. For graph mode include exactly one implementation with concrete write_paths within scope, research/design dependencies, and verification/review gates depending on the implementation. For direct mode include exactly one implementation with concrete write_paths. Never include completed or attempts; the runner owns progress. Return only the JSON object."

const plannerGraphAssignmentV4 = "As planner, produce a bounded engineering task graph as strict JSON matching the engineeringplan v1 schema (version 1, mode direct|graph, summary, tasks with id/kind/title/scope_paths/write_paths/expected_evidence/estimated_seconds and optional parent_id/dependencies). Use kinds research, design, implementation, verification, review. In graph mode, plan one initial implementation task for work that should remain together. Use two initial implementation tasks only when the work can be split into genuinely independent paths after any shared research/design: each task must own concrete, disjoint write_paths within scope and depend on the shared research/design it needs. Do not duplicate work to fill a slot. Include native verification and review gates in graph mode; both gates must depend on both implementation tasks when there are two, and on the implementation when there is one. In direct mode include exactly one implementation with concrete write_paths and no verification or review tasks; the runner performs native gates. Never include completed or attempts; the runner owns progress. Return only the JSON object."

// plannerInvocation binds the configured planner contract to the exact
// objective. The empty contract deliberately retains the historical raw input
// identity so journals created before planner contracts remain replayable.
// plan-graph-v1 binds the strict autonomous graph instruction and still
// hashes the exact planner result as the PlanID; no extra approval authority
// is introduced.
func plannerInvocation(c config.Config, objective string) (runtime.Invocation, error) {
	input := objective
	if c.PlannerContract != "" || c.ReviewerContract == "json-v1" {
		assignment := plannerAssignmentV1
		var schema json.RawMessage
		var verificationChecks []config.Check
		switch c.PlannerContract {
		case "":
		case plannerContractV1:
			assignment = plannerAssignmentV1
		case plannerContractGraphV1:
			assignment = plannerGraphAssignmentV1 + fmt.Sprintf(" Use at most %d research/design tasks in total. Keep the initial plan to at most 40 tasks, leaving task slots for bounded repairs.", c.MaxExplorationRecords())
		case plannerContractGraphV2:
			assignment = plannerGraphAssignmentV1 + fmt.Sprintf(" Use at most %d research/design tasks in total. Keep the initial plan to at most 40 tasks, leaving task slots for bounded repairs.", c.MaxExplorationRecords()) + " In direct mode, include exactly one task total: the implementation task. Do not include verification or review tasks; the runner performs those native gates. expected_evidence MUST contain objects with nonempty kind and description strings, never plain strings. scope_paths may use a single dot for the repository root; write_paths must name concrete relative files or directories, never dot or parent paths. The review gate must depend on verification."
			schema = engineeringplan.PlannerJSONSchema()
		case plannerContractGraphV3:
			readBudget := c.MaxExplorationRecords() - 8
			if readBudget < 0 {
				readBudget = 0
			}
			assignment = plannerGraphAssignmentV1 + fmt.Sprintf(" Use at most %d research/design tasks in total. Keep the initial plan to at most 32 tasks and at most %d research/design tasks, reserving worst-case capacity for eight four-task repair slots.", readBudget, readBudget) + " In direct mode, include exactly one task total: the implementation task. Do not include verification or review tasks; the runner performs those native gates. expected_evidence MUST contain objects with nonempty kind and description strings, never plain strings. scope_paths may use a single dot for the repository root; write_paths must name concrete relative files or directories, never dot or parent paths. The review gate must depend on verification. For the initial implementation, treat scope_paths as the maximum repair area and write_paths as only the initial change. Inspect generated-source directives, templates, and extension points, and include relevant files in scope_paths even when they are not initially changed; do not add them to write_paths unless the first implementation edits them."
			schema = engineeringplan.PlannerJSONSchema()
		case plannerContractGraphV4:
			readBudget := c.MaxExplorationRecords() - 8
			if readBudget < 0 {
				readBudget = 0
			}
			assignment = plannerGraphAssignmentV4 + fmt.Sprintf(" Use at most %d research/design tasks in total. Keep the initial plan to at most 32 tasks and at most %d research/design tasks, reserving worst-case capacity for eight four-task repair slots.", readBudget, readBudget) + " expected_evidence MUST contain objects with nonempty kind and description strings, never plain strings. scope_paths may use a single dot for the repository root; write_paths must name concrete relative files or directories, never dot or parent paths. The review gate must depend on verification. For the initial implementation, treat scope_paths as the maximum repair area and write_paths as only the initial change. Inspect generated-source directives, templates, and extension points, and include relevant files in scope_paths even when they are not initially changed; do not add them to write_paths unless the first implementation edits them."
			schema = engineeringplan.PlannerJSONSchema()
		default:
			return runtime.Invocation{}, errors.New("unsupported planner contract")
		}
		if c.ReviewerContract == "json-v1" {
			verificationChecks = append([]config.Check(nil), c.Verification...)
			assignment += " The verification gate may claim only the configured verification checks supplied in verification_checks. Treat additional checks as planned recommendations only; never imply they ran or passed without recorded observations."
			assignment += " Prefer direct mode for localized changes with known scope. Use graph mode for substantial or dependent work; independent read-only tasks may run concurrently, and research/design tasks are not mandatory for every small change."
		}
		wrapped, err := canonical.Bytes(struct {
			OutputSchema       json.RawMessage `json:"output_schema,omitempty"`
			Role               string          `json:"role"`
			Instruction        string          `json:"instruction"`
			Objective          string          `json:"objective"`
			VerificationChecks []config.Check  `json:"verification_checks,omitempty"`
		}{schema, "planner", assignment, objective, verificationChecks})
		if err != nil {
			return runtime.Invocation{}, err
		}
		input = string(wrapped)
	}
	profile, err := modelProfileForInput(c, "planner", input, 0)
	if err != nil {
		return runtime.Invocation{}, err
	}
	return runtime.NewInvocation(profile, input)
}
