// Package memoryadmission turns bounded host-memory observations into
// conservative limits for future worker claims.
package memoryadmission

import (
	"errors"
	"fmt"
)

// Observation statuses, decision modes, supported sources and bounded policy
// defaults use MiB for memory values. Defaults are estimates, not measured usage.
const (
	ObservationObserved    = "OBSERVED"
	ObservationUnavailable = "UNAVAILABLE"
	ModeNormal             = "normal"
	ModePressured          = "pressured"
	ModeUnknown            = "unknown"
	SourceWindowsGlobal    = "windows_global_memory_status_ex"
	SourceLinuxProcMeminfo = "linux_proc_meminfo"
	MaxObservedMemoryMiB   = int64(1 << 30)
	DefaultReserveMiB      = int64(1024)
	DefaultHysteresisMiB   = int64(512)
)

// Observation contains host-reported available physical memory. It is not a
// measurement of the memory a worker will consume.
type Observation struct {
	Status       string `json:"status"`
	Source       string `json:"source,omitempty"`
	AvailableMiB int64  `json:"available_mib,omitempty"`
	TotalMiB     int64  `json:"total_mib,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// Decision records the future-claim ceiling derived from an observation and
// the previous durable decision.
type Decision struct {
	Mode                  string `json:"mode"`
	EffectiveWorkers      int    `json:"effective_workers"`
	ConfiguredMaxWorkers  int    `json:"configured_max_workers"`
	EstimatedPerWorkerMiB int64  `json:"estimated_per_worker_mib"`
	ReserveMiB            int64  `json:"reserve_mib"`
	HysteresisMiB         int64  `json:"hysteresis_mib"`
}

// Observe reads available physical memory through a native OS interface. A
// failed or unsupported observation is returned as UNAVAILABLE, never as zero.
func Observe() Observation {
	return observe()
}

// ValidateObservation checks the bounded public shape before an observation is
// placed in a durable admission record.
func ValidateObservation(observation Observation) error {
	switch observation.Status {
	case ObservationObserved:
		if (observation.Source != SourceWindowsGlobal && observation.Source != SourceLinuxProcMeminfo) || observation.AvailableMiB < 0 || observation.AvailableMiB > MaxObservedMemoryMiB || observation.TotalMiB < 1 || observation.TotalMiB > MaxObservedMemoryMiB || observation.AvailableMiB > observation.TotalMiB || observation.Reason != "" {
			return errors.New("invalid observed memory facts")
		}
	case ObservationUnavailable:
		if observation.Source != "" || observation.AvailableMiB != 0 || observation.TotalMiB != 0 || observation.Reason != "unavailable" {
			return errors.New("invalid unavailable memory observation")
		}
	default:
		return errors.New("invalid memory observation status")
	}
	return nil
}

// Next chooses a worker ceiling. Lower limits take effect immediately. A
// previously pressured or unknown state returns to normal only above the
// configured exit threshold, which avoids oscillating around the entry line.
func Next(previous *Decision, observation Observation, configuredMax int, estimatedPerWorkerMiB, reserveMiB, hysteresisMiB int64) (Decision, error) {
	if err := ValidateObservation(observation); err != nil {
		return Decision{}, err
	}
	if configuredMax < 1 || configuredMax > 8 || estimatedPerWorkerMiB < 1 || estimatedPerWorkerMiB > MaxObservedMemoryMiB || reserveMiB < 0 || reserveMiB > MaxObservedMemoryMiB || hysteresisMiB < 0 || hysteresisMiB > MaxObservedMemoryMiB {
		return Decision{}, errors.New("invalid memory admission bounds")
	}
	decision := Decision{
		Mode: ModeNormal, EffectiveWorkers: configuredMax,
		ConfiguredMaxWorkers: configuredMax, EstimatedPerWorkerMiB: estimatedPerWorkerMiB,
		ReserveMiB: reserveMiB, HysteresisMiB: hysteresisMiB,
	}
	if previous != nil {
		if err := validateDecision(*previous); err != nil {
			return Decision{}, fmt.Errorf("invalid previous memory decision: %w", err)
		}
		if previous.ConfiguredMaxWorkers != configuredMax || previous.EstimatedPerWorkerMiB != estimatedPerWorkerMiB || previous.ReserveMiB != reserveMiB || previous.HysteresisMiB != hysteresisMiB {
			return Decision{}, errors.New("memory admission policy changed within one run")
		}
	}
	if observation.Status == ObservationUnavailable {
		decision.Mode = ModeUnknown
		decision.EffectiveWorkers = 1
		return decision, nil
	}

	availableForWorkers := observation.AvailableMiB - reserveMiB
	safeWorkers := 1
	if availableForWorkers > 0 {
		safeWorkers = int(availableForWorkers / estimatedPerWorkerMiB)
		if safeWorkers < 1 {
			safeWorkers = 1
		}
		if safeWorkers > configuredMax {
			safeWorkers = configuredMax
		}
	}
	if previous != nil {
		if previous.Mode == ModePressured || previous.Mode == ModeUnknown {
			exitThreshold := reserveMiB + estimatedPerWorkerMiB*int64(configuredMax) + hysteresisMiB
			if observation.AvailableMiB < exitThreshold {
				decision.Mode = ModePressured
				decision.EffectiveWorkers = previous.EffectiveWorkers
				if safeWorkers < decision.EffectiveWorkers {
					decision.EffectiveWorkers = safeWorkers
				}
				return decision, nil
			}
		}
	}
	if safeWorkers < configuredMax {
		decision.Mode = ModePressured
		decision.EffectiveWorkers = safeWorkers
	}
	return decision, nil
}

func validateDecision(decision Decision) error {
	if decision.ConfiguredMaxWorkers < 1 || decision.ConfiguredMaxWorkers > 8 || decision.EffectiveWorkers < 1 || decision.EffectiveWorkers > decision.ConfiguredMaxWorkers || decision.EstimatedPerWorkerMiB < 1 || decision.EstimatedPerWorkerMiB > MaxObservedMemoryMiB || decision.ReserveMiB < 0 || decision.ReserveMiB > MaxObservedMemoryMiB || decision.HysteresisMiB < 0 || decision.HysteresisMiB > MaxObservedMemoryMiB {
		return errors.New("invalid memory decision limits")
	}
	switch decision.Mode {
	case ModeNormal:
		if decision.EffectiveWorkers != decision.ConfiguredMaxWorkers {
			return errors.New("normal memory decision must retain configured capacity")
		}
	case ModePressured, ModeUnknown:
	default:
		return errors.New("invalid memory decision mode")
	}
	if decision.Mode == ModeUnknown && decision.EffectiveWorkers != 1 {
		return errors.New("unknown memory decision must be serial")
	}
	return nil
}
