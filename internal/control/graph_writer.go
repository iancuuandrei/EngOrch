package control

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/codexhost"
	"harness.local/engorch/internal/engineeringplan"
	"harness.local/engorch/internal/fileeffects"
	"harness.local/engorch/internal/journal"
	"harness.local/engorch/internal/runtime"
	"harness.local/engorch/internal/safepath"
	"harness.local/engorch/internal/taskscheduler"
	"harness.local/engorch/internal/worktree"
	"harness.local/engorch/internal/writercontract"
)

type graphWriterTaskContextKey struct{}

var graphWriterActiveMu sync.Mutex
var graphWriterActiveCalls = map[string]int{}

func graphWriterActivityKey(path, taskID string) string { return path + "\x00" + taskID }

func markGraphWriterActive(path, taskID string) func() {
	key := graphWriterActivityKey(path, taskID)
	graphWriterActiveMu.Lock()
	graphWriterActiveCalls[key]++
	graphWriterActiveMu.Unlock()
	return func() {
		graphWriterActiveMu.Lock()
		if graphWriterActiveCalls[key] <= 1 {
			delete(graphWriterActiveCalls, key)
		} else {
			graphWriterActiveCalls[key]--
		}
		graphWriterActiveMu.Unlock()
	}
}

func graphWriterCurrentlyActive(path, taskID string) bool {
	graphWriterActiveMu.Lock()
	defer graphWriterActiveMu.Unlock()
	return graphWriterActiveCalls[graphWriterActivityKey(path, taskID)] > 0
}

func autonomousGraphWriterDispatchBlocked(s Snapshot, path, taskID string) error {
	check := s
	if len(s.GraphWriterHosts) > 0 {
		check.GraphWriterHosts = make(map[string]WriterHostState, len(s.GraphWriterHosts))
		for id, host := range s.GraphWriterHosts {
			if id != taskID && graphWriterCurrentlyActive(path, id) {
				continue
			}
			check.GraphWriterHosts[id] = host
		}
	}
	return autonomousDispatchBlocked(check)
}

func withGraphWriterTask(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, graphWriterTaskContextKey{}, taskID)
}

func graphWriterTaskFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	taskID, _ := ctx.Value(graphWriterTaskContextKey{}).(string)
	return taskID
}

func graphWriterProjectedSnapshot(s Snapshot, taskID string) Snapshot {
	if taskID != "" {
		s.WriterHost = graphWriterHostState(s, taskID)
	}
	return s
}

func appendWriterHostTransition(path, taskID, kind string, payload any) error {
	if taskID == "" {
		return Append(path, kind, payload)
	}
	event := GraphWriterHostEvent{TaskID: taskID}
	graphKind := "graph.writer." + strings.TrimPrefix(kind, "writer.")
	switch kind {
	case "writer.host-intent", "writer.host-ready":
		intent, ok := payload.(WriterHostIntent)
		if !ok {
			return errors.New("graph writer host intent payload mismatch")
		}
		event.Intent = &intent
	case "writer.host-observed":
		receipt, ok := payload.(codexhost.Receipt)
		if !ok {
			return errors.New("graph writer host receipt payload mismatch")
		}
		event.HostReceipt = &receipt
	case "writer.runtime-observed":
		receipt, ok := payload.(WriterRuntimeReceipt)
		if !ok {
			return errors.New("graph writer runtime receipt payload mismatch")
		}
		event.RuntimeReceipt = &receipt
	default:
		return errors.New("unsupported writer host transition")
	}
	return appendGraphWriterHostEvent(path, graphKind, taskID, event)
}

// GraphWriterHostEvent attributes one isolated Codex host transition to a
// frozen implementation task. It never carries write authorization.
type GraphWriterHostEvent struct {
	TaskID         string                `json:"task_id"`
	Intent         *WriterHostIntent     `json:"intent,omitempty"`
	HostReceipt    *codexhost.Receipt    `json:"host_receipt,omitempty"`
	RuntimeReceipt *WriterRuntimeReceipt `json:"runtime_receipt,omitempty"`
}

// GraphWriterRecord retains one task's actual runtime result and independently
// prepared proposal against the common frozen candidate.
type GraphWriterRecord struct {
	TaskID string       `json:"task_id"`
	Writer WriterRecord `json:"writer"`
}

// GraphWriterMember binds one aggregate member to its exact graph task and
// actual writer invocation.
type GraphWriterMember struct {
	TaskID       string `json:"task_id"`
	InvocationID string `json:"invocation_id"`
}

