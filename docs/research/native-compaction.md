# Automatic Codex compaction: source evidence and qualification boundary

The source pin below establishes that Codex can automatically compact inside
one model turn while it is handling tool calls. It does not establish that a
Fabric run or a live provider/model will reach the configured threshold.

- Repository: [`openai/codex`](https://github.com/openai/codex)
- Commit: [`b172810921f89847cd310ecc496f9c901760e933`](https://github.com/openai/codex/commit/b172810921f89847cd310ecc496f9c901760e933)
- License: Apache-2.0, as declared by the repository's [`LICENSE`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/LICENSE)
- Same-turn sampling and tool loop: [`codex-rs/core/src/session/turn.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/core/src/session/turn.rs)
- Threshold calculation: [`codex-rs/core/src/session/context_window.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/core/src/session/context_window.rs)
- Thread-start config: [`codex-rs/app-server-protocol/src/protocol/v2/thread.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server-protocol/src/protocol/v2/thread.rs)
- Config key declaration: [`codex-rs/config/src/config_toml.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/config/src/config_toml.rs)
- Upstream app-server compaction test: [`codex-rs/app-server/tests/suite/v2/compaction.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server/tests/suite/v2/compaction.rs)

In `session/turn.rs`, after a sampling step and its tools complete, Codex
checks whether the turn needs a follow-up and whether the token limit was
reached. If both are true, it runs automatic mid-turn compaction and continues
the same turn loop. This is the relevant path for a single long writer turn
with ordinary tool iterations; a second `turn/start` is not required. The
threshold is calculated from `model_auto_compact_token_limit` (or model
metadata) and the configured scope, independently of the hard full-context
window limit. `ThreadStartParams.config` is a JSON key/value override map, and
the app-server thread-start handler loads those overrides. A future opt-in can
therefore set the threshold per thread without editing the private CODEX_HOME
config.

The upstream test at this pin uses mocked responses and multiple user turns to
exercise notification pairing. Its 200,000-token setting and synthetic usage
are fixture values, not a recommended/live model limit. They do not qualify
automatic compaction on a production model. Live qualification must bind the
actual binary, model/provider/effort, thread config, invocation and run; it
must observe the exact same compaction item ID on started and completed
notifications and a completed turn. If the threshold is not reached, the
result is “not observed,” not zero compactions or a successful compaction test.

## Proposed bounded option (not implemented here)

Keep the current no-option behavior and serialized identities unchanged. Add
an omitted-by-default, versioned immutable run option such as
`Execution.AutoCompact = { version: 1, token_limit: N }`, selected only by an
explicit init CLI flag. Validate `N > 0`; do not infer a threshold from bytes,
token estimates, or an assumed model window. Bind this option through the
existing immutable run creation/config hash and include the exact thread-start
override in the thread/runtime intent and observed thread binding. The
Codex-host's generated CODEX_HOME config remains fixed and hash-validated.

Qualification should use one bounded coding task with a real native test,
`max_parallel=1`, no repair allowance, and one writer invocation. The task
should require several useful inspect/edit/test tool iterations, not filler
turns. Set a conservative explicit threshold only after validating it against
the selected model's effective window. Require exact compaction lifecycle
pairing inside the same invocation, completed final turn, accepted candidate,
native verification, review, and held-out acceptance. Enforce existing
invocation/token budgets and stop on UNKNOWN; do not issue manual compact RPCs
or resend an uncertain turn.

Fabric currently sets no auto-compaction threshold in `StartThreadWithTools`;
its Codex-host private config is intentionally fixed. The current source-level
evidence and mock tests are not live qualification. This note proposes no
runtime or controller changes.

## Fresh-context rounds after READY

An accepted `READY` run remains terminal. The existing local commit path can
integrate its exact verified candidate without reopening that run:
`prepare-commit` captures a candidate only when the run is READY and the
workspace outcome is confirmed; `commit` requires the matching preview intent
ID and explicit actor. A following `run --autonomous` rediscovers the exact
current repository HEAD/tree and creates a new run identity from that source,
the same harness config, and a new objective. Its role invocations start fresh
runtime threads. This provides a separate, auditable long-horizon round using
existing operations, while retaining both prior run journals and their
candidate/review/verification bindings. It is distinct from compaction within
one writer turn and does not continue a terminal READY controller.
