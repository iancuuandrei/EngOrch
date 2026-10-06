# Finite evidence acquisition integration

v1.1.22 adds explicit `evidence-acquire`: finite JEV selection can enter existing
bounded explorer source context admission. It records the complete declared
model and targets, producing prefix, canonical request identity and recomputable
decision. No provider-bearing action is implemented here.

Targeted deterministic checks exercise positive selection reaching explorer
input, decision replay, changed reports/targets/identities/prefix rejection,
no-value and unavailable-cost stopping without context acquisition, malformed
frontiers without journal mutation and controller stop precedence. Initial
controller checks passed in 27.206 s. Review identified the gap between the
decision and source admission: the candidate/stop state could change there.
Decision-linked admission now rechecks that binding before capture, after
capture and during append replay; discriminating gap regressions passed with
the full targeted acquisition selection (control 7.693 s; CLI wire/stops 0.120 s).
The reviewer initially also questioned append atomicity, then withdrew that
finding after inspecting both backends' proposed-history validation.

No model provider was invoked. The existing real Git CLI lifecycle with legacy
acquisition rejection/no-journal-mutation assertions passed in 20.491 s;
selected documentation/API checks passed in 0.513 s. Final independent GPT 6
Luna High static review: APPROVE after the decision-bound admission repair.
The reviewer ran no tests/providers. Acquisition plus selected-text, immutable
replay, forgery and byte-identical legacy context regressions passed in 15.475 s.
Full documentation checks passed in 0.422 s. Vet for evidencevalue/control/CLI
and `git diff --check` passed. This is targeted deterministic integration
evidence, not a full repository suite, race-detector or live-provider result.

Declared costs/priors are not calibrated observations; this increment does not
prove token, time, money or engineering-quality improvements. Autonomous
acquisition policy, observed cost ingestion, broader action execution and a live
matched downstream comparison remain NOT RUN. Phase C and Fabric v2 remain
incomplete.