// GraphWriterBatchRecord is a runner-owned deterministic composition of actual
// member proposals. It contains no fabricated model Result.
type GraphWriterBatchRecord struct {
	Version     int                 `json:"version"`
	GraphDigest string              `json:"graph_digest"`
	Revision    int                 `json:"revision"`
	CandidateID string              `json:"candidate_id"`
	Members     []GraphWriterMember `json:"members"`
	Prepared    PreparedFiles       `json:"prepared"`
}

type graphWriterAggregateIdentity struct {
	Version     int                 `json:"version"`
	GraphDigest string              `json:"graph_digest"`
	Revision    int                 `json:"revision"`
	CandidateID string              `json:"candidate_id"`
	Members     []GraphWriterMember `json:"members"`
}

func expectedWriterHostForTask(s Snapshot, taskID string) (WriterHostIntent, error) {
	i, err := writerInvocationForTask(s, taskID)
	if err != nil {
		return WriterHostIntent{}, err
	}
	c := s.Creation.Config.Codex
	if i.Profile.Runtime != "codex-app-server" || c == nil {
		return WriterHostIntent{}, errors.New("configured Codex graph writer required")
	}
	rel, err := filepath.Rel(s.Creation.Repository.Root, c.StateRoot)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return WriterHostIntent{}, errors.New("writer host state must be outside source repository")
	}
	l, err := codexhost.Expected(filepath.Join(c.StateRoot, s.RunID, "writer-"+i.ID), c.Executable, c.ExecutableHash)
	return WriterHostIntent{i, l}, err
}

func graphWriterHostState(s Snapshot, taskID string) *WriterHostState {
	state, ok := s.GraphWriterHosts[taskID]
	if !ok {
		return nil
	}
	copy := state
	return &copy
}

func appendGraphWriterHostEvent(path, kind, taskID string, event GraphWriterHostEvent) error {
	if taskID == "" || event.TaskID != taskID {
		return errors.New("graph writer host task binding required")
	}
	return Append(path, kind, event)
}

func replayGraphWriterHost(s *Snapshot, e journal.Event) error {
	var event GraphWriterHostEvent
	if err := canonical.Decode(e.Payload, &event); err != nil {
		return err
	}
	if event.TaskID == "" {
		return errors.New("graph writer host task missing")
	}
	expected, err := expectedWriterHostForTask(*s, event.TaskID)
	if err != nil {
		return err
	}
	host, exists := s.GraphWriterHosts[event.TaskID]
	switch e.Kind {
	case "graph.writer.host-intent":
		if exists || event.Intent == nil || *event.Intent != expected || event.HostReceipt != nil || event.RuntimeReceipt != nil {
			return errors.New("graph writer host intent substitution or duplication")
		}
		host = WriterHostState{Intent: expected}
	case "graph.writer.host-ready":
		if !exists || host.Ready || host.Intent != expected || event.Intent == nil || *event.Intent != expected || event.HostReceipt != nil || event.RuntimeReceipt != nil {
			return errors.New("graph writer host preparation transition rejected")
		}
		host.Ready = true
	case "graph.writer.host-observed":
		if !exists || !host.Ready || host.Intent != expected || event.HostReceipt == nil || event.Intent != nil || event.RuntimeReceipt != nil {
			return errors.New("graph writer host observation transition rejected")
		}
		if err := validateCodexHostReceipt(*event.HostReceipt, expected.Launch, s.Creation.Config.Codex); err != nil {
			return err
		}
		host.Receipt = event.HostReceipt
	case "graph.writer.runtime-observed":
		if !exists || !host.Ready || host.Receipt == nil || host.Intent != expected || host.RuntimeReceipt != nil || event.RuntimeReceipt == nil || event.Intent != nil || event.HostReceipt != nil {
			return errors.New("graph writer runtime receipt transition rejected")
		}
		receipt := event.RuntimeReceipt
		if receipt.InvocationID != expected.Invocation.ID || strings.TrimSpace(receipt.ThreadID) == "" || len(receipt.ThreadID) > 256 || strings.TrimSpace(receipt.TurnID) == "" || len(receipt.TurnID) > 256 {
			return errors.New("graph writer runtime receipt identity mismatch")
		}
		for _, digest := range []string{receipt.JournalHead, receipt.ResultHash} {
			if err := safepath.RequireDigest(digest); err != nil {
				return err
			}
		}
		host.RuntimeReceipt = receipt
	default:
		return errors.New("unsupported graph writer host event")
	}
	if s.GraphWriterHosts == nil {
		s.GraphWriterHosts = make(map[string]WriterHostState)
	}
	s.GraphWriterHosts[event.TaskID] = host
	return nil
}

