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

## Checkpoint — v1.0.13 integrated; bounded cache and routing evidence candidate

- Public main/dev advanced to `2459ab9d863386cae16b158ddc505f65b9397951`
  after clean full Go, independent review, hosted security and Sonar passed.
  The concurrent directory race correction belongs to that same increment.
  See its qualification record for exact pins and retained original failure.
- The completed frozen context comparison has fixed-six 5/6 in both arms,
  with different held-out failures. Contract godotenv passed; context-v2
  godotenv blocked at repair 1/2. Numeric-text exhausted two repairs in both.
  No passing sets are pooled; no new default context is adopted. Typed usage
  includes failed work, with cost and physical provider calls UNKNOWN.
- The next opt-in candidate reuses exact per-file candidate review syntax
  observations while recomputing source/candidate/module/generator bindings.
  Its RI cache has bounded 1,024-entry / 64 MiB cooperative storage and safe
  eviction. No verification, review authority or external effect is cached.
- Versioned routing decision evidence now binds the serialized input and
  configured route to the original invocation, including settlement paths.
  It records decisions without claiming a calibrated allocation policy.
- Independent reviews and focused Go/Rust/PowerShell checks approved these
  slices. Clean full Go/Rust, hosted checks and actual cache/topology use are
  still required before promotion. Native compaction/fresh-round guidance is
  documented but not live qualification.
- Old v1.0.13 topology preparations had zero effects and are superseded by
  the source amendment. Fresh matched topology runs will use a qualified
  frozen increment. A read-only native model catalog confirmed Sol High for
  a predeclared fixer pilot; no pilot model invocation has been sent.
## Checkpoint — v1.0.14 integrated; strict writer and fixer pilot candidate

- Public main/dev advanced to `66252573e4377f6c9f409bc72cc22a4c034e9d01`
  after clean full Go/Rust, focused Native harnesses, independent review,
  hosted security and Sonar passed. See its qualification record.
- Fresh topology preparations use that exact clean binary and RI b189 with
  matched context/cache/isolation policies. Baseline native checks pass and
  held-out assertions reject the baseline. The serial arm is evaluating;
  parallel is not dispatched until serial completes. Each arm runs once.
- Read-only diagnosis found the earlier godotenv blocker was an existing-file
  entry with no anchored edits. A strict Codex writer contract now prevents
  that form in the provider-facing schema and prompt, retaining historical
  V1/V2 replay and the original fail-closed decoder. The old blocked run is
  preserved; no new recovery or repair permission is introduced.
- The next runner binds optional fixed fixer model/effort and a bounded public
  access policy across Prepare/Evaluate before init. This enables the
  predeclared Luna-versus-Sol fixer pilot; configuration alone is not actual
  route exposure or adaptive-policy qualification.
- Independent reviews and focused regressions approved the strict contract
  and runner, including rejecting conflicting init flags before writes. Their
  clean full qualification and actual model exercises remain pending.
- The logr failure has a concrete retrieval/decomposition cause: root API
  source was omitted and write ownership covered only the example sink.
  A generic receiver/declaration selection repair is being scoped. The
  difflib frozen representation failure remains unchanged; hidden acceptance
  details are not placed in prompts or used to retroactively claim PASS.
- Raw notifications from an exact completed Luna High writer thread report
  an effective context window of 258,400 tokens. Fresh-round and organic
  compaction experiments still need predeclared settings and live evidence.

## Checkpoint — v1.0.15/16 final batch qualification pending

- Public main remains qualified v1.0.14. The original v1.0.15 clean full Go
  suite passed; subsequent PowerShell variable and test fixture duplication
  fixes belong to that same increment. A final clean v1.0.16 suite will cover
  both increments before integration. Old zero-effect v1.0.15 pilot preparations
  are superseded and will not be evaluated.
- Independent review approved strict writer output and the optional fixer
  route, and the receiver-aware v2 contract context. Receiver ranking repairs
  cover unrelated method counts and generic receiver syntax. Root declaration
  admission passes offline; fresh live task acceptance remains pending.
- The v1.0.14 serial topology run blocked after a workspace intent, before a
  workspace outcome or candidate. No materialized workspace was observed; the
  effect remains unresolved and its underlying error UNKNOWN. The old run is
  preserved and will not be resent. Parallel has not been dispatched.
