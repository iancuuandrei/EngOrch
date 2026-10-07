package cli

import (
	"errors"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/control"
	"harness.local/engorch/internal/runtime"
	"os"
)

type autonomousCapabilities struct {
	dynamicExplorers                                     bool
	agentContext                                         bool
	workingContext                                       bool
	parallel                                             bool
	isolation                                            *isolatedWriterPolicyFile
	plannerContext, parser, parserHash                   string
	parseCache, plannerPPR, reviewImpact, candidateCache int
	autoCompact                                          int64
	evidence                                             *control.EvidenceAutoPolicy
	contextSelector                                      string
	fallbacks                                            []control.CapabilityFallback
}

// Resolve preferences only before run creation. Invalid policies, mismatched
// binary hashes, unsafe state paths and all existing-run inputs stay strict.
func (o *autonomousCapabilities) resolve(cfg config.Config) error {
	if (o.workingContext || o.dynamicExplorers) && (cfg.Version != 2 || cfg.Explorer == nil || cfg.Explorer.Runtime != "codex-app-server" && cfg.Explorer.Runtime != "fake") {
		return errors.New("working-context requires configuration v2 with a Codex or fixture explorer")
	}
	add := func(capability, reason, selected string) {
		o.fallbacks = append(o.fallbacks, control.CapabilityFallback{Capability: capability, Disposition: control.GateFallback, Reason: reason, Selected: selected})
	}
	if o.parser != "" {
		if _, err := os.Stat(o.parser); errors.Is(err, os.ErrNotExist) {
			add("planner_context", "parser_unavailable", autonomousPlannerContextSourceBoundedV1)
			o.plannerContext, o.parser, o.parserHash = autonomousPlannerContextSourceBoundedV1, "", ""
			o.parseCache, o.plannerPPR, o.reviewImpact, o.candidateCache = 0, 0, 0, 0
		} else if err != nil {
			return err
		}
	}
	if o.parallel || o.isolation != nil {
		reason := ""
		if cfg.Writer != nil && cfg.Writer.Runtime != "codex-app-server" && cfg.Writer.Runtime != "fake" {
			reason = "runtime_unsupported"
		}
		if o.isolation != nil {
			if cfg.ControllerStateRoot == "" {
				reason = "external_state_unavailable"
			}
			c, e := o.isolation.Capacity, o.isolation.Estimate
			if e.CPUMilli > c.CPUMilli || e.MemoryMiB > c.MemoryMiB || e.VerificationSlots > c.VerificationSlots || e.RuntimeSlots > c.TotalRuntimeSlots || e.RuntimeSlots > c.RuntimeSlots {
				reason = "capacity_insufficient"
			}
		}
		if reason != "" {
			add("parallel_writers", reason, "serial_writer")
			o.parallel, o.isolation = false, nil
		}
	}
	if o.autoCompact != 0 {
		supportsCompaction := cfg.Planner.Runtime == "codex-app-server"
		for _, p := range []*runtime.Profile{cfg.Writer, cfg.Fixer, cfg.Explorer, cfg.Reviewer} {
			if p != nil && p.Runtime != "codex-app-server" {
				supportsCompaction = false
			}
		}
		if !supportsCompaction {
			add("auto_compaction", "runtime_unsupported", "disabled")
			o.autoCompact = 0
		}
	}
	return nil
}
