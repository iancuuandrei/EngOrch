package control

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/fileeffects"
	"harness.local/engorch/internal/journal"
	"harness.local/engorch/internal/safepath"
	"harness.local/engorch/internal/taskscheduler"
	"harness.local/engorch/internal/worktree"
)

func decodeForkContent(change fileeffects.Change) ([]byte, error) {
	if change.ContentBase64 == nil {
		return nil, nil
	}
	raw := *change.ContentBase64
	if len(raw) > 350000 {
		return nil, errors.New("staged fork content bound exceeded")
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(raw)
	if err != nil || base64.StdEncoding.EncodeToString(decoded) != raw {
		return nil, errors.New("staged fork content is not canonical base64")
	}
	return decoded, nil
}

// StagedCohortArchive retains one confirmed staged cohort's immutable evidence.
// Original identities are retained; journal history is never rewritten and
// completed task history, repair budgets and evidence are preserved.
type StagedCohortArchive struct {
	CohortIndex     int                    `json:"cohort_index"`
	PreparationID   string                 `json:"preparation_id"`
	BaseCandidateID string                 `json:"base_candidate_id"`
	CandidateID     string                 `json:"candidate_id"`
	TaskIDs         []string               `json:"task_ids"`
	Batch           GraphWriterBatchRecord `json:"batch"`
	FileIntent      FileIntent             `json:"file_intent"`
	FileReceipt     FileReceipt            `json:"file_receipt"`
}

// StagedCohortAdvance is the compact durable stage-advance event. It carries
// only exact identities and content hashes; the full structured archive is
// reconstructed during replay from already replay-validated current state
// (confirmed batch, file intent, receipt, evidence). This keeps the journal
// envelope far below canonical.MaxBytes even when a near-bound stage would
// double Before/BeforeFiles/Changes/After in a full archive payload.
// Legacy full-archive payloads remain replayable; new advances append this
// compact form. Journal history is never rewritten.
type StagedCohortAdvance struct {
	Version         int      `json:"version"`
	CohortIndex     int      `json:"cohort_index"`
	PreparationID   string   `json:"preparation_id"`
	BaseCandidateID string   `json:"base_candidate_id"`
	CandidateID     string   `json:"candidate_id"`
	TaskIDs         []string `json:"task_ids"`
	BatchProposalID string   `json:"batch_proposal_id"`
	FileProposalID  string   `json:"file_proposal_id"`
	ObservationHash string   `json:"observation_hash"`
}

// stagedAdvanceFromArchive derives the compact advance identities from a
// validated full archive. It performs no filesystem access.
func stagedAdvanceFromArchive(archive StagedCohortArchive) (StagedCohortAdvance, error) {
	if archive.PreparationID == "" || archive.BaseCandidateID == "" || archive.CandidateID == "" || len(archive.TaskIDs) == 0 {
		return StagedCohortAdvance{}, errors.New("staged advance identity is incomplete")
	}
	batchID, err := archive.Batch.Prepared.Proposal.ID()
	if err != nil {
		return StagedCohortAdvance{}, err
	}
	fileID, err := archive.FileIntent.Prepared.Proposal.ID()
	if err != nil {
		return StagedCohortAdvance{}, err
	}
	if archive.FileReceipt.Receipt.ObservationHash == "" {
		return StagedCohortAdvance{}, errors.New("staged advance observation hash is missing")
	}
	tasks := append([]string(nil), archive.TaskIDs...)
	sort.Strings(tasks)
	return StagedCohortAdvance{Version: 1, CohortIndex: archive.CohortIndex, PreparationID: archive.PreparationID, BaseCandidateID: archive.BaseCandidateID, CandidateID: archive.CandidateID, TaskIDs: tasks, BatchProposalID: batchID, FileProposalID: fileID, ObservationHash: archive.FileReceipt.Receipt.ObservationHash}, nil
}

// expectedStagedAdvance derives the compact advance for the current confirmed
// cohort without persisting the duplicated full payload.
func expectedStagedAdvance(s Snapshot) (StagedCohortAdvance, error) {
	archive, err := expectedStagedArchive(s)
	if err != nil {
		return StagedCohortAdvance{}, err
	}
	return stagedAdvanceFromArchive(archive)
}

// StagedForkIntent binds one staged leaf task to its exact parent candidate
// and the parent-authorized delta to copy. It is persisted BEFORE destination
// creation or overlay; materialization reuses existing worktree/fileeffects.
type StagedForkIntent struct {
	Version         int                  `json:"version"`
	CohortIndex     int                  `json:"cohort_index"`
	TaskID          string               `json:"task_id"`
	BaseCandidateID string               `json:"base_candidate_id"`
	ParentBinding   worktree.Binding     `json:"parent_binding"`
	ParentCandidate worktree.Candidate   `json:"parent_candidate"`
	SourceCommit    string               `json:"source_commit"`
	Delta           []fileeffects.Change `json:"delta"`
	ScopePaths      []string             `json:"scope_paths"`
	WritePaths      []string             `json:"write_paths"`
	Request         worktree.Request     `json:"request"`
}

// StagedForkReceipt records the exact confirmed forked child and its candidate.
type StagedForkReceipt struct {
	Intent    StagedForkIntent   `json:"intent"`
	Binding   worktree.Binding   `json:"binding"`
	Candidate worktree.Candidate `json:"candidate"`
}

// StagedForkState is the replayed durable fork state for one staged task.
type StagedForkState struct {
	Intent    StagedForkIntent    `json:"intent"`
	Binding   *worktree.Binding   `json:"binding,omitempty"`
	Candidate *worktree.Candidate `json:"candidate,omitempty"`
	Outcome   string              `json:"outcome"`
}

func stagedCohortIndex(s Snapshot) int {
	return len(s.GraphStagedCohorts)
}

// expectedStagedPreparation freezes the exact CURRENT ready implementation
// subset with its exact graph/candidate/head/resources. It uses the existing
// exact-ready SelectResourceCohort scope with bounded W1 waves when more
// current ready tasks exceed capacity. No new solver, generic memory or
// recovery framework is introduced.
func expectedStagedPreparation(s Snapshot) (GraphIsolationPreparation, error) {
	if !stagedIsolationEnabled(s) {
		return GraphIsolationPreparation{}, errors.New("staged isolation not enabled")
	}
	if s.State != "IMPLEMENTING" || s.Graph == nil || s.Workspace == nil || s.Candidate == nil || s.FileIntent != nil {
		return GraphIsolationPreparation{}, errors.New("pristine staged cohort candidate required")
	}
	if graphHasUnknown(s) || s.FileOutcome == "UNKNOWN" {
		return GraphIsolationPreparation{}, errors.New("staged cohort blocked on UNKNOWN; explicit reconciliation required")
	}
	for _, isolation := range s.GraphIsolations {
		if isolation.Outcome == "UNKNOWN" {
			return GraphIsolationPreparation{}, errors.New("staged cohort blocked on UNKNOWN child isolation")
		}
	}
	for _, fork := range s.GraphStagedForks {
		if fork.Outcome == "UNKNOWN" {
			return GraphIsolationPreparation{}, errors.New("staged cohort blocked on UNKNOWN fork")
		}
	}
	if s.GraphIsolationPreparation != nil {
		return GraphIsolationPreparation{}, errors.New("staged cohort already prepared")
	}
	cohortIndex := stagedCohortIndex(s)
	// Past archives must be contiguous and their candidates must chain to the
	// current parent candidate. Completed evidence is preserved, never reset.
	if cohortIndex > 0 {
		if len(s.GraphStagedCohorts) != cohortIndex {
			return GraphIsolationPreparation{}, errors.New("staged cohort history is not contiguous")
		}
		last := s.GraphStagedCohorts[cohortIndex-1]
		currentID, err := s.Candidate.ID()
		if err != nil || currentID != last.CandidateID {
			return GraphIsolationPreparation{}, errors.New("staged parent candidate differs from archived cohort candidate")
		}
		// Every archived member must retain completed graph evidence with its
		// archived candidate and invocation; no synthetic settlement is allowed.
		for _, taskID := range last.TaskIDs {
			evidence, ok := s.Graph.Evidence[taskID]
			if !ok || evidence.Outcome != "completed" || evidence.CandidateID != last.CandidateID {
				return GraphIsolationPreparation{}, errors.New("staged prior cohort lacks completed evidence")
			}
			found := false
			for _, member := range last.Batch.Members {
				if member.TaskID == taskID && member.InvocationID == evidence.WriterInvocationID {
					found = true
					break
				}
			}
			if !found {
				return GraphIsolationPreparation{}, errors.New("staged prior cohort member evidence mismatch")
			}
		}
	} else {
		if len(s.GraphStagedCohorts) != 0 {
			return GraphIsolationPreparation{}, errors.New("initial staged cohort requires empty history")
		}
	}
	ready, err := graphReadyTasks(s)
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	var implementations []engineeringplan.Task
	for _, task := range ready {
		if task.Kind == engineeringplan.Implementation {
			implementations = append(implementations, task)
		}
	}
	if len(implementations) == 0 || len(implementations) > 8 {
		return GraphIsolationPreparation{}, errors.New("staged cohort requires one to eight ready implementations")
	}
	// Staged cohorts admit only currently ready tasks; dependencies grant
	// readiness only when observed completed (Ready already enforces this).
	// C3 and lower coupling is advisory and never grants ownership here.
	if s.Creation.Execution.IsolationCapacity == nil || s.Creation.Execution.IsolationEstimate == nil {
		return GraphIsolationPreparation{}, errors.New("isolation capacity and template required")
	}
	template := *s.Creation.Execution.IsolationEstimate
	profileID, err := canonical.Hash("harness.isolation-writer-profile.v1", *s.Creation.Config.Writer)
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	route := engineeringplan.RuntimeResourceKey{ProfileID: profileID, Provider: s.Creation.Config.Writer.Provider, Model: s.Creation.Config.Writer.Model}
	demands := make([]engineeringplan.TaskResourceDemand, len(implementations))
	for i, task := range implementations {
		demands[i] = engineeringplan.TaskResourceDemand{TaskID: task.ID, CPUMilli: template.CPUMilli, MemoryMiB: template.MemoryMiB, VerificationSlots: template.VerificationSlots, Runtime: route, RuntimeSlots: template.RuntimeSlots}
	}
	// Exact-ready scope with bounded W1 waves on remainder.
	waves, err := engineeringplan.SelectResourceWaves(s.Graph.Graph, implementations, demands, *s.Creation.Execution.IsolationCapacity, s.Creation.Execution.EffectiveMaxParallel())
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	if len(waves) == 0 {
		return GraphIsolationPreparation{}, errors.New("staged waves selected no implementation tasks")
	}
	baseID, err := s.Candidate.ID()
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	ids := make([]string, 0, len(implementations))
	for _, task := range implementations {
		ids = append(ids, task.ID)
	}
	sort.Strings(ids)
	waveIDs := make([][]string, 0, len(waves))
	waveEstimated := make([]engineeringplan.ResourceTotals, 0, len(waves))
	waveBlocked := make([][]engineeringplan.ResourceBlockReason, 0, len(waves))
	seen := map[string]bool{}
	for _, wave := range waves {
		if len(wave.Tasks) == 0 {
			return GraphIsolationPreparation{}, errors.New("staged waves contain an empty wave")
		}
		waveIDList := make([]string, 0, len(wave.Tasks))
		for _, task := range wave.Tasks {
			if seen[task.ID] {
				return GraphIsolationPreparation{}, errors.New("staged waves duplicated a task")
			}
			seen[task.ID] = true
			waveIDList = append(waveIDList, task.ID)
		}
		waveIDs = append(waveIDs, waveIDList)
		waveEstimated = append(waveEstimated, wave.Estimated)
		waveBlocked = append(waveBlocked, append([]engineeringplan.ResourceBlockReason{}, wave.Blocked...))
	}
	if len(seen) != len(implementations) {
		return GraphIsolationPreparation{}, errors.New("staged waves do not cover all ready implementations")
	}
	total := engineeringplan.SumResourceDemands(demands)
	prep := GraphIsolationPreparation{Version: 3, PlanID: s.Graph.PlanID, GraphDigest: s.Graph.Digest, Revision: s.Graph.Revision, BaseCandidateID: baseID, Capacity: *s.Creation.Execution.IsolationCapacity, Demands: demands, SelectedTaskIDs: ids, Estimated: total, Waves: waveIDs, WaveEstimated: waveEstimated, WaveBlocked: waveBlocked, CohortIndex: cohortIndex}
	prep.PreparationID, err = canonical.Hash("harness.graph-staged-preparation.v1", prep)
	return prep, err
}

func admitStagedPreparation(ctx context.Context, s Snapshot) (GraphIsolationPreparation, error) {
	if s.Workspace == nil || s.Candidate == nil {
		return GraphIsolationPreparation{}, errors.New("parent workspace required")
	}
	lease, err := worktree.AcquireRead(s.Workspace.Request)
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	defer lease.Close()
	current, _, err := worktree.Capture(ctx, *s.Workspace)
	if err != nil || current != *s.Candidate {
		return GraphIsolationPreparation{}, errors.New("base candidate changed")
	}
	return expectedStagedPreparation(s)
}

// PrepareStagedCohort freezes the current ready staged subset before any child
// fork effect is admitted for that stage.
func PrepareStagedCohort(ctx context.Context, path string) (GraphIsolationPreparation, error) {
	s, err := Inspect(path)
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	if !stagedIsolationEnabled(s) || s.GraphIsolationPreparation != nil {
		return GraphIsolationPreparation{}, errors.New("staged cohort already prepared or disabled")
	}
	prep, err := admitStagedPreparation(ctx, s)
	if err != nil {
		return GraphIsolationPreparation{}, err
	}
	if err := Append(path, "graph.staged-prepared", prep); err != nil {
		return GraphIsolationPreparation{}, err
	}
	return prep, nil
}

// stagedParentDelta returns the exact parent-authorized delta to copy into a
// forked child: the union of all prior staged batch changes. Regular
// bytes/mode only; RI/Go overlay and lexical views never authorize a copy.
func stagedParentDelta(s Snapshot) ([]fileeffects.Change, error) {
	delta := []fileeffects.Change{}
	seen := map[string]bool{}
	for _, archive := range s.GraphStagedCohorts {
		for _, change := range archive.Batch.Prepared.Proposal.Changes {
			if seen[change.Path] {
				return nil, errors.New("staged parent delta reuses a path without ownership versioning")
			}
			seen[change.Path] = true
			delta = append(delta, change)
		}
	}
	sort.Slice(delta, func(i, j int) bool { return delta[i].Path < delta[j].Path })
	return delta, nil
}

func expectedStagedForkIntent(s Snapshot, taskID string) (StagedForkIntent, engineeringplan.Task, error) {
	if !stagedIsolationEnabled(s) || s.State != "IMPLEMENTING" || s.Graph == nil || s.Workspace == nil || s.Candidate == nil || s.FileIntent != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("pristine staged fork candidate required")
	}
	if s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("frozen staged cohort required")
	}
	prep := s.GraphIsolationPreparation
	if prep.CohortIndex != stagedCohortIndex(s) {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged cohort index mismatch")
	}
	ready, err := graphReadyTasks(s)
	if err != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, err
	}
	var task engineeringplan.Task
	found := false
	for _, candidate := range ready {
		if candidate.ID == taskID {
			task, found = candidate, true
			break
		}
	}
	if !found || task.Kind != engineeringplan.Implementation {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("ready staged implementation task required")
	}
	if !containsGraphIsolationTask(prep.SelectedTaskIDs, taskID) {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("task is not in frozen staged cohort")
	}
	baseID, err := s.Candidate.ID()
	if err != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, err
	}
	if prep.BaseCandidateID != baseID {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged parent candidate changed")
	}
	if s.Candidate.Head != s.Creation.Repository.Commit {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged source commit mismatch")
	}
	delta, err := stagedParentDelta(s)
	if err != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, err
	}
	// Bound the fork delta before any effect: paths, files, bytes, encoding.
	if len(delta) > 64 {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged fork delta exceeds file bound")
	}
	totalBytes := 0
	seenPaths := map[string]bool{}
	for _, change := range delta {
		if err := safepath.Writable(change.Path); err != nil {
			return StagedForkIntent{}, engineeringplan.Task{}, err
		}
		if change.Path == "" || seenPaths[change.Path] {
			return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged fork delta path is foreign or duplicated")
		}
		seenPaths[change.Path] = true
		if change.ContentBase64 != nil {
			content, err := decodeForkContent(change)
			if err != nil {
				return StagedForkIntent{}, engineeringplan.Task{}, err
			}
			totalBytes += len(content)
			if totalBytes > 256<<10 {
				return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged fork delta exceeds byte bound")
			}
		}
		if change.Executable {
			return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged fork executable changes unsupported on Windows")
		}
	}
	runID, err := canonical.Hash("harness.graph-staged-fork.v1", struct {
		ParentRunID, Digest, TaskID, BaseCandidateID string
		Revision, CohortIndex                        int
		PreparationID                                string
	}{s.RunID, s.Graph.Digest, task.ID, baseID, s.Graph.Revision, prep.CohortIndex, prep.PreparationID})
	if err != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, err
	}
	req, err := worktree.Prepare(runID, s.Creation.Repository)
	if err != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, err
	}
	req.CandidateIdentity = s.Creation.Config.CandidateIdentity
	stateRoot, err := controllerNamespace(s.Creation)
	if err != nil || stateRoot == "" {
		return StagedForkIntent{}, engineeringplan.Task{}, errors.New("staged fork requires external controller state root")
	}
	req.ControllerStateRoot = filepath.Join(stateRoot, "graph-stages", runID)
	if err := req.Validate(); err != nil {
		return StagedForkIntent{}, engineeringplan.Task{}, err
	}
	intent := StagedForkIntent{Version: 1, CohortIndex: prep.CohortIndex, TaskID: task.ID, BaseCandidateID: baseID, ParentBinding: *s.Workspace, ParentCandidate: *s.Candidate, SourceCommit: s.Creation.Repository.Commit, Delta: delta, ScopePaths: append([]string(nil), task.ScopePaths...), WritePaths: append([]string(nil), task.WritePaths...), Request: req}
	return intent, task, nil
}

