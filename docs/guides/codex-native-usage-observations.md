# Native Codex context observations and consumed tokens

Codex can emit a `thread/tokenUsage/updated` notification whose `last`
contains a positive total and zero input, cached input, output, reasoning
output, and cache-write components. The pinned upstream
[`recompute_token_usage` implementation](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/core/src/session/mod.rs#L4785)
constructs precisely this shape from an estimated context size while retaining
the existing cumulative usage. This source reference explains an observation
shape; it does not attest the source of the installed runtime executable.

Fabric recognizes only that narrow last-observation form when the cumulative
`total` independently passes its existing validation. The estimate contributes
no input, cached input, output, reasoning, cost, or consumed-token budget.
Missing required fields, nonzero components with inconsistent totals, invalid
cumulative totals, identity mismatches, regressions, and token-budget ceilings
retain their existing checks. Cached tokens remain a subset of input; reasoning
tokens remain a subset of output. Unknown cost remains unknown.

New normalized observations record version 2. Historical records without a
version replay with the original validation order and failure text, preserving
previous interruption and UNKNOWN outcomes. Recognizing the native shape in
new work does not authorize resending a historical invocation.

A configured compaction threshold or a total-only context estimate does not
prove successful native compaction. That claim requires a completed runtime
turn and matching, identity-bound compaction lifecycle evidence. Coding task
acceptance separately requires candidate-bound verification and approved review.

## Compaction lifecycle notifications

The pinned Codex [started notification](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server-protocol/schema/typescript/v2/ItemStartedNotification.ts)
and [completed notification](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server-protocol/schema/typescript/v2/ItemCompletedNotification.ts)
include method-specific timestamps and a complete
[ThreadItem](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server-protocol/schema/typescript/v2/ThreadItem.ts).
Ordinary user-message text can exceed the 16 KiB bound used for compaction
metadata. Fabric classifies known non-compaction items under the existing
1 MiB RPC message bound and excludes them from compaction observations.
Actual compaction items retain the smaller bound, strict identity checks,
matched lifecycle pairs, and event limits. Method-specific timestamps are
optional for historical shapes and never substitute for completed-turn proof.
Unknown or malformed item types retain UNKNOWN coverage.

The interrupted v1.0.17 run recorded generic INVALID_NOTIFICATION coverage
without the offending payload. Its exact notification shape cannot be
reconstructed. The schema and large-message regressions demonstrate observer
incompatibilities; they do not retroactively qualify that run's compaction.