- Nine identity-audited candidate-cache measurements confirm exact warm reuse
  and bounded fixture storage. Overlapping timings and host contention provide
  no speedup or memory improvement evidence. The cache remains opt-in.
- Fixed-six 6/6, generated-code ownership, real topology comparison, calibrated
  allocation, accepted fresh rounds/native compaction and release gates remain
  open. Product fixes do not substitute for those outcomes.

## Checkpoint — v1.0.17 final batch candidate

- v1.0.15/16 remain pending integration over public v1.0.14. Their original
  clean full Go suites passed. The corrected init harness belongs to v1.0.15;
  all ten PowerShell harnesses passed on the corrected v1.0.16 source. Exact
  tested and corrected identities are in the preintegration record.
- The next product correction gives worktree Git commands an explicit
  `core.longpaths=true` under their existing isolated configuration. A fresh
  Windows regression exceeds 260 characters in the branch-lock path and
  passes ordinary creation/observation. Historical uncertain state is untouched.
- An opt-in explicit-manifest Go formatting observation reuses only exact
  deterministic artifacts. It does not satisfy verification, review or READY.
  Generic `go test` remains ineligible for result reuse because its complete
  process/input closure is not controlled. Final checks remain fresh.
- Final code/documentation review and a clean combined batch suite precede
  integration. Superseded preparations have zero provider effects and are not
  dispatched. The next productive live steps are unchanged coding acceptance,
  matched topology/model trials and useful accepted fresh rounds; all broader
  qualification and release gates remain open.

## Checkpoint — v1.0.17 integrated; native usage compatibility candidate

- Public main/dev contain `1e0084198637fba7f347d00a96378695e4f61a2c`.
  Fresh complete Go/Rust checks, all ten PowerShell harnesses, independent
  review, hosted security and Sonar passed. PR #24 merged. The qualification
  record preserves the earlier documentation-check failure and same-increment
  correction; hosted checks alone did not qualify the original source.
- Nine synthetic formatter measurements demonstrate exact warm reuse and
  lower allocated bytes/counts. Process-level memory and CPU are separately
  sampled; no task-latency, model-token, provider-cost or speedup is claimed.
- Fresh fixed-eight preparation passed native baseline checks with zero
  provider calls. The unchanged acceptance cohort remains unexecuted.
- The first useful ParseBytes/native-compaction task blocked during
  implementation when a native last-usage observation had a total-only
  context estimate but zero typed counters. Cumulative consumption remained
  coherent. The old invocation remains active/uncertain and is never resent.
- The next narrow product correction recognizes that native observation
  without changing consumed-token accounting or budget enforcement. Versioned
  normalization preserves historical failure/interrupt replay; an actual
  read-only replay of the old journal returned byte-identical observations
  and the unchanged nonaccepted controller checkpoint.
- Fixed-six 6/6, generated-code acceptance, topology/model comparisons, accepted
  fresh rounds and completed native compaction still require live evidence.
  Code, fixture and preparation progress is not substituted for these gates.

## Checkpoint — v1.0.18 integrated; first completed native compaction

- Public main/dev and the clean frozen source contain
  `605529818af04b9809f25239b80bdd3e3f2d9f9b`; bot-authored PR #25 merged.
  The complete fresh Go suite passed on `76d13a9`; the sole final-source change
  corrected a qualification sentence. Product source, scripts and fixtures are
  identical. Final-source documentation, focused usage checks, all ten
  PowerShell harnesses, hosted checks and independent review passed.
- A fresh unchanged ParseBytes run completed its writer and native verification.
  Retained usage plus read-only SQLite metadata independently bind one completed
  native compaction and its typed counters. The task then failed before reviewer
  dispatch because a required projection array encoded an empty deletion list
  as null. The checkpoint is not accepted and has no active/uncertain intent.
  The exact result is recorded in the v1.0.18 native-compaction evaluation note.
