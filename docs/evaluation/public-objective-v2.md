# Public objective clarification, suite v2

The original public difflib task already specified:

> Nonempty text without a trailing newline must retain the current convention of returning its last line with a newline; existing newlines must be preserved.

That task document predates these evaluations (introduced in commit
`2db168d5c106c1266c7d585a9d20e1f255d98ab9`). The manifest's shorter objective
omitted this convention. The pinned implementation and upstream regression
also preserve it. Independent review confirmed an ambiguity in the text given
to the model, rather than an unsupported held-out expectation.

`fabric-v1-real-repository-2026-10-03-objective-v2` includes this existing public
requirement in the model objective. Task pins, native commands, held-out
checks, acceptance criteria and repair limits are unchanged. No hidden input
or expected result is disclosed in the objective. The objective contract
harness checks the public sentence, suite identity and unchanged task bindings.

Historical FAIL/BLOCKED results and fixed-six 5/6 remain unchanged. New cohorts
must freeze the new manifest and runner provenance. A comparison must give
both baseline and candidate the same clarified objective; an old-objective
baseline is not a matched control. A new 6/6 claim requires six actual accepted
results in that frozen cohort, followed by the applicable review and integration.
