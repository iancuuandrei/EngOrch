# Evidence acquisition High canary 2026-10-06 (frozen, UNKNOWN preserved)

Private frozen canary at exact clean Fabric source
`f03d083867c389bcb9987135026785e5631141d4`, binary SHA
`ed601c88418e53b9ba0b3080cb9683679239239b8eb5e2ab2468050e32737ff2`.
Public task pmezard/go-difflib source
`5d4384ee4fb2527b0a1256a821ebfc92f91efefc`, config SHA
`da3ec5544400e0dcdc9b56fbc7667f9392f772505626ecaf891cf846daff1e4c`,
unchanged oracle SHA
`2567bdfa8af39cc405dd7fc1dc733f9484dc66a0a7fdd911705d0bcb2bb7d938`.

All five configured roles `muse-spark-1.3-contributor` high over the
opencode-http/Responses/Go route with SystemRole `developer`, reasoning
effort high, summary auto, include `reasoning.encrypted_content`. Bounded
dispatch: max-parallel 1, max-repairs 2, max calls per invocation 16,
invocation timeout 600 s, overall frozen wall limit 3600 s.
`GO111MODULE=off` with pinned Go 1.27.1 because the upstream task has no
go.mod. Frozen plan SHA
`cfa71873dfb8213b139c4879dc81af35fbe89f3d5522777fd5e380dc4a802bf2`.

Run `e85e56be265bfa2cebf52c219e1dc74a9b5988d06bd357e9ca586a74883f9183`:
exit 1 after 40.656773 s, state PLANNING, lifecycle ACTIVE, planner dispatch
UNKNOWN with a `.failure.json` diagnostic of `transcript_decode` /
`validation_or_runtime_failure`. The gateway journal holds two call intents
and two accepted call receipts with observed model
`muse-spark-1.3-contributor`, first `tool_calls` then `stop`. Zero accepted
planner result; writer, native, review, held-out, and EVC decision NOT RUN.
Typed gateway usage: input 26390 (cached 12386, uncached 14004), output 3237
(reasoning 2830, ordinary 407); these are subsets and are never
double-counted. Monetary cost UNKNOWN.

The persisted SDK transcript held two reasoning parts in the first
generation and three in the second; within each generation all parts share
one item ID with a single final encrypted carrier, and the earlier parts are
item-only summaries. The source repair now shares one strict
same-generation parser for the ordinary and composite decoders; unpaired,
foreign-generation, invalid, and duplicate-carrier shapes remain strict, and
the original transcript and raw opaque material stay private. No source fix
can settle the old UNKNOWN record, and there is no resend or settlement
authorization.

This record is not task acceptance, not a benefit experiment, not a fixed
cohort, not runtime High qualification, and not v2 completion. Private
snapshots and receipts are retained; no raw prompts, ciphertext, or local
secret paths appear here. Source qualification is reported separately.
