package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"harness.local/engorch/internal/control"
	"harness.local/engorch/internal/controllerstate"
	"harness.local/engorch/internal/repository"
)

// runEligibility reports whether a fully validated journal snapshot may be
// selected as the latest run. Validation (replay, filename and repository
// binding) always runs before eligibility so ineligible journals never mask
// a corrupt or foreign journal.
type runEligibility func(s control.Snapshot) bool

// latestRunCandidate is one validated journal eligible for latest selection.
type latestRunCandidate struct {
	id string
	at int64
}

// selectLatestValidatedRun validates every journal bound to this checkout and
// then selects the latest eligible entry by filesystem modification time.
// Tied durable timestamps are rejected rather than silently picking an
// arbitrary run. noneErr names the empty-eligible case; tieErr names the
// timestamp-tie case so interactive and autonomous callers keep distinct
// diagnostics while sharing validation and ordering.
func selectLatestValidatedRun(root string, eligible runEligibility, noneErr, tieErr string) (string, error) {
	cfg, err := configuration(root)
	if err != nil {
		return "", err
	}
	identity, err := repository.Discover(context.Background(), root, cfg.Repository)
	if err != nil {
		return "", err
	}
	paths, err := controllerstate.Resolve(cfg.ControllerStateRoot, identity)
	if err != nil {
		return "", err
	}
	if paths.External {
		if err := controllerstate.Validate(paths, identity); err != nil {
			return "", err
		}
	}
	entries, err := filepath.Glob(filepath.Join(paths.Runs, "*.jsonl"))
	if err != nil {
		return "", err
	}
	candidates := make([]latestRunCandidate, 0, len(entries))
	for _, path := range entries {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		s, err := control.Inspect(path)
		if err != nil {
			return "", err
		}
		if filepath.Base(path) != s.RunID+".jsonl" || filepath.Clean(s.Creation.Repository.Root) != filepath.Clean(root) {
			return "", errors.New("run journal repository binding mismatch")
		}
		if eligible != nil && !eligible(s) {
			continue
		}
		candidates = append(candidates, latestRunCandidate{id: s.RunID, at: info.ModTime().UnixNano()})
	}
	if len(candidates) == 0 {
		return "", errors.New(noneErr)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].at == candidates[j].at {
			return candidates[i].id < candidates[j].id
		}
		return candidates[i].at > candidates[j].at
	})
	if len(candidates) > 1 && candidates[0].at == candidates[1].at {
		return "", errors.New(tieErr)
	}
	return candidates[0].id, nil
}

// latestRunID selects only a journal which fully replays and is bound to this
// checkout. Tied durable timestamps are rejected rather than silently picking
// an arbitrary run.
func latestRunID(root string) (string, error) {
	return selectLatestValidatedRun(root, nil, "no durable runs found for this repository", "multiple durable runs have the same latest timestamp; specify RUN explicitly")
}

// latestAutonomousRunID selects the latest validated autonomous run bound to
// this checkout by explicit filesystem modification timestamp. Interactive
// runs never qualify; no new persistence machinery is introduced.
func latestAutonomousRunID(root string) (string, error) {
	return selectLatestValidatedRun(root, func(s control.Snapshot) bool {
		return s.Creation.Execution != nil && s.Creation.Execution.Mode == "autonomous-v1"
	}, "no durable autonomous runs found for this repository", "multiple durable autonomous runs have the same latest timestamp; specify RUN explicitly")
}

func inspectRunArgument(root string, args []string) (id string, export bool, err error) {
	switch len(args) {
	case 0:
		id, err = latestRunID(root)
		return id, false, err
	case 1:
		if args[0] == "--export-jsonl" {
			id, err = latestRunID(root)
			return id, true, err
		}
		return args[0], false, nil
	case 2:
		if args[1] != "--export-jsonl" {
			return "", false, errors.New("inspect accepts [RUN] [--export-jsonl]")
		}
		return args[0], true, nil
	default:
		return "", false, errors.New("inspect accepts [RUN] [--export-jsonl]")
	}
}
