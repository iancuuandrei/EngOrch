# Public objective: atomic numeric text support

Add `encoding.TextMarshaler` and `encoding.TextUnmarshaler` for `Int64` and
`Uint64` with canonical base-10 text, `ParseInt`/`ParseUint` base 10 and 64-bit
bounds, unchanged values on invalid or overflow input, and regression coverage
for endpoints, zero, signs, round trips, and invalid input. These wrappers are
generated: inspect and update the shared `internal/gen-atomicint` template,
regenerate both wrapper files from that template, and keep generated output
exactly regeneration-consistent. Own the shared template integration as one
cohesive task or establish it as a prerequisite before splitting per-wrapper
implementation; do not describe the end-to-end change as file-disjoint while
the shared template remains unowned. Preserve existing concurrency and JSON
behavior.

This text is copied from the existing `go-atomic-numeric-text` manifest entry
at the pinned source revision. It does not alter that entry or its held-out
assertions.
