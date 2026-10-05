package cli

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"

	"harness.local/engorch/internal/canonical"
	"harness.local/engorch/internal/control"
)

func diagnoseCommand(ctx context.Context, root string, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("diagnose requires RUN")
	}
	flags := flag.NewFlagSet("diagnose", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	anchor := flags.Bool("anchor", false, "validate current candidate bytes")
	previous := flags.String("previous", "", "use prior report anchors as untrusted localization hints")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || (*previous != "" && !*anchor) {
		return errors.New("diagnose expects RUN [--anchor] [--previous REPORT_JSON]")
	}
	path, err := runPath(root, args[0])
	if err != nil {
		return err
	}
	s, err := control.Inspect(path)
	if err != nil {
		return err
	}
	if s.RunID != args[0] || filepath.Clean(s.Creation.Repository.Root) != filepath.Clean(root) {
		return errors.New("journal/run repository binding mismatch")
	}
	if !*anchor {
		r, err := control.DiagnoseRepair(s)
		if err != nil {
			return err
		}
		return output(out, r)
	}
	hints, err := previousRepairAnchors(*previous, s.RunID)
	if err != nil {
		return err
	}
	r, err := control.DiagnoseRepairAnchored(ctx, path, s, hints)
	if err != nil {
		return err
	}
	return output(out, r)
}

func previousRepairAnchors(path, runID string) ([]control.RepairAnchor, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("previous repair report unavailable")
	}
	defer f.Close()
	const limit = 1 << 20
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || len(raw) > limit {
		return nil, errors.New("previous repair report outside read bounds")
	}
	var r control.RepairDiagnosis
	if err := canonical.Decode(raw, &r); err != nil {
		return nil, errors.New("invalid previous repair report")
	}
	if r.Version != 1 || r.RunID != runID || len(r.Anchors) > 128 {
		return nil, errors.New("previous repair report binding mismatch")
	}
	for _, a := range r.Anchors {
		if a.CandidateID != r.CandidateID {
			return nil, errors.New("previous repair anchor candidate mismatch")
		}
	}
	return r.Anchors, nil
}
