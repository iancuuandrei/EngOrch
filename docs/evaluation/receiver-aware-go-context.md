# Receiver-aware Go context selection

`--planner-context go-contract-context-v2` opts into bounded declaration-aware
selection. The existing default and `go-contract-context-v1` remain unchanged.
The option requires the exact pinned RI executable. It reuses the existing RI,
graph, context compiler and parse cache.

The selector extracts dotted receiver/method identifiers from the public task
objective and inspects digest-checked committed Go source. Exact declarations
rank before unrelated declarations, independent of their count. Pointer and
generic receiver forms are supported. Calls and example filenames do not prove
an exact declaration. This is syntax evidence, not semantic call resolution.

Discovery admits at most 48 paths within the existing 8 MiB source budget;
the retained graph still admits at most 24 files and 512 KiB. Omitted source
and graph coverage remain explicit. Persisted planner record v5, hash domain
v4 and selection version 2 bind the recipe for replay. Review impact and the
candidate facts cache accept both contract context versions with their existing
source/candidate/module validation.

## Evidence and limits

The preserved logr contract cohort failed after root `logr.go` was omitted and
the accepted plan assigned implementation to an example sink. This motivates
the generic declaration selector; no held-out assertion is added to prompts.
Synthetic tests cover more than 24 decoys, exact ranking against 32 unrelated
methods, generic receivers and resource bounds. A pinned committed logr
fixture admits the root API declaration. Independent core and CLI review
approved the slice after correcting count-based ranking and generic receivers.

These checks establish selection behavior, not a successful new coding task.
The historical fixed-six cohorts remain 5/6 with different failing tasks.
Their passing sets are not pooled. The difflib held-out gate remains unchanged.
Fresh full qualification and actual coding-task evaluation are pending.