- The next product increment fixes that field with a regression and adds
  opt-in empirical fixer calibration, bounded CLI inspection/initialization,
  frozen configuration and exact objective-bound routing evidence. Matched
  train/holdout outcomes and shared non-fixer policy identity are required;
  UNKNOWN, unexercised, drifted, unmatched and insufficient data retain the
  configured baseline. Static escalation keeps priority. No default model or
  measured quality/cost improvement is claimed from fixtures.
- v1.0.18 fixed-eight preparation passed baseline checks with zero provider
  effects. Final-source product qualification precedes further dispatch.
  Fixed-six 6/6, generated-code acceptance, real topology/model comparisons,
  accepted fresh rounds and pause/resume/release gates remain open. The previous
  v1.0.17 uncertain invocation is preserved and never resent.

## Checkpoint — v1.0.19 integrated; corrected evaluation bindings

- Public main/dev contain `44d935ba833f23d1832dd01a66946a2acee36bb6`;
  bot-authored PR #26 merged. The exact clean source passed the complete fresh
  Go suite, all ten PowerShell harnesses, independent source review, hosted
  checks and Sonar (zero unresolved issues and zero new duplication). Rust
  source is unchanged; no fresh v1.0.19 Rust suite is claimed.
- The optional empirical calibration CLI, immutable objective-bound routing
  and conservative fallback are integrated. They are mechanisms, not measured
  model-quality, cost or default-selection improvements.
- The settled v1.0.18 ParseBytes candidate completed its first effective review
  using v1.0.19 and reached an accepted, quiescent READY checkpoint. Its writer
  and completed native compaction remain v1.0.18 evidence; the mixed-version
  continuation is not an entirely frozen v1.0.19 coding-task qualification.
- The first fresh v1.0.19 fixed-eight evaluation was 0/8 PASS: all tasks stopped
  in planning before host/model-access intent. Its explicit subscription access
  file used `session`, whereas the supported Codex mode is `chatgpt-session`.
  Token consumption and physical provider calls remain unobserved, not zero.
  Those run journals are preserved and are not retried.
- A new access file changes only that auth-mode spelling and preserves all
  budget ceilings. The candidate-copy helper is rebuilt from clean v1.0.19
  because the older helper rejected the newer compaction fields. Fresh trials
  bind both corrected files explicitly; previous provider-free preparations
  remain preserved and are not dispatched.
- The frozen evaluation runner does not request planner parse caching. Earlier
  planning notes that assumed it was enabled are superseded by this observation;
  candidate-facts caching is a separate requested policy. The direct long-horizon
  run did explicitly request planner parse caching.
- The documented subscription example is corrected with a focused regression.
  This documentation/test correction does not change product runtime source.
  Fixed-six 6/6, generated-code acceptance, topology/model comparisons, accepted
  fresh rounds, pause/resume and final release gates remain open.
## Checkpoint — v1.0.20 integrated; large candidate graph review correction

- Public main/dev contain `cf729529e776a734a0f53cbc91bcbc059f486b4d`.
  Bot-authored PR #27 merged. The documentation/example regression and ledger
  passed fresh doccheck, independent review, Sonar and security checks. Product
  runtime source is identical to qualified v1.0.19; no fresh complete v1.0.20
  Go suite is claimed.
- Correctly configured fresh v1.0.19 evaluations are active. Humanize and
  go-atomic completed native/held-out/review gates; broader cohort acceptance
  remains pending. Numeric-text exhausted its two repairs and is preserved.
- Afero and multierror stopped before reviewer dispatch. A single authorized
  normal Afero review continuation, after a quiescent candidate-bound preflight,
  retained the exact error `JSON size or UTF-8 invalid`. It left the candidate
  and controller unchanged and created no reviewer effect. Other generic
  review blocks are not asserted to share this cause without specific evidence.
- Graph sealing used the generic 1 MiB canonical limit before review impact
  could emit its existing 900 KiB unavailable summary. The next narrow product
  correction applies an explicit 32 MiB closed-typed bound at graph sealing,
  matching the existing graph input ceiling. Generic/untrusted boundaries and
  small graph identities are unchanged. Large graph seal/replay and bounded
  unavailable-summary regressions pass with independent review; complete clean
  qualification and integration remain pending.
