# Native Codex automatic compaction

Autonomous runs can opt Codex app-server threads into the native automatic
in-turn compaction threshold with `run --autonomous --auto-compact-token-limit N`.
The value must be between 1 and 10,000,000. Omit the flag to preserve the
existing thread configuration and invocation identities.

The threshold is part of the immutable run policy and each Codex invocation
identity. The runtime records the requested value before `thread/start`, then
sends it as `config.model_auto_compact_token_limit`. The recorded thread field
describes the requested config; it is not evidence that compaction happened.
The setting does not request manual compaction and does not create another
turn or retry.

The upper bound is a configuration-input bound, not a claim about a model's
context capacity. Whether native compaction occurred must be established from
matched lifecycle observations; merely enabling the threshold does not prove
that it triggered.

The v1 evaluation runner exposes the Native-only matched treatment as
`-AutoCompactTokenLimit N` on both Prepare and Evaluate. Prepare records the
requested threshold in `run.json`; Evaluate requires the same value, checks
the inspected immutable execution policy, and records requested/observed
threshold fields in its receipt. Omit the parameter on both actions for the
legacy argv. PR5Matched rejects this option because its baseline invocation is
kept unchanged.

## Observed continuation evidence

Two coding rounds reached accepted, independently checked candidates. The
sustained second round recorded six native compaction events, then
completed a run-scoped pause and resume without another model invocation.
These rounds used mixed Fabric versions and shared the recorded parent runtime
state home. They qualify this observed continuation, rather than a frozen
single-version experiment or every Codex release. See the
[current product checkpoints](../roadmap/v2-product-plan.md) for exact evidence
and the remaining release gates.
