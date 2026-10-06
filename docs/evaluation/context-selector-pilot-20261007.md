# Context-selector pilot 2026-10-07 (frozen, scoped)

Frozen experiment completed before the v1.1.34 usage-accounting repair.
This record describes its exact source, artifacts and runs; it does not
define current behavior. Prior historical evidence is preserved unchanged;
this cohort is not rebuilt, redispatched or pooled with any other cohort.
Original journals and reports were never rewritten.

## Frozen scope

Compare default context selection (control) with the opt-in
`rrf-coverage-v1` selector treatment on two public tasks: difflib and
numeric. All configured roles used Muse Spark 1.3 Contributor High over
OpenCode Go (`opencode-go/muse-spark-1.3-contributor`, high variant), with
max-parallel 1, max-repairs 2, max 16 calls per invocation, invocation
timeout 600 s and overall frozen wall limit 3600 s. One attempt per arm, no
successors or retries, same task, source and configuration per pair except
the root and the selector flag. Balanced dispatch order:
difflib-control, difflib-treatment, numeric-treatment, numeric-control.
Config paired normalized hashes are equal.

Build source `49afbc8daa6e4140c11ae20260b36f0efe3ebbfa`, tree
`12256ad29cee5dc8b8b5ecc35717e0c65a54d79e`, clean binary SHA-256
`ad72b8375db9e0ca00171eff420d1ef3957c676a9d43a99f99a7e72a7c04a1a7`,
frozen protocol SHA
`f4d69702aab462080f1a1caccf1d892c41fe62cb4c089bb2430b97827f4520bd`.
Both original native baselines PASS; copied unchanged public v1 oracles
record baseline targeted-assertion failures. Frozen v2 hidden material was
never consulted.

Task pins: difflib `5d4384ee4fb2527b0a1256a821ebfc92f91efefc`; numeric
`2d2bdbd262f95d0a890c696fb6f91d4776cb142c`. On Windows native checks,
only `TestNocmpIntegration` skips under existing public policy; held-out
checks never skip.

Acceptance used the existing reviewed candidate copy `copy-b58a649.exe`
under a read lease plus the exact source `Get-RunGate`, with separate
source-bound native and held-out copies and exact candidate, verification
and review binding; no original candidate was mutated. Copy binding was
strengthened in `acceptance-binding.json` after dispatch and before the
first acceptance, and applied equally; the original protocol was retained.
External proof is scoped to the named public v1 tasks, not to final v2 or
to general quality or causal gain.

## Observed arms

Machine-readable summary:
[context-selector-pilot-20261007.json](../../evals/v1/results/context-selector-pilot-20261007.json).
All wall times run from dispatch to CLI process termination, not exact
READY timestamps. Controller state and CLI process exit are reported
separately. Unsupported categories are null, never zero.

1. difflib-control
   (`19c5598413d42c16f3fe151eafe1049b28e1b4c001ae32b63cc5b1560f73dd65`):
   final controller REPAIRING, terminal CLI exit 1, wall 552.6685334 s,
   repairs 2 exhausted, native FAIL (string literal not terminated), no
   accepted review, external NOT ATTEMPTED. Final candidate
   `8b33820909c89371f520536be19752a2ad0e8d5712935b6b399726389690ed51`.
   6 verified completed invocations, 28 calls and 28 receipts; input
   596380 (cached 356188, uncached 240192); output 36577 (reasoning
   30667, ordinary 5910); cache-write, money, agent-active-time and
   tool-read counts null. Final checkpoint 0 active and 0 uncertain;
   no settlement.
2. difflib-treatment
   (`463a36032badb34f63d82ce44d49c0beeae3df01d88b0f49540655fe8ba1b615`):
   final controller READY, CLI exit 0, wall 298.8209011 s, repairs 0,
   native PASS, review approve with 0 findings, external PASS. Candidate
   `76c8d88dc9b01670ecb01e6d388ca927d57006a598ff9c4d36cc6c5f028478f1`.
   3 verified invocations, 12 calls and 12 receipts; input 226960
   (cached 130124, uncached 96836); output 11918 (reasoning 8949,
   ordinary 2969); cache-write, money, agent-active-time and tool-read
   counts null. Actual admitted writer and reviewer manifests select
   `rrf-coverage-v1`: writer 10810 selected bytes across 3 files,
   reviewer 19002 bytes across 4 files. Final 0 active and 0 uncertain;
   no settlement.
