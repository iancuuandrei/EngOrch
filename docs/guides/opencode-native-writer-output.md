# Native OpenCode writer output

The opt-in `opencode.native_writer_output = true` binds the selected writer
proposal schema to OpenCode's native `format: json_schema` request. `utf8-v2`
and `utf8-replace-v3` share the strict UTF-8 changes-array schema;
`utf8-scoped-v4` uses the same shape with an exact candidate-ID enum.
`anchored-edits-v1` uses exact text edits and explicit new-file content, with
the same candidate-ID binding. `changes-json-v1` has a separate wire shape.
Historical live qualification below covers its original frozen contract and
does not qualify the newer v3/v4 or anchored-edit contracts. The mode is
restricted to the bounded writer/fixer path using Responses. Planner, reviewer,
read-only and composite agent turns keep their existing output contracts.

This mode does not claim provider-native strict JSON output. Stock OpenCode
advertises one additional `StructuredOutput` function with the bound schema and
uses required tool choice. Source and candidate reads retain their existing
broker authority. The additional function captures a result; it cannot write a
file or invoke a broker effect.

The controller records the schema with the invocation and dispatch identities.
Readback must show the same format and one final completed capture. Its arguments,
native structured object, call identity and provider receipt must agree. The
structured value is validated against the schema and then passed through the
ordinary writer-proposal and candidate checks. A matching native object alone is
not permission to apply a proposal.

The provider's actual `tool_calls` finish is preserved. A separately bound terminal
capture closes the gateway; it is not relabeled as a provider `stop`. Intermediate
source calls remain legal, but mixed, duplicate, invalid or substituted captures
are rejected. A terminal capture on the last permitted provider call must not be
misclassified as call-budget exhaustion.

`retryCount: 0` is recorded in the native format intent. It does not prove that
stock OpenCode disables its independent session retry policy. EngOrch's gateway
admission and terminal state prevent another upstream dispatch after terminal
success or a failed provider operation; local runtime retry observations must be
reported separately from upstream dispatches.

Provider receipts remain the usage authority. Raw native text and transcript
evidence remain separate from canonical structured result JSON. Reasoning tokens
that are a subset of output are not added again. Malformed output is never repaired
by appending delimiters, stripping fields or replaying the same live invocation.

Qualification is bounded to the exact stock binary and runner/config identities
in the external R26 evidence bundle. A localhost fixture proves native protocol
shapes and dispatch counts; it is not a live model-quality or M2 success claim.
