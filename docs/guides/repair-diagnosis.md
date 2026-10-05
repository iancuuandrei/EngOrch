# Inspect a failed candidate

Development source adds a read-only repair report:

```powershell
fabric diagnose RUN
```

Run this in the repository owning RUN, as with `fabric inspect RUN`. It replays
the existing journal and returns canonical JSON. It does not call a model,
read candidate files, append an event, consume a repair slot or authorize retry.
It also works for an exhausted run; a report does not reopen its budget.

## Validate current file evidence

```powershell
fabric diagnose RUN --anchor
fabric diagnose RUN --anchor > C:\private-evidence\prior-repair.json
# After an admitted candidate change, use prior code as localization hints:
fabric diagnose RUN --anchor --previous C:\private-evidence\prior-repair.json
```

Keep saved reports outside the candidate worktree, because an unadmitted new
report file would itself change candidate identity. `--anchor` acquires the
existing shared read lease, validates the recorded candidate and journal head,
and reads at most 24 files in sorted path order with 32 KiB retained per file.
Each file hash covers the complete file; code excerpts do not establish whole
file coverage. No lease is acquired by the default projection.

The optional `anchors` array preserves the original finding separately from
current `candidate_id`, `file_hash`, byte column, line and bounded code. Matching
candidate positions are labelled `range_verified`; a reviewer file without a
line is `file_verified`. Source excerpts can contain private code or embedded
secrets: keep these reports private. Public diagnostics do not include them.

For a stale finding, a prior report is an **untrusted hint**. Its content hash
detects accidental alteration; it does not authenticate historical authority.
The hint must match the same run, finding, original candidate and path. Only one
exact complete-line match in the fully retained current file can become
`code_reanchored`. The original finding's candidate and position stay unchanged.
Missing, repeated, substring-only, foreign and incompletely covered anchors do
not silently relocate. No guessed symbol resolution or cross-file search occurs.

The strategy suggestion becomes `localized_llm` only when every spec finding
has a current verified anchor inside that ready task's exact write paths. It
does not dispatch a fixer or prove the original defect persists or is closed.
If file observation is unavailable, `anchoring_status` is `unavailable` and the
base diagnosis remains available. Individual unavailable or out-of-bound ranges
remain explicit; they do not erase useful failures or weaken readiness.

## Read the result

| Field | Meaning |
| --- | --- |
| `controller_head` | Exact validated journal head used for the projection |
| `candidate_id` | Latest recorded candidate; not a new observation of disk |
| `findings` | Recorded native-check failures/unavailable checks and reviewer concerns, with source-specific categories |
| `candidate_binding` | Matching recorded candidate, stale candidate or no recorded candidate |
| `location_status` | Unlocalized, reported but unvalidated, or ambiguous |
| `evidence_id` | Hash of the complete source record, including omitted diagnostic tails |
| `gate_id` | Original verification plan or review invocation |
| `closure_oracle_id` | Original failed check/review identity; a reference, not proof of later closure |
| `specifications` | Advisory investigation specs for ready repair tasks, using only their existing write paths and exact failed gate |
| `verification_pending` | An unresolved native launch remains visible; it creates no invented finding |
| `verification_closure_recorded` | Operator-recorded quiescence evidence exists; it does not settle the missing check outcome |
| `repair_attempts`, `max_repairs` | Existing consumed slots and immutable ceiling; unavailable ceiling is `null` |

Native output locations are recognized conservatively from `.go:line:column`
diagnostics. Relative Windows separators are normalized; absolute paths,
traversal, invalid numbers and ambiguous locations are not silently relocated.
Compiler messages are not copied into findings. Review messages retain bounded,
UTF-8-safe text and explicit shortening. Treat this as private run evidence;
review text remains untrusted and is not automatically safe for public sharing.

## Use the evidence

For explicitly obtained individual Go test coverage, use
[spectrum-based localization](repair-localization.md). The optional
`--spectrum` report validates candidate source bytes and keeps caller-supplied
test outcomes untrusted. It does not change findings, admitted scope or closure.

Inspect the reported candidate file using the existing candidate tools before
editing. A reported line alone does not validate its current bytes. Stale
findings require explicit localization against the current candidate. Empty
specifications mean no ready admitted repair task is represented, not permission
to create one or widen ownership.

After a permitted repair, the new candidate still requires fresh configured
native checks and independent review. UNKNOWN provider effects remain UNKNOWN.
This command never resends them or infers their completion from missing findings.

## Give a fixer structured repair evidence

For a new development-source run, enable the optional integration explicitly:

```powershell
fabric run --autonomous --repair-intelligence "Fix a concrete coding task"
```

