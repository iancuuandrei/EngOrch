# Imported spectrum in fixer inputs

## Scope

v1.1.17 adds explicit pre-invocation spectrum admission through existing
`task.context-admitted` evidence. The same ready task, candidate, failed gate,
query and complete selected source bytes are validated during admission and
replay. Profiles remain caller-supplied and untrusted. This capability does not
collect coverage, execute tests, grant retry rights or close findings.

## Requested checks

Targeted deterministic regressions cover exact reuse, frozen-context rejection,
foreign candidate/run/task/gate, malformed recorded source, missing selected
bytes, advisory write-scope filtering, raw-profile omission, serial/task-bound
query matching and stricter 128 KiB admission bounds. A real temporary Git
repair lifecycle consumes the admitted evidence and still requires fresh native
verification and independent review to reach READY. Runtime role responses in
that fixture are deterministic; it is not live model qualification.

## Executed evidence

Targeted Go checks: **PASS** for faultlocalization (4.400 s), control (59.457 s)
and CLI (0.191 s). The set includes the new imported-context repair lifecycle,
existing repair/reviewer-recheck lifecycles, spectrum parser/source-binding
regressions, compact writer intelligence and generated reference checks.
Additional admission-bound, scoped-projection and CLI argument checks: **PASS**
(control 0.158 s; CLI 0.098 s). Vet on the three changed packages and diff
whitespace checks: **PASS**. Documentation checks: **PASS** (0.580 s).
Compatibility checks for legacy invocation bytes, task-context replay/forgery,
scoped writer contracts and real CLI plan/approval/resume: **PASS** (control
43.404 s; CLI 39.189 s; documentation 0.575 s).

Independent GPT 6 Luna High static review: **APPROVE**, with the normal selected
source limits explicitly noted. The reviewer did not execute checks or provider
calls and does not establish live model benefit.

Live matched repair quality, time/token benefit, automatic acquisition and
release/installed acceptance remain **NOT RUN**. The earlier WorkingContext
pilot does not qualify this separate repair capability.

See [usage](../guides/repair-localization.md) and
[contract](../specifications/repair-diagnosis.md).