func replayGraphWriterProposal(s *Snapshot, e journal.Event, seen map[string]bool) error {
	var record GraphWriterRecord
	if err := canonical.Decode(e.Payload, &record); err != nil {
		return err
	}
	if record.TaskID == "" || s.Graph == nil || s.GraphWriterResults[record.TaskID].TaskID != "" {
		return errors.New("graph writer proposal task missing or duplicated")
	}
	task, ok := s.Graph.Graph.Task(record.TaskID)
	if !ok || task.Kind != "implementation" || task.Completed {
		return errors.New("graph writer proposal task is not an incomplete implementation")
	}
	base, err := writerInvocationForTask(*s, record.TaskID)
	if err != nil {
		return err
	}
	i, err := resolveScheduledRecordedInvocation(*s, base, record.Writer.Invocation)
	if err != nil || i != record.Writer.Invocation {
		return errors.New("graph writer proposal invocation mismatch")
	}
	if err := runtime.ValidateResult(i, record.Writer.Result, true); err != nil {
		return err
	}
	host, ok := s.GraphWriterHosts[record.TaskID]
	if i.Profile.Runtime != "codex-app-server" || !ok || host.Intent.Invocation != i || host.RuntimeReceipt == nil {
		return errors.New("graph writer proposal requires exact runtime receipt")
	}
	hash, err := canonical.Hash("harness.writer-result.v1", record.Writer.Result)
	if err != nil || hash != host.RuntimeReceipt.ResultHash {
		return errors.Join(errors.New("graph writer proposal differs from runtime result"), err)
	}
	if s.Candidate == nil {
		return errors.New("graph writer candidate unavailable")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return err
	}
	var reply WriterProposal
	if writercontract.IsAnchoredEdits(s.Creation.Config.WriterContract) {
		anchored, err := decodeAnchoredProposal(record.Writer.Result.Output)
		if err != nil || anchored.CandidateID != candidateID {
			return errors.Join(errors.New("graph writer candidate mismatch"), err)
		}
		changes, err := composeAnchoredProposal(anchored, record.Writer.Prepared.Proposal.BeforeFiles, record.Writer.EditPreimages)
		if err != nil {
			return err
		}
		reply = WriterProposal{CandidateID: candidateID, Changes: changes}
	} else {
		if len(record.Writer.EditPreimages) != 0 {
			return errors.New("unexpected graph writer edit preimages")
		}
		reply, err = decodeWriterProposal(s.Creation.Config.WriterContract, record.Writer.Result.Output)
		if err != nil || reply.CandidateID != candidateID {
			return errors.Join(errors.New("graph writer candidate mismatch"), err)
		}
	}
	prepared, err := preparedFiles(*s, record.Writer.Prepared.Proposal)
	if err != nil {
		return err
	}
	a, err := canonical.Bytes(prepared)
	if err != nil {
		return err
	}
	b, err := canonical.Bytes(record.Writer.Prepared)
	if err != nil || string(a) != string(b) {
		return errors.Join(errors.New("graph writer prepared effect substitution"), err)
	}
	if !graphWriterChangesWithinTask(task, reply.Changes) {
		return errors.New("graph writer proposal exceeds its task write paths")
	}
	a, err = canonicalWriterChanges(reply.Changes)
	if err != nil {
		return err
	}
	b, err = canonicalWriterChanges(record.Writer.Prepared.Proposal.Changes)
	if err != nil || string(a) != string(b) {
		return errors.Join(errors.New("graph writer reply differs from prepared changes"), err)
	}
	key := "graph-writer-proposal:" + i.ID
	if seen[key] {
		return errors.New("duplicate graph writer proposal invocation")
	}
	seen[key] = true
	if s.GraphWriterResults == nil {
		s.GraphWriterResults = make(map[string]GraphWriterRecord)
	}
	s.GraphWriterResults[record.TaskID] = record
	return nil
}

func graphWriterChangesWithinTask(task engineeringplan.Task, changes []fileeffects.Change) bool {
	if len(changes) == 0 {
		return false
	}
	for _, change := range changes {
		admitted := false
		for _, writePath := range task.WritePaths {
			if change.Path == writePath || strings.HasPrefix(change.Path, writePath+"/") {
				admitted = true
				break
			}
		}
		if !admitted {
			return false
		}
	}
	return true
}

