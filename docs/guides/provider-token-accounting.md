# Harness accounting and provider technical limits

`unlimited_tokens = true` with `tokens = 0` disables the EngOrch token
consumption ceiling for that run or role. Observed input/output usage remains
recorded. Cost, concurrency, exact identity, terminal settlement and technical
route constraints are not disabled.

For an unlimited provider-backed invocation, gateway evidence records:

- `engorch_budget.mode = "unlimited"` and legacy `reserved_tokens = 0`;
- `provider_reservation.tokens`, derived from the validated model contract;
- `provider_reservation.reason = "model_contract_conservative_maximum"`;
- `provider_reservation.hard_limit = true`.

The technical bound is the existing explicit model-contract calculation:
`max_calls * (context_window_tokens + max_output_tokens)`. It is not an
invented admission budget or an observation of a provider's advertised maximum.
These route fields are operator declarations and remain bound to model/config
identity. Missing or invalid bounds fail configuration/admission. No default
numeric reservation is invented for unlimited accounting.

Per-call output, total response, request/response byte, call count, and pricing
constraints remain enforced. An observation exceeding a hard technical contract
is rejected as a provider contract violation, not budget exhaustion. There is no
soft reservation mode in this bounded implementation. Unlimited accounting alone
can accept arbitrary representable observed totals; it does not override a hard
technical contract. Reasoning and cache detail remain subsets, never additional
tokens.

Historical finite bindings retain their serialized identity and finite budget
semantics. The new explicit fields are absent on those bindings. Access intent,
gateway binding, request/response receipts and terminal access receipt remain
linked by exact hashes; the numeric provider reservation is never copied into an
unlimited access reservation.

## OpenCode usage (v1.1.31, read-only)

`fabric usage RUN` additionally exposes admitted OpenCode invocations through
`open_code_invocations` in actual controller admission order. Legacy receipts
without `AgentDispatch` order by their receipt event position, preserving
deterministic controller order. `RunUsage.Scope` is unchanged; the additive
field is omitted when absent. Since v1.1.32, `evidence-feedback` additionally
admits matched completed OpenCode entries into read-only advisory token
costs through the same four axes; with no OpenCode invocations its scope
stays `receipt_matched_completed_codex_typed_tokens`, otherwise it is
`receipt_matched_completed_codex_and_opencode_typed_tokens`. See the
[observed-feedback guide](evidence-value.md#apply-observed-token-feedback)
and the [normative contract](../specifications/evidence-value.md).

Each completed entry is bound by existing controller, gateway and runtime
journal APIs only:

- controller authority is reconstructed at the correct historical prefix
  through `resolveScheduledInvocationID`, never trusting runtime intent;
- receipt matches invocation, source, run and access identity;
- runtime and gateway exact admitted journal heads match;
- result hash, observed model and canonical gateway aggregate match;
- `opencoderuntime.Inspect` (v1/v2) or `InspectComposite` with the existing
  journal-only verifier from `composite_backend.go` (v3) revalidates
  referenced subjournals; v3 additionally gates through the sealed
  `ValidateCompositeFinalGateway` transcript;
- `providergateway.Inspect` supplies the replay-validated normalized aggregate.

Display shows input, cached, uncached, output, reasoning and ordinary plus
call, receipt and pending counts where trustworthy. Cached and cache write
remain subsets of input and reasoning a subset of output, with checked
subtraction. Unknown optional axes remain null, never zero. Money remains
unknown; no cost is derived from subscription or price estimates and no SDK
normalized zero becomes provider evidence.

Pending or uncompleted admitted work remains explicitly `UNKNOWN` with no
totals and never infers completion from gateway receipts or zero spend.
Partial receipts may appear with clearly unmatched provenance only when
binding can be proven; missing totals are never synthesized. Completed
effects are never reported as effect `UNKNOWN` because accounting was
omitted. A known receipt with a missing, corrupt, changed or detached source
rejects with a safe bounded diagnostic that carries no paths, prompts or
ciphertext.

Scope of evidence is the current normalized gateway receipts. Raw
native-usage receipt migration and provider gateway rewrites are explicitly
out of scope. Direct-provider adapters remain outside current scope and are
omitted. Composite v3 turns verify through the existing journal-only
verifier only with an exact scheduler binding: `fabric usage RUN --schedule
SCHEDULE_ID` validates the repository, schedule, and run binding with the
existing schedule checks, then additionally requires the schedule's static
or dynamic tasks to reference the exact selected run ID and controller
path, and measures through `MeasureRunUsageWithScheduler`, which threads
the exact scheduler into the v3 verifier without searching directories or
changing effects. A same-repository schedule for another run is rejected
without emitting output or mutating journals. Plain `fabric usage RUN` is
unchanged and fails safely for composite turns with no exact scheduler. No
token accounting becomes authority and no new durable schemas or journal
writes are introduced.
