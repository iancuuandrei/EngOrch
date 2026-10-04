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