func replayStagedPreparation(s *Snapshot, e journal.Event) error {
	if !stagedIsolationEnabled(*s) {
		return errors.New("staged isolation not enabled")
	}
	var prep GraphIsolationPreparation
	if err := canonical.Decode(e.Payload, &prep); err != nil {
		return err
	}
	if prep.Version != 3 {
		return errors.New("invalid staged preparation version")
	}
	if s.GraphIsolationPreparation != nil {
		return errors.New("staged preparation already recorded")
	}
	if len(prep.Waves) == 0 || len(prep.WaveEstimated) != len(prep.Waves) || len(prep.WaveBlocked) != len(prep.Waves) {
		return errors.New("staged waves are incomplete")
	}
	if prep.CohortIndex != len(s.GraphStagedCohorts) {
		return errors.New("staged cohort index mismatch")
	}
	expected, err := expectedStagedPreparation(*s)
	if err != nil || !sameCanonical(expected, prep) {
		return errors.New("staged preparation substitution")
	}
	s.GraphIsolationPreparation = &prep
	return nil
}

func replayStagedFork(s *Snapshot, e journal.Event) error {
	if !stagedIsolationEnabled(*s) {
		return errors.New("staged isolation not enabled")
	}
	if s.GraphStagedForks == nil {
		s.GraphStagedForks = map[string]StagedForkState{}
	}
	switch e.Kind {
	case "graph.staged-fork-intent":
		var intent StagedForkIntent
		if err := canonical.Decode(e.Payload, &intent); err != nil {
			return err
		}
		if _, exists := s.GraphStagedForks[intent.TaskID]; exists {
			return errors.New("staged fork already admitted")
		}
		expected, _, err := expectedStagedForkIntent(*s, intent.TaskID)
		if err != nil || !sameCanonical(expected, intent) {
			return errors.New("staged fork intent substitution")
		}
		s.GraphStagedForks[intent.TaskID] = StagedForkState{Intent: intent, Outcome: "UNKNOWN"}
	case "graph.staged-fork-confirmed":
		var receipt StagedForkReceipt
		if err := canonical.Decode(e.Payload, &receipt); err != nil {
			return err
		}
		state, ok := s.GraphStagedForks[receipt.Intent.TaskID]
		if !ok || state.Outcome != "UNKNOWN" || !sameCanonical(state.Intent, receipt.Intent) || receipt.Binding.Request != state.Intent.Request {
			return errors.New("staged fork receipt substitution")
		}
		if _, err := receipt.Binding.ID(); err != nil {
			return err
		}
		if err := receipt.Candidate.ValidateBinding(receipt.Binding); err != nil {
			return err
		}
		if s.Candidate == nil || s.GraphIsolationPreparation == nil {
			return errors.New("staged fork base candidate missing")
		}
		baseID, err := s.Candidate.ID()
		if err != nil || receipt.Intent.BaseCandidateID != baseID || s.GraphIsolationPreparation.BaseCandidateID != baseID {
			return errors.New("staged fork base candidate substitution")
		}
		parent := *s.Candidate
		child := receipt.Candidate
		childID, err := child.ID()
		if err != nil {
			return err
		}
		_ = childID
		workspaceID, err := receipt.Binding.ID()
		if err != nil {
			return err
		}
		if child.WorktreeID != workspaceID {
			return errors.New("staged fork worktree binding mismatch")
		}
		if child.Version != parent.Version || child.Head != parent.Head || child.IndexHash != parent.IndexHash || child.FilesHash != parent.FilesHash || child.FileCount != parent.FileCount {
			return errors.New("staged fork child is not the exact parent candidate")
		}
		if child.Head != s.Creation.Repository.Commit {
			return errors.New("staged fork source commit mismatch")
		}
		state.Binding, state.Candidate, state.Outcome = &receipt.Binding, &receipt.Candidate, "CONFIRMED"
		s.GraphStagedForks[receipt.Intent.TaskID] = state
	default:
		return errors.New("unknown staged fork event")
	}
	return nil
}

