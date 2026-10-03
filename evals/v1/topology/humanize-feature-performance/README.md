# Humanize disjoint-feature topology

Design-only matched-pair recipe for the existing
`go-humanize-feature-performance` manifest task. It adds no manifest row and
does not modify the objective or held-out checks.

The pinned source confirms two disjoint implementation clusters: integer
separator parsing in `bytes.go` with `bytes_test.go`, and floating-point
formatting/allocation behavior in `comma.go` with `comma_test.go`. They do not
call each other's implementation. The real shared hub is the package contract
and complete package test surface, including the shared `testList` helper in
`common_test.go`; it is not a shared implementation algorithm. A read-only
design task can check those boundaries before the two implementation leaves
become ready. If the accepted plan does not actually contain that read-only
hub and two disjoint leaves, record “topology not qualified”; do not edit or
replace the plan to force the shape.

- [Pinned objective and boundary](objective.md)
- [Fixture descriptor](fixture.json)
- [Matched acceptance recipe](acceptance.md)
