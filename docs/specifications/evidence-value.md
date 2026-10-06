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