func replayStagedAdvance(s *Snapshot, e journal.Event) error {
	if !stagedIsolationEnabled(*s) {
		return errors.New("staged isolation not enabled")
	}
	// Compact advances carry only identities; legacy full archives remain
	// replayable. Detection uses the durable payload shape, not heuristics
	// about current state.
	if !bytes.Contains(e.Payload, []byte(`"batch"`)) {
		var advance StagedCohortAdvance
		if err := canonical.Decode(e.Payload, &advance); err != nil {
			return err
		}
		if advance.Version != 1 {
			return errors.New("invalid staged advance version")
		}
		if advance.CohortIndex != len(s.GraphStagedCohorts) {
			return errors.New("staged advance cohort index mismatch")
		}
		expectedArchive, err := expectedStagedArchive(*s)
		if err != nil {
			return err
		}
		expected, err := stagedAdvanceFromArchive(expectedArchive)
		if err != nil || !sameCanonical(expected, advance) {
			return errors.New("staged advance substitution")
		}
		// Reconstruct the full structured archive ONLY from already
		// replay-validated current state; retain prior evidence/history.
		s.GraphStagedCohorts = append(s.GraphStagedCohorts, expectedArchive)
		// Advance current slots only via this exact finite completion record
		// that retains immutable past evidence. Journal past is never
		// rewritten. Memory admission decision/pressure history is retained
		// across cohorts; active reservations block advance elsewhere.
		s.GraphIsolationPreparation = nil
		s.GraphWriterBatch = nil
		s.FileIntent = nil
		s.FileReceipt = nil
		s.FileOutcome = ""
		return nil
	}
	var archive StagedCohortArchive
	if err := canonical.Decode(e.Payload, &archive); err != nil {
		return err
	}
	if archive.CohortIndex != len(s.GraphStagedCohorts) {
		return errors.New("staged advance cohort index mismatch")
	}
	expected, err := expectedStagedArchive(*s)
	if err != nil || !sameCanonical(expected, archive) {
		return errors.New("staged advance substitution")
	}
	s.GraphStagedCohorts = append(s.GraphStagedCohorts, archive)
	// Advance current slots only via this exact finite completion record that
	// retains immutable past evidence. Journal past is never rewritten.
	// Legacy path retains prior nil-reset behavior for byte-identical replay
	// of pre-compact journals; new compact advances retain memory history.
	s.GraphIsolationPreparation = nil
	s.GraphWriterBatch = nil
	s.FileIntent = nil
	s.FileReceipt = nil
	s.FileOutcome = ""
	s.GraphMemoryAdmission = nil
	return nil
}

