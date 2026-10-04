package engineeringplan

import (
	"encoding/json"
	"errors"
)

// AgentContextPlannerJSONSchema adds bounded skill selection to the immutable
// graph planner schema. It is shared by prompt construction and wire admission;
// it neither admits skill names nor grants task ownership.
func AgentContextPlannerJSONSchema() (json.RawMessage, error) {
	var schema map[string]any
	if err := json.Unmarshal(PlannerJSONSchema(), &schema); err != nil {
		return nil, err
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid graph planner properties")
	}
	tasks, ok := properties["tasks"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid graph planner tasks")
	}
	items, ok := tasks["items"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid graph planner task items")
	}
	taskProperties, ok := items["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid graph planner task properties")
	}
	required, ok := items["required"].([]any)
	if !ok {
		return nil, errors.New("invalid graph planner required fields")
	}
	taskProperties["skills"] = map[string]any{"type": "array", "maxItems": 4, "items": map[string]any{"type": "string", "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$", "maxLength": 64}}
	items["required"] = append(required, "skills")
	return json.Marshal(schema)
}