func replayGraphWriterBatch(s *Snapshot, e journal.Event) error {
	var batch GraphWriterBatchRecord
	if err := canonical.Decode(e.Payload, &batch); err != nil {
		return err
	}
	if batch.Version != 1 || s.Graph == nil || s.Candidate == nil || s.GraphWriterBatch != nil || batch.GraphDigest != s.Graph.Digest || batch.Revision != s.Graph.Revision {
		return errors.New("graph writer batch identity or transition rejected")
	}
	if !parallelImplementationEnabled(*s) {
		return errors.New("graph writer batch requires the immutable parallel implementation policy")
	}
	candidateID, err := s.Candidate.ID()
	if err != nil || candidateID != batch.CandidateID {
		return errors.Join(errors.New("graph writer batch candidate changed"), err)
	}
	if len(batch.Members) < 1 || len(batch.Members) > 2 {
		return errors.New("graph writer batch must contain one or two members")
	}
	memberIDs := make([]string, 0, len(batch.Members))
	seenTasks := map[string]bool{}
	for _, member := range batch.Members {
		if seenTasks[member.TaskID] {
			return errors.New("duplicate graph writer batch member")
		}
		seenTasks[member.TaskID] = true
		record, ok := s.GraphWriterResults[member.TaskID]
		if !ok || record.Writer.Invocation.ID != member.InvocationID || record.Writer.Prepared.Proposal.Before != *s.Candidate {
			return errors.New("graph writer batch member evidence mismatch")
		}
		memberIDs = append(memberIDs, member.TaskID)
	}
	expected, err := buildGraphWriterBatch(*s, memberIDs)
	if err != nil {
		return err
	}
	a, err := canonical.Bytes(expected)
	if err != nil {
		return err
	}
	b, err := canonical.Bytes(batch)
	if err != nil || string(a) != string(b) {
		return errors.Join(errors.New("graph writer aggregate differs from deterministic composition"), err)
	}
	s.GraphWriterBatch = &batch
	return nil
}

func buildGraphWriterBatch(s Snapshot, taskIDs []string) (GraphWriterBatchRecord, error) {
	if s.Graph == nil || s.Candidate == nil || len(taskIDs) == 0 || len(taskIDs) > 2 {
		return GraphWriterBatchRecord{}, errors.New("invalid graph writer batch request")
	}
	ids := append([]string(nil), taskIDs...)
	sort.Strings(ids)
	ready, err := graphReadyTasks(s)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	var readyIDs []string
	for _, task := range ready {
		if task.Kind == "implementation" {
			readyIDs = append(readyIDs, task.ID)
		}
	}
	sort.Strings(readyIDs)
	if len(ids) != len(readyIDs) {
		return GraphWriterBatchRecord{}, errors.New("aggregate must include every ready implementation")
	}
	for i := range ids {
		if ids[i] != readyIDs[i] {
			return GraphWriterBatchRecord{}, errors.New("aggregate implementation cohort changed")
		}
	}
	var members []GraphWriterMember
	var changes []fileeffects.Change
	var beforeFiles []worktree.FileState
	for _, id := range ids {
		record, ok := s.GraphWriterResults[id]
		if !ok || record.Writer.Prepared.Proposal.Before != *s.Candidate {
			return GraphWriterBatchRecord{}, fmt.Errorf("graph writer task %q lacks a candidate-bound proposal", id)
		}
		if beforeFiles == nil {
			beforeFiles = append([]worktree.FileState{}, record.Writer.Prepared.Proposal.BeforeFiles...)
		} else if !sameCanonical(beforeFiles, record.Writer.Prepared.Proposal.BeforeFiles) {
			return GraphWriterBatchRecord{}, errors.New("graph writer proposal manifests differ for the frozen candidate")
		}
		members = append(members, GraphWriterMember{TaskID: id, InvocationID: record.Writer.Invocation.ID})
		changes = append(changes, record.Writer.Prepared.Proposal.Changes...)
	}
	if len(ids) == 2 {
		first, ok1 := s.Graph.Graph.Task(ids[0])
		second, ok2 := s.Graph.Graph.Task(ids[1])
		if !ok1 || !ok2 || !tasksIndependentAndDisjoint(first, second) {
			return GraphWriterBatchRecord{}, errors.New("graph writer tasks are not independent and disjoint")
		}
	}
	candidateID, err := s.Candidate.ID()
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	identity := graphWriterAggregateIdentity{Version: 1, GraphDigest: s.Graph.Digest, Revision: s.Graph.Revision, CandidateID: candidateID, Members: members}
	nonce, err := canonical.Hash("harness.graph-writer-aggregate.v1", identity)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	proposal := fileeffects.Proposal{Version: 1, Nonce: "graph-writers-" + nonce, Before: *s.Candidate, BeforeFiles: beforeFiles, Changes: changes}
	proposal.After, err = composeGraphWriterCandidate(proposal.Before, proposal.BeforeFiles, proposal.Changes)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	prepared, err := preparedFiles(s, proposal)
	if err != nil {
		return GraphWriterBatchRecord{}, err
	}
	return GraphWriterBatchRecord{Version: 1, GraphDigest: s.Graph.Digest, Revision: s.Graph.Revision, CandidateID: candidateID, Members: members, Prepared: prepared}, nil
}

