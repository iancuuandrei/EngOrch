package control

import (
	"errors"

	"harness.local/engorch/internal/access"
	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/modelpolicy"
	"harness.local/engorch/internal/runtime"
)

// modelProfileForInput selects only among profiles frozen in configuration.
// ContextBytes is the exact final UTF-8 input size; it is not token usage.
func modelProfileForInput(c config.Config, role, input string, failures int) (runtime.Profile, error) {
	// Keep the pre-policy builder path independent of full config validation.
	// Older planner callers intentionally construct only the configured role
	// profile, and nil policy must preserve their exact invocation identity.
	base, err := configuredRoleProfile(c, role)
	if err != nil {
		return runtime.Profile{}, err
	}
	if c.ModelPolicy == nil {
		return base, nil
	}
	if _, err := c.Route(role); err != nil {
		return runtime.Profile{}, err
	}
	if _, configured := c.ModelPolicy.Rules[role]; !configured {
		return base, nil
	}
	decision, err := modelpolicy.Select(*c.ModelPolicy, modelpolicy.Request{
		Role: role, ReadOnly: role == "explorer", Complexity: modelpolicy.LevelMedium, Risk: modelpolicy.LevelMedium,
		Uncertainty: modelpolicy.LevelMedium, ContextBytes: int64(len(input)), Failures: failures,
	})
	if err != nil {
		return runtime.Profile{}, err
	}
	selected := runtime.Profile{Runtime: decision.Profile.Runtime, Provider: decision.Profile.Provider, Model: decision.Profile.Model, Effort: decision.Profile.Effort, Role: role}
	if !c.AllowsProfile(selected) {
		return runtime.Profile{}, errors.New("model policy selected an unconfigured profile")
	}
	return selected, nil
}

func configuredRoleProfile(c config.Config, role string) (runtime.Profile, error) {
	var profile *runtime.Profile
	switch role {
	case "planner":
		profile = &c.Planner
	case "writer":
		profile = c.Writer
	case "fixer":
		profile = c.Fixer
	case "explorer":
		profile = c.Explorer
	case "reviewer":
		profile = c.Reviewer
	default:
		return runtime.Profile{}, errors.New("unknown routing role")
	}
	if profile == nil {
		return runtime.Profile{}, errors.New("role has no explicit runtime configuration")
	}
	if err := profile.Validate(); err != nil {
		return runtime.Profile{}, err
	}
	return *profile, nil
}

func modelProfileForSnapshot(s Snapshot, role, input string) (runtime.Profile, error) {
	return modelProfileForInput(s.Creation.Config, role, input, modelFailureCount(s, role, input))
}

func modelFailureCount(s Snapshot, role, input string) int {
	failures := 0
	if s.Graph != nil {
		var taskID string
		switch role {
		case "explorer":
			for _, task := range s.Graph.Graph.Tasks {
				if task.Kind != "research" && task.Kind != "design" {
					continue
				}
				question, err := explorerQuestionForTask(s, task)
				if err == nil && question == input {
					taskID = task.ID
					break
				}
			}
		case "writer", "fixer":
			if task, ok := graphImplementationTask(s); ok {
				taskID = task.ID
			}
		case "reviewer":
			for _, task := range s.Graph.Graph.Tasks {
				if task.Kind == "review" && !task.Completed {
					taskID = task.ID
					break
				}
			}
		}
		if taskID != "" {
			if task, ok := s.Graph.Graph.Task(taskID); ok {
				for _, attempt := range task.Attempts {
					if attempt.Outcome == "failed" {
						failures++
					}
				}
			}
		}
	}
	if role == "writer" || role == "fixer" {
		candidateID := ""
		if s.Candidate != nil {
			candidateID, _ = s.Candidate.ID()
		}
		if candidateID != "" && s.Verification != nil && !s.Verification.Pending && s.Verification.Plan.CandidateID == candidateID {
			for _, observation := range s.Verification.Observations {
				if observation.Result.Status == "FAIL" && observation.After.Candidate != nil && *observation.After.Candidate == *s.Candidate {
					failures++
				}
			}
		}
		if candidateID != "" && s.Review != nil {
			var verdict ReviewVerdict
			if err := canonical.Decode([]byte(s.Review.Result.Output), &verdict); err == nil && verdict.Decision == "changes_requested" && verdict.CandidateID == candidateID {
				failures++
			}
		}
	}
	if failures > modelpolicy.MaxFailures {
		return modelpolicy.MaxFailures
	}
	return failures
}

func modelChoiceForProfile(c config.Config, role string, profile runtime.Profile) (*access.ModelChoice, error) {
	base, err := c.Route(role)
	if err != nil {
		return nil, err
	}
	if profile == base {
		return nil, nil
	}
	if !c.AllowsProfile(profile) {
		return nil, errors.New("selected model profile is not allowed")
	}
	return &access.ModelChoice{Model: profile.Model, Effort: profile.Effort}, nil
}
