package cli

import (
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
	// Even advisory acquisition recommendations must not route around UNKNOWN.
	if stop := control.ClassifyAutonomousFailure(state, errors.New("evidence advisory inspection")); stop.Disposition() == control.GateReconcile {
		report.Selected = ""
		report.StopReason = stop.PublicReason()
	} else if state.State == "READY" {
		report.Selected = ""
		report.StopReason = "accepted_checkpoint_requires_no_additional_research"
	} else if state.Lifecycle.Status != "" && state.Lifecycle.Status != control.LifecycleActive {
		report.Selected = ""
		report.StopReason = "lifecycle_requires_attention"
	}
}