func composeGraphWriterCandidate(before worktree.Candidate, beforeFiles []worktree.FileState, changes []fileeffects.Change) (worktree.Candidate, error) {
	states := make(map[string]worktree.FileState, len(beforeFiles)+len(changes))
	for _, state := range beforeFiles {
		states[state.Path] = state
	}
	for _, change := range changes {
		current, exists := states[change.Path]
		if change.BeforeHash == nil {
			if exists {
				return worktree.Candidate{}, errors.New("aggregate expected an absent path")
			}
		} else if !exists || current.Hash != *change.BeforeHash {
			return worktree.Candidate{}, errors.New("aggregate file precondition differs from frozen candidate")
		}
		if change.ContentBase64 == nil {
			if !exists || change.Executable {
				return worktree.Candidate{}, errors.New("aggregate contains an invalid deletion")
			}
			delete(states, change.Path)
			continue
		}
		content, err := base64.StdEncoding.Strict().DecodeString(*change.ContentBase64)
		if err != nil || base64.StdEncoding.EncodeToString(content) != *change.ContentBase64 {
			return worktree.Candidate{}, errors.Join(errors.New("aggregate content is not canonical base64"), err)
		}
		hash := sha256.Sum256(content)
		states[change.Path] = worktree.FileState{Path: change.Path, Hash: hex.EncodeToString(hash[:]), Executable: change.Executable}
	}
	files := make([]worktree.FileState, 0, len(states))
	for _, state := range states {
		files = append(files, state)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	hash, err := worktree.FilesID(files)
	if err != nil {
		return worktree.Candidate{}, err
	}
	after := before
	after.FilesHash = hash
	after.FileCount = len(files)
	return after, nil
}

func graphContainsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func tasksIndependentAndDisjoint(first, second engineeringplan.Task) bool {
	for _, dependency := range first.Dependencies {
		if dependency == second.ID {
			return false
		}
	}
	for _, dependency := range second.Dependencies {
		if dependency == first.ID {
			return false
		}
	}
	for _, a := range first.WritePaths {
		for _, b := range second.WritePaths {
			a, b = strings.ToLower(strings.TrimSuffix(a, "/")), strings.ToLower(strings.TrimSuffix(b, "/"))
			if a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/") {
				return false
			}
		}
	}
	return len(first.WritePaths) > 0 && len(second.WritePaths) > 0
}

func recordGraphWriterBatch(path string, batch GraphWriterBatchRecord) error {
	return Append(path, "graph.writer.batch-proposed", batch)
}

func recordGraphWriterProposal(path string, taskID string, writer WriterRecord) error {
	return Append(path, "graph.writer.proposed", GraphWriterRecord{TaskID: taskID, Writer: writer})
}

func graphWriterInvocation(s Snapshot, taskID string) (runtime.Invocation, error) {
	return writerInvocationForTask(s, taskID)
}

func graphWriterTaskQuestion(s Snapshot, taskID string) (string, error) {
	if s.Graph == nil {
		return "", errors.New("graph unavailable")
	}
	task, ok := s.Graph.Graph.Task(taskID)
	if !ok || task.Kind != "implementation" {
		return "", errors.New("implementation task unavailable")
	}
	question := fmt.Sprintf("[%s] %s | scope: %s | write paths: %s | objective: %s", task.ID, task.Title, strings.Join(task.ScopePaths, ","), strings.Join(task.WritePaths, ","), s.Creation.Objective)
	if len(question) > 4096 {
		return "", errors.New("writer task context exceeds bound")
	}
	return question, nil
}

