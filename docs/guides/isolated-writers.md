# Resource-bounded isolated writers

Autonomous runs can opt into separate initial writer worktrees with
`--isolated-writers --isolation-policy PATH --max-parallel N`. This mode is mutually exclusive
with `--parallel-writers`; it records `isolated_implementation_version: 1` and
the exact validated capacity and per-writer estimate in immutable run
creation. It requires graph execution, bounded task context, and repair
planning, an explicit writer route, an anchored-edits writer contract, and a
JSON v2 explorer contract. The CLI binds the plan-graph-v7 contract for this
mode.

The policy file is strict JSON, version 1, and limited to 32 KiB. It must
contain every field shown below; duplicate, unknown, missing, or null fields
are rejected. Capacity CPU and memory are in millicores and MiB. Slot values
are positive integers from 1 to 64. Estimate CPU and memory are positive and
bounded; runtime slots are 1 to 64; verification slots are an explicit value
from 0 to 64. These estimates are declared scheduler inputs, not observed
resource measurements.

```json
{
  "version": 1,
  "capacity": {
    "cpu_milli": 4000,
    "memory_mib": 8192,
    "verification_slots": 1,
    "total_runtime_slots": 4,
    "provider_slots": 2,
    "model_slots": 2,
    "runtime_slots": 2
  },
  "estimate": {
    "cpu_milli": 1000,
    "memory_mib": 2048,
    "verification_slots": 0,
    "runtime_slots": 1
  }
}
```

The provider, model, and exact runtime ceilings apply to the configured writer
route. The CLI derives those keys, including the canonical writer profile ID,
from the run's bound configuration; the policy file cannot name or substitute a
different route. The estimates are applied uniformly to each initially ready
implementation task. `--max-parallel` remains the maximum cohort size.

The isolated mode uses the v7 planner contract, which permits up to the
configured `--max-parallel` number of initial implementation tasks (capped by
the controller's eight-task bound). Those implementations must have disjoint
write ownership and shared read-only prerequisites. Every initially ready
implementation task must fit the declared resource capacity; an incomplete fit
blocks admission rather than silently changing the graph's task set. Later
verification is not charged as writer runtime. Separate worktrees isolate the
writers' proposals, while parent-candidate integration remains a distinct
controller step.

This is an opt-in scheduling and isolation policy. It does not infer host
capacity from environment data, measure actual CPU or memory use, qualify a
provider route, or authorize unrestricted operating-system sandboxing.
