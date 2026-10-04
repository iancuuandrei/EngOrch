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

## Implemented opt-in and qualification boundary

The CLI exposes `run --autonomous --auto-compact-token-limit N`, bounded to
1–10,000,000. Omission preserves prior run and invocation identities. The
requested limit is bound to immutable execution policy and each Codex
invocation; the runtime sends it as
`config.model_auto_compact_token_limit` at thread start. The thread receipt
proves which value was requested, not that the runtime reached it. The
generated Codex host configuration remains fixed and hash-validated. See the
[operator guide](../guides/native-auto-compaction.md) for the exact binding.

Qualification must use a useful coding objective and its unchanged repair,
native, review, and held-out gates. Keep `max_parallel=1`, use a threshold
validated against the configured model's effective window, and retain the
existing invocation and token budgets. Require the same compaction item ID in
both started and completed notifications for one invocation/thread/turn and a
completed final turn. A configured limit without that lifecycle evidence is
“not observed,” not a pass. Do not issue manual compact RPCs or resend an
uncertain turn. The upstream mock and local protocol fixtures remain protocol
evidence, not live provider qualification.

## Fresh-context rounds after READY

An accepted `READY` run remains terminal. `prepare-commit` and `commit` create
and confirm a commit on that run's isolated `harness/<RUN>` branch. They do not
move the configured source checkout's `HEAD`; the source checkout is
deliberately preserved. Therefore, starting another run at the same `--root`
still discovers the old source `HEAD`, not the accepted candidate commit.

To base a new run on the accepted result, confirm the exact commit in the
original run's `COMMITTED` snapshot, then create a separate detached source
worktree at that commit. Copy the operator's `harness.toml` policy into that
checkout and give it a distinct external `controller_state_root`. Start the
next run with that checkout as `--root`. The new journal binds the accepted
commit/tree and its roles start fresh runtime threads, while the old run stays
terminal and its journal is retained. The [verified fresh rounds guide](../guides/verified-fresh-rounds.md)
provides a not-executed Humanize recipe. A fresh round and in-turn compaction
are separate: only the latter requires a compaction lifecycle pair within one
long-lived role invocation. Resume only the exact nonterminal run, preserve
`UNKNOWN`, and never resend an uncertain invocation.
