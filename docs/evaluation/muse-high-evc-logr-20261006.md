# Muse High EVC logr acceptance 2026-10-06 (one task, scoped)

A genuinely distinct frozen public go-logr/logr task, not a replacement or
retry for any UNKNOWN attempt. Exact clean Fabric source
`b58a649d81c468e92656966d54b02c951d7a5eae`, v1.1.29 clean binary SHA-256
`98b2e5444030f30842eff7bb87c84d1b2e424a93c11644e59c53ff6ccdade0e8`,
pinned Go 1.27.1. Earlier UNKNOWN attempts are preserved untouched: no
settlement, resend, or cohort pooling.

## Configuration and limits

Public task source `e99cde667024c0da5e20141f88fb7952e5f4e582`, config SHA
`c651a309b4cd6ca3f48fb88ac7b3e1cf6a99d948b921635c57f15649766b439f`,
policy SHA `8175b51927ab8380ac1f423a4cf24a65ffaf4c0c940f1b117c11564d562b7b33`,
unchanged public-v1 oracle SHA
`290eb12c1448c4b611dfb0ed2aa84a945fa6e4fdfd98746c0aa9435e0710bf51`,
frozen plan SHA
`e267dea720cdc656d5b243b797a2ea8983ebdc46023ec9d84fa9f35bdc271a86`.

All five configured roles `muse-spark-1.3-contributor` high over the
opencode-http/Responses/Go route with SystemRole `developer`, reasoning
effort high, summary auto, include `reasoning.encrypted_content`. Bounded
dispatch: max-parallel 1, max-repairs 2, max 16 calls per invocation,
invocation timeout 600 s, overall frozen wall limit 3600 s. Native baseline
PASS 0.190 s; `doctor` and `--inspect-plan` PASS after the private state root
was initialized before the freeze.

## Finite EVC handoff

One finite `source_read` acquisition with action ID `inspect-withvalues`,
selected once before the initial writer; caller-estimated local compute is
unvalidated and claims no benefit. The acquired `logr.go` (exact hash
`cce9b9499b3111c29e6cc9a816618b4af3a70fa6d44667134215477191f976bc`,
excerpt hash
`2b15b03ef2c9424847e81762d27c5460d5663c44af22a73df4d39740bf25424b`)
appears in the writer context at the same initial candidate.

## Run, candidate and oracle

Run `5dc719f58cb3d71309ef6d644502bc4e49ef64404a83294eb7b1f291197c7728`:
wall 177.8714694 s, state READY. Planner 3 calls/receipts, writer 6,
reviewer 5; all succeeded with observed model
`muse-spark-1.3-contributor`. No fixer turns and zero repairs. Native PASS,
independent review APPROVE with zero findings. Candidate
`2401f06c2996603f5ff7ba1597b3f24665f38f14f2c1abbf7d64aad6dee6e480`:
exact reviewed-candidate copy guard PASS, then the unchanged public-v1
external oracle plus the full copied package test PASS 0.199 s. The model
never received the oracle.

## Observed usage (gateway receipts)

| Measure | Tokens |
| --- | --- |
| Input | 386751 |
| Cache-read input (subset of input) | 289966 |
| Uncached input | 96785 |
| Output | 8470 |
| Reasoning output (subset of output) | 6037 |
| Ordinary output | 2433 |

14 provider calls with 14 exact receipts; no unmatched call in this fresh
run. Cache-write tokens unavailable, monetary cost UNKNOWN, and
agent-active-time and tool-read totals unavailable. Subsets are never
double-counted.

## Evidence and non-claims

Machine-readable receipt:
[muse-high-evc-logr-20261006.json](../../evals/v1/results/muse-high-evc-logr-20261006.json).
Private raw full snapshots and receipts are retained; no paths, prompts,
ciphertext, or secrets appear here. This is one task's evidence at the exact
source and observed roles only: not a canonical cohort, not model-quality
generalization, not an efficiency improvement, not calibrated JEV, not final
v2, and not installed-release qualification.