func prepareGraphWriterFiles(ctx context.Context, path string, taskID string, invocation runtime.Invocation, result runtime.Result) (WriterRecord, error) {
	s, err := Inspect(path)
	if err != nil {
		return WriterRecord{}, err
	}
	expected, err := graphWriterInvocation(s, taskID)
	if err != nil || expected != invocation {
		return WriterRecord{}, errors.Join(errors.New("graph writer invocation is stale or substituted"), err)
	}
	prepared, preimages, err := prepareWriterFilesForTask(ctx, path, taskID, invocation, result)
	if err != nil {
		return WriterRecord{}, err
	}
	task, _ := s.Graph.Graph.Task(taskID)
	var reply WriterProposal
	if writercontract.IsAnchoredEdits(s.Creation.Config.WriterContract) {
		anchored, err := decodeAnchoredProposal(result.Output)
		if err != nil {
			return WriterRecord{}, err
		}
		changes, err := composeAnchoredProposal(anchored, prepared.Proposal.BeforeFiles, preimages)
		if err != nil {
			return WriterRecord{}, err
		}
		reply = WriterProposal{CandidateID: anchored.CandidateID, Changes: changes}
	} else {
		reply, err = decodeWriterProposal(s.Creation.Config.WriterContract, result.Output)
		if err != nil {
			return WriterRecord{}, err
		}
	}
	if !graphWriterChangesWithinTask(task, reply.Changes) {
		return WriterRecord{}, errors.New("graph writer proposal exceeds declared task write paths")
	}
	return WriterRecord{Invocation: invocation, Result: result, Prepared: prepared, EditPreimages: preimages}, nil
}

func validateGraphWriterInvocationID(id string) error { return safepath.RequireDigest(id) }

func isStaticGraphWriterCohort(s Snapshot, claim taskscheduler.Claim) bool {
	if claim.Task.Operation != taskscheduler.OperationWriter || claim.AgentTurn != nil || !parallelImplementationEnabled(s) || s.State != "IMPLEMENTING" || s.Graph == nil || s.Candidate == nil || s.Workspace == nil || claim.Task.RunID != s.RunID || claim.Task.Input != "" {
		return false
	}
	ready, err := graphReadyTasks(s)
	if err != nil {
		return false
	}
	for _, task := range ready {
		if task.ID != claim.Task.ID || task.Kind != engineeringplan.Implementation {
			continue
		}
		invocation, err := writerInvocationForTask(s, task.ID)
		return err == nil && invocation.Profile.Runtime == "codex-app-server" && invocation.ID == claim.Task.InvocationID
	}
	return false
}