// expectedStagedArchive derives the exact archive for the current confirmed
// cohort. New FileIntent is allowed only after prior CONFIRMED with all member
// progress and no UNKNOWN effects; no synthetic settlement is permitted.
// Advance rejects any active memory reservation so unsettled grants are never
// dropped; adaptive decision/pressure history is retained across cohorts by the
// compact advance path.
func expectedStagedArchive(s Snapshot) (StagedCohortArchive, error) {
	if !stagedIsolationEnabled(s) || s.Graph == nil || s.Candidate == nil || s.Workspace == nil {
		return StagedCohortArchive{}, errors.New("staged advance requires graph and candidate")
	}
	prep := s.GraphIsolationPreparation
	if prep == nil || prep.Version != 3 || prep.CohortIndex != len(s.GraphStagedCohorts) {
		return StagedCohortArchive{}, errors.New("staged advance requires the current frozen cohort")
	}
	if s.GraphMemoryAdmission != nil && len(s.GraphMemoryAdmission.ActiveTaskIDs) != 0 {
		return StagedCohortArchive{}, errors.New("staged advance blocked on active memory reservation")
	}
	if s.FileIntent == nil || s.FileOutcome != "CONFIRMED" || s.FileReceipt == nil {
		return StagedCohortArchive{}, errors.New("staged advance requires a CONFIRMED parent effect")
	}
	if s.GraphWriterBatch == nil {
		return StagedCohortArchive{}, errors.New("staged advance requires the current aggregate")
	}
	if graphHasUnknown(s) {
		return StagedCohortArchive{}, errors.New("staged advance blocked on UNKNOWN graph attempt")
	}
	for _, isolation := range s.GraphIsolations {
		if isolation.Outcome == "UNKNOWN" {
			return StagedCohortArchive{}, errors.New("staged advance blocked on UNKNOWN isolation")
		}
	}
	for _, fork := range s.GraphStagedForks {
		if fork.Outcome == "UNKNOWN" {
			return StagedCohortArchive{}, errors.New("staged advance blocked on UNKNOWN fork")
		}
	}
	if unresolved := lifecycleUnresolved(s); len(unresolved) != 0 {
		return StagedCohortArchive{}, errors.New("staged advance blocked on active invocation")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return StagedCohortArchive{}, err
	}
	afterID, err := s.GraphWriterBatch.Prepared.Proposal.After.ID()
	if err != nil || afterID != candidateID {
		return StagedCohortArchive{}, errors.New("staged aggregate candidate differs")
	}
	beforeID, err := s.GraphWriterBatch.Prepared.Proposal.Before.ID()
	if err != nil || beforeID != prep.BaseCandidateID {
		return StagedCohortArchive{}, errors.New("staged aggregate parent candidate changed")
	}
	// All members must carry completed graph progress with this candidate and
	// invocation; intermediate hubs may integrate without final READY.
	ids := append([]string(nil), prep.SelectedTaskIDs...)
	sort.Strings(ids)
	for _, taskID := range ids {
		evidence, ok := s.Graph.Evidence[taskID]
		if !ok || evidence.Outcome != "completed" || evidence.CandidateID != candidateID {
			return StagedCohortArchive{}, errors.New("staged advance requires completed member progress")
		}
		memberFound := false
		for _, member := range s.GraphWriterBatch.Members {
			if member.TaskID == taskID && member.InvocationID == evidence.WriterInvocationID {
				memberFound = true
				break
			}
		}
		if !memberFound {
			return StagedCohortArchive{}, errors.New("staged advance member evidence mismatch")
		}
	}
	tasks := append([]string(nil), ids...)
	return StagedCohortArchive{CohortIndex: prep.CohortIndex, PreparationID: prep.PreparationID, BaseCandidateID: prep.BaseCandidateID, CandidateID: candidateID, TaskIDs: tasks, Batch: *s.GraphWriterBatch, FileIntent: *s.FileIntent, FileReceipt: *s.FileReceipt}, nil
}

