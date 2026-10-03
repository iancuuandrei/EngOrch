package engineeringplan

import (
	"encoding/json"
	"testing"
)

func TestPlannerSchemaStrictObjectProperties(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(PlannerJSONSchema(), &schema); err != nil {
		t.Fatal(err)
	}
	var check func(map[string]any)
	check = func(node map[string]any) {
		if _, ok := node["type"].(string); !ok {
			t.Fatal("wire schema node lacks explicit type", node)
		}
		if properties, ok := node["properties"].(map[string]any); ok {
			required, ok := node["required"].([]any)
			if !ok || len(required) != len(properties) || node["additionalProperties"] != false {
				t.Fatal("strict object must require every property", node)
			}
			keys := map[string]bool{}
			for _, key := range required {
				keys[key.(string)] = true
			}
			for key, child := range properties {
				if !keys[key] {
					t.Fatal("property missing from required", key)
				}
				check(child.(map[string]any))
			}
		}
		if items, ok := node["items"].(map[string]any); ok {
			check(items)
		}
	}
	check(schema)
}
