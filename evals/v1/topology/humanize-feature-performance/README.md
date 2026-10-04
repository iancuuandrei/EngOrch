# Humanize disjoint-feature topology

Runner-supported, not-yet-qualified matched-pair recipe for the existing
`go-humanize-feature-performance` manifest task. It adds no manifest row and
does not modify the objective or held-out checks.

The pinned source confirms two disjoint implementation clusters: integer
separator parsing in `bytes.go` with `bytes_test.go`, and floating-point
formatting/allocation behavior in `comma.go` with `comma_test.go`. They do not
call each other's implementation. The real shared hub is the package contract
and complete package test surface, including the shared `testList` helper in
`common_test.go`; it is not a shared implementation algorithm. A read-only
design task can check those boundaries before implementation becomes ready.
The two-leaf topology is specifically the cap-2 treatment outcome. The cap-1
serial control may validly select one cohesive implementation task because
the cap is also a planner input; that is a valid control result but does not
qualify the two-leaf topology. Do not edit or replace either accepted plan to
force the shape.

- [Pinned objective and boundary](objective.md)
- [Fixture descriptor](fixture.json)
- [Matched acceptance recipe](acceptance.md)

Both arms enable the same candidate-bound reviewer-impact context using the
manifest task unchanged, `go-contract-context-v1`, and one explicit pinned RI
executable/hash. Both arms use the isolated implementation contract and the
same resource policy; the treatment changes only the scheduler cap from one
writer to two. The reviewer context is held constant in both arms and is only
an advisory topology projection; it does not prescribe the task plan or prove
independence. Prepare records zero provider calls, and the existing native
full-package and two composed held-out checks remain mandatory.

For the cap-2 treatment, qualification additionally requires a read-only hub
followed by two ready, disjoint implementation leaves and positive overlap of
their controller dispatch intervals. The cap-1 control is compared on normal
acceptance outcomes and non-overlap; it is not required to emit the same two
implementation tasks. A one-task cap-1 plan is reported as such, not
replanned.
