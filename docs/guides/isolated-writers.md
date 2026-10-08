# Resource-bounded isolated writers

Autonomous runs can opt into separate initial writer worktrees with
`--isolated-writers --isolation-policy PATH --max-parallel N`. This mode is mutually exclusive
with `--parallel-writers`; it records `isolated_implementation_version: 1` and
the exact validated capacity and per-writer estimate in immutable run
creation. It requires graph execution, bounded task context, and repair
planning, an explicit writer route, an anchored-edits writer contract, and a
JSON v2 explorer contract. The CLI binds the plan-graph-v7 contract for this
mode.

Configured `opencode-http` writers are admitted for the opt-in isolated
(v1/v2) and staged (v3) cohorts only. Each turn executes against its exact
confirmed child workspace/candidate with per-invocation OpenCode
runtime/gateway journals, immutable policy/scopes, scheduler reservations and
source ownership; the parent candidate is never bound as the tool candidate.
The journal namespace is exactly `<role>.invocation-<invocationID>` (for
example `writer.invocation-<64-hex-ID>`), derived from the
controller-admitted task/invocation binding before any journal read or write,
so two distinct task invocations derive distinct journals even when neither
journal exists yet. Execution, usage verification and resumed observation
resolve the same namespace. Legacy serial, scheduled turn-bound and Codex
paths keep their historical stems byte-for-byte, and old journals are never
renamed. Codex behavior and serialization are unchanged, and
shared-workspace parallel writers remain Codex-only: the isolation flag alone
never routes a shared-workspace opencode task through the scheduler.
Dispatch rechecks current host admission immediately before runtime dispatch
and re-confirms the child workspace/candidate binding and parent identity
after dispatch before observing successful receipt; a stale or mutating
binding fails safely with the effect identity retained and no resend. UNKNOWN
stops with no resend, and old active uncertain effects are never probed.

Measured qualification is offline only. The sealed-fixture chain covers a
two-task deterministic-namespace regression (distinct stems before either
journal exists, legacy bare `writer` stem preserved), a true UNKNOWN
dispatch/resume test through the actual wrapper (durable admission, UNKNOWN
observation, one runtime intent, zero gateway calls, frozen counts on legal
resume), and full offline `RunWriter` chains for an isolated task and a
staged hub task (dispatch to exact runtime/provider receipt to proposal
record/replay to inspectable usage validation with genuine sealed journals).
Scope, swapped-receipt, legacy Codex and shared-fallback regressions remain
strict. Full staged READY with a real Muse route requires live provider
qualification; offline fixtures do not claim live acceptance, latency, cost,
or per-invocation performance.

Before creating an isolated run, `harness.toml` must set `controller_state_root`
to an absolute external directory separate from the repository and its Git
control paths. The CLI validates it with the repository-bound controller-state
resolver before writing a run journal or dispatching the planner. This
namespace owns durable child-worktree state; it is not inferred from the task
checkout.

Task-bound writer questions reuse the pre-existing 256 KiB full-query admission bound; selection text is capped at 16 KiB of UTF-8 with the full digest and length retained, the total invocation bound stays 256 KiB, and the full mandatory objective is retained separately with scopes and budgets immutable. Historical short (≤4096-byte) question bytes and identifiers are unchanged. The staged long-objective check is a focused offline shared-stage v8 regression only; the original fresh v9 run at 18cd118 stopped pre-writer with one completed planner call and makes no live accepted-writer claim.

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

### Staged lexicographic wave optimum (`lexicographic-v1`)

`--isolated-writer-staged --isolation-policy PATH --cohort-selector lexicographic-v1 --max-parallel N`
opts a staged run into the exact finite wave optimum over the same hard
gates as the default greedy derivation. It freezes
`isolation_cohort_selector_version: 1` in immutable run creation; absent
preserves the greedy derivation byte-for-byte. Each wave maximizes admitted
task count, then summed downstream critical-path length (declared
`EstimatedSeconds` estimates, not measurements), then minimizes summed
estimated CPU, memory, verification, and runtime slots, with sorted task-ID
tie-breaks. These objectives do not establish measured wall-clock makespan,
and coupling stays a hard independence gate rather than an optimized score.
The version 3 preparation records `cohort_selector_version: 1` (omitted for
greedy runs so legacy journals replay unchanged); journal replay recomputes
with the frozen selector and rejects a substituted value. The flag is
rejected for serial, parallel, version 1, and version 2 runs, and capability
fallback cannot silently drop it.

Limitation stated honestly: the frozen per-writer estimate template applies
identical demands to every ready implementation, so on current runs the
greedy and lexicographic wave derivations agree. Wave-order divergence is
demonstrated at the selector unit level with unequal declared demands, not
with invented per-writer estimates on a live run. No measured quality,
latency, or cost benefit is claimed.

### Staged coupling-aware wave optimum (`coupling-aware-v1`)

