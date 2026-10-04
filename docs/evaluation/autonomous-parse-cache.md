# Autonomous planner parse cache

`ExecutionPolicy.planner_parse_cache_version: 1` opts `go-source-context-v2`
and `go-contract-context-v1/v2/v3` into reuse of the pinned RI's validated
committed-Go syntax facts. Zero or an omitted field keeps the uncached
collection path and its existing run identity.
The opt-in is local to planner-context preparation; it does not enable reuse
for candidate observations, native checks, review, or final verification.

The controller derives a private cache directory beneath the operating
system's user cache directory. Its namespace binds the stable checkout
identity (checkout name/root/Git common directory/object format) and the
pinned RI executable digest. A new producer digest therefore gets a separate
leaf and does not migrate or overwrite another producer's cache. Individual
facts are keyed by source path, source digest, parser/schema version, and
producer digest. A new commit reuses only facts whose complete key still
matches. The absolute cache path and hit/miss counts are not written into the
journal or planner prompt. A run that already has an admitted planner context
replays that immutable record without querying the cache again.

The cache is a trusted-local performance optimization, not an authority or
integrity boundary. RI revalidates cached fact shape and bindings before use;
candidate source is separately observed without this cache, and all native,
review, and final checks remain fresh. A user with write access to the private
cache may replace an entry and recompute its hashes, so the cached syntax facts
are advisory just like freshly parsed syntax facts. Cache corruption is
discarded and reparsed; errors do not cause a provider call or grant file-write
permission.

Each per-checkout/per-producer RI leaf retains at most 1,024 published fact
files and 64 MiB across those files plus its small policy marker. Temporary
files are included in the byte reservation while a managed writer publishes;
the zero-byte lock file is not. A one-time locked migration trims older leaves
to these limits before publishing the marker. When an insertion would exceed
either bound, the oldest fact by file modification time is evicted; equal
timestamps use the cache filename as a stable tie-break. The marker is never
evicted. A bounded scan rejects symlink, non-regular, or malformed entries for
storage, and a create-new local lock serializes cache publications. If another
writer holds the lock during migration/publication or the cache cannot safely
be scanned, the freshly parsed facts are returned without publishing that
result. A leftover lock after process interruption is not recovered; writes
remain unpublished while the lock exists, while already-published exact hits
remain readable if the policy marker is valid. These limits govern files
managed by cooperative RI writers. An external process
with access to the cache directory can add unrelated files after the fast-path
policy check; that does not expand cache authority, and the next managed write
will reject an unsafe scan. Eviction, lock contention, and cache write
failures never turn cached data into verification evidence.

Focused evidence is collected with the pinned RI executable:

```powershell
$env:ENGORCH_RI_BINARY = 'C:\path\to\pinned\engorch-ri.exe'
$env:ENGORCH_RI_BENCH_REPOSITORY = 'C:\path\to\clean\go-humanize-checkout'
go test ./internal/control -run '^TestPlannerParseCache' -count=1
go test ./internal/ri -run '^TestCollectCommittedGoCorpusParseCache' -count=1
go test ./internal/ri -run '^$' -v -bench '^BenchmarkGoCorpusResource$' -benchmem -benchtime=1x
```

For an opt-in process-tree sample, provide an existing output directory:

```powershell
.\scripts\measure-ri-cache.ps1 `
  -RiBinary $env:ENGORCH_RI_BINARY `
  -HumanizeRepository $env:ENGORCH_RI_BENCH_REPOSITORY `
  -GoExecutable 'C:\path\to\go.exe' `
  -OutputPath 'C:\path\to\existing-evidence-dir\parse-cache.json'
```

The resource benchmark includes a deterministic 24-file, 64-KiB-per-file
fixture and optionally the exact pinned `dustin/go-humanize` checkout. It
reports Go allocation/time measurements and cache disk size/hit metrics; it
does not establish an end-user latency claim. The benchmark does not invoke a
model or provider. Final candidate and native checks are outside the benchmark
and remain uncached.

One five-sample `-benchtime=1x` run on Windows amd64, Go 1.27.1, the pinned RI
executable, and go-humanize commit
`a1b4e66b9a6d890e9e15e7091cf16c8032367d6e` produced these median
per-collection times:

| Input | Uncached | Cold cache | Warm cache | Warm hits | Cache bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| Synthetic 24 × 64 KiB | 0.759 s | 0.782 s | 0.606 s | 24 | 16,440 |
| Pinned go-humanize | 0.951 s | 0.999 s | 0.952 s | 19 | 135,024 |

The synthetic source set was 24 files and 1,571,842 bytes; the selected
go-humanize corpus was 19 files and 52,828 bytes. Graph and context construction
were outside the timer, and their digests matched across uncached, cold, and
warm runs for each input. Allocations remained about 642,000/op for the
synthetic fixture and 4,997,000/op for go-humanize across all three modes; the
cache did not materially reduce Go allocations. Five samples are still small
and variable, so these timing differences are observations, not a qualified
performance claim. The
optional `scripts/measure-ri-cache.ps1` adds a 100-ms process-tree working-set
and CPU sample; that instrumentation can perturb runtime and can miss short
process peaks between samples.

For the pinned go-humanize input, the benchmark logged source-set SHA
`fb94f80b532caa7ef03585430f0d21fc938ebcab785596932da20597318354d5`,
graph digest `532c97fc645e3a90d1fb691da71d2103c9462cdb98c2452adb3947d8f839542e`,
and context digest
`7e472ff99e392ef36e6486b114f28e55005892485a36148da4c3c2e36bc5a261`. The
benchmark output also logged matching per-input source, graph, and context
digests across all three modes.
