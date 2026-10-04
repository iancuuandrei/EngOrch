# Fabric v1 performance and resource audit

Baseline: `de61e6fc5c372b8b4161995a053497d888bc658b`, Windows amd64, Go 1.27.1.

The subsequent author-only history rewrite maps this baseline to
`637e64665426e49824c8aebf9e14ef9c1f547eef` with the identical source tree.
Frozen build/evaluation receipts retain their original commit IDs. The
published `v1.0.0` tag is unchanged.

This audit covers the current product implementation and the planner-context
increment. Historical G0 artifacts are retained evidence, not active product
work. Source-level estimates below are not measured speedups or memory usage.

## Priorities

### Whole-workflow efficiency qualification for v1.1

RAM is one dimension of qualification. Compare the same pinned task, model,
candidate, acceptance checks and concurrency settings in fresh run roots;
record cold and warm cache conditions separately. An improvement must retain
the verified result and external-effect guarantees. A faster failed run is
not an improvement in productive throughput.

| Dimension | Record for each real workflow |
| --- | --- |
| Latency | End-to-end wall time, time to first useful result, planning/execution/review/verification durations, queue and provider wait time. |
| CPU | CPU time and profiles for Fabric; separately observe child processes. Distinguish computation from waiting. |
| Memory | Allocated bytes, allocations, GC count/pause, peak resident/private memory and child-process peaks. End heap is not peak RAM. |
| I/O | File bytes/operations, candidate scans, journal reads/appends, process starts and transport bytes where measured. Unobserved counters remain UNKNOWN. |
| Scheduling | Useful concurrency, queue delay, idle capacity, contention and completed verified tasks per unit time. |
| Cache/context | Cold/warm behavior, observed cache hits/misses, invalidation correctness, retained context size and repeated source reads. Provider caching is reported separately from local caches. |
| Model usage | Unique invocations, corrections/retries, input, cached input, uncached input, output and reasoning output. Preserve provider-specific counter semantics and missing usage. |
| Cost | Actual supported provider prices or billing receipts and currency; UNKNOWN when unavailable. Cached input and reasoning subsets are not counted twice. |
| Reliability | Acceptance outcome, useful work retained after failure, recovery latency and additional model calls needed to obtain the verified result. |

The optional Fabric process profile does not measure child CPU/RSS, provider
latency, disk traffic or billing. Combine it with existing runtime receipts,
journal phase evidence and host observations. Profile successful, bounded
repair and isolated parallel workflows before prioritizing further changes.
Use repeated comparable samples when claiming gains; report measurement
overhead and host contention. These workflow measurements remain NOT RUN
until fresh real-task evidence is recorded.

| Area | Finding | Action | Evidence boundary |
| --- | --- | --- | --- |
| Candidate context | Each selected-file `ReadSource` captures the whole candidate before and after the read; up to 24 files multiply whole-tree hashing. It also base64-encodes bytes immediately decoded by the caller. | Batch bounded reads under the existing lease, validate every selected file against its manifest, verify the complete candidate before and after. Benchmark and test drift rejection. | Retain the public single-file read API and candidate identity checks. |
| Explicit `inspect RUN` | The CLI validates the same journal twice without an intervening operation. | Reuse the first validated snapshot for that read-only command. | Resume, reconciliation, and external-effect paths keep their observations. |
| Journal append | Every append reads and checks the history, clones events, and replays the controller. Per-append work grows with history; a complete growing run can accumulate quadratic work. | Measure production-shaped append/inspect cases before considering an authority or storage change. | Integrity and effect ownership remain required. |
| Candidate verification and acceptance copy | Full-tree observations and copy verification incur repeated byte reads. Capture streams file bytes and retains a bounded manifest. | Measure bytes and phase timings before changing these guards. | An accepted copy must still bind to the reviewed candidate. |
| RI lexical calls | Two context queries can launch two Rust processes and rehash the executable. Index manifests/shards are verified per search. | Consider admission-scoped streaming only after a representative measurement. | No cross-run cache or skipped integrity verification is justified by this audit. |
| Full journal export | Validated history is retained alongside a complete output buffer. | Measure peak allocation; streaming requires an explicit partial-output contract and regression coverage. | A speculative memory improvement is not an implemented capability. |
| Scheduling and processes | Reviewed Codex hosts are closed/reaped; explorer and writer concurrency have explicit bounds. | Measure startup and peak resident memory with those bounds before proposing reuse. | No process-leak conclusion can be inferred from source inspection alone. |
| Planner evidence | Planner previously received no persisted bounded source selection. Current run RI bindings require later approved/workspace states. | Add an independent opt-in source pack and evaluate fixed tasks A/B. | Partial committed-source evidence is not semantic coverage; empty results prove no absence. |

## Fixed-six model evaluation, 2026-10-03

