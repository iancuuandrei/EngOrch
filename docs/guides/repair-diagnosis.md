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
The fixer still proposes changes through its existing contract and candidate
tools. Every configured native check and independent review remains required.
See [executed integration evidence](../evaluation/repair-fixer-context.md).

Repair Intelligence supplies typed diagnosis, exact scope inspection, bounded
preimage validation and explicit unique code relocation. Deterministic
transformations, strategy execution and finding closure are not implemented
by this report.
No live repair-quality or token-efficiency improvement is claimed. See the
[contract](../specifications/repair-diagnosis.md) and
[evidence status](../evaluation/status.md).
