# Real repository comparison and observed product failures

## Latest six-task evaluation

The fresh `eval-20261003T010038Z-dbbadab3` evaluation used clean Fabric
`af34d3ffb39c0812c4f818822397e480b0e5d5a4`, the same pinned repositories,
Codex executable (`1722907aa64401bcc9b34467ef5c2af43f6aef9a5045ef04d11d19dfae4589fb`),
model (`gpt-6-luna`) and effort (`high`) as the initial comparison below.
All six rows reached a terminal evaluation outcome.

| Task | Latest Native | Earlier PR #5 baseline | Repair attempts in successful run |
| --- | --- | --- | --- |
| go-humanize | PASS | BLOCKED | 2 |
| afero | PASS | BLOCKED | 1 |
| go-multierror | PASS | PASS | None recorded |
| go-atomic | BLOCKED | BLOCKED | Not established |
| go-difflib | BLOCKED | BLOCKED | Not established |
| logr | PASS | BLOCKED | None recorded |

This run completed **4/6**, compared with the earlier baseline's **1/6**.
Each PASS includes READY, native checks, an approving candidate-bound review
with zero findings, and independent candidate-bound native and held-out tests.
The successful humanize and Afero runs repaired failed gates without a human
intervention recorded by the runner. Atomic and difflib stopped in implementation;
their missing metrics remain unknown. The earlier baseline was not rerun, and
these six observations do not establish a general success rate or a completed
v1 release. Later CLI run-listing fixes and documentation changes were verified
separately from this immutable evaluation binary.

## Harder parser task on public main

A separate fresh `godotenv` attempt used clean public-main binary
`10fee6bafa4ecc28ae8a56f52c08a4001ecd8b28`, the same pinned runtime, and
`gpt-6-luna` with high effort. It completed research and design but ended
`BLOCKED` in implementation before any file effect. The returned paths,
candidate ID and file hashes were valid, but the proposed `godotenv.go`
before-anchor occurred zero times in the current file. The other two anchors
occurred once. Exact anchored-edit validation rejected the proposal;
verification and review were not run, and the candidate remained clean.

The [sanitized diagnostic](../../evals/v1/results/godotenv-main10-20261003.json)
retains exact evaluation/result identities and unknown metadata. The original
run was not resumed or edited. This is a separate unsuccessful harder-task
attempt, not an added PASS or a seventh row in the six-task comparison.

The later clean `525d53a686cd56961de4dcd5a2f74c74ec887193` opt-in trial
used `--parallel-writers --max-parallel 2` but the planner selected one
implementation task. Its writer completed the SDK turn and made two candidate
reads; all three proposed file hashes matched. Four edit anchors occurred
once, while an 801-byte source anchor occurred zero times. Admission rejected
the proposal before aggregation, file application, verification or review.
The [sanitized trial summary](../../evals/v1/results/godotenv-parallel-525d53a-20261003.json)
records this separate BLOCKED result and its immutable evaluation hash.
Outer finish metadata and unavailable usage metrics remain UNKNOWN. The run
was not resumed; it supplies no successful parallel or latency comparison.

The clean `e973cd50b89964eb6ffe791b2fcfadc4b3bc7deb` trial selected the
opt-in `anchored-edits-v2` contract. Its initial proposal was accepted and
the file outcome was CONFIRMED. Native verification then failed in
`TestMultilineQuotedValues`: `Can't separate key from value`. Repair design
completed, but the repair implementation remained at zero attempts and no
repair writer host intent was recorded. This trial is BLOCKED, with no
successful acceptance or latency claim. Actual validation-tool call counts
are not established by occurrences of tool names in schemas or instructions.
The original run and candidate were preserved without a retry.

## Initial matched six-task run

Two fresh sets of pinned repositories used the same Codex executable, model
(`gpt-6-luna`) and reasoning effort (`high`). Native Fabric was built from clean
commit `bf20b2d3b6fa794c1d1ca2e2bf936c423cd020de`; the PR #5 baseline was
`395e3b9f55b5262dfda81ab3484b0d113849a7cb`. These results describe those binaries,
not later fixes or a completed v1 release.

