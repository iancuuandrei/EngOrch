# Topology validation allocation experiment

The bounded topology query previously hashed supplied graph content and source
pairs, then rebuilt and hashed the same deterministic graph again. The candidate
rebuilds once and compares **all supplied content** with the reconstructed
content, as well as the canonical digest. Caller-owned graphs are never cached.
Stale digests and self-rehashed altered nodes, edges, generator directives and
source digests remain rejected. Existing canonical graph identities are unchanged.

## Local measurements

Windows amd64, Go 1.27.1, AMD Ryzen 9 9955HX; three benchmark samples per size,
300 ms minimum each. Fixtures have eight files per observed package, no calls
or imports, and one changed path. Construction occurs outside the timed region;
each query includes fresh validation, impact selection and topology serialization.

| Files | Before median ms/query | Candidate median ms/query | Before MiB allocated/query | Candidate MiB allocated/query |
|---|---:|---:|---:|---:|
| 24 | 3.320 | 2.364 | 1.510 | 1.134 |
| 256 | 28.749 | 23.186 | 16.429 | 12.160 |
| 512 | 52.612 | 45.156 | 33.033 | 24.361 |

Allocation counts fall from approximately 17,180/177,469/354,199 to
13,800/142,741/284,935 respectively. These are allocation totals, not peak RAM.
The small synthetic corpus does not establish provider token savings, coding
success, real repository graph latency, or a persistent cache benefit.

Private receipts: `D:\fabric-ci2-tools\topology-query-benchmark.log`,
`D:\fabric-ci2-tools\topology-query-optimized-benchmark.log` and
`D:\fabric-ci2-tools\topology-validation-focused.log`. Independent review and
clean final-candidate qualification remain required before integration.