- The first long-horizon round passed its original Humanize held-out gate and
  was committed locally through Fabric as `20bad9c4bbee3fb0acdb318e5e25b5361107a4ea`.
  A separately added Commaf supplement failed and remains recorded; it was not
  part of that round's preregistered ParseBytes gate. Round two starts from the
  confirmed commit, with a distinct run/thread/CODEX_HOME and no copied
  conversation. Its shared runtime parent is a private protocol deviation,
  not proof of the entire original private protocol. Round-two acceptance and
  pause/resume remain pending.

## Checkpoint — v1.0.21 integrated; observed parallel and token evidence

- Public main/dev contain `d3c1b581c8ea98c182b257eae7cc9762fee9196c`.
  Automation PR #28 merged. The exact clean complete Go suite exited zero
  (log SHA256 `567fc4d97b06c5beed09383130ca9d71243992a25166a34039ddcc518cc3ba49`),
  all ten frozen PowerShell harnesses passed, and independent source review,
  Sonar and security checks passed. The graph-size correction preserves the
  generic 1 MiB boundary and historical small graph identities.
- The corrected frozen v1.0.19 eight-task cohort completed **2/8 PASS**:
  Humanize and go-atomic. Numeric-text exhausted its two repairs; the other
  five stopped at review admission. The original failed cohort and individual
  run records are preserved. A later version's normal review continuation does
  not change that frozen result or establish fixed-six 6/6.
- The actual Humanize topology comparison completed: serial **BLOCKED** after
  an in-run repair and parallel **PASS** with two isolated writers, one serial
  aggregate integration, candidate-bound native and held-out verification,
  independent review approval and an accepted, quiescent checkpoint. The
  independent review approved the final private paired receipt SHA256
  `5e3ef1536c2600cf8e3f9f7d6133b2b78f2bdfe941f5e3f0161c3919aea49a78`.
  Both arms used candidate-facts cache v1; neither requested planner parse
  caching or explicit native compaction. Different plans and review outcomes
  prevent a causal speedup or token-saving claim.

| Observed Codex token category | Serial | Parallel |
| --- | ---: | ---: |
| Input, including cached | 1,587,874 | 1,013,808 |
| Cached input | 1,350,400 | 775,424 |
| Uncached input | 237,474 | 238,384 |
| Output, including reasoning | 20,533 | 14,132 |
| Reasoning output | 10,363 | 5,913 |

The accounting deltas are receipt-matched, completed and OBSERVED. Cached
input is a subset of input; reasoning output is a subset of output. Cost and
physical provider-call count remain unknown. Sampled process CPU is a lower
bound, sampled RSS can miss peaks, and evaluator elapsed time includes its
acceptance checks rather than measuring exact time to READY.

An independently reviewed metadata addendum observes 94,964.7481 ms of overlap
between the two isolated writer dispatch-wrapper intervals. These intervals
include setup; provider-turn-only overlap and reservation activation timing
remain unknown. Its SHA256 is
`87c2eff243a49df13b36f4aa0cfce6c5db1b76cd7757a765d871cd54ee6a74da`.

- The generator ownership problem is now localized: discovery and graph edges
  correctly bind numeric outputs to `gen.go`, the generator tool and template,
  but contract-context v5 exposes opaque graph endpoints and exhausts its
  source-excerpt slots before showing the actionable generation chain. A new
  explicit context recipe will project the existing admitted metadata, while
  preserving v5 prompt identities and replay. No generator commands or extra
  write authority are implied.
- The real six-arm fixer comparison and the second long-horizon round remain
  active. A task accepted without any fixer invocation is NOT_EXERCISED for
  fixer quality. Failed or incomplete task evidence is not a successful role
  outcome. Neither adaptive defaults nor v2 release readiness is claimed.

## Candidate checkpoint — explicit generation chains and review arrays

- The opt-in `go-contract-context-v3` writes record v6 and exposes only relevant,
  already-admitted generator chains with owner/output/tool/template paths and
  source hashes. The projection has eight-binding, eight-tool-reference and
  16 KiB bounds. Source excerpts, write scopes and generator execution remain
  governed by their existing rules. Tests cover durable replay, legacy v5
  prompt shape and hash domain, substituted source evidence, filtering and
  truncation. Fresh real generated-code acceptance is still required.
