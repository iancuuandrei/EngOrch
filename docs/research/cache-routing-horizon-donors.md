# Cache, context, routing, and long-horizon donors

Research snapshot: 2026-10-03. Repository refs below are pinned to the default-branch commit SHA observed for each public repository on this date; license IDs come from GitHub repository license metadata and are cross-linked to the license file at that pinned ref. This is a mechanism map, not a code-adoption decision. Donor code was inspected read-only and not executed.

## Findings

### Build/action caching is about complete deterministic input closure

Bazel's remote cache separates an action cache (action digest → result metadata) from a content-addressable store (output blobs). Its action identity includes declared inputs, command, environment, and expected outputs. Bazel explicitly cautions that inputs changed during an action can produce invalid cache entries. This is a useful pattern for **pure build/check artifacts**: cache only when the action's full source/tool/command/environment closure is represented and the result is replay-safe. See [Bazel remote caching](https://bazel.build/versions/7.1.0/remote/caching) and the [Remote Execution Action definition](https://github.com/bazelbuild/remote-apis/blob/main/build/bazel/remote/execution/v2/remote_execution.proto).

Nix models a derivation as a command over declared inputs in a controlled build environment. Its store is content-addressed, while derivation outputs can use different addressing modes. The manual stresses that sandboxing and input closure are necessary but not alone sufficient: builds can still leak nondeterminism. Nix is good prior art for deterministic dependency closure, immutable materialized outputs, and comparing independent builds; it is not a model-response cache. See the [Nix build manual](https://github.com/NixOS/nix/blob/2ab29c63d3273d4e2b4d1346b1aefede9b5af350/doc/manual/source/store/building.md) and [Nix reproducibility guidance](https://reproducible.nixos.org/).

**EngOrch fit:** keep any future local cache limited to deterministic projections/checks whose key includes the exact candidate/source manifest, command, environment, toolchain/runtime identity, and relevant parameters. A cache hit can supply an artifact or prior observation only after its full key and content digest are verified against the current candidate. It must not create a new verification pass, reviewer decision, permission, runtime receipt, or settled effect. Never cache an uncertain external action or a model answer as if it were fresh evidence.

### Context condensers are derived views, not verified task state

OpenHands' current `OpenHands/software-agent-sdk` defines condensers over a conversation `View`; a condenser returns a smaller view or a `Condensation` event that is applied to later views. The base contract says the view is read-only because it may be a cached projection. This is a useful interface for reducing **in-session model context** while keeping the source event history; it does not turn summaries into independently verified repository facts or a cross-run memory authority. See [`CondenserBase`](https://github.com/OpenHands/software-agent-sdk/blob/08af1d5aa3c7ec75813b01de94b26d8af9996752/openhands-sdk/openhands/sdk/context/condenser/base.py), [conversation state/view](https://github.com/OpenHands/software-agent-sdk/blob/08af1d5aa3c7ec75813b01de94b26d8af9996752/openhands-sdk/openhands/sdk/conversation/state.py), and [MIT license](https://github.com/OpenHands/software-agent-sdk/blob/08af1d5aa3c7ec75813b01de94b26d8af9996752/LICENSE).

**EngOrch fit:** if adding context compaction, bind the derived view to an exact journal head and keep the original events available. On resume, rebuild/validate from the durable source rather than trusting a detached summary. Keep verified candidate identity, current task graph, active effect intents, and unknown outcomes as structured non-condensable state. The context layer may summarize explanation/history, not rewrite controller authority.

### Long-horizon loops separate execution context from verified progress

`AMAP-ML/LongHorizon-Harness` describes a loop that rebuilds each bounded step from the original goal and verified progress, executes with fresh context, checks the real environment, then checkpoints accepted progress or carries failure evidence forward. It is the closest donor for context budget and task-state ergonomics. Treat its public claims as project documentation, not as qualification evidence for EngOrch. See [README at pinned ref](https://github.com/AMAP-ML/LongHorizon-Harness/blob/a1dd930614972b92361c1b9cd6aac441a6db5a65/README.md) and [MIT license](https://github.com/AMAP-ML/LongHorizon-Harness/blob/a1dd930614972b92361c1b9cd6aac441a6db5a65/LICENSE).

OpenAI's `openai/symphony` is a language-neutral tracker/workspace/agent runner specification plus an experimental Elixir implementation. Its spec puts workflow policy in-repository and calls for isolated workspaces, bounded concurrency and observability. It intentionally leaves implementations' trust, approval, and sandbox posture open; the current Elixir README also labels the implementation a prototype and documents blocked state as in-memory across restart. Borrow repository-owned workflow configuration and scheduler observability only; do not borrow its restart assumptions as durable effect settlement. See [Symphony SPEC](https://github.com/openai/symphony/blob/be10a1b79df723d6d7612b5651c8522704dafb2e/SPEC.md), [Elixir implementation README](https://github.com/openai/symphony/blob/be10a1b79df723d6d7612b5651c8522704dafb2e/elixir/README.md), and [Apache-2.0 license](https://github.com/openai/symphony/blob/be10a1b79df723d6d7612b5651c8522704dafb2e/LICENSE).

### Sandboxed execution and model routing are separate concerns

`SWE-agent/SWE-ReX` provides a runtime interface for sandboxed shell environments locally or in cloud backends, with parallel sessions. It is useful prior art for execution-boundary adapters and resource caps, not for treating a transcript as verification. Any EngOrch check result still needs the exact candidate and test invocation binding. See [SWE-ReX README](https://github.com/SWE-agent/SWE-ReX/blob/5c995c365dfb1fd5bc56fda688be5d8538f9931f/README.md) and [MIT license](https://github.com/SWE-agent/SWE-ReX/blob/5c995c365dfb1fd5bc56fda688be5d8538f9931f/LICENSE.txt).

The names in the routing prompt resolve to two concrete, distinct projects:

- [`fubak/ultraswarm`](https://github.com/fubak/ultraswarm/tree/fe113098eeafce8b800b94efe64e8bff97a52f5f) is a Node-based multi-CLI orchestrator. Its README describes repository/capability-aware worker routing, SQLite calibration, structured-usage-only reporting, explicit budgets, worktrees and approvals. Its route estimates and spend parsing are operational estimates, not provider billing truth or correctness evidence. See its [MIT license](https://github.com/fubak/ultraswarm/blob/fe113098eeafce8b800b94efe64e8bff97a52f5f/LICENSE).
- [`LanceZPF/agent-as-a-router`](https://github.com/LanceZPF/agent-as-a-router/tree/e43839edb0d5d0a9feec2f7078019406ab4d64bd) is the official ACRouter implementation and CodeRouterBench reproduction for cost/performance model selection on coding tasks. Its benchmark matrix includes measured model results and pricing, and its inference interface selects a model then optionally verifies the response. Benchmark routing is a hypothesis to test on held-out EngOrch tasks; it does not establish this product's model quality, current prices, or acceptable verifier behavior. See the [paper](https://arxiv.org/abs/2606.22902) and [MIT license](https://github.com/LanceZPF/agent-as-a-router/blob/e43839edb0d5d0a9feec2f7078019406ab4d64bd/LICENSE).

**EngOrch fit:** calibration can inform an explicit route choice, but the selected model/effort must be persisted in the invocation and charged against the same provider admission/reservation as any other call. Track observed latency and provider-reported token/cache usage by exact provider, model, effort, role, and request recipe; retain `unknown` when usage is absent. Route quality must be measured against candidate-bound native checks and review on held-out tasks, with failure and escalation costs included. Do not route away from required review/verification or alter effect authority according to a cost estimate.

## Safe adoption and evaluation

1. **Prefix caching:** measure common-prefix bytes locally first. Then run a matched provider experiment on one fixed prompt recipe/model/provider, keeping source, task context, schema, tool catalog, effort, and concurrency pinned. Record provider-reported cache read/write usage where supplied; report missing metrics as unknown. Compare quality, latency, and cost independently. Never claim a cache hit from a cache key or prefix alone.
2. **Derived context:** compare full history vs. condensed view on the same pinned tasks. Require exact journal-head binding and reconstructability; measure input tokens/latency and task/native-gate outcomes. Keep the original event log authoritative and leave identity/effect state uncompressed.
3. **Action cache:** cache only deterministic read-only build/check actions with full content/command/environment/toolchain closure. Include candidate ID/manifest, cwd/path remaps, runtime version, timeout and declared inputs in the key. Verify returned artifact hashes and candidate binding before reporting a hit. Cold-cache and cache-hit runs must both pass the same independent native acceptance gates.
4. **Model routing:** compare fixed high-quality baseline to a frozen deterministic policy on a stratified held-out suite. Report completion/native-gate pass, intervention, abstention, latency, input/output/cache usage and actual provider cost separately. A routing policy can only select among already-admitted provider/model/effort combinations; unavailable usage or price remains unknown, not zero.

Do not add a cross-run cache of model outputs, summaries, approvals, or runtime outcomes in this increment. Cache the provider's reusable prompt prefix where the provider exposes it; keep controller state, durable journals, and permission decisions outside that cache.

## Pinned donor revisions and licenses

| Donor | Ref observed 2026-10-03 | License metadata |
|---|---|---|
| Bazel | [`bazelbuild/bazel@9e19427`](https://github.com/bazelbuild/bazel/tree/9e19427a2a5e782d4a106f9b1a4b7b92651b0e18) | Apache-2.0 ([LICENSE](https://github.com/bazelbuild/bazel/blob/9e19427a2a5e782d4a106f9b1a4b7b92651b0e18/LICENSE)) |
| Nix | [`NixOS/nix@2ab29c6`](https://github.com/NixOS/nix/tree/2ab29c63d3273d4e2b4d1346b1aefede9b5af350) | LGPL-2.1 ([COPYING](https://github.com/NixOS/nix/blob/2ab29c63d3273d4e2b4d1346b1aefede9b5af350/COPYING)) |
| LongHorizon-Harness | [`AMAP-ML/LongHorizon-Harness@a1dd930`](https://github.com/AMAP-ML/LongHorizon-Harness/tree/a1dd930614972b92361c1b9cd6aac441a6db5a65) | MIT |
| OpenHands SDK | [`OpenHands/software-agent-sdk@08af1d5`](https://github.com/OpenHands/software-agent-sdk/tree/08af1d5aa3c7ec75813b01de94b26d8af9996752) | MIT |
| SWE-ReX | [`SWE-agent/SWE-ReX@5c995c3`](https://github.com/SWE-agent/SWE-ReX/tree/5c995c365dfb1fd5bc56fda688be5d8538f9931f) | MIT |
| Symphony | [`openai/symphony@be10a1b`](https://github.com/openai/symphony/tree/be10a1b79df723d6d7612b5651c8522704dafb2e) | Apache-2.0 |
| Ultraswarm | [`fubak/ultraswarm@fe11309`](https://github.com/fubak/ultraswarm/tree/fe113098eeafce8b800b94efe64e8bff97a52f5f) | MIT |
| Agent-as-a-Router | [`LanceZPF/agent-as-a-router@e43839e`](https://github.com/LanceZPF/agent-as-a-router/tree/e43839edb0d5d0a9feec2f7078019406ab4d64bd) | MIT |

License metadata is a source-screening aid, not a legal conclusion. Recheck licenses and transitive dependencies at the exact ref before copying code; prefer adapting mechanisms and citing source over vendoring donor implementations.