// AdvanceStagedCohort moves the current confirmed batch to immutable past
// history, retaining original identity. It appends the compact derived
// stage-advance event carrying exact identities/hashes after an honest
// pre-effect envelope check; replay reconstructs the full archive from
// validated current state. It never rewrites journal past,
// completed history, or repair budgets.
func AdvanceStagedCohort(path string) error {
	s, err := Inspect(path)
	if err != nil {
		return err
	}
	advance, err := expectedStagedAdvance(s)
	if err != nil {
		return err
	}
	raw, err := canonical.Bytes(advance)
	if err != nil {
		return err
	}
	if len(raw) == 0 || len(raw) > canonical.MaxBytes-1024 {
		return errors.New("staged advance envelope exceeds journal bound")
	}
	return Append(path, "graph.staged-advanced", advance)
}

func validateStagedCohort(s Snapshot, tasks []engineeringplan.Task) error {
	if !stagedIsolationEnabled(s) || s.State != "IMPLEMENTING" || s.Graph == nil || s.Candidate == nil || s.FileIntent != nil || s.Creation.Execution == nil {
		return errors.New("staged writer requires a pristine cohort candidate")
	}
	if graphHasUnknown(s) || s.FileOutcome == "UNKNOWN" {
		return errors.New("staged cohort blocked on UNKNOWN")
	}
	ready, err := graphReadyTasks(s)
	if err != nil {
		return err
	}
	readyByID := map[string]engineeringplan.Task{}
	for _, task := range ready {
		if task.Kind == engineeringplan.Implementation {
			readyByID[task.ID] = task
		}
	}
	if len(tasks) == 0 || len(tasks) > 8 {
		return errors.New("staged cohort requires one to eight ready implementations")
	}
	seen := map[string]bool{}
	for _, task := range tasks {
		if seen[task.ID] {
			return errors.New("duplicate staged cohort task")
		}
		seen[task.ID] = true
		readyTask, ok := readyByID[task.ID]
		if !ok {
			return errors.New("staged task is not currently ready")
		}
		if !sameEngineeringTask(task, readyTask) {
			return errors.New("staged cohort differs from exact ready set")
		}
	}
	if len(readyByID) != len(tasks) {
		return errors.New("staged cohort must include the entire current ready implementation subset")
	}
	bare := s
	bare.GraphIsolationPreparation = nil
	preparation, err := expectedStagedPreparation(bare)
	if err != nil {
		return err
	}
	if preparation.Version != 3 {
		return errors.New("staged cohort version differs from policy")
	}
	if len(preparation.SelectedTaskIDs) != len(tasks) {
		return errors.New("staged cohort does not fit declared resource capacities")
	}
	if s.GraphIsolationPreparation != nil && !sameCanonical(*s.GraphIsolationPreparation, preparation) {
		return errors.New("recorded staged preparation differs from the current ready cohort")
	}
	for _, task := range tasks {
		if !containsGraphIsolationTask(preparation.SelectedTaskIDs, task.ID) {
			return errors.New("staged cohort differs from ready implementations")
		}
	}
	for i := range tasks {
		for j := i + 1; j < len(tasks); j++ {
			if !tasksIndependentAndDisjoint(tasks[i], tasks[j]) {
				return errors.New("staged concurrent tasks are not independent and disjoint")
			}
		}
	}
	return nil
}

func sameEngineeringTask(a, b engineeringplan.Task) bool {
	ab, err := canonical.Bytes(a)
	if err != nil {
		return false
	}
	bb, err := canonical.Bytes(b)
	if err != nil {
		return false
	}
	if len(ab) != len(bb) {
		return false
	}
	for i := range ab {
		if ab[i] != bb[i] {
			return false
		}
	}
	return true
}

func stagedGraphWriterIDs(s Snapshot, tasks []engineeringplan.Task) ([]string, error) {
	if err := validateStagedCohort(s, tasks); err != nil {
		return nil, err
	}
	ids := taskIDs(tasks)
	sort.Strings(ids)
	return ids, nil
}