- The normal Afero review through qualified v1.0.21 reached an accepted READY
  candidate; fresh candidate-bound native and unchanged held-out checks passed.
  This mixed-version continuation does not replace the frozen v1.0.19 result.
  Its acceptance receipt SHA256 is
  `19f7db2aa47885c91f4cd0346034a1e6c21e8f59d5b9b0b47bef4452b5cbd6d4`.
- The second long-horizon candidate's first v1.0.21 review stopped before any
  reviewer effect: an empty required projection `omissions` array became null.
  The producer now preserves `[]`; the regression also checks all required
  projection arrays. Strict decoding is unchanged. The failed attempt and
  unchanged candidate/journal remain preserved; later acceptance is pending.
- A new, independently reviewed `go-wordwrap-tabs` task extends the pinned MIT
  package with an explicit tab-stop API. Native baseline passes and unchanged
  held-out baseline fails on the missing API. It is unseen relative to the
  retained Fabric task histories, not universally unseen. Existing ten task
  definitions and oracles are unchanged. No new task dispatch is claimed.

## Checkpoint — v1.0.22 integrated; empirical treatment preparation

- Main/dev contain `c46c5647d713dde198646925f84d7e598975f983`, merged through
  automation PR #29. Fresh clean full Go verification and all ten PowerShell
  harnesses passed, independent source review approved, and hosted security
  and Sonar checks passed. Sonar reported zero unresolved issues and about
  0.44% new duplication. The clean Go log SHA256 is
  `2b5cb004ce3fb0c4e9ab3a6f04132876a7344817a6d0b8dd0418568b8698fcc8`.
- A fresh frozen eight-task evaluation uses this exact qualified source with
  contract-context v3. Its first Humanize task reached approved READY; the
  cohort remains running. Fixed-six 6/6 and the two additional gates remain
  pending; older cohorts are not pooled with this evaluation.
- The second long-horizon review completed normally through v1.0.22 and
  requested changes. Native verification passed and active/uncertain intents
  were both zero, but the candidate is not accepted. A bounded normal repair
  must satisfy the unchanged performance oracle before continuation evidence.
- The real six-arm fixer calibration was independently checked against the
  retained source/task/route/invocation receipts. The actual calibrator selected
  no profile: `fallback_unknown_outcome`. Both training arms were UNKNOWN;
  the hold-out arms include one accepted exercised Luna fixer, one accepted
  Sol task without a fixer (NOT_EXERCISED), and two UNKNOWN tasks. This evidence
  supports conservative fallback, not a model-quality improvement claim.
- The next runner increment adds optional `-ModelPolicyPath` for Native runs,
  with prepared file bindings and confirmation of the embedded policy. Default
  runs retain their previous arguments and receipt shape. A genuinely unseen
  Wordwrap evaluation with the declared fallback policy remains pending.

## Checkpoint — v1.0.23 integrated; seeded review projection repair

- Main/dev contain `88fb02c448ee737ed17950dc82074f20ca3da122`, merged through
  automation PR #30. The optional model-policy runner treatment received
  independent review, and all ten PowerShell harnesses passed on the clean
  final commit (receipt SHA256
  `2073d30c78cb0f055306afc5293bf6bf5872674a8e6e4e5e280a95271d705d2d`).
  Hosted Sonar/security checks passed with zero unresolved issues and zero
  new duplication. Sonar and harness fixes remain in the same increment.
- Go/Rust source and dependencies are unchanged from qualified v1.0.22;
  its fresh full Go result is the runtime qualification reference. No fresh
  full final-v1.0.23 Go or Rust suite is claimed. An additional full Go run on
  the earlier Go-identical v1.0.23 source passed, log SHA256
  `df4c12ddc63a60efc367a196402113381c7aa10c76eb6f3369a7a3b10966d6d4`.
  Script files in that checkout were briefly amended and restored during the
  run; Go files were unchanged. It is not a continuously clean-checkout claim.
- The independent policy review approved the declared baseline fallback,
  not a measured improvement. Its calibration is unselected because of
  UNKNOWN outcomes; Wordwrap is outside its objective scope. Astra is only
  configured for static escalation, without live qualification evidence.
  The single Wordwrap preparation passed native baseline and the unchanged
  held-out baseline failed on the missing API as expected. Its new evaluation
  is authorized once on frozen v1.0.23; acceptance remains pending.
