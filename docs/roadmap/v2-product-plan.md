# Fabric v2 product implementation

## Goal and publication

Complete the attached [user brief](v2-user-brief.md) through usable, qualified
engineering capabilities. The brief is a proposal; donor claims and its old
`de61e6f` baseline are not current product qualification evidence.

Publish independently reviewed increments on `main` with commit subjects
`v1.x.x` (or `v1.x`), author and committer
`automation:fabric-v1-sol-supervisor <automation@fabric.invalid>`.
Keep fixes, including Sonar fixes, with the increment they repair. Preserve
existing release identities and frozen evaluations. Publish `v2.0.0` only
after the full applicable capability and release gates below pass.

## Capability gates

| Increment | Deliverable | Required evidence |
| --- | --- | --- |
| v1.0.x | Finish current token/prompt hardening and eliminate demonstrated regressions | Full Go/PowerShell tests, legacy journal replay, independent review, Sonar/security checks, actual input/cached/uncached/output counters |
| v1.1 | Unified engineering intelligence over existing RI: per-file facts, immutable base plus candidate graph delta, import/call/generator/test/impact relations, semantic queries, bounded context compiler and planner evidence | Retrieval fixtures before end-to-end evaluation; fresh unchanged fixed-six 6/6 plus godotenv and generated-code ownership; provenance/coverage/corruption tests; measured benefit before default adoption |
| v1.2 | Coupling-aware decomposition and resource-vector scheduling, isolated cluster worktrees, serial integration, impact-aware bounded review | Current-versus-new topology A/B; shared-hub/generator ownership, resource ceilings, conflict isolation, combined verification/review, no duplicate external effects |
| v1.3 | Content-addressed parse/fact/context reuse and opt-in deterministic verification observations | Exact keys/input closure/toolchain/environment binding; corruption/invalidation/eviction tests; cold/warm latency, CPU, RAM, disk, token types and time-to-READY; final release checks fresh |
| v1.4 | Evidence-calibrated model allocation with conservative fallback | Real task/role/risk traces, measured success/cost/latency, calibration and unseen-task evaluation; no invented priors presented as measured results |
| v1.5 | Verified checkpoints and fresh-context long-horizon continuation | Resume from accepted durable state, interrupted/pending/UNKNOWN effects preserved, independent verification, runtime-native compaction qualification on a sustained coding workload |
| v2.0.0 | Integrated autonomous engineering release | All applicable rows qualified, fresh full regression/evaluation evidence, independent final review, green hosted checks, documented install/task/inspect/resume happy paths and published release artifacts |

Research repomap/SCIP/Serena/Aider, Co-Coder/Agentless, Pact/pi-subagent-tasks/CAID,
OpenCodeReview, Bazel/Nix, Ultraswarm/router work, LongHorizon, OpenHands/SWE-ReX
and Symphony individually. Record pinned sources, license compatibility,
mechanisms and limitations. Reuse concepts without creating a second RI.
JEV and later runtime integrations require demonstrated benefit; speculative
vector databases, graph databases, distributed caches, agent DSLs, owned
sandbox/conversation-memory systems and governance-only work are excluded.

## Validation and invariants

Run Go with the retained toolchain:

```powershell
& 'D:\dev\EngOrch-toolchains\go\1.27.1\go\bin\go.exe' test -p 2 -timeout 20m ./...
```

Run relevant Rust and PowerShell regressions for each changed capability.
Real evaluations retain pinned repositories/objectives, model/runtime identity,
repair budget, native verification, independent review and unchanged held-out
acceptance. Never delete, skip, weaken or narrow tests to obtain PASS. A changed
test contract needs explicit objective support and independent evidence.

Direct product edits are allowed. Self-hosting is optional. Preserve exact
effect ownership; UNKNOWN stays UNKNOWN and uncertain effects are never
resent. Cache deterministic observations rather than authority or LLM truth.
Version invocation-affecting recipes to preserve legacy replay. Every candidate
must satisfy review/integration as well as executable verification.

## Starting checkpoint — 2026-10-03

- Public main: `637e64665426e49824c8aebf9e14ef9c1f547eef`, source-identical to
  the old `de61e6f` baseline after the author-only history rewrite.
- Public dev and local main: `84e8a7732174faf5bd7125e06f808d99facb691d`
  (`v1.0.1`); draft PR #14 was created by `github-actions[bot]`. The local
  main checkout has additional RI work in progress. Public main is unchanged.