The completed feature-only A/B returned **5/6 PASS** for the unchanged
baseline and **4/6 PASS** for `source-bounded-v1`. This does not qualify the
experimental planner context for default promotion. Both arms used GPT 6 Luna
High, writer edit validation, the same six pinned repositories, native checks,
and unchanged held-out acceptance. One sample per task cannot establish a
causal quality or token improvement.

Baseline go-difflib exhausted its two repair attempts. Treatment go-difflib
failed held-out acceptance; treatment go-humanize was blocked before applying
files because the final proposal changed one nibble of its exact preimage hash
after a successful validation. The preimage guard correctly rejected it.
These completed runs are preserved; new evaluation must use fresh run roots.
Separate godotenv/numeric-atomic supplements returned 1/2 PASS in each arm
and are not counted in the fixed six.

Frozen runtime-result counters sum to 3,580,465 input / 32,209 output tokens
for baseline and 3,973,881 input / 32,174 output for treatment. These include
known-partial counters from blocked runs; provider request counts remain
unknown. Complete-counter rows alone sum to 2,892,496 / 22,970 and
3,578,872 / 25,926 respectively, over different successful task sets. Input
counters can include repeated context/prefill and cached tokens, so neither
sum measures unique source text. The experiment has not demonstrated a token
saving.

| Arm, all six including known-partial counters | Input total | Cached input | Uncached input | Output | Reasoning output |
| --- | ---: | ---: | ---: | ---: | ---: |
| Baseline | 3,580,465 | 2,673,664 | 906,801 | 32,209 | 15,091 |
| Planner treatment | 3,973,881 | 3,041,024 | 932,857 | 32,174 | 13,435 |

Cached input is a subset of total input; reasoning output is a subset of
output. Neither is added again. Uncached input is total minus cached, using
the locally verified Codex counter semantics. Exactly one final usage delta
per unique runtime journal is summed, rather than cumulative notifications.
The baseline and treatment therefore already received substantial automatic
prompt caching. Cache-write counters and provider request counts remain
unknown; no pricing or billing conclusion is inferred from these counters.

The next bounded product iteration targets duplicate model-visible source
representations and metadata copying. A 32,768-byte ASCII read currently sends
both UTF-8 and Base64. A presentation-only UTF-8 projection would remove
43,712 serialized bytes from a representative response (about 57%). This is
a payload-byte measurement, not an observed token saving. Binary fallback,
source/candidate identity, offsets, hashes, and legacy durable evidence must
remain intact.

### Hardening iteration

UTF-8-first projection is confined to model-tool adapters; repository/worktree
observations retain their original shapes and binary pages retain Base64.
Projection rejects malformed/noncanonical Base64 and differing UTF-8 bytes.
Regression cases cover split multibyte pages, empty EOF, and preserved identity,
size, hash, offset, and continuation metadata. A maximum ASCII page with a
non-null next offset measures 76,836 to 33,124 bytes for source and 76,722 to
33,010 for candidate: 43,712 bytes removed in each case.

The opt-in `cache-prefix-v1` prompt recipe places stable role instructions
ahead of variable run/candidate data. Exact output schemas remain bound and
are placed last. Empty recipe preserves historical prompt identities. This
improves reusable prefix structure; it does not guarantee a provider cache
hit. New-recipe guidance requires exact validated writer metadata and
justification for changing existing test expectations.

Difflib diagnosis confirms that the original newline case was visible to the
writer, but the candidate changed its expected result. Native tests passed
after that change, while unchanged held-out acceptance rejected it. No excerpt
budget increase or documentation exclusion is justified by this evidence.

## Deferred process-profile prototype

The opt-in process profiler was removed from the delivered v1.1 code after
profiling was deferred and Sonar flagged its debug instrumentation. Prototype
source and prior measurements remain local development evidence. Production
Fabric does not enable profiling through `FABRIC_PROFILE_DIR`. The following
notes describe the retained prototype, not the published executable.

Analyze the sampled stacks with `go tool pprof -top cpu.pprof` and
`go tool pprof -top -alloc_space allocs.pprof`. The CPU profile covers the opted-in
execution. The allocation profile is Go's cumulative process sample and may lag
recent allocations; `runtime.json` records execution-interval allocation/count,
GC count/pause and end-of-interval heap counters. The default runtime allocation
sampling rate is retained. End heap is neither peak memory nor RSS. Every metric
excludes child processes and external providers. CPU profiling and profile
serialization add overhead; compare runs under the same profiling conditions.

Bind profiles to the exact executable/source and workload receipt. Combine
stack attribution with journal stage times, external process resource counters,
verification/subprocess measurements and separately observed model token types.
Do not add overlapping nested stages or cumulative usage notifications twice.
Successful, repaired and isolated real workloads remain required before
claiming a hotspot or an optimization gain; profiling fixtures alone prove none.

## Measurement policy

Record executable/source identities, toolchain, fixture size, selected file count,
elapsed time, allocations, and bytes allocated. Report individual benchmark
results rather than extrapolating a microbenchmark to overall agent-task latency.
Model-task outcomes require the exact candidate, native checks, independent
review, and held-out acceptance; an optimization benchmark proves none of these.

