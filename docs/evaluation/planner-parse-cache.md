# Opt-in committed Go parse cache

`CollectCommittedGoCorpus` keeps its legacy no-cache behavior and rejects a
nonempty cache directory. Callers that own an explicit local cache may opt in
through `CollectCommittedGoCorpusWithOptions` with
`GoCorpusOptions{EnableParseCache: true}` and an absolute, clean, non-root
cache directory. The existing RI cache validates entries against source path,
source bytes, and parser producer identity. A changed committed source misses
its old entry; corrupt entries are reparsed by the RI implementation.

The directory must be writable only by trusted users/processes. Validation
checks structure, source/producer bindings and content digests; it is not
authentication or proof that cached facts are complete. A writer able to alter
and re-hash a cache entry can omit facts. Shared or untrusted cache writers are
outside this opt-in's trust model. Facts remain partial advisory evidence and
cannot authorize candidate effects or replace verification/review.

The cache only reuses validated Go syntax facts. Corpus selection still reads
and binds the committed Git objects, and graph construction, candidate file
effects, native checks, review, and final acceptance are not skipped. Cache
hit/parse counters are observability fields; they do not change graph identity.

## Local benchmark

Run the bounded fixture benchmark with the pinned Go and RI executable:

```powershell
$goDir = 'D:\dev\EngOrch-toolchains\go\1.27.1\go\bin'
$env:PATH = "$goDir;$env:PATH"
$env:ENGORCH_RI_BINARY = 'D:\fabric-ci2-tools\ri-facts-target\debug\engorch-ri.exe'
& "$goDir\go.exe" test ./internal/ri -run '^$' -bench '^BenchmarkCollectCommittedGoCorpusParseCache$' -benchtime=1x -count=3 -benchmem
```

Observed on Windows/amd64 with Go 1.27.1 and RI executable SHA-256
`c86244aa6becf439eb770b1f4d4c0076ae65411ef15353d4f9300e1a74520ec0`: the
24-file fixture contained 31,437 committed source bytes. Cold runs had 24
misses, 0 hits, and 71,676 cache bytes; elapsed time ranged from 647 ms to
691 ms, with 165–167 MB/op of Go allocations. Warm runs had 24 hits, 0
misses, and the same cache size; elapsed time ranged from 559 ms to 656 ms,
with 165–168 MB/op of Go allocations. Cache priming, cache-directory creation,
and hit/disk accounting are excluded from both measurements. These three
single-iteration samples show lower warm elapsed time in this fixture, but are
too small to establish a general latency benefit; they show no material
allocation reduction. They verify reuse and unchanged source/graph identity,
not a production performance claim.

The benchmark reports elapsed wall time and cumulative Go allocations. It does
not measure CPU time, peak resident memory, or provider/model performance. This
small synthetic corpus is not a production throughput result; larger and
representative repositories need separate measurement before enabling the
cache in a product workflow.
