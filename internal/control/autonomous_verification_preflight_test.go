package control

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAutonomousMissingVerificationExecutableStopsBeforeRoleEffects(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		c := autonomousCreation(t, 2)
		c.Config.Verification[0].Name = "required-go-check"
		c.Config.Verification[0].Argv = []string{"fabric-test-missing-go-executable-48219"}
		path, s := autonomousAwaitingApproval(t, c)
		machineAuthorizePlan(t, path, s)

		result, err := RunAutonomous(context.Background(), path)
		if err == nil || !errors.Is(err, ErrAutonomousVerificationNotRun) || !strings.Contains(err.Error(), `check "required-go-check"`) || !strings.Contains(err.Error(), "unavailable before implementation") {
			t.Fatalf("missing required executable did not produce a bounded preflight error: %v", err)
		}
		latest, inspectErr := Inspect(path)
		if inspectErr != nil {
			t.Fatal(inspectErr)
		}
		if result.State != "IMPLEMENTING" || latest.WorkspaceOutcome != "CONFIRMED" || latest.ExplorerHost != nil || latest.WriterHost != nil || latest.FileIntent != nil || latest.RepairAttempts != 0 {
			t.Fatalf("missing executable reached role or file work: result=%+v snapshot=%+v", result, latest)
		}
		for _, kind := range []string{"explorer.host-intent", "writer.host-intent", "files.intent", "autonomous.repair-started"} {
			if countJournalKind(t, path, kind) != 0 {
				t.Fatalf("missing executable produced %s", kind)
			}
		}
	})

	t.Run("graph", func(t *testing.T) {
		c := graphCreation(t, 1)
		c.Config.Verification[0].Name = "required-go-check"
		c.Config.Verification[0].Argv = []string{"fabric-test-missing-go-executable-48219"}
		path, s := graphAwaitingApproval(t, c)
		machineAuthorizePlan(t, path, s)

		result, err := RunAutonomous(context.Background(), path)
		if err == nil || !errors.Is(err, ErrAutonomousVerificationNotRun) || !strings.Contains(err.Error(), `check "required-go-check"`) || !strings.Contains(err.Error(), "unavailable before implementation") {
			t.Fatalf("missing required executable did not produce a bounded preflight error: %v", err)
		}
		latest, inspectErr := Inspect(path)
		if inspectErr != nil {
			t.Fatal(inspectErr)
		}
		if result.State != "IMPLEMENTING" || latest.WorkspaceOutcome != "CONFIRMED" || len(latest.ExplorerRuns) != 0 || len(latest.GraphWriterHosts) != 0 || latest.WriterHost != nil || latest.FileIntent != nil || latest.RepairAttempts != 0 {
			t.Fatalf("missing executable reached graph role or file work: result=%+v snapshot=%+v", result, latest)
		}
		for _, kind := range []string{"explorer.host-intent", "graph.writer.host-intent", "writer.host-intent", "files.intent", "autonomous.repair-started"} {
			if countJournalKind(t, path, kind) != 0 {
				t.Fatalf("missing executable produced %s", kind)
			}
		}
	})
}

func TestUnstartedVerificationDoesNotConsumeRepairButStartedFailureDoes(t *testing.T) {
	t.Run("not-run", func(t *testing.T) {
		c := autonomousCreation(t, 2)
		c.Config.Verification[0].Name = "required-go-check"
		c.Config.Verification[0].Argv = []string{"fabric-test-missing-go-executable-48219"}
		path, s := autonomousAwaitingApproval(t, c)
		machineAuthorizePlan(t, path, s)
		applyWriterOutput(t, path, "writer output\n")
		failed, err := Verify(context.Background(), path)
		if err != nil || failed.State != "REPAIRING" || failed.Verification == nil || len(failed.Verification.Observations) != 1 {
			t.Fatalf("fixture did not persist the unstarted native result: %+v %v", failed, err)
		}
		observation := failed.Verification.Observations[0].Result
		if observation.Started || observation.Status != "NOT_RUN" {
			t.Fatalf("fixture result is not unstarted NOT_RUN: %+v", observation)
		}
		beforeRepair := countJournalKind(t, path, "autonomous.repair-started")
		beforeWriter := countJournalKind(t, path, "writer.host-intent")
		_, err = RunAutonomous(context.Background(), path)
		if err == nil || !errors.Is(err, ErrAutonomousVerificationNotRun) || !strings.Contains(err.Error(), "NOT_RUN") || !strings.Contains(err.Error(), `check "required-go-check"`) {
			t.Fatalf("unstarted check was treated as semantic failure: %v", err)
		}
		if countJournalKind(t, path, "autonomous.repair-started") != beforeRepair || countJournalKind(t, path, "writer.host-intent") != beforeWriter {
			t.Fatal("unstarted check consumed a repair slot or dispatched another writer")
		}
	})

	t.Run("started-failure", func(t *testing.T) {
		path, failed := repairingAutonomousFixture(t, 1)
		if failed.Verification == nil || len(failed.Verification.Observations) != 1 {
			t.Fatal("failure fixture lacks a native observation")
		}
		observation := failed.Verification.Observations[0].Result
		if !observation.Started || observation.Status != "FAIL" {
			t.Fatalf("fixture result is not an actual native failure: %+v", observation)
		}
		if err := rejectUnstartedAutonomousVerification(failed); err != nil {
			t.Fatalf("started native failure was blocked as NOT_RUN: %v", err)
		}
		_, err := RunAutonomous(context.Background(), path)
		if countJournalKind(t, path, "autonomous.repair-started") != 1 {
			t.Fatalf("actual native failure did not remain eligible for repair (run error %v)", err)
		}
	})
}