3. numeric-treatment
   (`d454226d6c81bbb5a9554adcd1c234ca8978b7d0c43c0c9534b2de9251a893f9`):
   final controller PLANNING, CLI exit 1, wall 36.4962985 s, candidate
   absent, no native, review or external proof. Terminal detail
   `PLANNING/operator_attention_required` after a rejected invalid
   planner JSON (`invalid character i after array element`) and one
   semantic correction; checkpoint 1 active intent and 0 uncertain; no
   run restart or settlement. The treatment never reached the writer
   selector, so this arm is not evidence of selector-caused failure.
   Original v1.1.33 observation (corrected): the usage report carries
   ONE OpenCode pending row, not no rows, for invocation
   `37c2d718500acf83d2f6828fcae6ca55dbde0d1a08293813d8c02f2293cda7f9`
   with status UNKNOWN, `receipt_matched` false, `journal_present`
   true, gateway calls 4 and receipts 4, and no usage totals. Those raw
   counts were never verified completed cost; the final projection was
   cleared by the semantic correction, and the original observation is
   preserved explicitly here.
4. numeric-control
   (`1d868270aaa8e820b8a32273c13b8951198f5a57b1007b0d43c4be39ed4817aa`):
   final controller READY, CLI exit 0, wall 577.4877629 s, repairs 1,
   native PASS, review approve with 0 findings, external PASS including
   regeneration consistency. Candidate
   `4e87bfb4476828441ca25782380103f5c6df0ca4f6d6345bc0d7b879375a634d`.
   8 verified invocations, 32 calls and 32 receipts; input 862547
   (cached 450720, uncached 411827); output 40829 (reasoning 29084,
   ordinary 11745); cache-write, money, agent-active-time and tool-read
   counts null. Final 0 active and 0 uncertain; no settlement.

## Later read-only remeasurement (parent-executed, PASS)

A parent pinned read-only observer remeasured the frozen journals without
redispatching the cohort and without rewriting any journal or report. All
178 logical journals were unchanged, including the original 35, and all 35
heads unchanged. All four arm usage totals were verified against exact
replay, runtime, gateway and composite authority. No provider was
dispatched. The three previously known rows are unchanged. The accepted
legacy run's 3 roles preserve their original typed totals, duplicate
feedback applies 0, and both previous UNKNOWN runs remain unavailable
without cursor changes.

New numeric-treatment historical costs from that later observer (a later
observer only, not a new model attempt or a changed cohort): 1 verified
invocation, 4 calls and 4 receipts; input 44664 (cached 25540, uncached
19124); output 4042 (reasoning 2270, ordinary 1772); cache-write, money,
agent-active-time and tool-read counts null. The new observer
independently verifies that historical receipt as completed
provider and runtime cost only, not semantic plan acceptance and not a
resolution of any current effect. The original run binary `49afbc8` and
its SHA remain intact. The new observer binds to base `49afbc8` plus
the collector repair under candidate binary SHA-256
`4bb4bdf7a9ec428ce99d04bb59cd05eff4395de05d5afe66549055edacfbc384`;
the dirty binary claims no committed v34 SHA.

## Measurement scope and non-claims

Cached input is a subset of input; reasoning output is a subset of
output. Cache-write tokens, monetary cost, agent-active-time and total
filesystem and tool-read counts are unavailable for all rows, not zero.
Measured wall time runs from dispatch to CLI process termination, not the
exact READY transition. Admitted manifest file counts are not total
filesystem reads. Outcomes are 1 of 2 accepted per arm; repairs differ
and one treatment never exposed the selector, so no aggregate
efficiency claim is made. KEEP THE LEGACY DEFAULT and RRF OPT-IN. No
general, causal, speedup, calibration, default-promotion, release or
eight-task qualification is claimed. Prior historical evidence stands.
This record is capability evidence linked to the v1.1.34 repair, not a
governance work package.

Independent Muse High review and parent retained-journal observation are
required before publication.