func stagedAggregateParts(s Snapshot, ids []string) (graphWriterAggregateIdentity, []GraphWriterMember, []fileeffects.Change, error) {
	identity, members, changes, err := isolatedGraphWriterAggregateParts(s, ids)
	if err != nil {
		return graphWriterAggregateIdentity{}, nil, nil, err
	}
	if s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 {
		return graphWriterAggregateIdentity{}, nil, nil, errors.New("staged aggregate requires version three preparation")
	}
	identity.Version = 4
	identity.CohortIndex = s.GraphIsolationPreparation.CohortIndex
	return identity, members, changes, nil
}

func buildStagedBatch(ctx context.Context, path string, taskIDs []string) (GraphWriterBatchRecord, error) {
	s, err := Inspect(path)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	ids, err := stagedIDsForResults(s, taskIDs)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	identity, members, changes, err := stagedAggregateParts(s, ids)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	nonce, err := canonical.Hash("harness.graph-writer-staged-aggregate.v1", identity)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	lease, err := worktree.AcquireRead(s.Workspace.Request)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	defer func() { _ = lease.Close() }()
	latest, err := Inspect(path)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	if latest.Candidate == nil || *latest.Candidate != *s.Candidate || latest.GraphIsolationPreparation == nil || latest.GraphIsolationPreparation.PreparationID != identity.IsolationPreparationID {
		return GraphWriterBatchRecord{}, errors.New("staged aggregate parent or frozen cohort changed")
	}
	proposal, err := fileeffects.Prepare(ctx, *latest.Workspace, "staged-graph-writers-"+nonce, changes)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	if err := fileeffects.Preflight(ctx, *latest.Workspace, proposal); err != nil {
		return GraphWriterBatchRecord{}, err
	}
	if proposal.Before != *latest.Candidate {
		return GraphWriterBatchRecord{}, errors.New("staged aggregate preimage differs from current parent candidate")
	}
	prepared, err := preparedFiles(latest, proposal)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	return GraphWriterBatchRecord{Version: 4, GraphDigest: latest.Graph.Digest, Revision: latest.Graph.Revision, CandidateID: candidateID, Members: members, Prepared: prepared, IsolationPreparationID: identity.IsolationPreparationID, CohortIndex: identity.CohortIndex}, nil
}

