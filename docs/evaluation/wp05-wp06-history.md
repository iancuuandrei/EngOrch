# WP05/WP06 program history and completion status

This is a retrospective of the self-hosted program's WP05 recursion and WP06
portable-package work, recorded on 2026-10-02. It separates product work from
bootstrap repairs and distinguishes local evidence from changes merged to the
public `main` branch. It is not a release note or a claim that Fabric v1 is
complete.

## At a glance

| Work | Recorded outcome | Public completion |
| --- | --- | --- |
| WP05 / WP05.1 | A bounded live-recursion scenario was exercised; WP05.1's guide-only increment passed its frozen four-check plan and was committed locally as `591ab62c099d29f2815f5a1327e1f3815bca8d56`. | Not merged to public `main`; it does not establish general recursive execution. |
| WP06.1 | The package-first-run candidate failed its frozen verification and the run was cancelled; its unintegrated candidate was not adopted. | Not complete. |
| WP06.2 | Package-first-run successor work exposed both a product verification gap and controller/recovery defects. The latest program record does not show a completed verification, independent approval, and integration for WP06.2. | Not complete. |
| Fabric v1 | Later bootstrap generations addressed defects in recovery, qualification, and execution infrastructure. They are not substitutes for completed product work packages. | Still open. |

## WP05: bounded recursion, not universal recursion

The first WP05 writer proposal was rejected before a file effect because its
guide description did not match the accepted child-task evidence. The successor
WP05.1 narrowed the change to `docs/guides/task-schedules.md`. Fabric applied
that exact file effect, ran the frozen checks (4/4 PASS), obtained an independent
APPROVE with zero findings, and recorded local commit
`591ab62c099d29f2815f5a1327e1f3815bca8d56` on
`feat/wp05-live-recursion`.

The guide's claim is deliberately narrow: one depth-1 child was accepted; two
depth-2 explorer tasks were queued, one returned an accepted result and one
ended `IRRECOVERABLE_UNKNOWN`; and a second turn used the same AgentID. This
does not demonstrate arbitrary-depth recursion, writer/reviewer children,
concurrent writers, or universal crash recovery. The program projection records
the local WP05.1 slice as closed but the full v1 goal as open. Its remote
readback at the time recorded no WP05 branch on `main`.

## WP06: package-first-run remains unfinished

WP06.1 proposed a portable package first-run flow. Its terminal fixer and
verification did not produce a qualifying candidate, so Fabric cancelled that
run and did not adopt its unintegrated candidate. A later bounded fixer proposal
was structurally valid but did not repair the frozen failing check; it was not
applied.

WP06.2 was created as a fresh package-first-run successor with the five-check
acceptance plan frozen. Its recorded result was 4/5 PASS: the
`package-first-run-independent` check failed because `scripts/package-local.ps1`
read `run`/`plan`, while Fabric's actual result uses `run_id`/`plan_id`. One
explicitly authorized additional fixer changed a PowerShell `$Args` parameter
collision, but the unchanged frozen check remained FAIL; it did not resolve the
field-name mismatch.

Further credential-free reproduction exposed adjacent smoke-contract gaps: the
smoke used an empty Git repository, so `doctor` failed because there was no
committed `HEAD`; subsequent smoke commands also did not match the documented
CLI contracts for `plan`, `approve`, `run`, and `inspect`. The independent
verifier did exercise a committed fake-route flow, but that did not make the
package smoke pass.

Separate planner-route experiments did not qualify an alternative provider:
retained HTTP 429 outcomes and a DeepSeek/OpenCode Go response-decoding failure
were recorded as failures, not successful plans. Those experiments did not
repair or complete WP06.2.

The WP06.2 path then required bootstrap work for exact invalid-proposal
settlement, replay/readback, and one-shot fixer authorization. This exposed
real infrastructure defects: for example, a promoted G0.221 command could not
replay the newly introduced event through the top-level controller, and its
added acceptance tests did not exercise the real persistence/CLI path.
Subsequent bootstrap packages attempted to strengthen those seams, but
bootstrap qualification or promotion alone is not evidence that the WP06.2
package itself passed.

As of this record, the authoritative program projection contains no WP06.2
closeout with the full frozen verification passing, independent review
approved, and product result integrated. Therefore WP06 must be described as
**attempted but not completed**.

## Why the generation count grew

WP identifiers describe product increments. G0 identifiers describe bootstrap
or qualification repairs needed to make the controller safely execute,
settle, or verify those increments. A G0 repair can be necessary and valuable,
but it does not satisfy a WP's product acceptance criteria. The WP06 history
shows how controller settlement and replay problems consumed substantial work
without delivering the intended portable first-run result. This distinction is
important when reporting progress: count verified, reviewed, integrated product
increments separately from bootstrap generations.

Historical failures and UNKNOWN outcomes remain failures/uncertainties; later
source repairs do not retroactively turn them into PASS or authorize resending
their effects.

## Public repository boundary and current position

The public PR history as of 2026-10-02 contains the v0.0.1 bootstrap checkpoint
(PR #1), the WP03 local portable-package work (PR #2), and its handoff record
(PR #3). PR #2 creates a local package and is explicitly not signing,
publication, or release qualification. Neither WP05.1 nor WP06 is represented
as a completed public product PR in that history.

The current program projection is later than WP05/WP06: it records G0.261 as
blocked during the external-user journey. Its verification is incomplete: some
checks passed, focused checks failed, and remaining checks were not run; a
verification-generated binary also changed the candidate workspace, preventing
a clean candidate readback. There is no commit or integration for that
candidate. This status does not change the WP05/WP06 outcomes above, and it is
not evidence that an external user can already complete the full setup and
first-run journey from a clean checkout.

## What would close the gap

For WP06 to be called complete, a fresh governed candidate must satisfy the
unchanged package-first-run contract from a clean checkout, pass every frozen
verification check with terminal evidence, receive independent review, and be
integrated. For the external-user claim, the public instructions and package
must also be exercised from a clean environment using explicitly selected
toolchains; local packaging must continue to be distinguished from a signed or
published release.
