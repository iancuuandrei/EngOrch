# Agent-context parser canary

**Predeclared experiment; initial live result NOT QUALIFIED.** This is a new fixture task,
not a modification of any canonical v1 pin or oracle. No performance or general
agent-quality claim follows from this single task.

## Hypothesis and frozen treatment

Source-bound scoped guidance may preserve decoder compatibility and safe errors
with no acceptance loss. Compare explicit agent-context Disabled against Enabled
on identical fresh committed copies of `fixture/`, using the same clean Fabric
binary, Codex model/effort, native checks, two-repair budget and held-out bytes.
The only intended treatment difference is the agent-context run flag. Run each
arm once; retain refusal/failure and uncertainty without retrying it for a score.

Objective (identical in both arms):

> Extend decoder.Decode to support JSON Lines records larger than the scanner's
> default capacity. Add DecodeWithLimit(io.Reader, int) ([]json.RawMessage, error)
> for configurable bounded record sizes. Preserve Decode's public signature,
> independent JSON record ordering, compatibility policy and safe errors. Add
> discriminating regression tests. Limit changes to decoder implementation/tests.

Root/scoped AGENTS.md provide project requirements, not an implementation answer.
The fixture has a real scanner-size and diagnostic defect. Native checks initially
pass, while the independent held-out oracle requires the new API and checks size
boundaries, reader errors, atomic batch failure and sanitized diagnostics.

Before dispatch retain source commit/tree, fixture/held-out hashes, exact clean
Fabric VCS revision and binary hash, runtime/model/effort, initialized config
hash and separate fresh run IDs. Keep held-out tests out of model-visible source;
run them only on an exact copy of the reviewed candidate. READY additionally
requires ordinary native verification and candidate-bound independent review.

Report raw outcomes for both arms: native/review/held-out acceptance, instruction
violations, role/tool calls, repairs, wall time and available input/cached/uncached/
output/reasoning usage. Unknown prices stay unknown. Inspect guidance selections
and skill references instead of inferring them from a requested flag.

This fixture exercises scope and guidance. Selected-skill quality needs a separate
matched task; this result cannot qualify every agent-context feature.

## Initial execution

The [retained report](../../../docs/evaluation/agent-context-parser-canary.md)
records both original arms on clean Fabric `24aa779`: Disabled exhausted two
repairs with native FAIL; Enabled stopped with unresolved dispatch state after
local schema rejection. No task acceptance or context benefit was established.
The wire correction receives separate regression evidence; do not relabel or
retry these original arms as a favorable matched result.

## Fresh comparison after the wire fix

The separately identified v1.1.11 pair freezes clean Fabric source
`335a8de8fb9869b187b3d73471cc8581a851e328`, binary SHA-256
`3711376f0e5a5a3a37ff03b4a2d955a92a23cf4a1bf6dfa1080490b0ec1790d5`,
and the original task commit/tree and held-out bytes above. Both native baselines
passed before dispatch. Model/effort remain `gpt-6-luna`/high, parallelism one
and repair budget two. Run Disabled first, then Enabled, once each on fresh
roots. Repair intelligence remains off to isolate the context treatment.

The pair has no retry or successor policy. Every arm uses an exclusive durable
dispatch marker; a missing process receipt is not permission to dispatch again.
Failure, refusal and UNKNOWN remain raw outcomes. Original runs remain intact.
Serial order and provider cache state are uncontrolled comparison limitations;
one pair cannot establish a general benefit or a statistical performance claim.

Private inputs and execution receipts reside at
`D:\fabric-ci2-tools\v2-agent-context-pair-v1111-20261005`.

Provider dispatch is **NOT RUN**: subsequent user steering prioritized the
[Adaptive Working Context A/B/C treatment](../../../docs/roadmap/adaptive-working-context.md)
before either fresh arm was dispatched. This preparation is not qualification
of guidance or editable context.
