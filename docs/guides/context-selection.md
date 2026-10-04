# Context selection

Opt-in experimental ranking for bounded role task context. This is product
context selection, not governance or a general memory framework. Self-hosting
is optional; effect, UNKNOWN, authority and controller behavior are unchanged.

## Use

```sh
fabric run --autonomous --context-selector rrf-coverage-v1 "Objective"
```

Empty (omitted) preserves the historical selector. `rrf-coverage-v1` requires
the existing `bounded-v1` context and is frozen at run creation. Unknown modes
are rejected before any provider dispatch.

## What changes

- Read prioritization and excerpt selection fuse independent rankers
  (changed paths, hints, path tokens, content terms, anchors when present)
  with equal unweighted RRF (`k=60`).
- A bounded coverage pass then prefers complementary excerpts by new concepts
  covered per excerpt byte, reducing near-duplicate selections.
- Each selected item records its ranker ordinals (`ranks`) and newly covered
  concepts (`covered`) for inspection. No fused score is stored.

Limits, secret-path filtering, UTF-8 handling, omission reporting and
invocation binding are unchanged. Inspect decisions with `fabric inspect RUN`
and the persisted `task.context-admitted` records.

## Status

Model-quality and downstream efficiency remain NOT QUALIFIED until the
predeclared matched historical experiment runs after freeze. Do not promote
this mode on fixture rankings alone. See the [selection contract](../specifications/task-context-selection.md)
and [product plan](../roadmap/v2-product-plan.md).