func expectedStagedBatch(s Snapshot, taskIDs []string, supplied PreparedFiles) (GraphWriterBatchRecord, error) {
	if s.Graph == nil || s.Candidate == nil || s.Workspace == nil || s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 {
		return GraphWriterBatchRecord{}, errors.New("staged aggregate state unavailable")
	}
	ids, err := stagedIDsForResults(s, taskIDs)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	identity, members, changes, err := stagedAggregateParts(s, ids)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	identityHash, err := canonical.Hash("harness.graph-writer-staged-aggregate.v1", identity)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	proposal := supplied.Proposal
	if proposal.Version != 1 || proposal.Nonce != "staged-graph-writers-"+identityHash || proposal.Before != *s.Candidate || !sameCanonical(proposal.Changes, changes) {
		return GraphWriterBatchRecord{}, errors.New("staged aggregate parent proposal identity mismatch")
	}
	if err := proposal.Validate(); err != nil {
		return GraphWriterBatchRecord{}, err
	}
	after, err := composeGraphWriterCandidate(proposal.Before, proposal.BeforeFiles, changes)
	if err != nil || after != proposal.After {
		return GraphWriterBatchRecord{}, errors.Join(errors.New("staged aggregate after-candidate mismatch"), err)
	}
	expected, err := preparedFiles(s, proposal)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	if !sameCanonical(expected, supplied) {
		return GraphWriterBatchRecord{}, errors.New("staged aggregate prepared effect identity mismatch")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	return GraphWriterBatchRecord{Version: 4, GraphDigest: s.Graph.Digest, Revision: s.Graph.Revision, CandidateID: candidateID, Members: members, Prepared: expected, IsolationPreparationID: identity.IsolationPreparationID, CohortIndex: identity.CohortIndex}, nil
}

func stagedIDsForResults(s Snapshot, requested []string) ([]string, error) {
	if s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 || s.Candidate == nil || len(requested) == 0 {
		return nil, errors.New("frozen staged writer results required")
	}
	ids := append([]string(nil), requested...)
	sort.Strings(ids)
	selected := append([]string(nil), s.GraphIsolationPreparation.SelectedTaskIDs...)
	sort.Strings(selected)
	if len(ids) != len(selected) {
		return nil, errors.New("staged aggregate must include the entire frozen cohort")
	}
	for i := range ids {
		if ids[i] == "" || (i > 0 && ids[i] == ids[i-1]) || ids[i] != selected[i] {
			return nil, errors.New("staged aggregate membership differs from frozen cohort")
		}
	}
	return ids, nil
}

func replayStagedBatch(s *Snapshot, batch GraphWriterBatchRecord) error {
	if batch.Version != 4 || !stagedIsolationEnabled(*s) || s.Graph == nil || s.Candidate == nil || s.GraphWriterBatch != nil || s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 || batch.GraphDigest != s.Graph.Digest || batch.Revision != s.Graph.Revision || batch.IsolationPreparationID != s.GraphIsolationPreparation.PreparationID || batch.CohortIndex != s.GraphIsolationPreparation.CohortIndex {
		return errors.New("staged aggregate identity or transition rejected")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil || batch.CandidateID != candidateID || s.GraphIsolationPreparation.BaseCandidateID != candidateID {
		return errors.Join(errors.New("staged aggregate parent candidate changed"), err)
	}
	ids := make([]string, 0, len(batch.Members))
	seen := map[string]bool{}
	if len(batch.Members) == 0 || len(batch.Members) > 8 {
		return errors.New("staged aggregate must contain one to eight members")
	}
	for _, member := range batch.Members {
		if seen[member.TaskID] || member.IsolationID == "" || member.ChildCandidateID == "" {
			return errors.New("invalid or duplicate staged aggregate member")
		}
		seen[member.TaskID] = true
		record, ok := s.GraphWriterResults[member.TaskID]
		if !ok || record.Isolated == nil || graphWriterRecordInvocation(record).ID != member.InvocationID || record.Isolated.IsolationID != member.IsolationID {
			return errors.New("staged aggregate member evidence mismatch")
		}
		childID, idErr := record.Isolated.Candidate.ID()
		if idErr != nil || childID != member.ChildCandidateID {
			return errors.Join(errors.New("staged aggregate child candidate mismatch"), idErr)
		}
		ids = append(ids, member.TaskID)
	}
	expected, err := expectedStagedBatch(*s, ids, batch.Prepared)
	if err != nil {
		return err
	}
	a, err := canonical.Bytes(expected)
	if err != nil {
		return err
	}
	b, err := canonical.Bytes(batch)
	if err != nil || string(a) != string(b) {
		return errors.Join(errors.New("staged aggregate differs from validated child changes"), err)
	}
	s.GraphWriterBatch = &batch
	return nil
}

func recordStagedImplementationProgress(path string, s Snapshot) error {
	if s.Graph == nil || s.Candidate == nil || s.GraphWriterBatch == nil || s.GraphWriterBatch.Version != 4 || s.FileOutcome != "CONFIRMED" || s.FileIntent == nil || s.FileIntent.Prepared.Intent != s.GraphWriterBatch.Prepared.Intent {
		return errors.New("staged implementation progress requires its confirmed aggregate effect")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return err
	}
	afterID, err := s.GraphWriterBatch.Prepared.Proposal.After.ID()
	if err != nil || afterID != candidateID {
		return errors.Join(errors.New("staged aggregate candidate differs"), err)
	}
	for _, member := range s.GraphWriterBatch.Members {
		latest, err := Inspect(path)
		if err != nil {
			return err
		}
		if evidence, ok := latest.Graph.Evidence[member.TaskID]; ok && evidence.Outcome == "completed" && evidence.CandidateID == candidateID && evidence.WriterInvocationID == member.InvocationID {
			continue
		}
		task, ok := latest.Graph.Graph.Task(member.TaskID)
		if !ok || task.Kind != engineeringplan.Implementation {
			return errors.New("staged writer task disappeared from graph")
		}
		record, ok := latest.GraphWriterResults[member.TaskID]
		if !ok || graphWriterRecordInvocation(record).ID != member.InvocationID || !graphWriterChangesWithinTask(task, graphWriterRecordChanges(record)) {
			return errors.New("staged writer task evidence is missing or substituted")
		}
		attemptID := "attempt-" + taskIDAttemptSuffix(len(task.Attempts)+1)
		p := GraphProgress{Version: 1, PlanID: latest.Graph.PlanID, Digest: latest.Graph.Digest, TaskID: task.ID, AttemptID: attemptID, Outcome: "completed", CandidateID: candidateID, WriterInvocationID: member.InvocationID}
		if _, err := recordGraphProgress(path, p); err != nil {
			return err
		}
	}
	return nil
}

func stagedAllImplementationsCompleted(s Snapshot) bool {
	if s.Graph == nil {
		return false
	}
	for _, task := range s.Graph.Graph.Tasks {
		if task.Kind != engineeringplan.Implementation {
			continue
		}
		if !task.Completed {
			return false
		}
		evidence, ok := s.Graph.Evidence[task.ID]
		if !ok || evidence.Outcome != "completed" {
			return false
		}
	}
	return true
}

func taskIDAttemptSuffix(n int) string {
	if n < 1 {
		return "x"
	}
	digits := "0123456789"
	out := ""
	for n > 0 {
		out = string(digits[n%10]) + out
		n /= 10
	}
	return out
}

func prepareStagedBindings(ctx context.Context, path string) ([]isolatedWriterBinding, error) {
	s, err := Inspect(path)
	if err != nil {
		return nil, err
	}
	if !stagedIsolationEnabled(s) {
		return nil, errors.New("staged isolation not enabled")
	}
	if s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 {
		if _, err := PrepareStagedCohort(ctx, path); err != nil {
			return nil, err
		}
		s, err = Inspect(path)
		if err != nil {
			return nil, err
		}
	}
	if s.GraphIsolationPreparation == nil || len(s.GraphIsolationPreparation.SelectedTaskIDs) == 0 {
		return nil, errors.New("frozen staged cohort is empty")
	}
	ids := append([]string(nil), s.GraphIsolationPreparation.SelectedTaskIDs...)
	for _, taskID := range ids {
		if _, err := CreateStagedTaskIsolation(ctx, path, taskID); err != nil {
			return nil, err
		}
		s, err = Inspect(path)
		if err != nil {
			return nil, err
		}
		question, err := graphWriterTaskQuestion(s, taskID)
		if err != nil {
			return nil, err
		}
		if err := maybeAdmitIsolatedWriterTaskContext(ctx, path, taskID, question); err != nil {
			return nil, err
		}
	}
	s, err = Inspect(path)
	if err != nil {
		return nil, err
	}
	bindings := make([]isolatedWriterBinding, 0, len(ids))
	for _, taskID := range ids {
		binding, err := stagedForkBindingForTask(s, taskID)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

func buildStagedWaveSpecs(controllerPath string, s Snapshot, waveTasks []engineeringplan.Task, waveIndex int) ([]taskscheduler.TaskSpec, error) {
	if !stagedIsolationEnabled(s) || s.State != "IMPLEMENTING" || s.Graph == nil || s.Candidate == nil || s.FileIntent != nil || s.Creation.Execution == nil {
		return nil, errors.New("staged wave writer requires a pristine cohort candidate")
	}
	if s.GraphIsolationPreparation == nil || s.GraphIsolationPreparation.Version != 3 || len(s.GraphIsolationPreparation.Waves) == 0 {
		return nil, errors.New("frozen staged waves are unavailable")
	}
	if waveIndex < 0 || waveIndex >= len(s.GraphIsolationPreparation.Waves) {
		return nil, errors.New("staged wave index is outside the frozen cohort")
	}
	frozen := append([]string(nil), s.GraphIsolationPreparation.Waves[waveIndex]...)
	sort.Strings(frozen)
	if len(waveTasks) == 0 || len(waveTasks) != len(frozen) {
		return nil, errors.New("staged wave membership differs from frozen cohort")
	}
	ready, err := graphReadyTasks(s)
	if err != nil {
		return nil, err
	}
	readyByID := map[string]engineeringplan.Task{}
	for _, task := range ready {
		if task.Kind == engineeringplan.Implementation {
			readyByID[task.ID] = task
		}
	}
	got := make([]string, 0, len(waveTasks))
	for _, task := range waveTasks {
		readyTask, ok := readyByID[task.ID]
		if !ok || !reflect.DeepEqual(task, readyTask) {
			return nil, errors.New("staged wave task is stale or substituted")
		}
		if !containsGraphIsolationTask(frozen, task.ID) {
			return nil, errors.New("staged wave task is outside its frozen wave")
		}
		got = append(got, task.ID)
	}
	sort.Strings(got)
	for i := range got {
		if got[i] != frozen[i] {
			return nil, errors.New("staged wave membership differs from frozen cohort")
		}
	}
	for i := range waveTasks {
		for j := i + 1; j < len(waveTasks); j++ {
			if !tasksIndependentAndDisjoint(waveTasks[i], waveTasks[j]) {
				return nil, errors.New("staged wave tasks are not independent and disjoint")
			}
		}
	}
	specs := make([]taskscheduler.TaskSpec, 0, len(waveTasks))
	for _, task := range waveTasks {
		if task.Kind != engineeringplan.Implementation {
			return nil, errors.New("writer cohort contains a non-implementation")
		}
		invocation, err := writerInvocationForTask(s, task.ID)
		if err != nil {
			return nil, err
		}
		if invocation.Profile.Runtime != "codex-app-server" {
			return nil, errors.New("parallel graph writer runtime unsupported")
		}
		specs = append(specs, taskscheduler.TaskSpec{ID: task.ID, RunID: s.RunID, ControllerPath: controllerPath, Operation: taskscheduler.OperationWriter, InvocationID: invocation.ID})
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
	return specs, nil
}

func runStagedWaves(ctx context.Context, controllerPath string, initial Snapshot, tasks []engineeringplan.Task) error {
	if !stagedIsolationEnabled(initial) || initial.Creation.Execution == nil || initial.State != "IMPLEMENTING" {
		return errors.New("staged wave writer batch is not enabled here")
	}
	if _, err := stagedGraphWriterIDs(initial, tasks); err != nil {
		return err
	}
	bindings, err := prepareStagedBindings(ctx, controllerPath)
	if err != nil {
		return err
	}
	current, err := Inspect(controllerPath)
	if err != nil {
		return err
	}
	if len(bindings) != len(tasks) {
		return errors.New("confirmed staged bindings differ from current ready cohort")
	}
	bindingIDs := map[string]bool{}
	for _, binding := range bindings {
		bindingIDs[binding.TaskID] = true
	}
	for _, task := range tasks {
		if !bindingIDs[task.ID] {
			return errors.New("confirmed staged binding membership differs from frozen task cohort")
		}
		if _, err := stagedForkBindingForTask(current, task.ID); err != nil {
			return err
		}
	}
	if _, err := stagedGraphWriterIDs(current, tasks); err != nil {
		return err
	}
	if current.GraphIsolationPreparation == nil || current.GraphIsolationPreparation.Version != 3 || len(current.GraphIsolationPreparation.Waves) == 0 {
		return errors.New("frozen staged waves are unavailable")
	}
	byID := map[string]engineeringplan.Task{}
	for _, task := range tasks {
		byID[task.ID] = task
	}
	preparationID := current.GraphIsolationPreparation.PreparationID
	for waveIndex, waveIDs := range current.GraphIsolationPreparation.Waves {
		latest, err := Inspect(controllerPath)
		if err != nil {
			return err
		}
		if latest.FileIntent != nil || latest.GraphWriterBatch != nil {
			return errors.New("staged waves require a pristine parent before aggregation")
		}
		if graphHasUnknown(latest) || latest.FileOutcome == "UNKNOWN" {
			return errors.New("staged waves blocked on UNKNOWN; explicit reconciliation required")
		}
		for _, fork := range latest.GraphStagedForks {
			if fork.Outcome == "UNKNOWN" {
				return errors.New("staged waves blocked on UNKNOWN fork; explicit reconciliation required")
			}
		}
		for _, isolation := range latest.GraphIsolations {
			if isolation.Outcome == "UNKNOWN" {
				return errors.New("staged waves blocked on UNKNOWN child isolation")
			}
		}
		waveTasks := make([]engineeringplan.Task, 0, len(waveIDs))
		for _, id := range waveIDs {
			task, ok := byID[id]
			if !ok {
				return errors.New("staged wave task is outside the frozen cohort")
			}
			waveTasks = append(waveTasks, task)
		}
		allProposed := true
		for _, task := range waveTasks {
			record, ok := latest.GraphWriterResults[task.ID]
			if !ok || record.Isolated == nil {
				allProposed = false
				break
			}
		}
		if allProposed {
			continue
		}
		for _, task := range waveTasks {
			if _, ok := latest.GraphWriterResults[task.ID]; ok {
				return errors.New("staged wave is partially proposed; UNKNOWN or tamper suspected")
			}
		}
		specs, err := buildStagedWaveSpecs(controllerPath, latest, waveTasks, waveIndex)
		if err != nil {
			return err
		}
		workers := latest.Creation.Execution.EffectiveMaxParallel()
		if workers > len(specs) {
			workers = len(specs)
		}
		if workers > 8 {
			workers = 8
		}
		if workers < 1 {
			return errors.New("staged wave has no scheduler capacity")
		}
		id, err := graphCohortID(latest.Graph.Digest, latest.Graph.Revision, specs)
		if err != nil {
			return err
		}
		waveNonce, err := canonical.Hash("harness.staged-wave-schedule.v1", struct {
			PreparationID string
			WaveIndex     int
			CohortID      string
		}{preparationID, waveIndex, id})
		if err != nil {
			return err
		}
		schedulePath := controllerPath + ".staged-graph-writers-wave-" + waveNonce[:16] + ".jsonl"
		definition := taskscheduler.Definition{Version: 1, Nonce: "staged-graph-writers-wave-" + waveNonce, Tasks: specs}
		if _, err := taskscheduler.Bind(schedulePath, definition); err != nil {
			return err
		}
		if err := runScheduledGraphWriterCohort(ctx, controllerPath, schedulePath, specs, workers, true); err != nil {
			return err
		}
		after, err := Inspect(controllerPath)
		if err != nil {
			return err
		}
		for _, task := range waveTasks {
			record, ok := after.GraphWriterResults[task.ID]
			if !ok || record.Isolated == nil {
				return errors.New("staged wave settled without a complete candidate-bound proposal set")
			}
		}
	}
	return nil
}

func applyStagedBatch(ctx context.Context, path string, s Snapshot) (Snapshot, error) {
	if s.GraphWriterBatch == nil || s.GraphWriterBatch.Version != 4 || !stagedIsolationEnabled(s) {
		return s, errors.New("staged aggregate is unavailable")
	}
	if s.FileOutcome == "UNKNOWN" {
		return s, errors.New("staged file effect is UNKNOWN")
	}
	if s.FileIntent != nil {
		return s, errors.New("staged aggregate already has a file attempt")
	}
	if err := autonomousDispatchBlocked(s); err != nil {
		return s, err
	}
	authorization, err := autonomousFileAuthorization(s, s.GraphWriterBatch.Prepared.Intent)
	if err != nil {
		return s, err
	}
	if _, err := ApplyFiles(ctx, path, s.GraphWriterBatch.Prepared, authorization); err != nil {
		return InspectOr(s, path, err)
	}
	latest, err := Inspect(path)
	return latest, err
}

var _ = time.Now
var _ = journal.Event{}
