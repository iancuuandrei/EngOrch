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

## Checkpoint — v1.0.6 frozen; v1.0.5 evaluation complete

- Module inventory/CLI increment `v1.0.6` is frozen at
  `a315e42acd6c2b4d401426c6be485a8b6e969b09`. Both Git identities are
  automation. Its full clean suite runs in a separate detached clone; current
  semantic-query/cache/preflight edits cannot change that candidate.
- Serial v1.0.5 evaluation completed 6/8 PASS: fixed-six 5/6, godotenv PASS,
  difflib held-out FAIL, numeric BLOCKED with two exhausted repairs and final
  changes-requested review. Journals and frozen results remain unchanged.
  [Measured categories and failed-arm diagnosis](../evaluation/v1.0.5-generation-and-isolation.md)
  distinguish model tokens, runtime receipts, unknown billing and failed tasks.
- No context default or model allocation is adopted from these outcomes.
  A fresh two-task single-variable model comparison is prepared with the same
  clean v1.0.5 binary, objectives and gates, using GPT 6 Sol High. It is a
  Fabric evaluation runtime, not a development subagent. Development subagents
  remain GPT 6 Luna High. No success or calibration is claimed before evidence.
- The next independent increment exposes graph-only semantic queries, explicit
  trusted-local parse-cache reuse and preflight rejection of isolated runs
  without an external state root. Final review/qualification/integration remain
  required. The complete v2 goal remains active.
- Clean v1.0.6 full-suite PASS: control 871.592 s, CLI 85.858 s, RI 30.544 s.
  Public dev now contains the exact candidate, and hosted run `37148085280`
  passed publish-draft, Sonar and security checks. Unresolved PR #16 Sonar
  issues were zero at the check. Public main still retains v1.0.2.
- Independent review approves the semantic adapter/CLI, isolated preflight
  and parse-cache slice under its explicit trusted-local cache assumption.
  Final combined focused regressions passed (CLI 10.869 s, RI 1.372 s), the
  NativeRunArgs PowerShell harness passed, and unfiltered doccheck passed
  (0.308 s). The cache benchmark excludes priming and setup: 24 cold misses
  versus 24 warm hits, similar allocation, and only a small fixture timing
  improvement. CPU and RSS are not yet measured; no general speedup is claimed.

## Checkpoint — v1.0.7 qualification and candidate graph work

- `v1.0.7` originally froze at `9762199f2a7b72bcc821974ca9b914b94a4c2265`.
  Its complete clean Go suite passed: control 786.264 s, CLI 95.519 s,
  RI 26.195 s. Hosted run `37149671650` passed all three checks.
- A direct Sonar issue query nevertheless found one PowerShell null-order
  issue. The same increment was amended to
  `09f71df18af1c86a8442b44f78a67d2809104256`, changing only that harness
  comparison. The harness rerun passed and independent review approved.
  All Go/product trees are identical; hosted checks for the amendment are
  tracked separately. This is not a rerun of the full Go suite at the new SHA.
- The separate Sol High experiment finished 1/2 PASS: numeric generated
  ownership passed with two repairs, difflib failed held-out acceptance.
  [Typed measurements](../evaluation/v1.0.5-generation-and-isolation.md)
  do not establish a calibrated routing default.
- Candidate module inventory and graph overlays now bind observations to
  actual candidate manifest hashes, files closure and committed base digest.
  Independent review approves these bounded library slices. CLI exposure and
  final combined qualification remain pending; no v1.1/v2 completion is claimed.
- [Public objective v2](../evaluation/public-objective-v2.md) repairs a missing
  pre-existing public difflib requirement. Historical results remain unchanged;
  fresh comparisons must use the same new objective on both arms.
- The amended v1.0.7 hosted run `37149801709` passed all three checks;
  direct PR #16 Sonar issue enumeration returned zero unresolved issues.
  Public main and dev now both contain `09f71df18af1c86a8442b44f78a67d2809104256`;
  bot-authored PR #16 is merged. This integrates the reviewed optional
  capabilities and normal-use preflight fix, without adopting experimental
  context defaults or claiming any milestone acceptance increase.
- The v1.0.8 candidate exposes `ri candidate-query` over a confirmed run's
  exact candidate, with module ownership rebuilt from candidate manifests.
  Omitted changed files cannot retain old base facts. Candidate reads are
  bounded to eight one-MiB Go files, held under verified read leases, with
  explicit PARTIAL coverage and redacted sensitive omissions.
  Independent combined review approves the module/overlay/corpus/CLI and
  objective clarification. Pinned-engine full CLI passed (99.699 s), candidate
  corpus regressions passed (26.815 s), vet and doccheck (0.308 s) passed, and objective/native-argument/usage
  PowerShell harnesses passed. Clean full-suite and hosted checks remain
  required before integration. New cache/checkpoint work is a later increment.
