# Inspect a failed candidate

Development source adds a read-only repair report:

```powershell
fabric diagnose RUN
```

Run this in the repository owning RUN, as with `fabric inspect RUN`. It replays
the existing journal and returns canonical JSON. It does not call a model,
read candidate files, append an event, consume a repair slot or authorize retry.
It also works for an exhausted run; a report does not reopen its budget.

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

This first Repair Intelligence increment supplies typed diagnosis and exact
scope inspection. Automatic re-anchoring, deterministic transformations,
strategy execution and finding closure are not implemented by this report.
No live repair-quality or token-efficiency improvement is claimed. See the
[contract](../specifications/repair-diagnosis.md) and
[evidence status](../evaluation/status.md).