| Task | Native | PR #5 | Observed limitation |
| --- | --- | --- | --- |
| go-humanize | PASS | BLOCKED | Baseline reviewer requested correction of comma/underscore handling. |
| afero | PASS | BLOCKED | Baseline exploration completed a runtime observation without an accepted explorer record; exact rejection cause is not established here. |
| go-multierror | PASS | PASS | Both candidates passed independent acceptance. |
| go-atomic | BLOCKED | BLOCKED | Native reviewer found missing generator support; repair stopped before writer dispatch. Baseline stopped before proposal acceptance. |
| go-difflib | BLOCKED | BLOCKED | Native writer removed unrelated APIs and exhausted its bounded repair attempts. Baseline failed SplitLines behavioral checks. |
| logr | BLOCKED | BLOCKED | Both reviewers requested missing empty-call regression coverage; native repair stopped before writer dispatch. |

Native completed **3/6**, versus **1/6** for PR #5. PASS requires READY, exact
reviewed and verified candidate identities, a candidate-bound independent copy,
native upstream checks and held-out behavioral acceptance. A model assertion,
successful process launch or incomplete run never counts as PASS.

The shared successful multierror task took 120.049 seconds with Native versus
286.921 seconds with PR #5. Observed input-token totals were 161,330 versus
482,080; completed runtime invocations were three versus four. A runtime
invocation is a role execution, not a provider-request count. This is one
observation with uncontrolled cache and service latency, not a general speed or
cost claim. Provider-request counts and monetary cost remain unknown.

Native runs have no mandatory plan/file approval gates. The baseline script used
two preauthorized gates; those records do not establish two actual human
interventions. Missing metrics on blocked rows remain unknown. Later repairs or
resumes must be reported separately from this initial matched result.

## Real decomposition and concurrency

A fresh Afero objective ran from the installed Windows bundle at
`7835ceb32aa7daf2f8bc1c76201f7a03fb9e8517`. Its planner delegated independent
ownership/semantics and regression-design investigations; implementation depended
on both, and its input bound both completed advisory records.

Actual runtime intervals overlapped by **21.890 seconds**. The observed research
span was **40.239 seconds**; the sum of the two individual service intervals was
62.129 seconds. That sum is only a counterfactual serial comparison. It is not an
observed sequential run or a claim about whole-task acceleration.

A separate sequential run at the same installed source observed 57.492 seconds
of research service span. Its graph and prompts differed, and its second result
was rejected because the model returned a candidate ID with one character missing.
It is therefore not a matched successful-performance comparison. That run remains
blocked; the newer explorer contract constrains future output to the exact ID and
does not rewrite the rejected result.

Run `38815e6370c8978b8e618ec2959c2c36b305588cd71d96bbdd65ad8d44d6f636`
reached READY, native verification passed, and independent review approved with
zero findings. Candidate
`222b0a17a1a268417d51018422c69fc90f1809b6ad014b01596d033a337d4caf`
was copied under a read lease and independently passed upstream and held-out
Afero checks. This establishes useful read-only parallelism and scoped
implementation, not concurrent overlapping writers.

## Reproduction and retained evidence

The [sanitized comparison record](../../evals/v1/results/comparison-20261003.json)
publishes the retained evaluation hashes, exact binary/runtime identities,
task pins and per-task outcomes. Unknown metrics remain null. It contains no
raw model output or credentials and does not include an executable archive
of the accepted candidate bytes; it is a result summary, not release proof.

Use the [pinned evaluation runner](../../evals/v1/README.md). Acceptance tests are
added only to the independently bound candidate copy after agent work completes.
The manifest records repository pins, task objectives and explicit verification
scope. Legacy difflib behavioral checks disable vet for pre-existing incompatibility;
Windows atomic checks explicitly exclude the unsupported Unix integration test.
Neither exception establishes a full upstream static-analysis or Unix result.

