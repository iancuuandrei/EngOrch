# Go format observation resource measurements

This bounded synthetic measurement exercises `ObserveGoFormat` in process. It
does not measure the Fabric CLI, a real repository, or end-to-end task latency.
All nine corrected samples exited successfully and passed an exact identity
audit: candidate, action, artifact, payload, observation digest, input count,
and payload byte count each had one value across the samples. The common
candidate ID was
`9a5f755cdb467485da7db0c1cfbd6e0b97250cf344c793b5ccc8b0465765759d`; the
fixture had 24 files and a 303,192-byte formatter payload.

## Provenance and method

- Frozen source: commit `1e0084198637fba7f347d00a96378695e4f61a2c`, tree
  `631a618533a6c4f3032877b2d951ecc1e7dff611`. The disposable overlay matched
  all 1,252 tracked files and added only the temporary benchmark test
  (`format_observation_resource_benchmark_test.go`, SHA-256
  `7db464091d01a73b81c7a10c24d7393d97724a3f416ffc1abc6828421c17cd09`).
- Pinned product binary SHA-256:
  `b0d6ca2893c23374332d602596e386401db07e79971f67472a975fc8d0d8d45e`. It is
  provenance only; it did not execute the formatter operation.
- The precompiled Go test executable ran `ObserveGoFormat` and had SHA-256
  `144591efc4e2a085573e347c7c6d91611ece0612be1248bdff8bab52dcf7280f`.
  Compiler: Go `1.27.1` on Windows/amd64.
- Fixture source commit/tree:
  `fac4eb9a6bdf1eb923b7ea2ff2d25146f0e94ca4` /
  `b72a4ebfca96f95a98ec7f7088f682d491a99bc2`; 24 Go files, 268,608 input
  bytes. Per-file manifest SHA-256:
  `2ea718fab34521889c67b0ae6fddb59c32ba50cb6c87a3c46177347828590fc2`.
- Three alternating samples each measured uncached direct formatting,
  cache-cold `ObserveGoFormat`, and cache-warm `ObserveGoFormat`. Each process
  reported one timed operation. The uncached path omits cache lookup and
  publication; cold/warm include those cache operations.

## Per-sample results

`ns/op`, allocated bytes, and allocations are Go benchmark metrics. Wrapper
elapsed includes process startup and setup; for warm samples it also includes
the cache-prime operation. RSS and CPU are process-tree samples taken every
100 ms; CPU is the maximum sampled cumulative process-tree CPU, not a peak-RAM
or instantaneous CPU measure. Cache bytes are the total files in the measured
leaf.

| Mode | Sample | ns/op | Allocated bytes/op | Allocs/op | Wrapper elapsed ms | Sampled peak RSS bytes | Sampled cumulative CPU ms | Cache bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| uncached | 1 | 625,693,400 | 18,735,024 | 237,248 | 1,846.5 | 20,340,736 | 359.4 | — |
| uncached | 2 | 621,708,900 | 18,758,264 | 237,295 | 1,929.0 | 20,250,624 | 328.1 | — |
| uncached | 3 | 627,861,800 | 18,825,128 | 237,284 | 1,925.4 | 20,340,736 | 453.1 | — |
| cache-cold | 1 | 649,684,400 | 19,041,752 | 239,048 | 1,905.8 | 20,168,704 | 234.4 | 314,424 |
| cache-cold | 2 | 615,463,400 | 18,950,392 | 239,036 | 1,901.3 | 20,606,976 | 578.1 | 314,424 |
| cache-cold | 3 | 625,177,200 | 19,000,288 | 239,060 | 1,932.5 | 20,336,640 | 343.8 | 314,424 |
| cache-warm | 1 | 580,338,900 | 8,672,808 | 29,383 | 2,598.3 | 20,946,944 | 500.0 | 314,424 |
| cache-warm | 2 | 610,759,800 | 8,657,368 | 29,365 | 2,549.0 | 19,951,616 | 421.9 | 314,424 |
| cache-warm | 3 | 618,382,100 | 8,653,688 | 29,349 | 2,569.7 | 19,673,088 | 468.8 | 314,424 |

| Mode | Median ns/op (range) | Median allocated bytes/op (range) | Median allocs/op (range) | Median wrapper ms (range) |
| --- | ---: | ---: | ---: | ---: |
| uncached | 625,693,400 (621,708,900–627,861,800) | 18,758,264 (18,735,024–18,825,128) | 237,284 (237,248–237,295) | 1,925.4 (1,846.5–1,929.0) |
| cache-cold | 625,177,200 (615,463,400–649,684,400) | 19,000,288 (18,950,392–19,041,752) | 239,048 (239,036–239,060) | 1,905.8 (1,901.3–1,932.5) |
| cache-warm | 610,759,800 (580,338,900–618,382,100) | 8,657,368 (8,653,688–8,672,808) | 29,365 (29,349–29,383) | 2,569.7 (2,549.0–2,598.3) |

The measured cache leaf occupied 314,424 bytes in each cold and warm sample.
Median sampled peak RSS was 20,340,736 bytes uncached, 20,336,640 cold, and
19,951,616 warm; ranges were 20,250,624–20,340,736, 20,168,704–20,606,976,
and 19,673,088–20,946,944 bytes, respectively. Median sampled cumulative CPU
was 359.4 ms uncached, 343.8 ms cold, and 468.8 ms warm; ranges were
328.1–453.1, 234.4–578.1, and 421.9–500.0 ms. These short samples and 100 ms
polling can miss peaks.

## Limits and retained setup failure

The warm median operation time is near the uncached and cold measurements,
while its wrapper time is higher because wrapper timing includes cache priming.
These differences do not establish a speedup. Allocation bytes are not peak
memory. This run did not measure model tokens, provider cost, or user-visible
task latency. Separate cache safety tests cover the configured 64-entry and
48 MiB bounds; this fixture's 314,424-byte leaf does not measure eviction or
capacity behavior.

An earlier harness attempt stopped at sample 2 because each test process
reused the same warm-cache leaf and the next process found a hit instead of a
fresh prime miss. That setup failure is preserved separately and is not part
of these nine results. The corrected temporary benchmark creates one fresh
warm leaf per process. The authoritative measurement receipt SHA-256 is
`4b295a68a30a0d166253e98c386fbaf8fa562208e98b3cd3fd28a00114b49bf1`.
