# Existing objective binding

Use the unchanged `go-humanize-feature-performance` task text from
`evals/v1/manifest.json`, pinned to
`github.com/dustin/go-humanize@a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`.
This fixture does not create an alternate task objective.

The target topology for the cap-2 treatment is:

1. One read-only research/design task examines the public API, package tests,
   existing helpers, and exact test/write boundaries, then identifies safe
   disjoint scopes. It produces no candidate-file effect.
2. Two ready implementation tasks address the existing ParseBytes and Commaf
   requirements. One owns only `bytes.go` and `bytes_test.go`; the other owns
   only `comma.go` and `comma_test.go`. Both may read the package and its full
   tests. Neither leaf owns `common_test.go` or the other leaf's files.

The design step is useful because the two features live in one Go package and
the existing tests share `testList`, while the implementation paths are
independent. There is no algorithmic dependency between `ParseBytes` and
`Commaf`; do not claim one. The full package test and existing public behavior
are the integration hub.

These are properties to measure, not fixed task IDs or a prescribed plan. A
single cohesive implementation task can still be a valid product outcome. It
is acceptable in the cap-1 control, whose purpose is to provide a serial
comparison, but it does not qualify the cap-2 treatment's two-leaf topology.
