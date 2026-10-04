# Go corpus graph-size check measurement

The committed Go corpus admits files in candidate order and checks the exact graph after each file. `BuildGoEngineeringGraph` already validates and canonical-hashes that graph. The admission check used to marshal the resulting typed graph, parse it through `canonical.Normalize`, and re-encode it only to obtain the byte length. It now counts the canonical byte length directly from deterministic JSON encoding, accounting for the HTML and U+2028/U+2029 escape differences. Graph construction, canonical validation, digest checks, the graph byte cap, and greedy admission order are unchanged.

`TestGoCorpusCanonicalJSONSizeMatchesNormalization` compares the new count with `canonical.Normalize` on typed graph-shaped values that include HTML characters, U+2028/U+2029, control characters, quotes, backslashes, and literal escape-like text. The committed-corpus tests exercise source admission and resulting graph/context identities.

## Humanize benchmark

The benchmark used the same pinned source, toolchain, and RI binary for the baseline and treatment:

- Repository: `dustin/go-humanize`, commit `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`, tree `6d08f76afdc588592be994934de82a283b125d2c`.
- Go: `go1.27.1 windows/amd64`.
- RI executable SHA-256: `c86244aa6becf439eb770b1f4d4c0076ae65411ef15353d4f9300e1a74520ec0`.
- Benchmark: `BenchmarkGoCorpusResource/pinned-go-humanize/uncached`, one collection per sample, three samples per revision.

The baseline was the clean `4ec8d0252332618c22a316cc9653d20eaef61497` checkout. The treatment used the same HEAD with the size-counting change. Median results:

| Revision | Time | Allocated bytes | Allocations |
| --- | ---: | ---: | ---: |
| Baseline | 890.3 ms/op | 384.8 MB/op | 4,996,916/op |
| Treatment | 771.8 ms/op | 309.9 MB/op | 3,551,909/op |

The treatment's medians were about 13.3% lower for benchmark time, 19.5% lower for allocated bytes, and 28.9% lower for allocations. A verbose one-sample run from each revision reported identical source-set SHA-256 (`fb94f80b532caa7ef03585430f0d21fc938ebcab785596932da20597318354d5`), graph digest (`532c97fc645e3a90d1fb691da71d2103c9462cdb98c2452adb3947d8f839542e`), and context digest (`7e472ff99e392ef36e6486b114f28e55005892485a36148da4c3c2e36bc5a261`), with 19 files and 52,828 source bytes.

These are local Go benchmark allocation and elapsed-time measurements, not process RSS or end-to-end task latency. They do not measure provider behavior or establish a universal speedup.

To reproduce on Windows, set `ENGORCH_RI_BINARY` to the executable with the SHA above and `ENGORCH_RI_BENCH_REPOSITORY` to a checkout at the pinned commit, then run:

```powershell
go test ./internal/ri -run '^$' -bench '^BenchmarkGoCorpusResource/pinned-go-humanize/uncached$' -benchtime=1x -count=3 -benchmem
```
