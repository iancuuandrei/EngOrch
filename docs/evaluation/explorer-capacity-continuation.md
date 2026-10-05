# Explorer capacity failure continuation

The unchanged eight-task v1.1.0 check on `7ac130d` finished **7/8 PASS**.
Its godotenv run retained a sealed Codex `failed` turn with
`codexErrorInfo=serverOverloaded`; the missing controller result receipt made
that known failure appear unresolved. This was a provider capacity failure,
not a successful exploration or an uncertain resend authorization.

## Repair

For configuration v1 only, Fabric can observe a sealed read-only explorer
capacity rejection offline and persist an explicitly failed receipt. The
receipt binds invocation, source, candidate, workspace, thread, turn and exact
runtime journal head. Incomplete streams, pending tools, successful or semantic
results, and missing or substituted identities do not qualify.

One subsequent invocation is permitted for the same graph question. Its
question suffix yields a new invocation and scheduler cohort identity. Replay
selects that identity consistently; a second capacity rejection exhausts the
retry. The original invocation is never resumed or erased. Configuration v2
keeps its separate model-access/usage settlement contract.

Failed receipts cannot establish exploration success. Usage measurement
revalidates the runtime head and failure hash and retains `Completed=false`.
This does not add code-repair slots or count bootstrap work as product progress.

## Evidence

- Sealed SQLite regression covers offline settlement, controller replay,
  idempotence, and honest usage measurement: PASS.
- Negative regressions cover UNKNOWN, incomplete stream, pending tools,
  existing result, semantic result, wrong turn and missing candidate binding:
  PASS. Deterministic restart and exhausted retry regressions: PASS.
- Graph-focused controller tests: PASS (60.819 seconds).
- Usage-focused controller tests: PASS (13.927 seconds).
- CLI autonomous failure/outcome/resume tests: PASS (3.056 seconds).
- Independent read-only review: APPROVE.
- Complete controller suite: TIMEOUT at Go's default 10-minute package limit;
  the active test had run for 24 seconds at that limit. This is not reported as
  a complete-suite PASS. The timeout log is retained separately.
- Final focused capacity, serial-correction lookup and the test active at the
  global timeout: PASS (27.152 seconds). Static vet for controller/CLI: PASS.

The retained run is
`b7040a4655bdc33bc30d9c330a493b36fda905bbbed4a98fcc184da4ea87fe9c`.
Its continuation uses the patched binary, so it is separate evidence from the
original unchanged 7/8 cohort. No historical rows or other runs are pooled into
a new unchanged 8/8 claim.

## Retained continuation result

The replacement explorer completed and the existing second repair writer
completed. An initial local verification attempt was NOT_RUN because the
continuation shell did not expose Go on PATH; that observation is retained.
After repairing the shell environment, the explicit verification command
passed on the exact candidate without another writer or repair slot.

The independent model reviewer requested changes to multiline quote/assignment
lookahead, compatibility of unquoted tabs, and quote-bearing comment boundaries.
The run stopped at **repair_budget_exhausted (2/2)** with candidate
`11c841d988a9089b1caab32e66b210e0ae705daa6ec9642ff003d794e6035468`.
It is not READY or accepted; candidate-bound heldout acceptance was NOT RUN.
The Fabric capacity-blocking defect is repaired, but the overall cohort remains
**7/8 PASS**. Another product repair requires explicit authorization beyond the
retained run's exhausted budget; the failure is not erased or relabeled.
