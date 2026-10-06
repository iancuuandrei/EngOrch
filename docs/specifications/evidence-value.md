# Finite advisory evidence value v1

`evidence-value RUN REQUEST_JSON` MUST read a bounded regular file and reject
duplicate/unknown JSON fields. Numeric estimate strings MUST decode to finite,
bounded nonnegative values; unavailable cost MUST remain null. The command MUST
validate run/source/candidate/journal-head identity against one replay-validated
controller snapshot. Its template MUST derive the same binding without dispatch.

Request and result MUST conform to canonical v1 without introducing float JSON
records or modifying historical schemas. The result MUST identify the canonical
request hash and mark its supplied decision model unvalidated and recommendation
advisory. This command MUST NOT mutate journals, repository files, working
contexts, policy, budgets, effect state, role capabilities or acceptance state.

Finite models MUST obey documented action/outcome/utility/resource/read bounds.
All action models MUST share marginal expected decision utilities. Calculations
MUST reject unsupported action/resource names and duplicate action/constraint
identities. Selection MUST be deterministic across input action/resource order.
Unknown active costs and insufficient remaining budgets MUST exclude actions;
Every supplied cost MUST have a declared resource constraint; otherwise the
request MUST be rejected. Unknown costs MUST NOT become known monetary zero.
No strictly positive known estimated JEV MUST produce no recommendation.

Unresolved admitted effects MUST suppress recommendations, including when another
stop condition also applies. Diagnostics MUST NOT settle UNKNOWN or authorize
resend. READY and inactive lifecycle MUST suppress additional research proposals.
Every later execution still requires existing task, role, ownership, source,
candidate, runtime, effect and budget admission. A positive score cannot replace
those gates or count as verified engineering quality.

Projected price updates MUST use the explicit allocation and accumulated
gradient magnitude, preserve resource limits/used values and reject invalid
observations. Supplied priors/counts/prices are not automatically validated
measurements. This increment MUST NOT claim automated acquisition or measured
quality/token/time benefit.

## Optional bounded source acquisition

`evidence-acquire RUN REQUEST_JSON` MUST accept a bounded canonical request
containing `model` and a query map matching the action frontier exactly. Only
`source_read` actions are admitted. Each query MUST be nonempty UTF-8 and at most
16 KiB. Input parsing MUST preserve existing strict field/duplicate-key rules.

The controller MUST bind the request to the exact run/source/candidate and
producing journal prefix. A decision append MUST recompute the entire supplied
model, targets, report and decision identity under journal semantic validation.
No-positive-value or unavailable-cost stopping MUST acquire no source evidence.
UNKNOWN, READY and inactive lifecycle MUST cause neither acquisition nor a
decision append. The existing UNKNOWN journal guards remain mandatory.

Only an eligible selected query MAY enter existing explorer TaskContext
admission. Existing candidate, lease, sensitive-path, file/byte and manifest
checks MUST remain unchanged. A new source record MUST bind its admitted
decision ID. Preparation before capture, the refreshed snapshot after capture,
and semantic append replay MUST reject changed run/source/candidate, selected
query or controller stop state. A compatible already admitted context MAY be
reused without changing its historical record. Selection alone MUST NOT count as acquisition,
task completion, acceptance or effect authority. Failure after decision append
MUST retain that decision and report its identity and legal next action.

The command MUST NOT dispatch a model, mutate working context, change scope,
permissions, repair budgets or acceptance. Resource costs remain declared
estimates, not measured expenditure. Historical absent-policy behavior and
absent-field serialization MUST remain compatible. Later autonomous JEV policy
and provider-bearing acquisition remain outside this initial integration.

### v1.1.25 advisory reuse in bounded task-context selection

Selected source paths from successful decision-associated explorer admissions
MAY inform existing bounded task-context selection for subsequent roles as
advisory path hints. The controller MUST reuse the existing exploration-hint
seam: hints MUST bias only read prioritization and selector `PathHints` within
unchanged file/count/byte limits, and MUST NOT inject manifests or summaries,
grant scope, remove required context, dispatch providers, repeat effects, or
change acceptance. Existing source/candidate checks and selector limits MUST
remain.

A hint MUST be admitted only from a positive selected decision with a matching
admitted explorer context for the exact current source, candidate and selected
query. Stale candidate/source, wrong role/query, no selected action, missing
context, malformed linkage and unavailable contexts MUST NOT contribute.
Same-query reuse MAY match one validated explicit decision to a historical
matching context without `EvidenceDecisionID` only when its identities, role
and exact query match; a successful read MUST NOT be treated as semantic truth.
Hint paths MUST be deduplicated and sorted deterministically with existing
exploration hints without filesystem reads. Historical runs without EVC
decisions MUST behave identically.

## Observed token feedback

`evidence-feedback` MUST bind a bounded model, per-axis allocations and consumed
invocation cursors to the exact inspected controller prefix. Verified run usage
MUST match that same run/head. Only matched completed Codex receipts with clean
OBSERVED typed accounting and no pending calls MAY supply token costs.

Costs MUST split accounting deltas into uncached/cached input and
ordinary/reasoning output without charging the parent totals again. Invalid
subsets MUST be rejected. Missing or non-token costs MUST remain unavailable;
they MUST NOT advance a cursor, change a price or invent zero consumption.

Each declared axis/invocation MUST update existing projected AdaGrad state at
most once per supplied cursor. Known observations MAY increment advisory used
quantities; limits MUST remain unchanged. The caller-supplied baseline/cursor
MUST NOT be represented as authenticated cumulative accounting or budget
authority. The result MUST identify its usage basis and reusable request.

Request size, observations, axes, cursor lengths and numeric ranges MUST be
bounded. Unknown/duplicate cursor identities or cursors for unavailable costs
MUST be rejected. The operation MUST NOT write a journal, change runtime policy,
settle UNKNOWN, dispatch any action or grant engineering acceptance.