- Released v1.0.0 tag remains `a4ca05692465bedc6b98d6ffbb8219b6d076f5d1`.
- Completed feature-only planner A/B: baseline 5/6, treatment 4/6; optional
  source-pack prototype is not a qualified intelligence architecture.
- Both supplements: godotenv PASS, numeric generated-code task BLOCKED.
- Fresh hardening evaluation of clean `84e8a773` with `cache-prefix-v1`:
  `20261003T143213Z-59b9afcd` / `eval-20261003T143556Z-5aaf753d`, **0/6 PASS**.
  Five blocked before writer dispatch because context batching selected
  protected workflow files. Difflib reached native verification and exhausted
  its two repairs with one remaining context-diff delimiter defect. Preserve
  these journals; none is resumed or resent. The workflow-selection repair
  passed focused task-context tests and independent review; fresh evaluation
  of the repaired candidate is still required.
- Full Go suite completed PASS (CLI 128.289 s, control 1012.984 s). This run
  began before the new RI edits and does not qualify those later edits.
- SonarCloud and GitGuardian checks on the public hardening candidate passed.
  Go facts, cache validation and incremental graph work still require focused
  tests, real cross-language integration and independent review.

This checkpoint is not v1.1 or v2 completion. Record implemented capabilities,
exact candidate/check identities, failures and remaining work after each
integrated increment, without resurrecting a G0 promotion ladder.

## Increment checkpoint — v1.0.1 and intelligence foundation

- `v1.0.1`, `61ed719c951e50b5d416ff6d73b36edbf6edcc78`, is integrated
  on public main. Author and committer are the automation identity above.
  Its clean full Go suite passed, as did independent review, SonarCloud and
  GitGuardian. PR #14 was created by `github-actions[bot]`.
- Fresh native evaluation `20261003T151841Z-919a36f9` /
  `eval-20261003T151947Z-fdd8d561` completed **5/6 PASS**, with unchanged
  held-out acceptance and two repairs. Difflib exhausted its budget; its final
  native failures are `ExampleGetUnifiedDiffCode` and
  `ExampleGetContextDiffCode`. No infrastructure fault or uncertain effect was
  observed. This journal remains terminal and is not resumed.
- The next increment is an opt-in RI foundation: Go syntax facts with explicit
  trusted-local derived caching; immutable source/candidate graphs; bounded
  symbol/import/call/impact queries; compact context compilation and committed
  source CLI commands. Rust, Go and cross-language fixtures and independent
  review qualify this foundation only. Calls remain unresolved and coverage
  partial. Automatic package discovery and planner integration, real task
  benefit and the v1.1 6/6 gate remain outstanding.
- Earlier frozen evaluations and release tags retain their identities. The
  workflow-file selection failure in the previous 0/6 run is repaired; the
  newer 5/6 result does not erase the older result or establish v1.1 completion.

## Current checkpoint — Go planner context candidate

- Public main includes the reviewed RI foundation `v1.0.2`,
  `cd5240d780de294d23cb34d648255af6960ae714`. Its clean full Go suite,
  Rust/cross-language tests, independent review and hosted checks passed.
  Detailed original and amended identities are in the foundation evaluation
  note; Go test cache reuse is stated explicitly.
- The next candidate wires immutable, opt-in `go-source-context-v1` into
  planning. It adds selected-only Git batch reads with size preflight,
  source-local package identities and context v2 caller/contract excerpts.
  Default and old source-context identities remain covered by regressions.
- All six committed corpus/controller admission and replay probes pass without
  provider effects. Graphs are at most 512 KiB; source excerpts together at
  most 48 KiB; complete admission records at most 768 KiB. This is offline
  preparation evidence, not six successful coding tasks.
- Fresh native fixed-six and supplement evaluation, measured model usage and
  benefit, and remaining v1.1 gates are next. Resource-vector selection is
  separate unintegrated work for v1.2; isolated worktrees, serial integration
  and topology A/B are still outstanding. Later roadmap rows remain active.

## Increment checkpoint — topology query

- The next independent increment exposes source-bound package dependencies,
  observed import hubs, explicit generator coupling and bounded impact review
  groups through `fabric ri topology`. Query and CLI independent review and
  focused actual-Rust integration tests pass. Full clean suite/hosted checks
  still gate public integration.
- `v1.0.3` completed its full clean Go suite. Its frozen eight-task evaluation
  completed 6/8 PASS: fixed suite 5/6 and godotenv PASS, difflib held-out FAIL,
  numeric generator task BLOCKED after two repairs and final review rejection.
  Matched-five token usage increased; the mode is not promoted to public main.