Planner-context A/B uses a feature-only treatment build so the separately
benchmarked candidate-read and inspect optimizations do not confound its timing.
The unchanged baseline preparation receipt retains its original runner hash;
both model-time arms use the same finalized evaluator. The prepare-only change
adds treatment metadata and does not alter task pins or check recipes.

## Initial measurements

These are single-iteration measurements on an AMD Ryzen 9 9955HX Windows host,
with concurrent development/evaluation activity. Timings are preliminary; B/op
is total allocated bytes per operation, not peak resident memory.

Candidate fixture: 256 regular files of 4,096 bytes each. Benchmark
`BenchmarkReadSourcesAgainstIndividualReads`, Go 1.27.1, `-benchtime=1x -benchmem`:

| Selected files | Individual reads | Batch reads | Individual B/op | Batch B/op |
| --- | ---: | ---: | ---: | ---: |
| 1 | 0.592 s | 0.967 s | 21,282,240 | 21,287,480 |
| 8 | 6.695 s | 0.671 s | 170,156,528 | 21,567,824 |
| 24 | 27.240 s | 1.570 s | 510,369,968 | 22,195,088 |

The helper performs two whole-candidate observations for a batch rather than
two per selected file. Admission retains its own initial and final observations:
approximately four whole-tree observations rather than `2N+2`. A single-file
case has no expected scan reduction and was slower in this sample.

Valid lifecycle histories, `BenchmarkControllerHistory`, `-benchtime=1x`:

| Fixture | One inspect | Two inspects | One B/op | Two B/op |
| --- | ---: | ---: | ---: | ---: |
| 16 events | 5.86 ms | 23.07 ms | 371,232 | 724,736 |
| 64 events | 7.53 ms | 50.34 ms | 1,311,936 | 2,666,432 |
| 64 events, 16 KiB objective | 19.37 ms | 51.29 ms | 2,442,360 | 4,935,976 |

These fixtures exercise real SQLite validation and controller replay but contain
lifecycle events, not full provider transcripts or task-context histories.
Append measured 8.87/22.17/18.79 ms respectively; this is insufficient to infer
a latency curve or justify controller caching. Near-doubled inspect allocations
support removing the repeated read; timing ratios require repeated samples.

## Further source findings

- The anchored-edit validator pages up to 256 KiB through the single-file
  candidate read API. Each 32 KiB page repeats candidate hashing and a full
  target-file hash. A future bounded full-source helper could reduce this;
  the current batch optimization does not change that API.
- Git object storage/readback launches processes per blob/tree. The admitted
  candidate limits can exceed what fits a two-minute overall deadline. Batch
  readback is a candidate for measurement; partial writes must retain their
  existing effect semantics.
- Provider transport freezes/copies a request body before adapter validation.
  An early existing-contract size guard avoids allocating an oversized copy.
  This is a defensive API allocation bound, not an observed normal-run failure.
- Reviewed provider SSE, Codex RPC, OpenCode readback, MCP exchange and config
  paths have explicit byte/event/cardinality bounds. No uncapped buffer or
  cancellation leak was confirmed in those inspected paths. This is a source
  inspection result, not a process-lifetime stress-test claim.

## Repeated samples

Planner path ranking now tokenizes the objective once rather than per visited
path and retains only the best 24 paths during traversal. On a 4,096-path pure
ranking fixture, re-tokenizing measured 6,757,227 ns/op, 852,085 B/op and 8,192
allocations; pretokenized scoring measured 2,548,993 ns/op with zero measured
allocations. This measures the ranking loop only, not Git reads or planning.
The traversal still inventories eligible paths, with bounded retained selection.

Three further single-iteration samples of the 24-file case measured individual
reads at 29.531/30.317/31.264 s and batch reads at 1.687/1.613/1.574 s.
Allocated bytes remained approximately 510.4 MB versus 22.19 MB per operation,
with approximately 1.15 million versus 49 thousand allocations. These samples
support the scan-amplification finding; they remain a local microbenchmark and
do not establish an overall autonomous-task speedup or peak-RAM reduction.

Three inspect repetitions measured 16-event histories at 362–364 KB allocated
for one inspection versus 745–767 KB for two, and 64-event histories at
1.328–1.340 MB versus 2.646–2.659 MB. Corresponding timing ranges were
2.86–3.81/12.77–14.67 ms and 6.45–8.81/19.68–22.43 ms respectively.
Concurrent host load and one-iteration samples limit timing precision.

Private local raw receipts:
`D:\fabric-ci2-tools\context-batch-read-bench-20261003.txt` and
`D:\fabric-ci2-tools\controller-history-bench-20261003.txt`.
Their sanitized fixture descriptions and measurements are recorded here;
provider transcripts are not publication artifacts.

Model-task A/B acceptance remains pending.
