# Agent-context parser canary — initial result

**NOT QUALIFIED — 2026-10-05.** The frozen [experiment](../../evals/v2/agent-context/README.md)
compares explicit Disabled/Enabled agent context. It is separate from the
canonical v1 cohort. Neither arm produced an accepted candidate, and the
treatment did not exercise model quality. No latency, token-saving or context
quality claim is supported by this pair.

## Frozen inputs

- Clean Fabric source `24aa779b27891b890e6be1df61b37e16b3b3cbc3`, embedded
  `vcs.modified=false`, Go 1.27.1; binary and per-arm configuration hashes are
  retained in the [machine-readable receipt](agent-context-parser-canary.json).
- Identical task source commit `a226ccf3eec4116694f6c25c5ccd19225a542281`,
  tree `630c24ad37264e902a1532feafff8c219916afb2`, model `gpt-6-luna`, high effort,
  max parallel 1 and two repairs. Fresh roots/run identities differ by design.
- Held-out SHA-256 `afa9213e52bd8c17c15809c264c552a3a4f4469af51d5c1d0a16c5cf9ba579a1`.
  Its initial missing-API negative control passed. Independent review requested
  UTF-8, default-cap and oversize-redaction assertions before any dispatch;
  the strengthened oracle received APPROVE.

## Raw outcomes

| Arm | Run | Outcome | Native | Review | Held-out |
| --- | --- | --- | --- | --- | --- |
| Disabled | `e3a33059c2ca56fe72ce09fb84cf86a938081aa2b14d0de1d12f4aa7a57118f9` | Repair budget exhausted; candidate retained | FAIL | NOT RUN | NOT RUN |
| Enabled | `3126a4109216629c1ff093d192bb8721e649743ca3e2b96579374aef7ad4cbf9` | Effect requires reconciliation; no candidate | NOT RUN | NOT RUN | NOT RUN |

Disabled first failed because its new decoder left an unused `strings` import.
Later native failures concerned generated tests and record limits. Seven runtime
invocations completed. Both repair opportunities were consumed; no extra repair,
manual candidate edit or favorable retry was performed.

Enabled retained a valid source-bound guidance bundle, host/thread preparation
and `runtime.turn-intent`, without a durable turn handle or result. The controller
therefore retains uncertainty. A deterministic wire regression reproduced the
local defect: the prompt's skills schema was outside the Codex adapter's old
schema allowlist. This diagnosis does not settle the original effect or authorize
resending it. The invocation has no admitted semantic result or usage.

## Observed resources

Disabled: 552,275 input tokens, including 347,136 cached and 205,139 uncached;
15,538 output, including 5,243 reasoning. Cached input and reasoning are subsets,
not additional totals. All seven invocation accounting rows were observed.
Elapsed run time was approximately 685 seconds, including two repairs.

Enabled: usage unavailable, one incomplete invocation; do not report zero tokens
or money. Provider request counts and monetary cost are unavailable in both arms.
No accepted-task efficiency comparison can be made.

## Correction and remaining evidence

The schema is now generated in `engineeringplan` and consumed by both controller
and Codex wire admission. Legacy schema bytes remain unchanged. The adapter's
finite allowlist includes exactly the previous skills augmentation; mutated
skill bounds are still denied before RPC. No recovery/settlement machinery or
authority exception is introduced.

The initial failing wire test was followed by passing legacy/skills wire,
substitution and controller parity regressions. Package tests and vet are scoped
source evidence. A new admissible real-task comparison is still required before
claiming agent-context model quality or selected-skill benefit. The original
uncertain run remains preserved and must not be resent.

Private artifacts are retained under
`D:\fabric-ci2-tools\v2-agent-context-canary-20261005`; no prompts, credentials or
raw provider responses are published in this report.
