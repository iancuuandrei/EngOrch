# Native autonomous repository pilot

Executed on Windows amd64 on 2026-10-02. This demonstrates one bounded native
coding journey, not full Fabric v1 acceptance or a release qualification.

## Identity and task

- Repository: `dustin/go-humanize`, source
  `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`.
- Task: validated underscore separators in `ParseBytes` integer inputs,
  preserving overflow detection and existing behavior, with regression tests.
- Fabric baseline: `5a4160211b06f3bf752e4e5ff57fbb708192f968` plus the reviewed
  native increment. The pilot executable was built before that increment was
  committed; this is local binary evidence, not an exact-commit release claim.
- Executable SHA-256:
  `062716e564142eacdd2c23fb60ba7d999a01eef607cf35688a3f593e7aaa377f`.
- Run: `070bd6f750bc7aea38f05a767e0ad6da47669db2ccd72219cec1498ab304a311`.
- Runtime: stock Codex 0.159.2; all four roles used `gpt-6-luna`, effort `high`.

## Observed result

`fabric run --autonomous --max-repairs 2` progressed through planning,
exploration, writing, verification and independent model review without an
intermediate operator approval. Terminal state was `READY`; file outcome was
`CONFIRMED`. The proposal changed `bytes.go` and `bytes_test.go`: 62 insertions,
one deletion. No commit or push was performed by the run.

The required `go test ./...` observation was `PASS`, exit 0, bound to candidate
`a374cd74498d511cad790e684f6046ddf796aa326ed5a0f0be2a6daea463e04c`.
Reviewer decision was `approve`, with zero findings and the same candidate ID;
verification plan was
`a4061dc19567d214cee07456b3e6987431f72517248c151f3f23cd2bcaa8ca6b`.
All four runtime journals had matched completed receipts and zero pending calls.
No repair was required.

Independent acceptance copied 39 source files byte for byte into a disposable
directory, checked each SHA-256 against the original candidate, then added the
hidden `humanize` test. `go test ./... -count=1` passed in that copy. The initial
copy attempt incorrectly filtered on the absolute workspace path and copied no
files; its setup failure was retained and excluded from product results. The
corrected copy used repository-relative exclusions. No provider work was resent.

## Evidence and limits

Private local evidence is retained under
`D:\dev\Fabric-v1-product-artifacts\native-pilot-20261002`: binary identity,
run snapshot, usage receipt summary, exact copy inventory and test stdout/stderr.
Raw runtime state and credentials are not included in this repository.

The native core/effects suite passed, as did focused restart, uncertainty and
repair-budget regressions. This pilot does not establish multi-task quality,
repair success, concurrent execution, adaptive routing or reproducible release
acceptance. The six-repository matched PR #5 comparison remains pending.
