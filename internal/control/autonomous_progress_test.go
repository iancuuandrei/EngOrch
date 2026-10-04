package control

import (
	"errors"
	"fmt"
	"testing"
)

func TestAutonomousTransitionBudgetIncludesSeparateContinuationAllowances(t *testing.T) {
	s := Snapshot{Creation: Creation{Execution: &ExecutionPolicy{MaxRepairs: 1}}}
	baseline := autonomousTransitionBudget(s)
	s.Creation.Execution.MaxScopeReplans = 2
	if got := autonomousTransitionBudget(s); got != baseline+12*4*2 {
		t.Fatal("replan allowance omitted from observation window", got)
	}
	withoutCorrection := autonomousTransitionBudget(s)
	s.Creation.Execution.SemanticCorrectionVersion = 1
	if autonomousTransitionBudget(s) <= withoutCorrection {
		t.Fatal("separate semantic allowance omitted from observation window")
	}
	if s.RepairAttempts != 0 || len(s.RoleCorrections) != 0 || len(s.ScopeReplans) != 0 {
		t.Fatal("budget calculation consumed continuation authority")
	}
}

func TestAutonomousProgressContinuesBeyondLegacyStepCaps(t *testing.T) {
	p := autonomousProgress{}
	s := Snapshot{State: "IMPLEMENTING"}
	for i := 1; i <= 1024; i++ {
		s.ControllerSequence = i
		s.ControllerHead = fmt.Sprintf("%064x", i)
		s.PlanID = fmt.Sprintf("%064x", i)
		if err := p.observe(s); err != nil {
			t.Fatalf("valid advancing progress stopped at %d: %v", i, err)
		}
	}
}

func TestAutonomousProgressStopsStallAndHeadSubstitution(t *testing.T) {
	s := Snapshot{State: "IMPLEMENTING", ControllerSequence: 1, ControllerHead: fmt.Sprintf("%064x", 1)}
	p := autonomousProgress{}
	if err := p.observe(s); err != nil {
		t.Fatal(err)
	}
	if err := p.observe(s); !errors.Is(err, ErrAutonomousNoProgress) {
		t.Fatal("unchanged head did not retain an attention outcome", err)
	}
	s.ControllerHead = fmt.Sprintf("%064x", 2)
	if err := p.observe(s); !errors.Is(err, ErrAutonomousUnsafe) {
		t.Fatal("same sequence with substituted head was not denied", err)
	}
}

func TestAutonomousTransitionWindowRequiresEngineeringProgress(t *testing.T) {
	p := autonomousProgress{}
	s := Snapshot{State: "IMPLEMENTING"}
	budget := autonomousTransitionBudget(s)
	for i := 1; i <= budget; i++ {
		s.ControllerSequence = i
		s.ControllerHead = fmt.Sprintf("%064x", i)
		err := p.observe(s)
		if i < budget && err != nil || i == budget && !errors.Is(err, ErrAutonomousNoProgress) {
			t.Fatalf("metadata-only transitions escaped the admitted progress window: %d %v", i, err)
		}
	}
}