- The second long-horizon candidate now contains the requested benchmark
  cases after one normal repair. Candidate-bound native verification passes,
  one of two repairs remains, and active/uncertain intents are zero, but a
  new reviewer result was not admitted. Frozen v1.0.22 Afero encountered the
  same review boundary; neither result is counted as accepted.
- Code inspection found a concrete seeded review-projection defect: copying
  empty coupling/test/group slices with a nil destination restores JSON null
  for required arrays. A direct producer correction and seeded regression
  passed focused regression tests and independent review. Complete clean
  qualification and integration remain pending. The blocked runs expose only
  `execution_blocked`, so their
  exact inner error is not yet proven; no uncertain effect was resent.

## Checkpoint — v1.0.24 integrated; unseen fallback acceptance

- Main/dev contain `34ae84f3b2473e81f8ac7dfab5477d70800711c6`, merged through
  automation PR #31. Independent review, clean full Go, ten PowerShell
  harnesses and hosted Sonar/security passed. The full Go log SHA256 is
  `13cc0f3a14a99f22ce57c79ea10787a459a0ec436af1908d6812aa4a7ac60615`.
- The frozen v1.0.22 eight-task run finished with four READY tasks: Humanize,
  go-atomic, go-multierror and logr. Afero and numeric stopped at review;
  Difflib stopped after its completed writer effect before proposal admission;
  godotenv exhausted two repairs. Blocked candidates are not accepted and no
  result is pooled with the fresh, separately prepared v1.0.24 cohort.
- The genuinely unseen Wordwrap trial on frozen v1.0.23 passed native checks,
  unchanged held-out acceptance and candidate-bound review with one repair.
  Independent artifact review approved its private acceptance-policy receipt
  SHA256 `3adc6d67316f53187e27820f35f129f8542555de8e1b9cd5b86d8433d21d66b6`.
  Five completed runtime invocations and all observed effects are matched;
  the accepted checkpoint has no unmatched or uncertain effect. The actual
  fixer was Luna High with `calibration-objective-out-of-scope`; the candidate
  profile was not selected or applied. This proves conservative fallback on
  this task, not routing gains or calibration generalization.
- Wordwrap input was 607,611 tokens: 482,560 cached and 125,051 uncached.
  Output was 8,620, including 2,735 reasoning tokens. Task elapsed was 333,629
  ms and includes acceptance checks. Physical provider calls, price and exact
  READY timing remain unknown.
- The v1.0.24 long-horizon review stopped without dispatch or journal change
  with `review impact context requires an undispatched verified candidate`.
  The current repaired candidate has native PASS and zero active/uncertain
  effects, but the guard rejects its settled review from the prior candidate.
  A narrow candidate-binding correction in admission/replay is in progress;
  no new recovery machinery or old-effect resend is authorized.

## Checkpoint — v1.0.25 integrated; accepted long-horizon continuation

- Main/dev contain `f365a49149705e2a799c916809888836fbf9121a`, merged through
  automation PR #32. Independent review, clean fresh full Go, ten PowerShell
  harnesses and hosted Sonar/security passed. Full Go log SHA256:
  `e0965e1995bb65c436ad59715296d9796567213f43f685947fe9b06d6ae123c0`.
  Sonar reported zero unresolved issues and zero new duplication. Rust source
  is unchanged; no fresh v1.0.25 Rust suite is claimed.
- The narrow review boundary now admits a repaired, natively verified candidate
  after a settled changes-requested review of its previous candidate. Pending,
  uncertain, malformed or current-candidate review effects remain blocked.
  Actual v1.0.24/v1.0.25 read-only replay of the old round-two journal produced
  identical inspect/checkpoint bytes without changing the journal.
- Round two received a normal v1.0.25 review: APPROVE, followed by fresh native
  and unchanged Commaf held-out PASS on separate candidate-bound copies.
  Candidate `821a9d6ece0d30b2561aef41445035dc0983494eb68c6d3d83f5b94081af96fa`
  remains READY and accepted, with zero active or uncertain intents. The native
  command was the saved `go test -count=1 .`; the held-out command was
  `go test -count=1 -run '^TestFabricV1Heldout$' .`.
