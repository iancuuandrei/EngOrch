# Muse Spark Contributor through OpenCode Go

The route uses `opencode-http`, model `muse-spark-1.3-contributor`, the
`openai-responses-sse-v1` adapter, and
`https://opencode.ai/zen/go/v1/responses`. Bind the provider credential to an
environment variable and require the `x-opencode-session` header. Subscription
cost remains unknown; do not declare zero API pricing.

The completed response observed on 2026-10-04 did not contain a trailing cost
ping. For this route, set `trailing_cost_ping_v1` to `false` in
`adapter_capabilities_json`. Keep `content_part_completes_text` and
`function_call_done_name_v1` enabled for the observed response dialect. This
changes the selected route declaration, not the decoder's global defaults.

## Pinned OpenCode Responses SDK High recipe (not yet live-qualified)

For the pinned High recipe, declare the OpenCode request controls
explicitly: `reasoning_effort` high, `reasoning_summary` auto, `include`
`[reasoning.encrypted_content]`, and SystemRole `developer`. The matching
provider declaration uses variant/profile effort `high` with variant
`system_role` `developer` and `reasoning_summary` `auto`. The 2026-10-04
record below used the default-effort recipe; the corrected High recipe has
not yet completed a live run. A frozen 2026-10-06 canary was rejected in
local preflight on the system/developer wire mismatch: no admitted
upstream call or accepted receipt in the gateway journal; runtime
ownership remains UNKNOWN. See the
[canary record](../evaluation/evidence-acquisition-canary-20261006.md).
Run `fabric doctor` first: it checks static configuration, not actual wire
serialization. Do not loosen the global decoder or acceptance gates to
tolerate the mismatch. The corrected recipe applies only to newly admitted
distinct runs and grants no resend or replacement authorization for an
existing UNKNOWN attempt; the local diagnostic settles nothing.

The ordinary and composite tool-turn decoders pair same-generation item-only
reasoning summary parts
with their fully validated encrypted carrier instead of rejecting the
transcript at decode. This source repair addresses the demonstrated SDK
summary shape (several summary paragraphs, encrypted content attached to the
final one) at the readback layer only. It does not claim task or EVC
acceptance, does not rewrite the frozen canary's UNKNOWN record, and the
corrected High recipe remains not yet live-accepted. See the
[High canary record](../evaluation/evidence-acquisition-high-canary-20261006.md).

With `native_writer_output = true`, writer and fixer require tools and
`structured_output = "UNSUPPORTED"` at the provider layer, with
`required_capabilities.reasoning = false`. The model may still advertise
reasoning capability and report reasoning tokens. The native StructuredOutput
tool captures the proposal; it does not request provider-native JSON output.

Advisory text may accompany that terminal tool. Fabric compares its digest to
the corresponding provider generation and separately binds the captured tool
ID, arguments and schema. Advisory text is not the proposal result.

All engineering roles use the existing bounded FIFO queue for parallel context
tool requests. Repeated OpenCode role invocations receive separate runtime journals;
reconciliation of the first invocation retains its historical journal path.

See [capability fallback and acceptance](graceful-capabilities.md) for optional
capability degradation and the distinction between attention and unsafe effects.

Create the selected private OpenCode state directory before dispatch. Run
`fabric doctor` first. A corrected declaration applies to a new run; do not edit
the bound configuration or resend a rejected response in an existing run.
One local run on 2026-10-04 completed the pinned go-humanize underscore task:
READY, native PASS, Muse review APPROVE with zero findings, and candidate-bound
copy plus unchanged held-out PASS. All eight provider calls had exact receipts
and observed model `muse-spark-1.3-contributor`. Planner, writer and reviewer were
executed; explorer and fixer were configured to Muse but not needed in that run.
See the [acceptance receipt](../../evals/v1/results/muse-go-humanize-20261004.json).
The custom binary was built from the v1.0.28 source plus local fixes; this is one
task's evidence, not general model, six-task suite or release qualification.
