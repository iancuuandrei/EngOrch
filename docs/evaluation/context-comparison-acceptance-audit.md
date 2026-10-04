# Context comparison acceptance audit

Read-only follow-up of the frozen comparison found two distinct failure modes.
Historical results and held-out inputs remain unchanged.

The context-v2 difflib candidate returns a nil zero-length slice for empty
input. The public objective says empty input has no lines. The held-out
assertion uses `reflect.DeepEqual` against a non-nil empty slice and therefore
requires a representation not specified by that wording. Independent source
review found no public API comment or existing test specifying non-nil empty
representation. The frozen result stays FAIL under its actual gate. It must
not be retroactively passed or used to claim 6/6.

A future protocol could explicitly specify non-nil representation publicly,
or test element count when representation is outside the contract. Neither
change is made here; current fresh comparisons retain their frozen gates.
Hidden acceptance details are not added to agent prompts.

The contract-arm logr candidate changed only the example sink in
`examples/tab_logger.go`; the required `Logger.WithValues` implementation in
`logr.go` remained unchanged. Its held-out failure is a real target-selection
failure. Source selection versus model choice requires separate inspection;
no outcome or successful fix is inferred from the example edit.