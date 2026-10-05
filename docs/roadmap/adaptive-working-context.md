# Adaptive Working Context — Fabric v2 steering decision

Status: experimental direction; primitive and dynamic explorer integration
implemented; live qualification pending. This decision extends the full v2 objective; it does not
replace repair, scheduling, model allocation or final release requirements.

## Architectural boundary

**JEV controls acquisition. Agents control retention. Fabric controls authority.**

Repository intelligence supplies grounded evidence. Evidence Value Controller/JEV
selects worthwhile acquisition or computation. The model organizes already
acquired reasoning state in a disposable per-agent projection. The topology
scheduler and model allocator retain their distinct responsibilities. Native
verification and independent review remain the acceptance boundary.

The inspiration is [Context Language Models, arXiv:2609.37725v1](https://arxiv.org/html/2609.37725v1),
especially editable active context and separate worker projections. Fabric's
implementation is independent code under its existing Apache-2.0 license; no
donor source, prompts, serving stack or orchestration implementation is imported.
Paper results do not establish a Fabric benefit. The paper synchronizes edited
context with inference; Fabric must establish its own runtime integration.

## Requirements retained from the steering instruction

1. Keep objective/source/candidate/task identities, graph dependencies,
   ownership, credentials, permissions, effect intents/receipts, UNKNOWN,
   repair budgets, checks, review, acceptance and journal truth controller-owned.
   Inject this structured envelope separately on every applicable invocation.
   Contradictory note content never overrides it.
2. Bind each bounded projection to run, agent, task, role, source, optional
   candidate and an admitted journal prefix. Children receive their own state,
   selected evidence and only explicitly admitted parent handoffs.
3. Permit explicit compare-and-swap replacement of reasoning content only,
   checking previous context identity/hash and encoding/size. No source writes,
   arbitrary note filesystem, external effects or generic memory API.
4. Reconstruct from journal evidence. Verify bindings on resume. Missing,
   corrupt or stale optional projections degrade to a fresh projection; corrupt
   authoritative journal truth remains a hard stop. No detached-summary trust.
5. Use existing source-bound Agent Skills for advisory retention guidance;
   suggested headings are optional prose, never state-machine fields. Preserve
   useful locations, hypotheses, unresolved findings, constraints, failures,
   observations and next steps. Remove redundant/stale excerpts and tool noise.
6. Acquisition and retention remain separate. Deleting notes never deletes
   evidence. Existing journal/artifact records remain available through their
   admitted interfaces.
7. Workers own separate projections. Supervisor handoffs are bounded and
   structured: status, completed work, exact proposal/candidate, invariants,
   risks, dependencies, evidence references and block reason. Task completion
   continues to require structured controller evidence.
8. Eventually compare spawning with other JEV actions. Keep localized work
   single-agent when additional workers offer no justified evidence value.
9. Retain bounded initial selection, localization, role inputs and hard wire
   limits. Defer sophisticated fixed compaction expansion until the comparison
   with model-maintained working context is complete.
10. Opt in initially. Freeze A=current behavior, B=existing conventional
    compaction where supported, C=editable working context. Hold task/source,
    model/effort/runtime, verification/review, repair budget, tools, maximum
    context budget and concurrency constant wherever the runtime permits.
11. Develop/debug on historical internal tasks; do not access new hidden v2
    benchmark material. Freeze mechanism/policy before final benchmark use.
    Preserve final Codex/Fabric fairness: same task/source/evaluator/model pool,
    maximum concurrency and comparable wall/runaway limits. Efficient workers
    are the starting route; escalation needs justification.
12. Do not import RL/GRPO, the donor serving stack, Suffix Cache Reuse, the
    donor's implementation or another model runtime.
13. Keep the initial mechanism small: separate envelope, bounded projection,
    replacement, journal binding, resume/fallback, advisory skill and metrics.
    No promotion or expansion without downstream evidence of useful value.
14. Incorporate this acquisition/retention/authority split into the v2 thesis.
15. Implement a role end-to-end, deterministic and stale/corrupt/bound tests,
    independent review, then the matched experiment. Document a negative result
    and retain the simpler baseline when appropriate. Do not reopen v1.1 work
    absent a demonstrated regression.

## Initial implementation

`internal/workingcontext` provides a pure bounded primitive: 16 KiB UTF-8 content,
128 KiB encoded-input cap, full content/projection hashes, caller-owned binding,
exact previous identity/hash replacement, and explicit missing/stale/corrupt
fallback. It has no filesystem, runtime, journal writer or acquisition code.
The byte cap is an explicit experimental resource constraint, not a token
estimate. Binding validation checks syntax; controller replay must prove
provenance/ancestry before supplying an expected binding.

The selected first seam is opt-in dynamic explorer child follow-ups:
`FollowUpExplorerAgent`/`prepareExplorerAgentTurn`, structured `ExplorerRecord`
admission, and the existing fresh scheduled invocation. Static graph cohorts
and other roles stay outside this first experiment. The first child turn starts
without an agent-bound projection because its agent identity derives from the
initial invocation/context hashes. Its admitted result may seed the projection
once the controller knows the identity. Later follow-ups know the target agent.

Retain model updates in the existing admitted result, derive projection state
only after exact result validation, and bind its origin to that event's hash.
Preserve a stable per-agent assignment identity separately from fresh scheduler
turn identities. Probe a completed turn using its exact admitted invocation
and agent-turn binding, never by rebuilding it from the latest note. Missing,
stale or corrupt projections may fall back while preparing new unclaimed work;
they must not replace the input of an existing claim. Existing caller/ownership
checks and dispatch guards remain authoritative.

The active projection must actually replace retained reasoning input between
invocations. A tool-only note that leaves the entire transcript active is
insufficient to claim the treatment implemented. Keep absent-policy invocation
bytes and effect semantics unchanged. No context update on host intent, running
state, invalid output or UNKNOWN; optional corruption never permits resending.

The primitive's focused regressions passed in 0.903 s; vet and documentation
checks passed. Independent review approved the primitive only and identified
the binding-cycle and frozen-claim constraints above. These receipts do not
establish role integration, controller replay or live benefit.

The prepared v1.1.11 Disabled/Enabled guidance pair is NOT RUN: user steering
arrived before either arm dispatched. Preserve its preparation; it does not
satisfy the A/B/C working-context comparison.

## Experiment evidence

Record accepted candidate and external acceptance, wall time, aggregate
agent-active time, input/cached/uncached/output/reasoning tokens, model turns,
source/tool reads, context rewrites, peak and average context size, repairs and
repeated acquisition where observable. Unknown metrics remain unavailable;
cached input and reasoning are subsets. A smaller context file alone is not
benefit. Conventional compaction applicability and actual observed lifecycle
must be distinguished from a requested setting.

Dynamic explorer integration now has deterministic journal/replay, accepted-child
follow-up factory, queued-before-acceptance stale CAS, isolation and bounds
regressions. Independent GPT 6 Luna High review approved the primitive and
factory/replay integration, then identified two execution-path issues: shared
runtime parity across arms and recovery after result settlement before exploration
admission. The current draft adds a separate immutable executor opt-in and
exact-receipt reuse without rerunning a successful turn. Independent review
approved these repairs. A Codex app-server fixture exercises the settlement
gap and asserts exact result reuse with an unchanged RPC method log; it does
not establish live provider qualification. The optional source-bound
`working-context` skill is advisory and explorer-only. Accepted supervisor and
writer handoffs omit private scratch updates.

The [user guide](../guides/working-context.md) documents the default-off
`--working-context` flag, supported routes, stale queue limitation and version-size
accounting. A/B/C must share the same dynamic executor policy; retention remains
separate. Existing frozen turns always retain their exact input. Matched live
comparison, final v2 benchmark and promotion remain NOT RUN.

The affected deterministic regressions passed on the final execution-path
draft (control 58.141 s, CLI 1.360 s, Codex RPC 0.272 s and working-context
primitive 0.206 s). The separate Codex cached-receipt regression passed in
16.844 s. A broader control-package attempt exceeded its ten-minute package
timeout during an existing Git finalization test; it is not a full-suite PASS.
The narrower results do not qualify the frozen v2 cohort or demonstrate a
token, wall-time or acceptance improvement.

The first historical live preparation on source `f672fc172cd6` completed planning
in 36.997 s, then stopped at `agent-list` because the Codex graph path had no
registered planner root. No dynamic explorer dispatched. Preserve that run as
a setup failure; it does not count as an A-arm treatment outcome. The bounded
repair registers the existing accepted planner result for the shared executor
opt-in, using the existing agent-tree mechanism. Default runs remain unaffected.

The next historical probe on `ae10cfcce03e` admitted a successful Codex explorer
result, but scheduler post-dispatch inspection failed when the observation's new
paths changed source-guidance selection. Runtime and access receipts were
completed; the provider effect was not UNKNOWN. Completed dynamic turns now use
their exact replay-admitted invocation, question and agent-turn binding for
inspection. New/unmatched turns still derive and validate their frozen identity.
Source-guidance regressions cover both retention-off and retention-on modes,
including rejection of changed question, invocation or agent identity. These
debugging probes are not pooled into a matched-treatment acceptance claim.