- Isolated writer/resource-vector integration is separate uncommitted work.
  Its new policy preserves legacy JSON with omitted pointer fields; runtime
  dispatch, combined verification and real topology A/B still require evidence.
  No v1.2 or v2 completion is claimed.

## Candidate checkpoint — isolated writers and bounded generation context

- Exact clean `v1.0.4` (`514616c4a31be65f5d71e127c395255b22ec9536`)
  passed the full Go suite and hosted run `37139837155`. Public dev contains
  it; public main retains `v1.0.2` while coding acceptance benefit remains unmet.
- The next candidate has explicit resource-vector admission, confirmed child
  worktrees, child-bound proposals, one parent aggregate, and serial parent
  repairs. Three-child combined gates and rejection/repair/restart fixtures
  reach READY. Independent isolation review approves the code. Accepted
  R17/R19 process qualification remains NOT RUN without a configured fixture.
- `go-source-context-v2` preserves old recipes and uses a separate record
  identity, at most 24 KiB Go context plus 8 KiB complete generation-supporting
  files. Source/seed/commit/graph bindings and durable JSON forgery regressions
  pass with the actual pinned RI engine. Large files remain explicit omissions.
- Fresh graph validation reconstruction avoids duplicate hashing, with about
  25% less allocation in bounded synthetic benchmarks and unchanged identities.
  This is not peak-RAM or model-token evidence.
- Full clean final-candidate suite, hosted checks, fresh unchanged eight-task
  evaluation, real topology A/B and remaining intelligence/cache/routing/horizon
  capabilities are still required. No v1.1, v1.2 or v2 completion is claimed.

## Frozen checkpoint — v1.0.5 and the next module capability

- `v1.0.5` is frozen at `856ccb0412090a8c065a2510794e1c01370b3b3c`,
  with automation as both author and committer. Its complete clean Go suite
  passed, including control (790.445 seconds), CLI and RI. Hosted run
  `37144231404` passed; unresolved Sonar issues for bot-authored draft PR #16
  were zero at the check. Public dev contains this candidate; public main
  remains at `v1.0.2` pending demonstrated coding acceptance benefit.
- A fresh eight-task evaluation uses the clean v1.0.5 binary
  `35c0057fa0088046a18107fb0993e36dc19539f5b9dbf6d4d7486bc324514399`
  and pinned RI engine
  `c86244aa6becf439eb770b1f4d4c0076ae65411ef15353d4f9300e1a74520ec0`.
  Preparation used no providers. Evaluation `eval-20261003T184316Z-3951966f`
  is in progress with unchanged objectives, held-out checks and repair limits;
  it does not enable isolated writers. No final PASS count is available yet.
- The next independent capability observes committed Go module/workspace/vendor
  manifests, derives conservative declared package ownership, and connects it
  to the existing RI graph and bounded corpus. It preserves the legacy mode
  and rejects module-backed candidate overlays until manifest closure exists.
  Dirty manifests cannot change committed-source answers. Independent review
  approves the helper/graph/corpus slice; full RI with the actual pinned engine
  passed in 17.884 seconds. CLI exposure, final combined qualification and
  public integration are still pending.
- A separate isolated-writer treatment prepared two pinned tasks without
  provider calls and used the same v1.0.5 binary, context mode and gates, with
  explicit resource estimates and `MaxParallel=3`. Frozen evaluation
  `eval-20261003T190119Z-b6bfcb7a` completed 0/2 PASS: both runs BLOCKED in
  IMPLEMENTING. This is a failed treatment, not a parallelism speed/quality
  improvement. Its journals are preserved; diagnosis is read-only and these
  runs are not relaunched. The serial eight-task evaluation remains separate.
- Module CLI review found that sensitive/protected manifests could enter the
  selected-source batch. A bounded path-policy fix and regression fixtures
  are required before this module increment can be integrated.
- The module path-policy repair is now implemented: sensitive/unsafe paths
  are redacted before source batching, and omitted manifests supply no content
  declarations or blob/hash metadata. Ownership remains ambiguous where a
  manifest was omitted. The final focused pinned-engine module/corpus/CLI
  regressions passed (CLI 2.272 s, RI 6.041 s); unfiltered doccheck passed
  (0.302 s). Module CLI previously passed its full suite (82.468 s).
- Final independent module/CLI review approves the privacy repair and committed
  inventory re-collection. The next increment is ready to freeze for a clean
  full-suite run. The separate graph-only semantic adapter is reviewed, but is
  retained for the following increment and does not change this freeze.
