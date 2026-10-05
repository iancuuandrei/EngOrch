# Editable working context: historical A/B/C pilot

Date: 2026-10-05. Decision: **retain opt-in; do not expand to other roles or
enable by default**. The explorer treatment worked end-to-end, but this single
task does not establish a token, latency or success advantage. A is incomplete.

## Frozen execution and method

Fabric execution source: `591df4759e773aaa43f5754d7efef710c7ef3b6e`.
Execution binary SHA-256:
`c1f6294c7dee65488d4297505429ea606064d14b6f203c9584e1287d81376672`.
Codex binary SHA-256:
`1722907aa64401bcc9b34467ef5c2af43f6aef9a5045ef04d11d19dfae4589fb`.

One historical internal JSONL decoder task was used: add configurable bounded
record decoding while preserving the public API, record order and safe errors.
Task source: `a226ccf3eec4116694f6c25c5ccd19225a542281`; tree:
`630c24ad37264e902a1532feafff8c219916afb2`.
The historical external evaluator hash was
`afa9213e52bd8c17c15809c264c552a3a4f4469af51d5c1d0a16c5cf9ba579a1`.
No new frozen v2 benchmark material was accessed.

All arms used Codex app-server, GPT 6 Luna High for each role, source-bound
agent guidance, the same tools and verification/review, maximum concurrency one,
two candidate repairs, and a 900-second watchdog per command stage. The existing
subscription route had explicitly unlimited token ceilings; this is not a
finite-token-budget or access-profile qualification. Existing input/wire bounds
were retained; C additionally bounds retained notes to 16 KiB. No stronger model
or extra worker was added. Arms executed in A/B/C order, without randomization.

| Arm | Retention / compaction policy |
|---|---|
| A | Shared dynamic explorer executor; bounded inputs; no editable notes or extra compaction setting |
| B | Same executor; no editable notes; existing Codex automatic compaction requested at 8,192 tokens |
| C | Same executor; editable explorer context; no extra compaction setting |

The planned sequence was accepted planner preparation, one explorer with three
successive questions, then autonomous implementation, native verification and
independent candidate review. Questions covered source localization, boundary
behavior, and a concise implementation handoff. Follow-ups waited for admission.
Each provider-bearing stage had an exclusive dispatch marker; uncertainty never
authorized replay. Earlier setup/debug runs are excluded from this cohort.

## Outcomes

| Metric | A | B | C |
|---|---:|---:|---:|
| Controller outcome | PLANNING / UNKNOWN | READY | READY |
| Native candidate verification | NOT RUN | PASS | PASS |
| Candidate review | NOT RUN | approve | approve |
| External acceptance | NOT RUN | PASS | PASS |
| Wall time to stop / completion, seconds | 16.813 (incomplete) | 656.779 | 608.491 |
| Sum of measured command-stage wall time, seconds | 16.813 | 656.637 | 319.588 |
| Observed input tokens | unavailable | 303,632 | 476,174 |
| Cached input tokens | unavailable | 190,720 | 355,328 |
| Uncached input tokens | unavailable | 112,912 | 120,846 |
| Output tokens | unavailable | 6,413 | 10,478 |
| Reasoning output tokens | unavailable | 1,368 | 4,406 |
| Receipt-matched completed runtime invocations | unavailable | 6 | 6 |
| Observed source/tool read requests | unavailable | 12 | 24 |
| Accepted context rewrites | NOT RUN | not enabled | 3 |
| Rejected context updates | NOT RUN | not enabled | 0 |
| Peak accepted context bytes | NOT RUN | not enabled | 1,246 |
| Average accepted-version bytes, integer floor | NOT RUN | not enabled | 750 |
| Total accepted-version bytes / version count | NOT RUN | not enabled | 2,251 / 3 |
| Observed conventional compactions | unavailable | 20 | unavailable |
| Candidate repairs | 0 | 0 | 0 |
| Planner semantic corrections | 1 | 0 | 0 |