When an admitted repair task becomes ready, its writer input includes
`repair_intelligence`: findings from its exact failed gate, the existing write
scope and advisory strategy. Source validation uses complete files already in
the recorded task context. It does not read extra workspace files or add an
admission event. Partial context cannot prove a complete preimage. A prior
admitted context from the same source and original candidate can support unique
line relocation; original findings retain their identities.

The prompt does not duplicate code already present in `task_context`. Reports
over 32 KiB become an explicit hashed omission record. If optional intelligence
would take the input over 256 KiB, that addition is omitted while the ordinary
input remains intact. These are byte limits, not measured token savings.

The flag is stored in immutable creation policy and cannot retrofit an old run.
It defaults off; ordinary run inputs retain their historical serialization.
With configured `json-v1` review, new runs also freeze explicit reviewer rechecks.
The fixer still proposes changes through its existing contract and candidate
tools. Every configured native check and independent review remains required.
See [executed integration evidence](../evaluation/repair-fixer-context.md).

## Inspect native recheck evidence after a repair

```powershell
fabric diagnose RUN --closure
```

This reads one validated journal prefix and reports historical findings separately
from later evidence. It does not call a model, read workspace files, append
events or reopen a budget. Use it on the repository that owns RUN. It cannot be
combined with `--anchor` or `--previous`.

`native_oracle_passed` means the exact configured check was observed PASS in a
chronologically later, distinct plan on the final accepted READY candidate.
The comparison excludes candidate identity and preserves every other frozen
invocation field: check arguments and timeout, directory, root executable path/
hash and complete selected environment. The report binds original finding and
oracle IDs to the new candidate, plan, invocation and observation hash.

This proves the native oracle passed, not that Fabric established a root cause,
that project test code was unchanged, or that every possible defect is absent.
A prior timeout, cancellation or output overflow can also be followed by a PASS;
that later receipt does not explain the earlier failure. Final native checks and
candidate-bound review remain required. Changed executable/environment/check
definitions produce `oracle_changed`; missing acceptance stays `open`.

Unavailable/NOT_RUN checks are `unavailable_not_semantic`, never invented
successful repairs. Ordinary approval does not close reviewer concerns.
An empty list does not settle UNKNOWN effects.

At most 128 findings from the most recent gate records are retained. Omitted
entries contribute to `omitted_findings` and `omitted_evidence_hash`.
The omission hash binds original finding IDs, oracle definitions and sequence;
it does not include derived closure statuses for omitted entries.
Original finding identity includes complete source evidence, including shortened tails.
Review text is private, untrusted run evidence. Check arguments, environment
values and raw process output are not copied into this report. See the
[scoped closure evidence](../evaluation/repair-closure.md).

## Require explicit reviewer rechecks

New `--repair-intelligence` runs with configured `json-v1` review preserve a
bounded cumulative history of reviewer concerns. Each subsequent review receives
original IDs, paths and bounded messages, with full source-evidence hashes.
The reviewer must answer each supplied ID exactly once with `closed` or
`unresolved` and a concrete rationale. Approval requires every supplied concern
closed. An unresolved answer must also appear in ordinary current findings with
the same path and rationale, so design/fixer feedback remains actionable.

`--closure` reports `reviewer_recheck_closed` only for an explicit recheck in
the final accepted candidate-bound review. It binds the original finding to that
review invocation, native verification plan and full review-record hash. This is
reviewer judgment, distinct from `native_oracle_passed` and semantic proof.

History retains at most 128 concerns and hashes/counts omitted original IDs.
Each admitted review occurrence has its own evidence-bound identity. Unresolved
feedback echoed into current findings is retained as a new occurrence; Fabric
does not guess semantic deduplication between reviewer messages.
Recheck requests cover at most 64 concerns to fit ordinary unresolved feedback;
larger histories use the ordinary review fallback.
Optional recheck context is bounded to 64 KiB; it is omitted if the full review
input would exceed 256 KiB. The ordinary review contract remains available.
`review_recheck_coverage` exposes `not_requested`, `requested` or
`partial_history`; unsupplied/omitted concerns are never inferred closed.
These limits are byte bounds, not measured token savings. Existing runs keep
their original review inputs and generic approval semantics. See the
[recheck evidence](../evaluation/reviewer-rechecks.md).

The [historical live pilot](../evaluation/repair-intelligence-short-root-20261005.md)
demonstrates two repairs reaching READY and external PASS, with separate native
oracle and explicit reviewer-recheck closure receipts for the accepted candidate.
The control remained blocked on a rejected file precondition. This single-task
integration result does not establish general quality or token/time improvement;
the treatment remains opt-in.

Repair Intelligence supplies typed diagnosis, exact scope inspection, bounded
preimage validation and explicit unique code relocation. Deterministic
transformations, strategy execution and finding closure are not implemented
by this report.
No live repair-quality or token-efficiency improvement is claimed. See the
[contract](../specifications/repair-diagnosis.md) and
[evidence status](../evaluation/status.md).
