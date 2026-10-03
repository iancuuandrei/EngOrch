# Native Codex automatic compaction

Autonomous runs can opt Codex app-server threads into the native automatic
in-turn compaction threshold with `run --autonomous --auto-compact-token-limit N`.
The value must be between 1 and 10,000,000. Omit the flag to preserve the
existing thread configuration and invocation identities.

The threshold is part of the immutable run policy and each Codex invocation
identity. The runtime records the requested value before `thread/start`, then
sends it as `config.model_auto_compact_token_limit`. The recorded thread field
describes the requested config; it is not evidence that compaction happened.
The setting does not request manual compaction and does not create another
turn or retry.

The upper bound is a configuration-input bound, not a claim about a model's
context capacity. Whether native compaction occurred must be established from
matched lifecycle observations; merely enabling the threshold does not prove
that it triggered.
