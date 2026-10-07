# Resource-bounded isolated writers

Autonomous runs can opt into separate initial writer worktrees with
`--isolated-writers --isolation-policy PATH --max-parallel N`. This mode is mutually exclusive
with `--parallel-writers`; it records `isolated_implementation_version: 1` and
the exact validated capacity and per-writer estimate in immutable run
creation. It requires graph execution, bounded task context, and repair
planning, an explicit writer route, an anchored-edits writer contract, and a
JSON v2 explorer contract. The CLI binds the plan-graph-v7 contract for this
mode.

Before creating an isolated run, `harness.toml` must set `controller_state_root`
to an absolute external directory separate from the repository and its Git
control paths. The CLI validates it with the repository-bound controller-state
resolver before writing a run journal or dispatching the planner. This
namespace owns durable child-worktree state; it is not inferred from the task
checkout.

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

## Resource-bounded initial writer waves (version 2)

`--isolated-writer-waves --isolation-policy PATH --max-parallel N` opts into
`isolated_implementation_version: 2` with the same v7 planner, capacity,
estimate, graph, repair, route, and external-state requirements as version 1.
It is mutually exclusive with `--parallel-writers` and `--isolated-writers`,
and it is not a default promotion. Version 1 remains unchanged.

Version 2 freezes all initial ready implementations (up to eight) with their
exact demands and deterministically derives nonempty resource-bounded waves
via repeated resource-cohort selection over each remainder. Each wave fits CPU,
memory, verification, global runtime, provider, model, and exact-route
ceilings and `--max-parallel`; hard write or dependency conflicts are rejected
under the existing independent-cohort contract. Child workspaces share the
same pristine parent source, execute serially with the existing
scheduler and memory admission (per-wave workers and admission ceilings, unique
deterministic wave schedules, no redundant provider calls), accumulate fully
validated per-child proposals, and perform a single parent aggregation and file
effect only after all required initial tasks have valid proposals. Graph tasks
complete only after that aggregate effect. Budgets do not reset, and UNKNOWN
stops without advancing, recreating, or retrying.

Dependent hub-to-leaves dirty-candidate isolation is explicitly pending and is
not implemented in version 2; all readiness, ownership, UNKNOWN, and budget
gates are retained. Cohort scope replanning is unsupported for version 2
waves: runs are created without a scope-replan policy, and a scope violation
remains a bounded rejection without silent widening.

## Dependent hub-to-leaf staged cohorts (version 3)

`--isolated-writer-staged --isolation-policy PATH --max-parallel N` opts into
`isolated_implementation_version: 3` with the `plan-graph-v8` planner, capacity,
estimate, graph, repair, route, and external-state requirements of version 2.
It is mutually exclusive with `--parallel-writers`, `--isolated-writers`, and
`--isolated-writer-waves`, and it is not a default promotion. Versions 1 and 2
remain unchanged.

Version 3 freezes each exact current ready implementation subset (hub then
leaves) with exact graph, candidate, HEAD, and resources, partitioning the
current subset into resource-bounded waves when it exceeds capacity. Each leaf
child is forked from its exact post-hub parent candidate through a durable fork
intent before destination creation, copying only the parent-authorized delta
with existing worktree and file-effect primitives; RI, overlay, and lexical
views never authorize the copy. Cumulative fork deltas enforce the existing
64-change, 256 KiB, and canonical-base64 bounds before intent or creation; full
preflight still needs the child binding and runs after intent, so its failure
stays UNKNOWN without redispatch. A streaming copy is a documented alternative
to the current bounded reuse. Each stage performs one parent aggregate and
file effect, advances the CONFIRMED cohort via a compact identity event (exact
cohort, preparation, candidate, task, proposal, and observation hashes,
pre-effect envelope checked) that reconstructs the full archive from validated
current state, and preserves completed graph
evidence, journal history, and repair budgets with stable version 3 memory
ceilings and retained adaptive history. Foreign, stale, manifest, mode,
path, or source drift is rejected; UNKNOWN stops without next dispatch or
redispatch. Intermediate hubs may integrate without final READY; final native
verification and review bind the exact last candidate with no stale receipt.
Cohort scope replanning is unsupported for version 3.

This is an opt-in scheduling and isolation policy. It does not infer host
capacity from environment data, measure actual CPU or memory use, qualify a
provider route, authorize unrestricted operating-system sandboxing, or claim
measured quality, latency, or cost benefit.

The Windows evaluation runner configures one external namespace per task under
that task's evaluation output directory before invoking `fabric run`. The
absolute setting is included in the task configuration hash and checked again
against the inspected run creation record. Ordinary evaluations do not add
this setting.

This is an opt-in scheduling and isolation policy. It does not infer host
capacity from environment data, measure actual CPU or memory use, qualify a
provider route, or authorize unrestricted operating-system sandboxing.

## Fresh-run failure note

A frozen v5 isolated-writer evaluation ended `BLOCKED` before any writer
proposal in both arms. The Humanize arm reached graph/isolation preparation but
had no external `controller_state_root`, so child-worktree admission stopped
before its intent was recorded. The atomic-numeric-text arm's planner graph
omitted required native verification and review gates and was rejected before
graph admission. Both parent workspaces were confirmed; neither journal
contains a child isolation intent or file intent. The runs were not resumed or
retried, and provider-call counts remain unknown. These outcomes are setup and
invalid-proposal evidence, not a performance result for isolated writers.
