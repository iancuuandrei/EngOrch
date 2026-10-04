# Advisory repair diagnosis v1

`fabric diagnose RUN` MUST derive its report from a repository-bound,
replay-validated snapshot. Without `--anchor`, it MUST NOT read candidate files.
It MUST NOT append journal events, modify invocation inputs, dispatch effects, settle UNKNOWN, reopen a
repair budget or grant write authority.

Findings MUST preserve their source candidate, producer, exact gate and complete
source-evidence hash. Shortened text MUST NOT collapse distinct full-evidence
identities. Native failure, unavailable native check and reviewer concern MUST
remain distinct. Pending native launches MUST remain pending and MUST NOT be
represented as observed failure or success.

Reported locations MUST be syntactically portable relative paths with positive
line/column numbers when supplied. The default projection does not validate file
preimages; it MUST label reported locations unvalidated, and MUST NOT choose
between distinct observed locations. Candidate matching means matching the
recorded identity, not freshly observed filesystem bytes.

An advisory repair specification MUST use the current recorded candidate,
the exact failed gate associated with an already ready implementation task,
and that task's declared write paths. It MUST exclude stale-candidate findings
without a current validated anchor, and unrelated gates. It MUST NOT substitute the broader original scope for
the admitted write paths. The strategy remains `localization_required` until
additional candidate-bound evidence is available.

The closure reference is the original check/review identity. It MUST NOT be
interpreted as a passing closure receipt. Acceptance still requires the existing
fresh candidate-bound verification and review gates. Findings are bounded to
128 entries and review-message excerpts to 384 UTF-8-safe bytes.

## Optional imported coverage localization

`--spectrum SPECTRUM_JSON` MAY add an advisory Ochiai ranking from explicitly
supplied individual Go coverage profiles. It MUST remain exclusive with
`--anchor`, `--previous` and `--closure`. Absent spectrum MUST preserve default
diagnosis serialization. No native test or model effect is executed by import.

The artifact MUST bind the exact run and current recorded candidate. It MUST
retain caller-supplied provenance rather than imply authenticated test outcomes
or oracle closure. PASS/FAIL are the only accepted assertions; UNKNOWN and
missing profiles MUST NOT become observed failures or uncovered blocks.
Profiles MUST have matching coverage mode, complete block inventories and
statement counts. Duplicate test IDs, source mappings or block coordinates
MUST be rejected. At least one asserted failed test MUST exist.

The exact squared Ochiai fraction MUST retain raw failed/passed test counts,
with deterministic range ordering for ties. Visit counts MUST NOT be counted
as independent tests. Unexecuted blocks MUST retain a zero denominator.

Candidate source hashes/ranges MUST be observed under the existing shared read
lease, bracketed by candidate observations and matching journal heads. A failed
guard close MUST withhold candidate-bound rankings. Optional observation
failure MUST preserve base diagnosis with explicitly unavailable source status;
it MUST NOT erase recorded findings or change write scope, repair budgets,
acceptance, effects or closure requirements. Hash/range observation validates
bytes, not authenticity of coverage production or causality.

Import is bounded to 1 MiB JSON, 64 tests, 24 sources, 64 KiB/profile, 4096
blocks/profile and 32 KiB complete source/file. It MUST NOT truncate partial
profiles into valid evidence. No imported ranking is persisted as controller
truth or injected into existing invocation inputs. Explicit pre-invocation
admission MAY retain untrusted raw evidence as described below.

## Optional candidate anchoring

`--anchor` MUST use the existing shared worktree read lease and ownership guard.
It MUST bind complete-file hashes and retained bytes to the recorded candidate,
with candidate observations before and after the source batch and matching
journal heads around the complete operation. Observation failure MUST preserve
the base diagnosis with unavailable anchoring; it MUST NOT infer a new candidate.

At most 24 paths, 32 KiB retained per path and 4096 bytes per code anchor are
admitted. Missing, protected, omitted or incomplete locations MUST remain
explicit. A retained prefix MUST NOT establish unique full-file relocation.

Previous reports are bounded to 1 MiB and 128 anchors and MUST match the run.
Their hashes MUST be checked, but MUST NOT be treated as authentication or
authority. A hint MUST match the original finding, source candidate and path.
Only a unique exact complete-line match in the current fully retained file can
relocate it. Missing or ambiguous matches MUST NOT select a guessed location.
Original finding fields and IDs MUST remain unchanged by anchoring.

`localized_llm` is an advisory strategy suggestion only when every specification
finding has a current validated anchor inside the existing admitted write paths.
It MUST NOT dispatch a model, bypass ownership or imply semantic finding closure.
Code and source hashes in anchored reports are private run evidence.

## Opt-in fixer input

`repair-context RUN SPECTRUM_JSON [--task TASK_ID]` MAY admit a spectrum through
the existing `task.context-admitted` event before freezing writer inputs. It
MUST require enabled repair intelligence, bounded task context, a ready repair
implementation and its exact failed parent gate. Empty task selection MUST
use the existing serial writer query; explicit task selection MUST use the
same query as the task-bound invocation. Isolated child admission is not
supported by this first integration.

The optional `repair_spectrum` field MUST bind task, failed parent gate, run,
candidate and invocation query. Absent field MUST preserve old serialization
and prompt bytes. Canonical spectrum bytes MUST NOT exceed 128 KiB. Replay
compatibility is additive for the current reader; older builds may reject this
optional field and MUST NOT silently treat it as authority. Replay
MUST validate complete recorded manifest source hashes/ranges and MUST NOT read
mutable files. Missing/partial source context MUST reject spectrum admission.
Normal selection ceilings remain 12 files, 48 KiB total and 8 KiB per file;
import MUST NOT expand them to retain a larger coverage preimage.
Identical re-admission MAY reuse the record; different or missing existing
evidence MUST NOT permit replacing or retrofitting a frozen context.

