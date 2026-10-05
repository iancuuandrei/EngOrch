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
