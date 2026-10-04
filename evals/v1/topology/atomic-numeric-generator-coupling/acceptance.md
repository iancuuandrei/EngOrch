# Topology A/B acceptance recipe

## Why this fixture exists

The pinned repository has a real shared generator boundary:
`gen.go` invokes `internal/gen-atomicint`, whose `wrapper.tmpl` emits the
integer wrapper files including `int64.go` and `uint64.go`. The requested
signed and unsigned contracts exercise distinct behavior branches, while a
template change and its generated outputs must remain one coherent owner.

The desired task shape is staged: one owner integrates the shared template and
generated outputs; after that parent-candidate change is confirmed, two
independent signed and unsigned implementation/test tasks become ready. Their
write sets must be disjoint, while the accepted graph records their dependency
on the shared generator integration. `engineeringplan.Task.Dependencies`
expresses that order; the scheduler executes the currently ready tasks and
does not rewrite a graph into this shape. Do not force this decomposition or
edit a planner answer. If inspection shows the work should remain one cohesive
task, record that result as “topology not available” rather than splitting it
to manufacture concurrency.

This fixture's parallel work is deliberately bounded: the shared generator
change remains one owner, and the signed/unsigned leaves focus on their
distinct semantic regressions. It measures staged cohort correctness, not
parallel implementation throughput across two large source clusters.

## Current product limitation

This staged treatment cannot run under the current isolated-implementation
version. Isolation preparation freezes the ready implementation cohort against
the pristine parent candidate; the run stores one `GraphIsolationPreparation`
and one `GraphWriterBatch`. `PrepareGraphIsolationCohort` rejects a second
preparation, while `CreateTaskIsolation` is explicitly initial-source only.
After the generator task is integrated, the parent candidate has changed and
the implementation-v1 preparation cannot admit the later signed/unsigned
cohort. Dependency graphs support the task ordering, but isolated writers do
not yet support repeated cohorts across confirmed parent effects.

The relevant current boundaries are
`internal/control/graph_isolation.go` (`PrepareGraphIsolationCohort`,
`expectedGraphIsolationPreparation`, and `CreateTaskIsolation`) and
`internal/control/graph_writer.go` (`replayIsolatedGraphWriterBatch`). The
preparation and batch are single snapshot fields, and the batch binds to one
candidate and graph revision.

There is also a planner-contract boundary: isolated mode selects
`plan-graph-v7` in `internal/control/planner_contract.go`. That instruction
requires each initial implementation to be ready after shared read-only
research/design and prohibits dependencies between initial implementations.
The generic graph model can represent dependencies, but v7 does not request a
changed-code hub implementation followed by implementation leaves. A future
staged topology therefore needs a versioned planner contract as well as the
repeated-cohort scheduler support below; the scheduler must not repair or
rewrite the planner's task graph.

The smallest product seam is a separately versioned staged-graph planner
contract plus repeated-cohort policy. Each appended cohort preparation must
bind the exact current parent candidate,
graph digest/revision, ready task IDs, route, and resource demand. It may be
created only after the previous cohort’s parent integration and progress are
confirmed. Every child starts from that current parent candidate; each cohort
has its own immutable preparation and one serial parent-side aggregate. An
UNKNOWN child or parent effect blocks later cohorts. Existing v1 behavior and
its replay bytes remain unchanged. Until that seam exists, this fixture is
design-only and must not be silently run as a single initial cohort.

## Matched arms after the product seam

Use fresh source checkouts and external controller-state roots. Keep the
objective, task pins, Go toolchain, model and effort, Codex runtime, RI binary
and digest, prompt recipe, edit validation, repair budget, verification scope,
and held-out test identical. The sole intended treatment difference is:

- Control: serial implementation, `MaxParallel=1`.
- Treatment: repeated isolated cohorts, `MaxParallel=2`, using the explicit
  resource policy in `fixture.json`.

Both preparations must record zero provider calls. Use the same source commit
`2d2bdbd262f95d0a890c696fb6f91d4776cb142c` for both arms. Evaluate sequentially
after the exact binary, runner, product tests, and hosted checks are frozen and
qualified. Do not resume any earlier numeric-text run.

## Evidence required for a topology result

Do not infer concurrency from `MaxParallel`, a planner proposal, or separate
task names. A successful treatment record must show all of the following in
the validated journal/checkpoint:

1. The generator-owning task and the exact parent candidate after its confirmed
   integration.
2. A later accepted graph revision with at least two ready implementation
   leaves, a shared dependency on the generator work, and disjoint concrete
   write paths. The planner may choose another valid graph; the test checks
   these properties rather than requiring fixed task IDs or an exact graph.
3. A separately confirmed child isolation identity and candidate-bound writer
   receipt for each leaf, both bound to the same post-generator parent
   candidate.
4. Positive overlap between the two controller-recorded writer-dispatch
   intervals, plus a single deterministic parent aggregate containing exactly
   those two leaf results. Dispatch intervals show overlapping task work; they
   do not prove provider-request overlap or provider-call counts.
5. Confirmed parent file effects, followed by fresh native verification and
   review on the combined candidate, with no unresolved external intent.

If any condition is absent, report that condition as `NOT PROVEN`; do not
rewrite the objective or plan. The control should show the same task contracts
integrated serially with no overlapping writer-dispatch intervals.

## Acceptance gates and reporting

Run the unchanged native package suite and the unchanged held-out test at
`../../heldout/atomic_numeric.heldout_test.go`. On a disposable copy of the
integrated candidate, record generated-output hashes, run the repository's
`make generate` target, and require those hashes to remain identical. This
target builds the generator binaries before invoking `go generate ./...`;
do not use `make generatenodirty` on a dirty candidate. Preserve the
existing Windows-only `^TestNocmpIntegration$` exclusion from
`evals/v1/manifest.json`; do not exclude the numeric held-out check. Do not
weaken tests to fit a proposal.

Report each arm’s READY state, candidate identities, repair usage, native and
held-out exit states, reviewer decision, child and aggregate receipts, and
typed usage fields. Keep missing metrics null. Runtime invocation count is not
a provider-call count. A single matched pair can establish that the topology
occurred; it cannot establish a general speedup. Any elapsed-time comparison
must use the same measured interval and be labeled as one sample.