- Run-scoped pause, explicit quiescence settlement, lifecycle resume and one
  autonomous resume completed successfully. The candidate stayed accepted,
  and all eight invocation IDs remained unchanged; the resume completed the
  existing review-progress record without a new model invocation. Private
  lifecycle receipt SHA256:
  `ee0f3d2239d59d93b4fa0dc1095292a82a356cce0e78fe26c39789658e8a9051`.
  This does not assert that unrelated Fabric workloads were stopped.
- The two accepted coding rounds used distinct threads/run state. Sustained
  round-two execution recorded six native compaction events. Their source
  versions are mixed, and their state homes share the recorded parent home;
  this is accepted continuation evidence, not a pure-v1.0.25 frozen experiment
  or completion of the original isolation protocol.
- The separately frozen v1.0.24 cohort completed 5/8 PASS: Humanize, Afero,
  multierror, atomic and Difflib. Logr passed native verification and review
  but failed unchanged held-out acceptance. Numeric and godotenv stopped at
  the review boundary. Fixed-six is 5/6, not 6/6. Later candidate continuation
  cannot retroactively change these frozen results. The numeric repair budget
  is exhausted; godotenv has one of two repairs remaining.
- v2 remains open pending the fresh complete coding gate, measured efficiency
  qualification and final integrated release/happy-path review. No adaptive
  routing gain, physical provider-call count or monetary saving is inferred
  from UNKNOWN outcomes or observed runtime token counters.
## Checkpoint — v1.0.26 integrated; semantic orientation candidate

- Main/dev contain `b5c7502a8982badd7daf54147036548701769dc2`, integrated
  through automation PR #33. Fresh clean full Go, fresh Rust workspace tests,
  ten PowerShell harnesses, disposable source-build/help/reference smoke,
  independent review and hosted Sonar/security passed. Source stayed clean.
  Go log SHA256: ae832b31eda37c25794ce4450c263c4a997f001b2ebbd1bde3863234faf5892a.
  Rust log SHA256: 35f4d1a02b05056f18e5d9c74d18bc0c708823e2be3441cb1b70a652e4c46109.
  Sonar reported zero unresolved issues and zero new duplication.
- The next independently reviewed capability exposes bounded source-bound
  objective ranking with observed import-degree centrality and query-scoped
  topology components, plus direct SCIP references and explicit IMPLEMENTS
  paging. It reuses the existing RI and preserves PARTIAL/UNKNOWN coverage.
  It changes neither planner defaults nor old graph/invocation identities.
  Real pinned-Rust CLI/library fixtures and focused regressions passed;
  clean full qualification and public integration are still required.
- The blocked frozen-v1.0.24 numeric candidate received its first current
  review through the normal v1.0.25 path: APPROVE, fresh native PASS and
  unchanged held-out PASS. The original evaluation row remains BLOCKED.
  Supplemental receipt SHA256:
  `5dcff85c1fa498a354b0fe21cabd8a98b1f72444b5aa77ea90d40ad4c534862a`.
  This demonstrates the repaired product review path, without pooling frozen
  results or granting another numeric repair.
- A separate fresh v1.0.25 eight-task evaluation declares Sol High for all
  roles, with the same unchanged tasks, verification, held-out fixtures and
  repair ceilings. It is a static model-quality treatment, not an adaptive
  policy promotion. Preparation passed; the evaluation is running. Overall
  model choice is bound by its external dispatch receipt because the existing
  prepare manifest records only the fixer override.

## Checkpoint — v1.0.27 integrated; fresh fixed-six accepted

- Main/dev contain `6205de9b6b18d503cd45969a8a64660cb6bc65bb`, integrated through automation PR #34. Fresh
  clean full Go with the pinned RI, ten PowerShell harnesses, focused semantic
  and ranking fixtures, independent review and hosted Sonar/security passed.
  Sonar reported zero unresolved issues and zero new duplication.
- The separately frozen v1.0.25 Sol High evaluation completed 7/8 PASS:
  fixed-six 6/6 plus generated numeric ownership PASS; godotenv BLOCKED.
  Each passing task has candidate-bound native verification, approved review,
  fresh unchanged held-out acceptance and READY evidence. Independent review
  approved the exact joins. This does not rewrite the v1.0.24 5/8 result,
  pool cohorts or establish a complete v2 qualification.