The external evaluation records are `eval-20261002T211827Z-a43bcf91` (Native) and
`eval-20261002T211827Z-cb426de5` (PR #5). They retain exact binary, runtime, runner,
helper and task provenance. Sanitized concurrency, candidate-copy and acceptance
receipts accompany the installed-bundle run; raw credentials are not published.

These failures led to role-bound repair-context admission and a new frozen
complete-file writer contract. Their offline regressions passed; real acceptance
of those repairs is still pending. Neither change weakens the verification
checks that detected the original defects.

The real logr resume at `e624738` successfully admitted repair context and
completed a new writer proposal. It remained blocked because the proposal added
`logr_withvalues_test.go` outside the declared `logr.go`/`logr_test.go` ownership.
No second file effect ran. This proves repair dispatch was restored, not that the
task completed; the ownership check was retained.

A fresh difflib evaluation with the `utf8-replace-v3` writer at `e624738`
also remained blocked. Its first proposal preserved the complete implementation
file unchanged and changed the existing tests; native verification failed three
`TestSplitLines` assertions because the source fix was still missing. The repair
proposed `difflib/splitlines_test.go` outside the two declared write paths and
was not applied. No review occurred. Complete-file preservation improved over
the earlier destructive rewrite, but this run does not establish successful
implementation or repair acceptance.

The recorded difflib context also omitted `SplitLines`: selection pivoted on
the common word `and` near the file header, while the target function was near
the end. Future selections prefer a less frequent matching token, then a longer
token, while retaining the same byte budgets. Existing recorded manifests are
reused. Future writers receive the active graph task and declared write paths
through `utf8-scoped-v4`; this does not widen ownership or prove live acceptance.

The first two-task `utf8-scoped-v4` evaluation at `000e907` reported **0/2
PASS**, both blocked before a writer proposal or verification result. The
Codex dispatch adapter still accepted only the older static writer schemas,
so it rejected the new candidate-bound schema. This is an integration failure,
not evidence about the models' implementation quality. The runtime schema
admission path must accept the exact invocation-bound candidate schema while
continuing to reject substituted identities and schema shapes.

After the adapter repair, the fresh two-task evaluation at `74ba8bb` also
reported **0/2 PASS**, with different failure evidence. Difflib started a real
writer turn and completed five candidate-read/list calls, but retained no terminal
result; no proposal, file effect, verification or review was recorded. Its
unresolved turn must not be retransmitted. Logr's initial proposal deleted both
`logr.go` and `logr_test.go` in the isolated candidate. Native compilation caught
the missing declarations. Its repair proposed deleting an already absent file
and was not accepted. The base repositories remain unchanged. These observations
do not establish successful live repair acceptance; they motivate localized
text edits instead of model-authored complete-file replacements for ordinary
implementation changes.

The clean `2942892` anchored-edit evaluation reported **1/2 PASS**. Logr
reached READY and passed native checks, review and candidate-bound held-out
acceptance. Difflib completed a writer turn but returned the old full-file
wire shape instead of anchored edits; it was rejected before any proposal or
file effect. This is improved observed acceptance for these two tasks, not
full v1 qualification. The console initially displayed an incorrect PASS count
for a single result because PowerShell counted dictionary fields; `eval.json`
retains the authoritative per-task states. The summary now counts result rows.

## Separate autonomous repair acceptance

Run `c78d41fc977e603b10673fd4363da2ebbe71f857da955b030cce069f92645b4f`
on the pinned go-atomic repository initially failed native compilation because
the generated test omitted `strconv`. It started under `3719c930545c16c9e7ffe20ef2eca7417918341d`.
The repair design was blocked before provider dispatch by a configuration-v1
scheduler guard. After the guard fix, an explicit resume under
`af34d3ffb39c0812c4f818822397e480b0e5d5a4` completed design, a scoped repair,
fresh native checks and approving review.

Candidate `6845faaa71695a92e1b0d6e62d3ec18283e19e1547af5416b5c7fbf0334e93bd`
then passed native and held-out acceptance on separate candidate-bound copies.
Only `bool_ext.go` and `bool_test.go` changed. This demonstrates the repaired
path and restart with the earlier failure preserved; it involved a product fix
and an explicit resume. It is a separate result, not a fifth PASS in the latest
six-task evaluation or a zero-intervention release qualification.