Raw profiles MUST remain private durable evidence and MUST NOT enter writer
inputs. The compact ranking MUST retain untrusted provenance and the full input
hash; locations MUST be filtered to exact task write paths and omitted blocks
counted. Source-byte validation MUST NOT authenticate test execution or closure.
Admission and retention MUST NOT change findings, write authority, effects,
repair budgets, verification, review or acceptance.

New autonomous creations MAY set `repair_intelligence_version: 1`. It requires
graph v1, repair planning v1 and bounded task context. Absent/zero policy MUST
preserve historical serialization and writer inputs.

The fixer projection MUST derive only from replay-validated records, filter to
the ready task's exact failed parent gate and preserve its existing write scope.
It MUST NOT read files during invocation reconstruction or append new events.
The diagnostic journal cursor MUST be omitted from invocation data because
host-intent appends cannot change invocation identity.

Preimage validation MUST require a complete selected file starting at zero,
matching content length and SHA-256, at most 32 KiB, and matching the current
candidate. Relocation MAY use complete prior admitted context for the same
source and original finding candidate. Missing or partial evidence MUST NOT
imply a verified preimage. Code already in task context SHOULD NOT be duplicated.

Derived repair reports over 32 KiB MUST become explicit omission records with
the full projection hash and omitted finding count. Optional repair intelligence
MUST be removed if it would exceed a 256 KiB writer input; mandatory existing
input MUST remain intact. Byte bounds MUST NOT be presented as token savings.
Strategy suggestions MUST remain advisory. Existing proposal validation,
repair budgets, provider authority, native checks and review remain unchanged.

## Historical native recheck projection

`--closure` MUST read and fully replay-validate one journal prefix, then check
the requested run/repository and configured controller path against that same
prefix. It MUST NOT replay every prefix, observe the workspace, append events,
dispatch effects or grant repair authority. It MUST remain exclusive with
`--anchor`/`--previous`; default diagnosis bytes MUST remain unchanged.

Historical native findings MUST use the complete terminal VerificationState
and original ordinal through the existing finding identity function. Review
findings MUST use the original full ReviewRecord context hash, shortening rules
and ordinal. Partial current plans and operator quiescence records MUST remain
truthful; quiescence MUST NOT become a successful result.

A native finding MAY report `native_oracle_passed` only when its failed plan
precedes a distinct final plan, that plan belongs to the accepted candidate-bound
READY checkpoint, and the original and final invocation definitions match with
only candidate identity excluded. All other invocation fields MUST be bound,
including full environment, executable hash, directory and timeout. Original
finding/gate/oracle identities MUST remain separate from closure candidate,
plan, invocation and observation hash. This status MUST NOT claim semantic
root-cause proof or immutable project test code. Timeout/cancellation/overflow
receipts remain original failures even if the later matching oracle passes.

Without final acceptance the status MUST remain `open`; a substituted final
oracle MUST yield `oracle_changed`. Unavailable original checks MUST remain
`unavailable_not_semantic`. Generic approved review MUST NOT close historical
review concerns, which remain `recheck_required` unless explicitly rechecked
under the opt-in contract below.

Derived history MUST retain at most 128 findings, count omissions and hash their
original finding identities, oracle definitions and journal sequence in order.
The report MUST NOT expose raw invocation arguments, environment or process
output. Bounded review messages retain the existing private-evidence boundary.

## Explicit reviewer rechecks

New structured-review creations MAY freeze `review_recheck_version: 1`, requiring
repair intelligence v1 and configured `json-v1` review. Absent policy MUST preserve
historical invocation bytes, verdict serialization and replay. No new journal
event is introduced: history MUST derive from admitted review records.

History MUST preserve original full-record finding IDs and hashes, including
shortened tails. It MUST retain at most 128 concerns; omitted original IDs MUST
be counted and bound in an ordered hash chain. Prompt candidate binding MUST
describe the current candidate without changing original finding identity.

The finite reviewer output schema MUST require one recheck per supplied original
ID, without duplicates or foreign IDs. Answers MUST be `closed` or `unresolved`
with a nonblank UTF-8 rationale of at most 512 bytes. Approval MUST require all
supplied concerns closed and ordinary findings empty. Every unresolved answer
MUST have an ordinary current finding with the exact original path and rationale
as message, preserving existing repair/design/fixer feedback and decision gates.

Recheck requests MUST admit at most 64 concerns, matching the existing current
finding limit; larger histories MUST use ordinary review fallback rather than
request an impossible unresolved verdict. Optional context MUST be omitted
beyond 64 KiB or when it takes review
input beyond 256 KiB. The exact ordinary invocation shape MUST identify fallback
during replay; omitted concerns MUST NOT acquire closure. Coverage MUST remain
explicit in the closure report. Native verification and review budgets remain
unchanged; incomplete recheck coverage MUST NOT imply overall finding closure.

`reviewer_recheck_closed` MUST require the final accepted review on the current
candidate/verification plan, an explicitly closed original ID, and an original
concern preceding the final native plan. Closure evidence MUST bind the full
admitted review record and its invocation. It is model judgment, not native
execution or proof of semantic correctness.
