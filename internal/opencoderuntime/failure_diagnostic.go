package opencoderuntime

import (
	"encoding/json"
	"errors"
	"os"

	"harness.local/engorch/internal/opencode"
)

// A sidecar preserves explanatory labels after the process exits. It is never
// read by replay, receipt admission, recovery, or provider dispatch.
func preserveDispatchDiagnostic(cfg ExecuteConfig, cause error) error {
	diagnostic := opencode.DispatchDiagnosticFromError(cause)
	if diagnostic == nil {
		return nil
	}
	record := struct {
		Version      int                         `json:"version"`
		InvocationID string                      `json:"invocation_id"`
		Status       string                      `json:"status"`
		Diagnostic   opencode.DispatchDiagnostic `json:"diagnostic"`
	}{1, cfg.Intent.Invocation.ID, "UNKNOWN", *diagnostic}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	// Exclusive creation preserves the first failure and refuses existing links.
	f, err := os.OpenFile(cfg.RuntimePath+".failure.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, writeErr := f.Write(append(data, '\n'))
	return errors.Join(writeErr, f.Sync(), f.Close())
}
