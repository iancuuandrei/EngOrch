# Typed Go graph sealing: paired allocation measurement

## Scope

This note measures one optional implementation detail in the Go engineering
graph: its internal digest serialization. It does not measure model quality,
provider behavior, task acceptance, or end-to-end user outcomes.

The treatment changes only the graph-content seal from `canonical.Hash` to
`canonical.TypedGeneratedHash`. The typed path accepts only closed tagged Go
values and falls back to `canonical.Bytes` for interfaces, marshal hooks, raw
JSON, byte slices, floats, non-string maps, embedding, invalid UTF-8, unsafe
integers, or unsupported nesting. Generic canonical normalization, decoding,
journal serialization, replay, and all other hash domains remain unchanged.

## Controlled paired run

- Frozen source: `e447adf4dceec58fba85dd19e5a390667ee79f6b`
- Disposable worktree: `D:\fabric-ci2-tools\v112-typed-overlay`, returned
  clean after the run.
- Benchmark: `BenchmarkPlannerGoContractResource/uncached`, `-benchtime=2x`.
- Pinned Humanize repository commit:
  `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`.
- Go toolchain: `D:\dev\EngOrch-toolchains\go\1.27.1\go\bin\go.exe`.
- RI executable SHA-256:
  `c86244aa6becf439eb770b1f4d4c0076ae65411ef15353d4f9300e1a74520ec0`.
- Measurement treatment overlay source hashes before later nil and byte-sequence
  fallback hardening:
  - `internal/canonical/typed_generated.go`:
    `77c2a88c6d8102a22d82e4b10ed5be09b7a1c199c488eb5f16dd05fc570101a4`
  - patched `internal/ri/engineering_graph.go`:
    `568f4057610e5758cb89a21f6f3ca49e29bc97efc394a2afd7cbd3aa06b09ad3`
- Final `internal/canonical/typed_generated.go` after the parity hardening:
  `d46ec6d419b2ff16e7e4e09c9b48b4623b1bd8add91c9ee156923fc8239ef95f`.
  The hardening routes nil and byte sequences to the unchanged generic path;
  graph content does not contain those shapes.

Three baseline and three treatment samples were run in alternating blocks.
The complete coarse receipt is private:
`D:\fabric-ci2-tools\v112-typed-paired-receipt.json`.

| Measure | Baseline mean | Treatment mean | Change |
| --- | ---: | ---: | ---: |
| Elapsed ns/op | 1,346,303,533 | 1,368,039,417 | +1.61% |
| Allocated B/op | 607,121,597 | 553,367,731 | -8.85% |
| Allocations/op | 7,224,866 | 6,555,159 | -9.27% |

The elapsed samples overlap (`1.328–1.367s` baseline,
`1.353–1.379s` treatment). This run supports the measured allocation reduction.
It does not establish a latency improvement or a material latency regression.

## Identity and replay checks

The treatment was checked with:

- `TestTypedGeneratedBytesParity`
- `TestTypedGeneratedBytesRetainsRawAndDepthBounds`
- `TestGraphTypedCanonicalSealMatchesGenericHash`
- `TestGoContractPlannerContextAdmissionAndReplay`
- `TestPlannerParseCacheOptInPreservesPlannerIdentityAndReplayDoesNotRequery`

These cover canonical byte/hash parity, graph reconstruction and validation,
and v4 contract-context admission, cold/warm parser-cache identity, and replay.
They do not replace independent review or product evaluation.
