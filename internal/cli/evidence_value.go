package cli

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/control"
	"harness.local/engorch/internal/evidencevalue"
)

func evidenceValueCommand(root string, args []string, out io.Writer) error {
	if len(args) != 2 {
		return errors.New("evidence-value requires RUN REQUEST_JSON")
	}
	path, err := runPath(root, args[0])
	if err != nil {
		return err
	}
	state, head, err := control.InspectWithHead(path)
	if err != nil {
		return err
	}
	if state.RunID != args[0] || filepath.Clean(state.Creation.Repository.Root) != filepath.Clean(root) {
		return errors.New("journal/run repository binding mismatch")
	}
	source, err := state.Creation.Repository.ID()
	if err != nil {
		return err
	}
	candidate := ""
	if state.Candidate != nil {
		candidate, err = state.Candidate.ID()
		if err != nil {
			return err
		}
	}
	binding := evidencevalue.Binding{RunID: state.RunID, SourceID: source, CandidateID: candidate, JournalHead: head}
	if args[1] == "--template" {
		return output(out, evidencevalue.EncodeRequest(evidencevalue.Request{Version: 1, Binding: binding}))
	}
	raw, err := readRegularPolicyJSON(args[1], evidencevalue.MaxBytes, "evidence value request", "64 KiB")
	if err != nil {
		return err
	}
	var wire evidencevalue.WireRequest
	if canonical.Decode(raw, &wire) != nil {
		return errors.New("invalid evidence value request")
	}
	request, err := wire.Decode()
	if err != nil {
		return err
	}
	if request.Binding != binding {
		return errors.New("evidence request binding is stale or substituted")
	}
	report, err := evidencevalue.Evaluate(request)
	if err != nil {
		return err
	}
	applyEvidenceControllerStop(state, &report)
	result := report.Wire()
	result.RequestHash, err = canonical.Hash("harness.evidence-value-request.v1", wire)
	if err != nil {
		return err
	}
	return output(out, result)
}

func applyEvidenceControllerStop(state control.Snapshot, report *evidencevalue.Report) {
	control.ApplyEvidenceStop(state, report)
}

func evidenceAcquireCommand(ctx context.Context, root string, args []string, out io.Writer) error {
	if len(args) != 2 {
		return errors.New("evidence-acquire requires RUN REQUEST_JSON")
	}
	path, err := runPath(root, args[0])
	if err != nil {
		return err
	}
	state, err := control.Inspect(path)
	if err != nil {
		return err
	}
	if state.RunID != args[0] || filepath.Clean(state.Creation.Repository.Root) != filepath.Clean(root) {
		return errors.New("journal/run repository binding mismatch")
	}
	input := args[1]
	if !filepath.IsAbs(input) {
		input = filepath.Join(root, input)
	}
	raw, err := readRegularPolicyJSON(input, evidencevalue.MaxBytes, "evidence acquisition request", "64 KiB")
	if err != nil {
		return err
	}
	var request control.EvidenceContextRequest
	if canonical.Decode(raw, &request) != nil {
		return errors.New("invalid evidence acquisition request")
	}
	result, err := control.AcquireEvidenceContext(ctx, path, request)
	if err != nil {
		return err
	}
	return output(out, result)
}
