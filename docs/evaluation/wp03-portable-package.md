# WP03 portable local package: scope and evidence

This record summarizes the changes merged by [PR #2](https://github.com/iancuuandrei/Fabric/pull/2). It is an implementation and handoff record, not a release note: the package produced by these scripts is local, unsigned, unpublished, and never release-qualified.

## Merged identity

- Merge commit: `c32ca7d3b35dac3010fd956d83fcef107dd3e859`
- PR head before merge: `d29ccef42bc739b9242c2fa256f2eb02a106faa1`
- PR: [WP03: make local package builds portable](https://github.com/iancuuandrei/Fabric/pull/2)

The six-commit series was:

1. `44d7794` — document the portable local package path.
2. `f0deb57` — add Windows PowerShell 5.1 package-path compatibility.
3. `b2eda91` — preserve portable package path handling.
4. `ae8ce51` — preserve source identity on PowerShell 5.
5. `231f953` — bind portable Go and Rust builds.
6. `d29ccef` — complete package inventory and verification.

## Changes

- `scripts/package-local.ps1` now requires absolute paths to the selected Go, Git, Cargo, and Rust compiler executables. It does not search `PATH` or rely on private operator toolchain folders. Output and build directories must be new, distinct, and non-overlapping; repository-local destinations must already be ignored by Git.
- The build uses the supplied Go and Rust executables, isolates the relevant environment settings, captures source identity before and after building, inventories the complete package payload, and rejects duplicate file entries. It writes checksums and a manifest, performs smoke checks, and invokes the package verifier.
- `scripts/package-local-compat.ps1` supplies path qualification and relative-path operations compatible with Windows PowerShell 5.1. The packaging and verification scripts use it instead of newer .NET path APIs.
- `scripts/verify-local-package.ps1` verifies package identity, exact file inventory, sizes, SHA-256 values, and smoke behavior. It rejects a manifest that claims a signed/published release or release qualification.
- `scripts/test-package-local-portable.ps1` adds offline negative coverage for private-toolchain assumptions, invalid tool and destination paths, tampered or extra package files, malformed checksum data, and false release claims.
- `README.md` and `docs/guides/local-packaging.md` expose the workflow and clarify that it creates a local package, not a release.

## Evidence and limits

GitGuardian and SonarCloud both reported success on PR #2. The added portable-path regression script also passed in Windows PowerShell `5.1.26100.9444` against the exact PR head before merge. It covered its path, inventory, tampering, and release-label negative cases.

That focused regression is not an end-to-end build using real Go and Rust toolchains. The follow-up validation for this handoff ran only that focused script; it is not evidence of a full clean-checkout package build. Do not infer installer, signing, publication, release qualification, or a clean-machine installation experience from these results. The authoritative usage instructions remain in [the local packaging guide](../guides/local-packaging.md).

Review note: the verifier's explicit lexical `../` pattern is over-escaped and does not match traversal segments as intended. The subsequent full-path containment check and exact on-disk inventory comparison still reject such paths; this finding did not identify an admitted path escape. Add a focused traversal-negative regression and correct the redundant lexical check before extending this verifier further. No product code was changed in this documentation follow-up.

## Next external-readiness step

On a clean checkout, run the documented package and verification commands with real, explicitly selected Go, Git, Cargo, and Rust compiler executables. Retain the terminal evidence and inspect the produced manifest and checksums. Keep any outcome labeled as local packaging unless a separate release process establishes signing, distribution, and release qualification.