`--isolated-writer-staged --isolation-policy PATH --cohort-selector coupling-aware-v1 --max-parallel N`
opts a staged run into the typed coupling-aware optimum under
`plan-graph-v9`. It freezes `isolation_cohort_selector_version: 2` in
immutable run creation; absent preserves the greedy derivation byte-for-byte
and version 1 keeps its frozen count-first order unchanged. Hard gates match
the frozen selectors plus the C4 hard-coupling gate: exact ready identity,
pairwise dependency/write-overlap independence, every explicit resource
ceiling, the 1..8 bound, and no C4 pair co-selected in one wave. Objectives
minimize C3 concurrency risk before admitted count, then admitted count, then
C2, then C1 co-scheduling, then summed downstream critical-path length
(declared `EstimatedSeconds`, not a measurement), then summed estimated CPU,
memory, verification, and runtime slots, with sorted task-ID tie-breaks.
A remainder that fits capacity but was excluded to avoid a C3/C2/C1 pair
renders the explicit bounded reason `coupling_risk` with the smallest coupled
partner; anything else that fits still fails closed.

Couplings are planner-declared advisory only (`planner_declared_advisory`).
Observed generator/topology labels are rejected as forged until a future
source-bound admission validates the exact source/candidate/path binding
against an existing admitted generator or topology record; that broader Phase
E admission is pending. Graph, source, candidate, and journal hashing binds
the declared risk inputs; it does not prove a cited fact. C4 requires one
owner or an explicit implementation dependency and rejects an unsafe split
before writer effects; serial waves forked from the same parent base do not
make it safe. C3/C2/C1 never grant readiness, ownership, or write authority,
and absent coupling (C0) never proves independence. Different files do not
prove independence, and partial evidence never proves absence.

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

## Writer cohort lifetime

Each scheduled writer cohort and each isolated or staged wave has its own
finite lifetime derived from existing frozen limits. For `opencode-http`
writer cohorts under isolated or staged isolation, the bound is the frozen
per-invocation runtime already enforced by `openCodeProviderRuntimeContext`
(the 15-minute default or the validated host `InvocationTimeoutSeconds`
1..7200) multiplied by the conservative serial worst case: every task's three
legal invocations (one initial plus at most two admitted semantic
corrections), executed one after another. The bound deliberately ignores the
advertised scheduler worker count: the memory admission gate parks future
claims once active grants reach `Decision.EffectiveWorkers`, and that
decision can fall to one worker under pressure while scheduler slots remain,
so guaranteed parallelism cannot be assumed and legal work must not be
cancelled for exceeding a parallelism-discounted wait. Claim and resource
contention park the same way. A 2-task leaves wave with a 600-second host
therefore waits 60 minutes, not five. Indefinite parking under sustained
pressure may still hit this bounded deadline; the bound covers only the
finite admitted serial work and promises nothing under infinite pressure.
Non-OpenCode cohorts (Codex, fake, serial legacy) keep the historical
five-minute wait because no comparable frozen per-invocation total exists for
those runtimes. The cohort loop reads the authoritative controller state
before starting any pump and fails closed on a read error with no silent
fallback limit. An earlier caller deadline or cancellation still wins,
`UNKNOWN` remains `UNKNOWN` with no resend, and retry, admission, role, and
budget semantics are unchanged. The unrelated five-minute research-cohort
wait in `autonomous_graph.go` is unchanged. Offline deterministic fixtures
cover the configured-versus-legacy mismatch, worker-independent serial
bounding, the configured-pair serializing to one effective worker under
pressure, corrections, fail-closed reads, and caller cancellation on the
cohort path; they do not claim live provider acceptance, latency, or cost. The motivating independent trial at
source `72a1343` configured a 600-second invocation timeout for a staged hub
plus two-leaf cohort at `--max-parallel 2`; the process exited `1` at
352.901 seconds with hub and sink proposals complete and one leaf writer
pending `UNKNOWN`, the public boundary canceled at controller sequence 39.
That pending effect was not retried and remains `UNKNOWN`; no live success is
claimed for the derived bound.

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

## Run listing and scheduler sidecars

Default `status`, `inspect` without `RUN`, and latest-run selection ignore
only the exact controller-produced scheduler journals in the runs directory:
`graph-schedule-<16-lower-hex>.jsonl`,
`graph-writers-<16-lower-hex>.jsonl`,
`isolated-graph-writers-<16-lower-hex>.jsonl`,
`isolated-graph-writers-wave-<16-lower-hex>.jsonl`, and
`staged-graph-writers-wave-<16-lower-hex>.jsonl` after `<run>.jsonl.`.
Any other spelling, including uppercase, wrong cohort length, an extra
suffix, or an unknown prefix, stays fail-closed and surfaces as a corrupt
or unknown journal instead of being skipped. Provider, model-access, and
runtime sidecar handling is unchanged.
