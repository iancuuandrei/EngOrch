# Writer behavior preservation

The writer/fixer instruction now explicitly requests preservation of established
behavior outside the requested feature, comparison of base/candidate boundary
behavior, and focused discriminating regressions. Parser/stateful-format work
must consider quote, escape, comment and end-of-input transitions using lexical
boundaries. This applies to every supported writer contract and the fixer role.

This is general implementation guidance. It embeds no benchmark identities,
heldout cases or expected outputs and changes no schemas, tool permissions,
effect ownership, retry limits or verification/review gates. Model quality is
still evaluated from actual candidates, not inferred from the instruction.

Focused regressions across all nine writer contracts and the fixer role: PASS
(14.329 seconds). Controller vet: PASS. Independent read-only review: APPROVE.
The earlier capacity failure evidence remains in
[the retained continuation report](explorer-capacity-continuation.md).
