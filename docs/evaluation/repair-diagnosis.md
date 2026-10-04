# Repair diagnosis increment — 2026-10-05

Development increment: v1.1.7, based on
`8ebffe564ba0a9d781698e994b3136ad055c7b86`. This is source and read-only
inspection evidence, not installed qualification or model-quality comparison.

## Executed checks

| Check | Result and scope |
| --- | --- |
| `go test ./internal/control -run '^TestRepairDiagnos' -count=1 -timeout=120s` | PASS: native/reviewer source truth, complete-evidence identity, candidate drift, exact gate/write scope, pending/unavailable distinction and path ambiguity |
| `TestRepairDesignPathRevisionWriterAndFreshGatesReachReady` | PASS: real Git, replay-validated repair design, exact ready repair scope and existing fresh verification/review path |
| `go test ./internal/cli -run '^TestLocalPlanApprovalAndResume$' -count=1 -timeout=120s` | PASS: repository-bound diagnosis, no invented findings and byte-identical journal export before/after |
| `go vet ./internal/control ./internal/cli` | PASS |
| Independent GPT 6 Luna High review | APPROVE after repairing the initial missing REPAIRING-state gate; reviewer did not execute tests |

The missing state gate was reproduced as FAIL with the original condition,
then corrected without changing graph readiness, finding/gate checks or scope.

## Retained real-run inspection

The command was applied to the two preserved
[parser-canary](agent-context-parser-canary.md) journals using `go run` from
the development working tree. No model was invoked and no run was resumed.

| Original run | Diagnosis observation |
| --- | --- |
| `e3a33059c2ca56fe72ce09fb84cf86a938081aa2b14d0de1d12f4aa7a57118f9` | 1 native finding, 2/2 consumed repairs, 0 ready repair specifications; budget remains exhausted |
| `3126a4109216629c1ff093d192bb8721e649743ca3e2b96579374aef7ad4cbf9` | No recorded candidate, 0 observed native/review findings, 0 specifications; original provider effect remains unresolved |

Private report SHA256 identities:

- Disabled: `54b2645f81352fb570c3fdf90a59e327a91777e43159cdd3488afac4c0edb91d`;
  controller head `670ff3f6c57426f1c66034bdb36e18e67edfb4935b882a6ee8562ac3e02d73a8`.
- Enabled: `0788b9a629ac65cb0ae71d391bba89657d2b2d02e9dcc9041689ac4fb7212f3f`;
  controller head `ee99529f040aeff49f50e3d7424343f2233ba4424f225928bf517935c9ddfdbe`.

An empty diagnosis does not prove provider completion or task success. Original
FAIL/UNKNOWN outcomes are retained. The report establishes inspectability;
automatic preimage validation, strategy execution, exact finding closure and
live repair-efficiency improvement remain NOT RUN/not implemented here.
