package control

import "errors"

// ErrScopeReplanRequired preserves the proposal without widening ownership.
// A new, candidate-bound scope decision is required before any file effect.
var ErrScopeReplanRequired = errors.New("proposal requires an authorized scope replan")

// GateDisposition describes a product decision, not permission to bypass an
// identity, ownership, journal, verification or uncertain-effect gate.
type GateDisposition string

const (
	// GatePass records that the requested gate was satisfied.
	GatePass GateDisposition = "PASS"
	// GateDeny prohibits continuation across an authority boundary.
	GateDeny GateDisposition = "DENY"
	// GateFallback selects an admitted baseline capability.
	GateFallback GateDisposition = "FALLBACK"
	// GateEscalate requests an explicit operator or scope decision.
	GateEscalate GateDisposition = "ESCALATE"
	// GateWarn reports an advisory limitation without granting authority.
	GateWarn GateDisposition = "WARN"
)

// CapabilityFallback binds an optional capability's absence to its selection.
type CapabilityFallback struct {
	Capability  string          `json:"capability"`
	Disposition GateDisposition `json:"disposition"`
	Reason      string          `json:"reason"`
	Selected    string          `json:"selected"`
}

// Validate permits only closed, documented capability fallback combinations.
func (f CapabilityFallback) Validate() error {
	if f.Disposition != GateFallback {
		return errors.New("invalid capability disposition")
	}
	switch f.Capability {
	case "planner_context":
		if f.Selected == "source-bounded-v1" && f.Reason == "parser_unavailable" {
			return nil
		}
	case "parallel_writers":
		if f.Selected == "serial_writer" && (f.Reason == "runtime_unsupported" || f.Reason == "external_state_unavailable" || f.Reason == "capacity_insufficient") {
			return nil
		}
	case "auto_compaction":
		if f.Selected == "disabled" && f.Reason == "runtime_unsupported" {
			return nil
		}
	}
	return errors.New("unsupported capability fallback")
}
