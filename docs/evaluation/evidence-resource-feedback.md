# Observed token resource feedback

## Scope

v1.1.23 adds read-only `evidence-feedback`, connecting existing receipt-verified
Codex accounting to the finite resource economy. It updates advisory resource
prices/used quantities with per-axis invocation cursors, preserving limits.
Caller baseline/allocations remain unvalidated policy inputs; this is not an
automatic budget ledger, route allocator or calibration qualification.

Focused tests cover clean typed-token partitions, missing/unmatched/incomplete
or anomalous accounting, invalid subsets, duplicate/unknown/unavailable cursors,
allocation validation, numeric bounds, caller immutability, underuse/overuse,
unknown money and repeat idempotence. The first compile attempt assumed an
accounting field named `Usage`; the actual receipt uses `Delta`. That assumption
was corrected before execution, preserving existing receipt schemas.

Initial focused checks passed: evidencevalue 0.268 s, control 0.127 s and CLI
0.132 s. Subsequent targeted checks passed (evidencevalue 0.194 s, control
0.118 s), and the real Git CLI lifecycle with fake-accounting absence,
no-journal-mutation and stale feedback rejection passed in 25.978 s.
The added finite-frontier regression demonstrates that observed price updates
can make an otherwise eligible positive acquisition negative (evidencevalue
0.229 s); generated reference/API/documentation checks passed (doccheck
0.550 s; CLI 0.112 s). Vet for evidencevalue/control/CLI and diff checks passed.
Independent GPT 6 Luna High static review: APPROVE, reconfirmed on the final
request-size guard, added tests and docs; no tests/providers were run by the
reviewer. Full documentation checks passed in 0.374 s. Full repository/race
checks and a new live cohort were not run for this targeted increment.

## Historical real-receipt inspection

The command was executed read-only against the accepted historical treatment
run `e409e32999edb9eedf812df50110eab1b5136db36e1d02812316aa8671a09c47`
from the [short-root repair pilot](repair-intelligence-short-root-20261005.md).
The initial guessed root was the experiment container, and was rejected by
repository binding. Reading the retained state identified the exact task repo;
no effect or repository identity was rewritten.

Ten existing matched completed invocations supplied clean typed accounting.
Usage-basis hash:
`7434a1046c856d5c6911009f2f7a55fa919da0858a456ab33d63c7766ee23646`.

| Observed category | Tokens |
|---|---:|
| Uncached input | 303,432 |
| Cached input | 865,792 |
| Ordinary output | 7,310 |
| Reasoning output | 12,336 |

Input components sum to the previously observed 1,169,224 input tokens; output
components sum to 19,646 output tokens. No cost category is counted twice.
The demonstrator used zero advisory baselines/prices, limits of 1e9 and a
10,000-unit allocation per observation solely to exercise feedback math. These
inputs do not establish optimal allocations or fair economic exchange rates.
Money had no observation and remained unconsumed; its preserved baseline zero
is not a zero-cost measurement.

Reapplying the returned request produced no newly applied invocation for any
axis. Canonical controller export was identical before/after. Private request,
result and repeat receipts remain under `D:\fr19\feedback-v1123`; no prompts or
provider responses are included in this report. No provider was called and no
historical run was resumed. This establishes observed-data integration and
cursor behavior, not a new benchmark cohort or downstream efficiency benefit.

Autonomous feedback persistence, non-Codex typed accounting integration,
monetary/wall-time/resource observations, calibrated reliability and final
Fabric v2 acceptance remain NOT RUN.
