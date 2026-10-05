# Experimental editable working context

New runs may opt into bounded reasoning retention for dynamic explorer children:

```powershell
fabric run --autonomous --working-context "Implement a concrete coding task"
fabric usage RUN
```

This requires configuration v2 with a Codex explorer. The fake runtime is allowed
for deterministic integration tests; it does not establish provider behavior.
The flag is off by default. Static graph research/design tasks and other roles
retain their existing context behavior. A run that never uses dynamic explorer
follow-ups has not exercised the treatment.

`--working-context` also enables the independent `--dynamic-explorers` execution
opt-in. The latter permits the same Codex explorer runtime without editable notes.
All A/B/C experiment arms must enable that shared executor; only C enables
working-context retention. Both immutable policies are absent on historical runs.
Autonomous preparation registers the accepted planner result as the read-only
parent for these dynamic children; registration does not invoke another model.

The parent's useful sequence is: spawn a child, wait for its accepted structured
result, then follow up that same child. The first turn has no retained notes.
Its accepted result may initialize the child's projection; a subsequent turn
receives those notes alongside the original controller-derived task input.
Each invocation uses a fresh runtime input, rather than appending the child's
entire prior transcript. Parent and sibling transcripts are not inherited.

## Replacement and authority

The invocation supplies `working_context`, `expected_context_id` and
`expected_context_hash`. The ordinary candidate-bound explorer result includes
`working_context_update` with `expected_id`, `expected_content_hash` and
`content`. This is the narrow read/replace interface; no note filesystem or
additional source-write permission is created.

Content is at most 16 KiB of UTF-8 bytes. Encoding/control characters, exact
preimage and scope are validated by Fabric. Notes are derived from the existing
admitted result event and bound to its journal hash, run, source, candidate,
agent and stable assignment identity. Task completion, permissions, effects,
UNKNOWN, repair budget and verification/review remain controller-owned.
Contradictory notes grant no authority. Skills can advise retention, never alter
the replacement contract or controller policy.

Malformed, oversized or stale updates are rejected separately from an otherwise
valid engineering result. Valid earlier notes remain available. Changed
source/candidate/run bindings do not inherit the old projection. Replay derives
notes from the validated durable journal; no detached summary is trusted.

An already queued or admitted turn retains its exact input. A follow-up queued
before the preceding child result is accepted can therefore receive older notes;
its later replacement may fail the stale compare-and-swap check. Wait for
acceptance before scheduling when current notes are needed. Fabric does not
rewrite that frozen turn or resend an uncertain effect to refresh context.
Missing optional projections can degrade new preparation. Missing preimages
needed to reconstruct an existing frozen invocation deny dispatch.

Supervisor wait results and writer evidence include the bounded structured
engineering handoff and original evidence references, without private scratch
updates. Removing notes never deletes original evidence.

## Measurement and qualification

`usage` reports accepted/rejected rewrites, peak bytes and average bytes across
accepted projection versions. That average is not time-weighted. Provider
input/cached/output/reasoning tokens remain the runtime's separately observed
usage; unavailable values are not zero.

Deterministic integration and independent code review are distinct from live
qualification. Independent review approved the execution path. A Codex
app-server fixture regression verifies exact cached-result recovery with no new
RPC calls; this is deterministic evidence, not live provider qualification.
The matched A/B/C provider experiment remains pending. No token,
latency or success advantage is established, and the treatment remains opt-in.
See the [architectural decision and experiment requirements](../roadmap/adaptive-working-context.md).
