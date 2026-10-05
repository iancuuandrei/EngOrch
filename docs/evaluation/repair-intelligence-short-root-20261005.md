# Structured repair: short-checkout historical pilot

## Scope and frozen controls

This is a **separate cohort** from the
[earlier Windows path failure](repair-intelligence-pilot-20261005.md).
Its two UNKNOWN workspace effects remain retained without resend. No outcomes
are pooled across the cohorts.

Product source: `5155ccd4f1c8e815a4ed26ca32968b527c60e41d`, clean embedded VCS
binding. Binary SHA-256:
`6366543bfcf58cc9e015c58219a8c2c85fbf130c4166f266d671aa9e1aab8492`.
The [machine-readable result](repair-intelligence-short-root-20261005.json)
binds runtime, runner and copy-helper hashes, runs and acceptance identities.
Private protocol, prepared receipts, validated journal exports, runtime usage,
candidate-bound native/held-out logs and process receipts are retained in
`D:\fr19`. The protocol was written before preparation and both preparations
were frozen before any provider dispatch.

Both arms use the unchanged historical godotenv objective, upstream
`ddf83eb33bbb136f62617a409142b74b91dbcff3`, native checks and external evaluator.
This public task is neither unseen nor part of the frozen v2 benchmark.
Both use the same binary and Codex runtime, GPT 6 Luna High, disabled agent
context, maximum concurrency one, two repairs and a 900-second wall limit.
Control runs first; treatment runs second. Each arm is dispatched once, with no
successors. Treatment enables `--repair-intelligence`, including reviewer
rechecks; this is a bundled comparison, not isolated coverage localization.
No imported spectrum is supplied.

An initial local preparation-script argument-binding failure occurred before
either checkout or provider existed. Its log is retained separately; it is not
a model attempt. Corrected preparation passed upstream tests and the expected
held-out baseline discriminator for both arms with zero provider calls.

## Acceptance and repair evidence

| Observation | Control | Treatment |
| --- | --- | --- |
| Evaluator outcome | BLOCKED | PASS |
| Durable run state | REPAIRING | READY |
| Repairs admitted | 1 | 2 |
| Completed repair writer observations | 1 | 2 |
| Candidate-native checks | PASS | FAIL, PASS, PASS |
| Reviews completed | 1 | 2 |
| Final accepted review | None | approve, zero findings |
| External candidate-copy native check | NOT RUN | PASS |
| External held-out acceptance | NOT RUN | PASS |
| Historical native finding closure | None | native_oracle_passed |
| Historical reviewer concern closure | recheck_required | reviewer_recheck_closed |

Control completed a writer result whose `godotenv.go` precondition hash did not
match the retained candidate bytes, as confirmed by read-only comparison of the
retained result and candidate. Fabric rejected the repair before file intent or
application. Its earlier candidate remains retained; the CLI reports
`operator_attention_required`. There is no accepted control repair or external
acceptance. No effect is resent to discover a different result.

Treatment's two repair writer inputs contain candidate-bound Finding/RepairSpec
projections, each with one finding, one specification and `recorded_context`
anchoring. The final candidate is
`9343d7edef0301e2a7879d57f007e320f8fc3f338dff7e697451f9fbda80c7ec`.
Review, verification, diff, native-copy and held-out-copy identities match it.
The accepted verification plan is
`aea24b5d177aa05bb374939035b53f4860ac4431cee922cd0638193b7329f8b4`.
`diagnose --closure` reports a historical native oracle passing and a reviewer
recheck closed against that candidate and plan, with distinct retained receipts.
These remain separate closure claims; prose does not replace either gate.

## Measured resource use

| Metric | Control | Treatment |
| --- | ---: | ---: |
| Matched completed runtime invocations | 7 | 10 |
| Input tokens | 599,772 | 1,169,224 |
| Cached input tokens | 398,336 | 865,792 |
| Uncached input tokens | 201,436 | 303,432 |
| Output tokens | 10,948 | 19,646 |
| Reasoning output tokens | 5,421 | 12,336 |
| Source-read requests | 4 | 10 |
| Source-list requests | 1 | 4 |
| Evaluation task wall seconds | 326.015 | 620.747 |
| Outer evaluator wall seconds | 327.737 | 619.870 |
| Timeout | No | No |

Usage coverage is OBSERVED for every matched completed invocation, with no
pending calls or accounting anomalies. Cached input is part of input;
reasoning is part of output. Runtime invocations are not physical provider-call
counts. Cost, aggregate agent-active time and exact time-to-READY are unavailable.
Task and outer wall times are independently recorded clock scopes; neither is
substituted for aggregate agent-active time. Private raw evidence is retained;
the public summary contains bounded metadata only.

## Decision

**Live integration and exact closure: demonstrated for this historical task.**
Treatment produced an accepted candidate after concrete native and review
failures. It consumed more tokens and time than the blocked control; one ordered
task with different outcomes establishes no causal efficiency improvement or
general acceptance-rate benefit. Retain structured repair as **opt-in**. Do not
expand recovery machinery or run another equivalent pilot to obtain a preferred
answer. Next product work is the finite Evidence Value Controller, using these
real action/resource observations while preserving independent acceptance.

This record does not qualify final v2, a new package or installed acceptance.

## Publication checks

Independent GPT 6 Luna High evidence review: **APPROVE** within this historical
pilot scope. It reconciled frozen identities, prepared hashes, dispatch order,
terminal process receipts, usage totals and candidate/closure bindings against
retained private records. The reviewer ran no providers or tests.
Documentation regressions: **PASS** (0.812 s); `git diff --check`: **PASS**.
Separate metadata checks confirmed distinct invocation identities within each
arm and valid input/cached/uncached and output/reasoning relationships.
