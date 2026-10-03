# Native Codex thread compaction: protocol evidence

The Codex app-server protocol has a typed `thread/compact/start` operation.
This is a source-level capability, not a claim that Fabric currently invokes
it. The source pin below is the exact upstream commit inspected for this note.

- Repository: [`openai/codex`](https://github.com/openai/codex)
- Commit: [`b172810921f89847cd310ecc496f9c901760e933`](https://github.com/openai/codex/commit/b172810921f89847cd310ecc496f9c901760e933)
- License: Apache-2.0, as declared by the repository's [`LICENSE`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/LICENSE)
- Protocol declaration: [`codex-rs/app-server-protocol/src/protocol/common.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server-protocol/src/protocol/common.rs)
- Handler: [`codex-rs/app-server/src/thread_processor.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server/src/thread_processor.rs)
- E2E behavior: [`codex-rs/app-server/tests/suite/compaction.rs`](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server/tests/suite/compaction.rs)

The v2 request names one exact `threadId`. Its empty-object response is an
acknowledgment that the handler submitted compaction work; it does not prove
that compaction finished. The upstream integration test waits for a matching
`ContextCompaction` item start and completion, then for `turn/completed` before
continuing. Therefore a future Fabric integration must bind the operation to
the persisted invocation/thread, wait for those completion facts, and preserve
the accepted checkpoint and runtime roots across continuation. An empty
response alone cannot advance the controller or resolve an UNKNOWN invocation.

Fabric does not currently expose a typed compaction call. Its Codex RPC
confinement rejects raw `thread/compact/start` through the generic request
surface. This research note does not change that policy or qualify sustained
long-horizon behavior.