- The increment's Sonar duplication gate required extracting shared test setup.
  A real pinned run also exposed an invalid empty-graph expectation in one
  fixture. The amended regression separately proves omitted changed facts are
  removed while unrelated facts survive, and an all-omitted candidate rejects
  with no usable graph. Independent review approves both changes. The original
  full-suite run does not qualify this amendment; a clean rerun is required.

## Checkpoint — v1.0.8 integrated; v1.0.9 candidate

- v1.0.8 at `16174ba460f7aacaea0238b87552e8ff9ec04214` passed the clean
  full Go suite (control 918.712 s, CLI 153.859 s, RI 72.078 s), hosted run
  `37151367477`, and direct Sonar enumeration (zero issues, 0.7% duplication).
  Public main/dev contain that SHA; automation-authored PR #17 is merged.
- v1.0.9 adds opt-in committed-source planner parse caching, payload-free
  `checkpoint RUN` inspection, and the evaluator's empty controller-root
  replacement fix. Independent review and focused regressions passed.
  Cache replay preserves admitted evidence; final candidate checks remain fresh.
- Five exploratory samples showed synthetic warm-cache median 605.9 ms versus
  uncached 759.3 ms; pinned humanize warm 951.6 ms versus uncached 951.4 ms.
  No material allocation reduction or representative speedup is established.
  The cache remains optional. CPU/RAM sampling and clean frozen qualification
  are pending. See the cache and checkpoint guides for limits.
- v1.0.7 serial/isolation evidence remains 1/2 PASS and 0/2 BLOCKED respectively.
  The latter stopped before run creation due to evaluator configuration;
  no failed or exhausted run was resumed. Future arms require new identities
  and the same public objective v2 on both arms.
- Contract evidence compilation and automatic compaction observation are
  separate unintegrated candidates. The v2 goal remains active and fixed-six
  acceptance remains 5/6 until fresh qualifying evidence proves otherwise.
- Frozen v1.0.9 hosted run `37153276332` and PR #18 Sonar/security checks
  passed; direct Sonar reported zero issues and 1.3% duplication. Main
  integration still awaits the clean full-suite result.
- Frozen CPU/RAM diagnostic (five samples) observed process-tree peak working
  set 290,803,712 bytes and maximum sampled aggregate CPU 31,734.4 ms.
  Sampling overlapped the full suite; it is not an isolated timing comparison.
  Pinned humanize allocated roughly 384–389 MB per collection despite only
  52,828 source bytes. A memory profile traced most allocation to rebuilding
  and hashing each candidate graph prefix; a separate repair is under review.
- The v1.0.9 checkpoint CLI read the retained real humanize READY journal
  `69712c626440b103279f224b9cb8d8a84c254bcc94eb58261f40d058a691760e`:
  accepted=true, three completed tasks, zero active/uncertain intents,
  approved review; journal main-file hash was unchanged. The exhausted numeric
  journal `54311360b3336553475f480104525c1f1d94b25f258be6a89014a1d1839dcff1`
  remained REPAIRING, accepted=false, zero active/uncertain intents. These are
  read-only compatibility observations, not a sustained resume qualification.
- v1.0.9 integrated on public main/dev at
  `4ec8d0252332618c22a316cc9653d20eaef61497`; bot-authored PR #18 merged.
  Clean full Go suite PASS with exit 0: control 882.817 s, CLI 106.204 s,
  RI 70.651 s, worktree 42.762 s; log `v109-clean-full-tests.log`.
  Frozen PowerShell native-argument/objective harnesses and doccheck passed.
  Fresh serial/isolation prepared cohorts now proceed under the unchanged
  objective-v2 task pins, gates and two-repair budgets.

## Checkpoint — v1.0.9 comparison complete; v1.0.10 candidate

- The fresh v1.0.9 serial arm passed Humanize and numeric-text (2/2),
  including generated-output ownership. The isolated arm passed Humanize but
  numeric-text exhausted two repairs with its last native check failing (1/2).
  No uncertain effects remain in that failed run; it is not resumed.
  See `../evaluation/v1.0.9-isolated-writers.md` for frozen identities and
  input/cached/uncached/output usage. These two tasks do not replace the
  fixed-six cohort or qualify isolated scheduling.
- A measured graph-size counting repair preserves the original greedy file
  selection and exact graph/context identities. Three local Humanize samples
  reduced median allocated bytes by 19.5% and allocations by 28.9%; a more
  aggressive batching prototype regressed time and was discarded. The repair
  still requires clean qualification of its released increment.
