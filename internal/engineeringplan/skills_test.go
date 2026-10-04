package engineeringplan

import "testing"

func TestAgentContextSkillReferencesAndStartedRevision(t *testing.T) {
	task := task("one", 3)
	task.Skills = []string{"tdd"}
	previous := graph(task)
	if err := previous.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, names := range [][]string{{"tdd", "tdd"}, {"../escape"}, {"one", "two", "three", "four", "five"}} {
		bad := graph(task)
		bad.Tasks[0].Skills = names
		if bad.Validate() == nil {
			t.Fatal("invalid task skills admitted", names)
		}
	}
	previous.Tasks[0].Attempts = []Attempt{{ID: "first", Outcome: AttemptFailed}}
	next := graph(previous.Tasks[0])
	next.Tasks[0].Skills = []string{"code-review"}
	if ValidateAutonomousRevision(previous, next) == nil {
		t.Fatal("started skill references changed")
	}
	previous.Tasks[0].Completed = true
	next = graph(previous.Tasks[0])
	next.Tasks[0].Skills = []string{"other"}
	if ValidateRevision(previous, next) == nil {
		t.Fatal("completed task skill references changed")
	}
}
