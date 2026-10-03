# Native Codex automatic-compaction observation

Fabric records automatic Codex context-compaction lifecycle metadata only when
the existing turn stream delivers matching `item/started` and
`item/completed` notifications for a `contextCompaction` item. The receipt is
bound to the already-recorded invocation, thread, turn, and item ID. It contains
no provider item payload or transcript.

The typed runtime observation reports a positive count only when every recorded
item has one start and one completion, the turn completed, and the stream reached
its terminal boundary. Missing phases, duplicate phases, identity mismatches,
the 64-item cap, interrupted turns, incomplete streams, and readback-only items
remain `UNKNOWN`; a `thread/read` item by itself is not proof of both lifecycle
phases. An append/read failure cannot produce a count. Existing invocation and
controller intents are unchanged. Fabric does not issue `thread/compact/start`.

## Protocol evidence

The inspected upstream source is Codex commit
[`b172810921f89847cd310ecc496f9c901760e933`](https://github.com/openai/codex/commit/b172810921f89847cd310ecc496f9c901760e933),
licensed Apache-2.0. Its
[automatic-compaction app-server test](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server/tests/suite/v2/compaction.rs)
observes `item/started` and `item/completed` for a matching context-compaction
item without calling the manual `thread/compact/start` method. The
[protocol item type](https://github.com/openai/codex/blob/b172810921f89847cd310ecc496f9c901760e933/codex-rs/app-server-protocol/src/protocol/v2/item.rs)
exposes the item ID. Upstream’s test is mock-backed; it demonstrates protocol
behavior, not sustained real-provider qualification.

## Qualification boundary

The local runtime protocol fixture verifies metadata capture, exact thread/turn/
item binding, unmatched and duplicate handling, stream interruption, bounded
overflow, readback-only ambiguity, and completed-invocation resume without a
new request. It does not prove that an organic long coding turn compacts under
the default native threshold. A graph made of fresh role threads is not such a
turn. Real qualification must retain the run/binary/model bindings, demonstrate
the matched notification pair on one actual long-lived thread, then pass the
existing candidate, native verification, review, and held-out gates. No lower
automatic-compaction threshold is enabled by this change.
