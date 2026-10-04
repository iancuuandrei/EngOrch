package cli

import (
	"encoding/json"
	"testing"
)

func TestWorkingContextCLIOptInInspection(t *testing.T) {
	fixture := newFixerInitFixture(t)
	root := fixture.root
	mustExecuteCLI(t, root, fixture.args("--fixer-model", "base-model", "--access-config", fixture.accessConfig(t))...)
	for _, enabled := range []bool{false, true} {
		args := []string{"run", "--autonomous", "--inspect-plan"}
		if enabled {
			args = append(args, "--working-context")
		}
		var result struct {
			Plan struct {
				Context struct {
					Enabled   bool   `json:"enabled"`
					Authority string `json:"authority"`
					MaxBytes  int    `json:"max_content_bytes"`
				} `json:"working_context"`
			} `json:"execution_plan"`
		}
		if err := json.Unmarshal(mustExecuteCLI(t, root, args...), &result); err != nil {
			t.Fatal(err)
		}
		if result.Plan.Context.Enabled != enabled || result.Plan.Context.Authority != "NONE" || result.Plan.Context.MaxBytes != 16384 {
			t.Fatal("inspection disagrees with opt-in policy", result)
		}
	}
}