- The next candidate adds opt-in contract evidence, explicit operator-owned
  fixer access configuration, bounded objective-focus topology queries and
  metadata-only native compaction observation. Contract replay must reproduce
  real source and generation inputs before provider evaluation. Native
  compaction remains unqualified until a useful live task demonstrates it.
- The next context comparison uses one frozen candidate binary and objective
  v2 on both arms, changing only the context mode across the six fixed tasks,
  godotenv and numeric-text. Budgets and held-out checks stay unchanged.

## Checkpoint — v1.0.10 integrated; v1.0.11 candidate

- v1.0.10 is public on main/dev at
  `fbd1e5d60760d213edc017761091dbaf4e414838`, automation PR #19 merged.
  Clean full Go, independent review, native-argument/doc checks, security and
  Sonar passed (zero issues, 2.1% duplication). See the increment qualification
  record for frozen CPU/RSS, cold/warm disk and allocation observations.
- The contract-context candidate now passes offline admission and durable
  replay on all eight pinned repositories with zero provider effects. Its
  record retains only the longest fitting ranked source prefix and the exact
  literal generator/tool/output source closure. This avoids duplicating
  irrelevant generator-discovery inputs while preserving source-bound replay.
- The numeric context retains the actual generator owner, template, tool and
  output evidence. The new planner prompt derives bounded coupling/test hints
  from that exact graph and contract selection. These are planning hints,
  not an independence certificate or write authority.
- An optional, invocation-bound Codex native-compaction threshold is under
  independent review. Default identities are unchanged; configured limits do
  not prove effective runtime thresholds or actual compaction. Live sustained
  qualification and all fresh provider comparisons remain required.
- A usable topology comparison is specified using the existing Humanize
  feature/performance task and both unchanged held-out checks. A changed
  generator hub followed by isolated leaf stages is a separate unsupported
  design, and is not allowed to drive new recovery machinery.

## Checkpoint — v1.0.11 integrated; v1.0.12 candidate

- Public main/dev contain `3aa15c41b65a0dbae62de54c1908c919fed1c56c`;
  automation-authored PR #20 is merged. Final clean full Go, independent
  review, native-argument/public-objective checks, Sonar and security passed.
  See the increment qualification record for exact candidate and binary pins.
- Two fresh eight-task context cohorts passed native preflights and expected
  held-out baseline discriminators with zero effects. The contract arm is now
  evaluating the frozen increment; no new task acceptance is claimed yet.
- The next increment extends existing exact per-file parse reuse to the
  contract context, with uncached/cold/warm identity and prompt parity. It
  also exposes the invocation-bound native-compaction threshold through the
  matched evaluation runner. Both remain opt-in, with live qualification
  and measured resource benefit required before broader adoption.
- A candidate-bound, bounded impact projection for reviewer guidance is being
  implemented separately. It supplies observed coupling and possible tests,
  without claiming complete dependencies, independence or successful checks.

## Checkpoint — v1.0.12 integrated; impact context candidate

- Public main/dev contain `e447adf4dceec58fba85dd19e5a390667ee79f6b`;
  automation PR #21 merged. Corrected clean full Go, independent review,
  native-argument/public-objective checks, Sonar and security passed. The
  obsolete cache-policy test correction is in that same increment.
- The frozen v1.0.11 contract cohort completed with fixed-six 5/6 PASS,
  godotenv PASS and numeric-text BLOCKED after two permitted repairs. Logr's
  native verification/review passed but its held-out acceptance failed. The
  exhausted numeric run is preserved. The matched v2 arm is evaluating; no
  new default or comparative quality improvement is established yet.
- Typed usage includes failed work: a read-only supplement captures numeric
  usage skipped by the runner's nonzero-run path, without modifying the
  frozen receipt or redispatching. Cost and physical provider-call counts
  remain UNKNOWN.
- Candidate-bound reviewer impact context has passed independent RI and
  control review. It remains opt-in and source-free, with durable bounded
  replay and explicit partial/unavailable coverage. Clean full qualification
  and a fresh provider exercise remain pending.
- A graph-sealing optimization under review reduced allocation bytes by
  8.85% and allocation count by 9.27% in three alternating paired samples
  against the same frozen source. Mean elapsed time increased 1.61%, with
  overlapping ranges; this is not a speed improvement claim.
- Real topology A/B preparation, routing-decision evidence and exact derived
  context reuse are the next product work. Adaptive policy calibration,
  sustained fresh-context/native compaction qualification and v2 release
  gates remain open.
