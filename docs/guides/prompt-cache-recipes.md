# Prompt cache recipes

Autonomous runs can opt into the versioned `cache-prefix-v1` prompt recipe:

```powershell
fabric run --autonomous --prompt-recipe cache-prefix-v1 "Describe the bounded change"
```

The recipe is recorded in the immutable execution policy. It preserves each
role prompt's JSON members and values, but emits the role instruction first,
then the variable run/candidate context, and the output schema last. Exact
candidate-bound schemas remain intact. Anchored-edit writers also receive
explicit guidance to reuse the exact validated edit call values; reviewers are
reminded to preserve test expectations and contracts unless a justified change
is part of the objective.

The empty recipe remains the legacy representation and is used by default.
Changing the recipe changes invocation identities, so an existing run retains
the recipe with which it was created.

Request ordering only makes a stable prefix available to a provider cache. It
does not guarantee cache admission, a cache hit, lower latency, or lower cost.
Evaluate provider-reported cache read/write usage on matched runs before making
performance claims. OpenCode's Responses path binds its cache key to the
generated tool-session identity; this is not a cross-run cache identity.