func verifyGraphWriterCohortDelta(controllerPath string, claim taskscheduler.Claim) error {
	events, err := journal.Read(controllerPath)
	if err != nil || len(events) == 0 {
		return errors.New("controller journal unavailable for writer cohort check")
	}
	idx := -1
	for i, event := range events {
		if event.Hash == claim.ControllerHead {
			idx = i
		}
	}
	if idx < 0 {
		return errors.New("writer cohort bound head not found")
	}
	if idx == len(events)-1 {
		return nil
	}
	bound, err := Replay(events[:idx+1])
	if err != nil {
		return err
	}
	current, err := Replay(events)
	if err != nil {
		return err
	}
	if bound.RunID != current.RunID || bound.PlanID != current.PlanID || bound.Graph == nil || current.Graph == nil || bound.Graph.Digest != current.Graph.Digest || bound.Graph.Revision != current.Graph.Revision || bound.Candidate == nil || current.Candidate == nil || *bound.Candidate != *current.Candidate || lifecycleStatus(bound) != lifecycleStatus(current) || bound.FileOutcome == "UNKNOWN" || current.FileOutcome == "UNKNOWN" || graphHasUnknown(bound) || graphHasUnknown(current) {
		return errors.New("writer cohort binding changed or became UNKNOWN")
	}
	ready, err := graphReadyTasks(bound)
	if err != nil {
		return err
	}
	cohort := map[string]engineeringplan.Task{}
	invocations := map[string]string{}
	for _, task := range ready {
		if task.Kind != engineeringplan.Implementation {
			continue
		}
		invocation, err := writerInvocationForTask(bound, task.ID)
		if err != nil {
			return err
		}
		if invocation.Profile.Runtime != "codex-app-server" {
			return errors.New("writer cohort runtime unsupported")
		}
		cohort[task.ID] = task
		invocations[invocation.ID] = task.ID
	}
	if len(cohort) < 1 || len(cohort) > 2 || cohort[claim.Task.ID].ID == "" || invocations[claim.Task.InvocationID] != claim.Task.ID {
		return errors.New("writer claim is outside frozen implementation cohort")
	}
	allowed := map[string]bool{"task.context-admitted": true, "graph.writer.host-intent": true, "graph.writer.host-ready": true, "graph.writer.host-observed": true, "graph.writer.runtime-observed": true, "graph.writer.proposed": true, "model.access-intent": true, "model.access-receipt": true}
	for _, event := range events[idx+1:] {
		if !allowed[event.Kind] {
			return fmt.Errorf("writer cohort sibling event %q is not attributable", event.Kind)
		}
		switch event.Kind {
		case "task.context-admitted":
			var record TaskContextRecord
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			matched := false
			for taskID := range cohort {
				question, err := graphWriterTaskQuestion(bound, taskID)
				if err == nil && record.Role == "writer" && record.Query == question {
					matched = true
					break
				}
			}
			if !matched {
				return errors.New("writer cohort task context is not bound")
			}
		case "graph.writer.host-intent", "graph.writer.host-ready":
			var record GraphWriterHostEvent
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			if _, ok := cohort[record.TaskID]; !ok || record.Intent == nil {
				return errors.New("writer host intent outside frozen cohort")
			}
			want, err := expectedWriterHostForTask(bound, record.TaskID)
			if err != nil || want != *record.Intent {
				return errors.Join(errors.New("writer host intent differs from cohort"), err)
			}
		case "graph.writer.host-observed":
			var record GraphWriterHostEvent
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			if _, ok := cohort[record.TaskID]; !ok || record.HostReceipt == nil {
				return errors.New("writer host receipt outside frozen cohort")
			}
			host, ok := current.GraphWriterHosts[record.TaskID]
			wantReceipt, wantErr := canonical.Bytes(host.Receipt)
			gotReceipt, gotErr := canonical.Bytes(record.HostReceipt)
			if !ok || host.Receipt == nil || wantErr != nil || gotErr != nil || string(wantReceipt) != string(gotReceipt) {
				return errors.New("writer host receipt not replayed")
			}
		case "graph.writer.runtime-observed":
			var record GraphWriterHostEvent
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			if _, ok := cohort[record.TaskID]; !ok || record.RuntimeReceipt == nil {
				return errors.New("writer runtime receipt outside frozen cohort")
			}
			host, ok := current.GraphWriterHosts[record.TaskID]
			if !ok || host.RuntimeReceipt == nil || *host.RuntimeReceipt != *record.RuntimeReceipt {
				return errors.New("writer runtime receipt not replayed")
			}
		case "graph.writer.proposed":
			var record GraphWriterRecord
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			task, ok := cohort[record.TaskID]
			if !ok {
				return errors.New("writer proposal outside frozen cohort")
			}
			invocation, err := writerInvocationForTask(bound, task.ID)
			if err != nil || invocation != record.Writer.Invocation {
				return errors.Join(errors.New("writer proposal invocation differs from frozen task"), err)
			}
			if record.Writer.Prepared.Proposal.Before != *bound.Candidate {
				return errors.New("writer proposal candidate differs from frozen cohort")
			}
		case "model.access-intent":
			var record modelAccessIntentEvent
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			taskID, ok := invocations[record.RuntimeInvocationID]
			if !ok {
				return errors.New("model access intent outside frozen writer cohort")
			}
			invocation, err := writerInvocationForTask(bound, taskID)
			if err != nil || invocation.ID != record.RuntimeInvocationID {
				return errors.Join(errors.New("model access intent invocation differs from frozen task"), err)
			}
			want, err := deriveModelAccessIntent(bound, invocation, 1)
			if err != nil || !sameCanonical(record.Intent, want) {
				return errors.Join(errors.New("model access intent differs from frozen task"), err)
			}
		case "model.access-receipt":
			var record ModelAccessTerminal
			if err := canonical.Decode(event.Payload, &record); err != nil {
				return err
			}
			if _, ok := invocations[record.RuntimeInvocationID]; !ok {
				return errors.New("model access receipt outside frozen writer cohort")
			}
			index := modelAccessIndex(current, record.RuntimeInvocationID)
			if index < 0 || current.ModelAccess[index].Terminal == nil || !sameCanonical(*current.ModelAccess[index].Terminal, record) {
				return errors.New("model access receipt differs from replayed cohort receipt")
			}
		}
	}
	return nil
}

