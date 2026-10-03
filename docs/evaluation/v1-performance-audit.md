# Fabric v1 performance and resource audit

Baseline: `de61e6fc5c372b8b4161995a053497d888bc658b`, Windows amd64, Go 1.27.1.

This audit covers the current product implementation and the planner-context
increment. Historical G0 artifacts are retained evidence, not active product
work. Source-level estimates below are not measured speedups or memory usage.

## Priorities

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
