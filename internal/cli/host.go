package cli

import (
	"context"
	"io"

	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/control"
	"harness.local/engorch/internal/hostenvironment"
	"harness.local/engorch/internal/memoryadmission"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/ri"
	"harness.local/engorch/internal/runtime"
	"harness.local/engorch/internal/verification"
)

type hostObserver func(context.Context) (hostenvironment.Observation, error)

func bindCurrentHost(ctx context.Context, creation control.Creation) (control.Creation, error) {
	observation, err := hostenvironment.ObserveDefault(ctx)
	if err != nil {
		return control.Creation{}, err
	}
	policy := control.DefaultHostPolicy()
	if creation.Config.HostPolicy != nil {
		policy = *creation.Config.HostPolicy
	}
	return control.BindHostAdmission(creation, observation, policy)
}

func writeDoctor(ctx context.Context, out io.Writer, identity repository.Identity, policy *hostenvironment.Policy, configurations ...config.Config) error {
	return writeDoctorWithObserver(ctx, out, identity, hostenvironment.ObserveDefault, policy, configurations...)
}

func writeDoctorWithObserver(ctx context.Context, out io.Writer, identity repository.Identity, observe hostObserver, policy *hostenvironment.Policy, configurations ...config.Config) error {
	host, err := observe(ctx)
	if err != nil {
		return err
	}
	selected := control.DefaultHostPolicy()
	if policy != nil {
		selected = *policy
	}
	if err := hostenvironment.Admit(selected, host); err != nil {
		return err
	}
	report := map[string]any{
		"status":           "PASS",
		"repository":       identity,
		"host_environment": host,
		"runtime_dispatch": "NOT_RUN",
		"verification":     "NOT_RUN",
	}
	if len(configurations) != 0 {
		plan, err := doctorExecutionPlan(identity, configurations[0])
		if err != nil {
			return err
		}
		report["execution_plan"] = plan
	}
	return output(out, report)
}

func doctorExecutionPlan(identity repository.Identity, cfg config.Config) (map[string]any, error) {
	id, err := identity.ID()
	if err != nil {
		return nil, err
	}
	roles := map[string]*runtime.Profile{"planner": &cfg.Planner, "writer": cfg.Writer, "fixer": cfg.Fixer, "explorer": cfg.Explorer, "reviewer": cfg.Reviewer}
	compactSupported := true
	for _, profile := range roles {
		if profile != nil && profile.Runtime != "codex-app-server" {
			compactSupported = false
		}
	}
	parallelSupported := cfg.Writer != nil && (cfg.Writer.Runtime == "codex-app-server" || cfg.Writer.Runtime == "fake")
	resources := doctorMemoryResources(memoryadmission.Observe())
	resources["shared_task_pool"] = "NOT_CONFIGURED"
	if cfg.TaskPool != nil {
		resources["shared_task_pool"] = "CONFIGURED_NOT_ADMITTED"
		resources["limits"] = cfg.TaskPool.Limits
	}
	checks := []map[string]string{}
	for _, check := range cfg.Verification {
		invocation, err := verification.Prepare(id, identity.Root, check)
		if err != nil {
			return nil, err
		}
		readiness := "AVAILABLE_NOT_RUN"
		if invocation.Executable == nil {
			readiness = "UNAVAILABLE"
		}
		checks = append(checks, map[string]string{"name": check.Name, "readiness": readiness})
	}
	return map[string]any{"roles": roles, "verification_checks": checks,
		"scope_replan": map[string]any{"configured_roles_available": cfg.Writer != nil && cfg.Explorer != nil, "default_topology": "serial_writer", "serial_policy_version": 1, "cohort_policy_version": 2, "max_replans": 2, "runtime_execution": "NOT_RUN"},
		"topology":     "serial_default", "context": "source_bounded_default",
		"run_options": "NOT_SELECTED", "provider_qualification": "NOT_RUN",
		"runtime_support": map[string]bool{"parallel_writers": parallelSupported, "auto_compaction": compactSupported},
		"resources":       resources, "cache": "RUN_OPTIONS_NOT_SELECTED"}, nil
}

func doctorMemoryResources(observation memoryadmission.Observation) map[string]any {
	return map[string]any{
		"live_memory_pressure": observation.Status,
		"memory_observation":   observation,
		"memory_admission":     "NOT_ADMITTED",
		"worker_memory_usage":  "NOT_MEASURED",
	}
}

func inspectAutonomousPlan(ctx context.Context, root string, options autonomousCapabilities, maxParallel int, out io.Writer) error {
	cfg, err := configuration(root)
	if err != nil {
		return err
	}
	if err := options.resolve(cfg); err != nil {
		return err
	}
	if options.parser != "" {
		parser := ri.Client{Executable: options.parser, ExecutableHash: options.parserHash}
		if err := parser.ValidateExecutable(); err != nil {
			return err
		}
	}
	identity, err := repository.Discover(ctx, root, cfg.Repository)
	if err != nil {
		return err
	}
	plan, err := doctorExecutionPlan(identity, cfg)
	if err != nil {
		return err
	}
	contextMode := options.plannerContext
	if contextMode == "" {
		contextMode = autonomousPlannerContextSourceBoundedV1
	}
	topology := "serial_writer"
	if options.parallel {
		topology = "parallel_writers"
	}
	if options.isolation != nil {
		topology = "isolated_writers"
		plan["isolation_capacity"] = options.isolation.Capacity
		plan["writer_estimate"] = options.isolation.Estimate
		resources := plan["resources"].(map[string]any)
		observation := resources["memory_observation"].(memoryadmission.Observation)
		decision, decisionErr := memoryadmission.Next(nil, observation, maxParallel, options.isolation.Estimate.MemoryMiB, memoryadmission.DefaultReserveMiB, memoryadmission.DefaultHysteresisMiB)
		if decisionErr != nil {
			return decisionErr
		}
		resources["memory_admission_preview"] = decision
		resources["memory_admission_preview_status"] = "ESTIMATED_NOT_ADMITTED"
		resources["memory_admission_recheck"] = "BEFORE_FUTURE_ISOLATED_WRITER_CLAIMS"
	}
	plan["context"] = contextMode
	plan["topology"] = topology
	scopeVersion := 1
	if topology != "serial_writer" {
		scopeVersion = 2
	}
	plan["scope_replan"] = map[string]any{"enabled": cfg.Writer != nil && cfg.Explorer != nil, "policy_version": scopeVersion, "max_replans": 2, "ownership_ceiling": "IMMUTABLE_SCOPE_PATHS", "design": "READ_ONLY_MODEL_DECISION", "runtime_execution": "NOT_RUN"}
	plan["max_parallel"] = maxParallel
	plan["run_options"] = "SELECTED_NOT_DISPATCHED"
	plan["cache"] = map[string]int{"parser_facts": options.parseCache, "candidate_facts": options.candidateCache}
	plan["auto_compact_token_limit"] = options.autoCompact
	plan["fallbacks"] = options.fallbacks
	return output(out, map[string]any{"execution_plan": plan, "runtime_dispatch": "NOT_RUN", "verification": "NOT_RUN"})
}
