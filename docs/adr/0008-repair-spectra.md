# 0008: Imported test spectra remain advisory

Status: ACCEPTED
Date: 2026-10-05

## Context and problem

Repair diagnosis preserves failed gates and source anchors, but cannot use
per-test coverage to discriminate locations executed by passing/failing tests.
Current journals do not authenticate per-test coverage production. Treating an
imported profile as native authority would invent evidence and closure.

## Decision and rationale

Add a pure bounded Go-profile/Ochiai analyzer and optional read-only diagnosis
interface. Validate complete source hashes/ranges using the existing read
lease and exact run/candidate binding. Keep outcomes and profile-to-source
mappings caller-supplied and untrusted even after byte validation. Retain raw
test counts and exact squared fractions rather than floating canonical output.
Never mutate findings, scopes, invocation inputs, budgets or acceptance.

## Alternatives and consequences

Automatic test execution/receipt admission and graph/ranking fusion are later
integrations, requiring their own evidence; they are not silently inferred by
this importer. A new coverage DSL, dependency or arbitrary weighted score is
unnecessary. Strict matching block inventories avoid treating omitted blocks
as uncovered. Bounded evidence can be unavailable; base diagnosis remains useful.

## Compatibility and validation

The optional spectrum field is omitted by default. Legacy diagnoses, persisted
records and runtime invocations remain unchanged. Validation includes actual
individual Go test profiles, deterministic counts/ties, encoding/allocation
bounds, foreign/stale candidate rejection, complete source/range observation,
guard-close failure and unchanged journal exports. This does not qualify live
repair quality or final v2 acceptance.

## References

- [Guide](../guides/repair-localization.md)
- [Contract](../specifications/repair-diagnosis.md)
- [Scoped evidence](../evaluation/repair-spectrum-localization.md)
- [Go coverage profile documentation](https://pkg.go.dev/golang.org/x/tools/cover)
