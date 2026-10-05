# Advisory repair diagnosis v1

`fabric diagnose RUN` MUST derive its report from a repository-bound,
replay-validated snapshot. It MUST NOT append journal events, read candidate
files, modify invocation inputs, dispatch effects, settle UNKNOWN, reopen a
repair budget or grant write authority.

Findings MUST preserve their source candidate, producer, exact gate and complete
source-evidence hash. Shortened text MUST NOT collapse distinct full-evidence
identities. Native failure, unavailable native check and reviewer concern MUST
remain distinct. Pending native launches MUST remain pending and MUST NOT be
represented as observed failure or success.

Reported locations MUST be syntactically portable relative paths with positive
line/column numbers when supplied. This projection does not validate file
preimages; it MUST label reported locations unvalidated, and MUST NOT choose
between distinct observed locations. Candidate matching means matching the
recorded identity, not freshly observed filesystem bytes.

An advisory repair specification MUST use the current recorded candidate,
the exact failed gate associated with an already ready implementation task,
and that task's declared write paths. It MUST exclude stale-candidate findings
and unrelated gates. It MUST NOT substitute the broader original scope for
the admitted write paths. The strategy remains `localization_required` until
additional candidate-bound evidence is available.

The closure reference is the original check/review identity. It MUST NOT be
interpreted as a passing closure receipt. Acceptance still requires the existing
fresh candidate-bound verification and review gates. Findings are bounded to
128 entries and review-message excerpts to 384 UTF-8-safe bytes.