func buildGraphWriterBatchSpecs(controllerPath string, s Snapshot, tasks []engineeringplan.Task) ([]taskscheduler.TaskSpec, error) {
	if len(tasks) < 1 || len(tasks) > 2 {
		return nil, errors.New("writer cohort must contain one or two implementations")
	}
	specs := make([]taskscheduler.TaskSpec, 0, len(tasks))
	for _, task := range tasks {
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

func runGraphWriterBatch(ctx context.Context, controllerPath string, initial Snapshot, tasks []engineeringplan.Task) error {
	if initial.Creation.Execution == nil || !parallelImplementationEnabled(initial) || initial.State != "IMPLEMENTING" {
		return errors.New("parallel graph writer batch is not enabled here")
	}
	// Freeze one exact context query per implementation before deriving the
	// invocation-bound schedule. Each task keeps its own recorded query/result.
	for _, task := range tasks {
		question, err := graphWriterTaskQuestion(initial, task.ID)
		if err != nil {
			return err
		}
		if err := maybeAdmitTaskContext(ctx, controllerPath, "writer", question); err != nil {
			return err
		}
	}
	initial, err := Inspect(controllerPath)
	if err != nil {
		return err
	}
	specs, err := buildGraphWriterBatchSpecs(controllerPath, initial, tasks)
	if err != nil {
		return err
	}
	workers := initial.Creation.Execution.EffectiveMaxParallel()
	if workers > len(specs) {
		workers = len(specs)
	}
	if workers > 2 {
		workers = 2
	}
	if workers < 1 {
		workers = 1
	}
	id, err := graphCohortID(initial.Graph.Digest, initial.Graph.Revision, specs)
	if err != nil {
		return err
	}
	schedulePath := controllerPath + ".graph-writers-" + id[:16] + ".jsonl"
	definition := taskscheduler.Definition{Version: 1, Nonce: "graph-writers-" + id, Tasks: specs}
	if _, err := taskscheduler.Bind(schedulePath, definition); err != nil {
		return err
	}
	pumpCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pumpResult := make(chan error, 1)
	go func() {
		pumpResult <- taskscheduler.Pump(pumpCtx, schedulePath, ScheduledDispatchAdapter{JournalPath: schedulePath}, taskscheduler.PumpOptions{Workers: workers, PollInterval: 10 * time.Millisecond})
	}()
	deadline := time.Now().Add(5 * time.Minute)
	for {
		schedule, err := taskscheduler.Inspect(schedulePath)
		if err != nil {
			cancel()
			return errors.Join(err, waitGraphPump(pumpResult, 5*time.Second))
		}
		ctrl, err := Inspect(controllerPath)
		if err != nil {
			cancel()
			return errors.Join(err, waitGraphPump(pumpResult, 5*time.Second))
		}
		done, failed := true, ""
		for _, spec := range specs {
			state, ok := schedule.Tasks[spec.ID]
			if !ok {
				done = false
				break
			}
			switch state.Status {
			case taskscheduler.StatusSucceeded:
				record, found := ctrl.GraphWriterResults[spec.ID]
				if !found || record.Writer.Invocation.ID != spec.InvocationID || state.Evidence == nil || state.Evidence.InvocationID != spec.InvocationID {
					done = false
				}
			case taskscheduler.StatusFailed, taskscheduler.StatusCancelled:
				failed = spec.ID
			default:
				done = false
			}
			if failed != "" {
				break
			}
		}
		if failed != "" {
			cancel()
			return errors.Join(fmt.Errorf("writer batch task %q failed", failed), waitGraphPump(pumpResult, 5*time.Second))
		}
		if done {
			cancel()
			break
		}
		if time.Now().After(deadline) {
			cancel()
			return errors.Join(errors.New("writer batch timed out"), waitGraphPump(pumpResult, 5*time.Second))
		}
		select {
		case <-ctx.Done():
			cancel()
			return errors.Join(ctx.Err(), waitGraphPump(pumpResult, 5*time.Second))
		case <-time.After(20 * time.Millisecond):
		}
		select {
		case err := <-pumpResult:
			if err != nil {
				return err
			}
			return errors.New("writer pump exited before terminal evidence")
		default:
		}
	}
	if err := waitGraphPump(pumpResult, 5*time.Second); err != nil {
		return err
	}
	return nil
}
