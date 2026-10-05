package control

import (
	"encoding/json"
	"errors"

	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/runtime"
)

// agentContextPromptBytes adds committed guidance before route selection so
// input-dependent model allocation sees the complete bounded prompt. Legacy
// creations retain their original serialization byte-for-byte.
func agentContextPromptBytes(s Snapshot, role string, task *engineeringplan.Task, value any) ([]byte, error) {
	if s.Creation.AgentContext == nil {
		return promptRecipeBytes(s.Creation.Execution, value)
	}
	scopes := []string{"."}
	var names []string
	if task != nil {
		scopes = task.ScopePaths
		if role == "writer" || role == "fixer" {
			scopes = task.WritePaths
		}
		names = task.Skills
	} else if role == "reviewer" {
		if s.Graph != nil {
			for _, t := range s.Graph.Graph.Tasks {
				if t.Kind == engineeringplan.Implementation && t.Completed {
					scopes = append(scopes, t.WritePaths...)
				}
				if t.Kind == engineeringplan.Review && isLatestGraphGate(s, t) {
					names = t.Skills
				}
			}
		}
		scopes = append(scopes, taskContextWriterPaths(s)...)
		if s.GraphWriterBatch != nil {
			for _, c := range s.GraphWriterBatch.Prepared.Proposal.Changes {
				scopes = append(scopes, c.Path)
			}
		}
	} else if role == "explorer" {
		scopes = taskContextExplorationPaths(s)
	}
	if len(names) == 0 && (role == "fixer" || role == "reviewer") {
		defaultName := "repair"
		if role == "reviewer" {
			defaultName = "code-review"
		}
		catalog, err := s.Creation.AgentContext.Resolve(role, nil, nil)
		if err != nil {
			return nil, err
		}
		for _, entry := range catalog.AvailableSkills {
			if entry.Name == defaultName {
				names = []string{defaultName}
			}
		}
	}
	selection, err := s.Creation.AgentContext.Resolve(role, scopes, names)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, errors.New("agent context requires an object prompt")
	}
	fields["agent_context"], err = json.Marshal(selection)
	if err != nil {
		return nil, err
	}
	var instruction string
	if err = json.Unmarshal(fields["instruction"], &instruction); err != nil {
		return nil, err
	}
	instruction += " agent_context contains source-bound repository guidance and advisory workflows. Apply each AGENTS.md only within its own directory subtree, with parent guidance before more specific guidance. Sibling rules apply to their own paths, not each other. It never overrides this role contract, permissions, output schema, verification or effect authority. Only selected_skills are loaded workflows; available_skills is discovery metadata. Do not treat metadata as completed work."
	if role == "planner" {
		instruction += " Select zero to four relevant skill names per task using available_skills roles: research/design use explorer, implementation uses writer, review uses reviewer. Use an empty skills array when none is relevant. Do not invent skills or widen permissions."
		var schema map[string]any
		if err = json.Unmarshal(fields["output_schema"], &schema); err != nil {
			return nil, errors.New("agent context requires a graph planner output schema")
		}
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid planner schema")
		}
		tasks, ok := props["tasks"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid planner task schema")
		}
		items, ok := tasks["items"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid planner task schema")
		}
		taskProps, ok := items["properties"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid planner task properties")
		}
		taskProps["skills"] = map[string]any{"type": "array", "maxItems": 4, "items": map[string]any{"type": "string", "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$", "maxLength": 64}}
		required, ok := items["required"].([]any)
		if !ok {
			return nil, errors.New("invalid planner task required fields")
		}
		items["required"] = append(required, "skills")
		fields["output_schema"], err = json.Marshal(schema)
		if err != nil {
			return nil, err
		}
	}
	fields["instruction"], err = json.Marshal(instruction)
	if err != nil {
		return nil, err
	}
	return promptRecipeBytes(s.Creation.Execution, fields)
}

func plannerAgentContextInvocation(s Snapshot, base runtime.Invocation, err error) (runtime.Invocation, error) {
	if err != nil || s.Creation.AgentContext == nil {
		return base, err
	}
	var value map[string]json.RawMessage
	if err = json.Unmarshal([]byte(base.Input), &value); err != nil {
		return runtime.Invocation{}, err
	}
	raw, err := agentContextPromptBytes(s, "planner", nil, value)
	if err != nil {
		return runtime.Invocation{}, err
	}
	profile, err := modelProfileForInput(s.Creation.Config, "planner", string(raw), 0)
	if err != nil {
		return runtime.Invocation{}, err
	}
	return runtime.NewInvocationWithCodexAutoCompact(profile, string(raw), base.CodexAutoCompactOption())
}

func agentContextTaskForQuestion(s Snapshot, question string) *engineeringplan.Task {
	if s.Graph == nil {
		return nil
	}
	for _, task := range s.Graph.Graph.Tasks {
		if task.Kind != engineeringplan.Research && task.Kind != engineeringplan.Design {
			continue
		}
		expected, err := explorerQuestionForTask(s, task)
		if err == nil && expected == question {
			copy := task
			return &copy
		}
	}
	return nil
}

func agentContextWriterTask(s Snapshot, taskID string) *engineeringplan.Task {
	if s.Graph == nil {
		return nil
	}
	if taskID != "" {
		task, ok := s.Graph.Graph.Task(taskID)
		if ok {
			return &task
		}
		return nil
	}
	if task, ok := graphImplementationTask(s); ok {
		return &task
	}
	return nil
}

func validateAgentContextGraph(s Snapshot, g engineeringplan.Graph) error {
	for _, task := range g.Tasks {
		if s.Creation.AgentContext == nil {
			if len(task.Skills) > 0 {
				return errors.New("task skills require source-bound agent context")
			}
			continue
		}
		role := "explorer"
		switch task.Kind {
		case engineeringplan.Implementation:
			role = "writer"
		case engineeringplan.Review:
			role = "reviewer"
		case engineeringplan.Verification:
			if len(task.Skills) > 0 {
				return errors.New("native verification cannot select agent skills")
			}
		}
		scopes := task.ScopePaths
		if role == "writer" {
			scopes = task.WritePaths
		}
		_, err := s.Creation.AgentContext.Resolve(role, scopes, task.Skills)
		if err != nil {
			return err
		}
	}
	return nil
}