Cached input is included in input; reasoning output is included in output.
Token sums for B/C require every invocation to be completed, receipt-matched
and provider-accounting coverage OBSERVED. Read requests count recorded tool
requests, not distinct files or all possible model-internal reads. Runtime
invocations are not internal model turns. Aggregate agent-active time, internal
turns, unnecessary/repeated acquisition and time-weighted context size are
unavailable. No cost is inferred from subscription token counts.

Only B explicitly observed compaction lifecycle events across all six
invocations. Missing compaction observations for A/C do not establish zero.
B's particular low threshold does not represent every conventional compaction
policy. C used more observed input, uncached input, output and reasoning tokens
than B; smaller retained notes do not establish a compute benefit.

### Retained identities

| Arm | Run identity | Accepted candidate identity |
|---|---|---|
| A | `564140b59ba804be74adf59d1077ad078c27d966486b4751d00d0eb6db8a9c3c` | none |
| B | `ed65f6029045fd45adb57f79aa0db33ce58f86dcfb42ed0db8b7d42693e7775f` | `6076af31d1173c727f7985e66cb8962e7851bba8e7ab2a5f54a59a7d7a8d10c6` |
| C | `3ec04b9fcff9faf79bd2bda05f0575ddf0e7a3f0b66f3f65e99a4df268aaa5dd` | `4d08c317e114a7063ffb9b1d0ef94049aec7f82689983f68867e83172d22ae05` |

External evaluation ran once per accepted candidate in a separate copied
`go.mod`/`decoder` tree with the historical evaluator injected only into that
copy, using Go 1.27.1 `go test ./... -count=1`. Original candidate files were
hashed before/after to check that evaluation did not change them. B passed in
1.202 seconds; C passed in 0.889 seconds. Candidate-copy manifest hashes:

- B: `1c91fe3db89cc0caa4ded0cd54c5289f0fcf5de883486ef7018a4870885e2a6d`.
- C: `a54144b0bde4ae3f48e3625191bbcd816f3e0c9528365bd079812ccc3f100a66`.

The [sanitized machine-readable summary](working-context-pilot-20261005.json)
records these measurements. Private journals, receipts, dispatch markers and
evaluator copies remain retained locally; raw prompts/provider responses and
credentials are not published with this report.

## Failures and limits

A stopped during planner correction with `effect_requires_reconciliation`.
An earlier completed planner receipt does not settle the current uncertain
effect. No resend was attempted. Aggregate `usage` rejected a runtime receipt
without a matching host intent, so its totals are unavailable. This is an
unresolved baseline/accounting issue, not evidence that retention improves
planning. A's stop time cannot be compared as time to accepted candidate.

C's first explorer result and context were admitted, then read-only `usage`
failed because its average used a float unsupported by canonical v1. The
bounded repair uses an integer floor and preserves numerator/count for the
exact mean. A targeted regression and independent review passed. C resumed
only subsequent new turns; the first explorer was not repeated. Its overall
wall time includes a manual telemetry repair pause; summed command stages
exclude that pause and do not measure agent-active time.

All provider-bearing execution used the frozen binary above. Read-only final
metrics collection for every arm used the telemetry repair candidate binary,
SHA-256 `83febbac5516704d65526f578645079ccaca2885bc4f092bbca0da9c74dd9f48`.
This reporting change did not alter prompts, runtime policies or effects. It
still limits treating the pilot as an uninterrupted timing comparison.

The integration's focused deterministic tests and independent GPT 6 Luna High
reviews passed. The final canonical-mean/replay regressions passed in 7.403
seconds. The earlier broad control-package attempt timed out after ten minutes
in an existing Git finalization test; no full-suite PASS is claimed.

## Decision

The mechanism supports real child follow-ups, admitted context replacements,
an accepted coding candidate and historical external acceptance. Keep explorer
retention experimental and off by default. Do not expand roles, claim a general
benefit, invest in more compaction machinery, or count this pilot as final v2
benchmark/release qualification. Reconsider expansion only with evidence of
downstream engineering value under an admissible comparable evaluation.
