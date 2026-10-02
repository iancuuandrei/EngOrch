# Real repository comparison and observed product failures

## Matched six-task run

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
