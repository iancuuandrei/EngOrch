# Fabric v1 first-task acceptance

Status: PASS for the documented Windows PowerShell path on 2026-10-02.
This is acceptance of one real coding journey, not a release or general sandbox
qualification. The implementation was developed and accepted in
[PR #5](https://github.com/iancuuandrei/Fabric/pull/5); the exact tested source
identity is recorded below.

## Executed journey

A new clone from GitHub at source commit
`395e3b9f55b5262dfda81ab3484b0d113849a7cb` built the executable with Go 1.27.1
on Windows amd64. Binary SHA-256:
`95ea1e2a634589ee8461b4a4cbcc3d951e4ba33b3d65c51d8b0a3885354dabdd`.
It used stock Codex CLI 0.159.2, `gpt-6.1-sol`, and low effort. The model was
observed exactly in each of the four distinct runtime invocations. Existing local
authentication was used without copying credential contents into project files.
No historical private runner, journal, qualification manifest or OpenCode patch
was needed. The host had the required tools and an authenticated account already;
installation of those third-party tools on a blank machine was not executed.

The sample task added name trimming and an empty-name fallback to `Hello`, plus
table-driven tests for normal, padded, empty and whitespace-only inputs. The
documented task script ran interactively: the operator inspected and approved the
exact plan and later the proposed source/test changes. The approval actor was
`codex:user-authorized-product-acceptance`; this was an automated acceptance
operator, not an independent human review.

| Step | Observed result |
| --- | --- |
| Clone/build | Clean public feature-branch checkout; successful source build |
| Configure/init | Four real roles generated; doctor PASS without model dispatch |
| Plan/delegate | Real source-reading planner; one admitted read-only explorer |
| Execute | Two files changed in an isolated worktree; file effect CONFIRMED |
| Verify | `go test ./...` PASS, child exit 0; source formatting and diff checks clean |
| Review | Separate real reviewer invocation; `approve`, zero findings |
| Inspect/restart | Fresh-process inspect/usage/export PASS; 30 validated controller events |

The result reached `READY`. All four runtime receipts matched, all reported
completed, and all had zero pending calls. The original project checkout and
the Fabric source checkout remained clean. The task script did not commit,
push or publish the sample changes. Monetary cost is unknown.

## Durable identities

- Run: `124891550ccaa8a01dd70aa9df51e06e983dff2536c94ccc74f1ea6fe35adcd1`.
- Approved plan: `42f0a9682e73a77c0bedbf8176588648a7a13030d842e1a918eafcc460c03fe0`.
- Result candidate: `9c65ca4be65ddb72b801cfa86857484331107807fd24350500705df03efe0da6`.
- Verification plan: `f7bddaee729a026bae2cb8cb8395e50e2d05850d9ac193daabd115dfdccbee65`.
- Controller head: `a016cdf078b69a9b421739697c7b89a953a33ff527dc8ece29041dd589f7b185`.
- Export SHA-256: `793fe5cad4a446c809c07fdbc869835276813814314b4734475635d2f219c3b7`.

The local project, isolated worktree and runtime journals are retained. A sanitized
receipt and the validated controller export are retained in the operator's local
acceptance artifacts. Raw transcripts and local authentication paths are not
published in this repository. Later README/evidence edits do not change the tested
implementation; the source identity above records exactly what was executed.

## Demonstrated failures and direct fixes

The first completed run exposed task-script verdict parsing and output-pipeline
bugs. Direct fixes were covered by a PowerShell regression test that reads a ready
run, returns one object, and dispatches only `inspect`.

A subsequent clean run, `e41a68e8b3932872d62a72484cb1faa48bf3a463aaa29bb1310d511bc71cd48f`,
stopped because its explorer returned an object where `summary` requires a string.
Its completed runtime evidence is preserved; it was not resent or relabeled as a
successful exploration. Re-entering that run through the task script stopped
before a new request and left its controller history unchanged.

The product fix declares the explorer's existing JSON shape and binds it to the
Codex output schema. New real-role configurations select this contract; existing
configurations retain their original invocation identities. Regression tests
check the wire type, reject substituted schemas before RPC, and reject object
summaries without admitting a semantic result. The final clean run above exercised
the fix successfully.

The model-generated sample regression tests were also executed against the old
implementation in a separate directory: padded, empty and whitespace-only cases
failed as expected. They therefore detect the behavior the task actually changed.

## Source checks and limits

CLI/configuration/documentation tests, focused race tests for the changed
setup/approval/explorer boundaries, full Go vet and diff checks passed. The full
`go test -timeout=15m ./... -count=1` suite passed on the tested implementation.

Its initial run failed because an old Muse recursive preflight unconditionally
required a private pinned OpenCode binary. That historical qualification is now
explicitly opt-in; its assertions are preserved. It is NOT RUN in default source
tests, and this acceptance makes no recursive Muse qualification claim.
Rust code was unchanged and Rust tests were NOT RUN for this change. Additional
platforms, model routes, signed releases and main-branch integration remain outside
this acceptance scope.