- Godotenv exhausted its two repairs after three changes-requested reviews.
  Its nine model invocations are completed and receipt-matched, with zero
  active or uncertain intents and three confirmed aggregate file effects.
  Native and held-out candidate-copy acceptance were not run. The final
  candidate has a real single-line compatibility defect in escaped-quote
  comment scanning; native checks alone did not establish acceptance.
- Across all eight tasks, 33 matched runtime observations report 9,885,990
  input tokens: 8,561,152 cached and 1,324,838 uncached. Output is 101,281,
  including 27,989 reasoning tokens. Costs, physical provider calls and exact
  durable READY transition times remain unknown. The reviewed private terminal
  receipt SHA256 is `ee772fc452ca5da371585f7a4be0cc2ce2d541e92bbbedc398d29e0068642ddc`.
- Two independent clean v1.0.27 Windows amd64 package builds are byte-identical.
  Archive SHA256 is `dc2e8dd8ccdff6c03deeb6aa7ee5c95fbc498813c23ab96e964f5fb644b0fada`; both release verifiers and offline help,
  version, reference, fake-init and doctor smoke checks passed. This is packaging
  evidence only: the manifest keeps release_qualified false, and no real
  installed-task or v2 release qualification is inferred. Private packaging
  receipt SHA256 is `9b524c3bc64faca936c8311e2c9ec3e1cd8de387626d4f6995b6188252268fe1`.
- A profiled full-batch corpus fast path was tested and independently reviewed,
  but the matched Humanize workload showed no useful latency gain. External-test
  prefix collisions require fallback; the candidate was archived and removed
  without integration. Product code remains unchanged. Caches stay opt-in.
- One separate fresh v1.0.27 godotenv quality pilot started with Sol High
  roles and an explicit gpt-6-astra/high fixer, the unchanged public task,
  source pin, native checks, held-out fixtures and two-repair ceiling. This
  tests a stronger static fixer; it is not a retry of the exhausted run or an
  adaptive-policy/default promotion. The exact runtime model catalogue and
  corrected embedded VCS provenance were checked before dispatch.
- The first preparation attempt used a wrong model identifier and a binary
  without embedded VCS metadata. Evaluation rejected before any provider call
  or journal creation. Those records are preserved; the corrected one-task
  preparation has a distinct identity. v2 remains OPEN.

## Paused checkpoint — documentation update v1.0.28

- The user paused development; the persistent goal is PAUSED. v1.0.28 updates
  README navigation, the capability catalogue, evaluation status, topology,
  continuation and packaging documentation. It adds no product-code change,
  qualification score, release tag or resumed provider work.
- The corrected one-task pilot is preparation run
  `20261004T104730Z-0cc1a0a6`, evaluation `eval-20261004T104918Z-ad174e13`,
  Fabric run `ed85fb8415da21396538a38c4de9c0a83b451d1f17ea94953b10b4e937a9e307`.
  Its run-scoped pause request was accepted. At the recorded pause snapshot,
  lifecycle was PAUSE_REQUESTED and workflow IMPLEMENTING: the planner was
  completed, one writer access remained pending without a terminal receipt or
  proposal, no candidate was accepted and zero file effects were confirmed.
  PID 38784 was observed alive. This is not quiescence, completion or permission
  to resend the pending writer.
- The pause handoff observed at 2026-10-04T10:58:13Z has SHA256
  `8ebf6c033db35a85e34dbd79dc35abae1464e967b45d9477b63037918864d18f`.
  It retains exact journal, invocation, access, binary, runtime and dispatch
  identities. No evaluator exit receipt was present at that snapshot. Do not
  start another run or resume this one while development is paused. At an
  explicitly authorized continuation, inspect the same journal and reconcile
  the admitted attempt's recorded outcome before any further action.
- The reviewed v1.0.25 cohort remains 7/8, fixed-six 6/6 and generated numeric
  PASS. Its godotenv row stays BLOCKED. The pending newer pilot cannot replace
  that frozen result, and v2.0.0 remains OPEN.
