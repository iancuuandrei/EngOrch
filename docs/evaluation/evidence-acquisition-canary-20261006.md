# Evidence acquisition canary 2026-10-06 (frozen, UNKNOWN preserved)

Private frozen canary at exact clean Fabric source
`31bc4cadbdc493037b7b0edf8c994a1eafd5b4aa`, binary SHA
`484596fb8be96b1c007abfd55012e7382c4557640fa9be53cdc6d07b1c15c8d0`.
Public task go-humanize source `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`,
unchanged oracle SHA
`8402e2584a029fcca0a1751be2445d476947ae4e773a2e6a81b0ece058a480ff`.

All five configured roles `muse-spark-1.3-contributor` high over the
opencode-http/Responses/Go route. Bounded dispatch: max-parallel 1,
max-repairs 2, invocation timeout 600 s, max calls 16, overall frozen wall
limit 3600 s. Policy: finite 1 `source_read` action `inspect-parsebytes`,
query Locate ParseBytes in `bytes.go` and its existing integer parsing and
overflow behavior, plus explicitly unvalidated illustrative
`local_compute_ms` estimates. Frozen plan SHA
`3c98698fe59ccc59dafb6d69e7e14c929f48d16cc6428a6070a5102e51758c00`.

Run `e6601fc73d01f4ae1115ccfb6bd8c9d75fd73747af277c91d1e3188262371e0d`:
exit 1 after 79.906123 s, state PLANNING, lifecycle ACTIVE, agent dispatch
unknown. Native, review, held-out, and EVC decision NOT RUN; tokens and cost
UNKNOWN. Provider gateway journal holds only `provider.bound`; zero accepted
calls and zero receipts. This task never reached writing.

Private local preflight stopped dispatch in phase `validate-request`: invalid
Responses system message, expected `system` but the SDK sent `developer` with
reasoning effort high, summary auto, include `reasoning.encrypted_content`.
Diagnostic SHA
`40363539a059a891432a254aff8674b355bd334de19bcec372c479231a5371a6`.
OpenCode log showed BadGateway then Conflict. Local preflight is diagnostic,
not a terminal runtime effect receipt; UNKNOWN is preserved and must not be
resent. This is not retry authorization. No provider success and no final EVC
acceptance are claimed.

Same retained run also proved the CLI sidecar listing bug: the old binary
failed `status` and implicit `inspect` with an unknown-controller event,
while explicit `inspect RUN` passed. With the dirty repair binary, `status`
and implicit `inspect` both exit 0 on the same retained run;
PLANNING/UNKNOWN unchanged.

Executed scope: pinned Go targeted three latest/listing tests PASS 4.191 s,
vet CLI PASS, doc check PASS 0.441 s; full CLI PASS 86.6 s in the writer
session. No full Go, race, Rust, cohort, or release qualification is
inferred. Baseline canary failures are unchanged. First independent Muse High
review returned REQUEST_CHANGES on the CLI sidecar repair scope; final
review is pending and no APPROVE is claimed.
