# Candidate facts cache resource observations

Nine identity-audited local measurements used frozen v1.0.14 source
`66252573e4377f6c9f409bc72cc22a4c034e9d01`, tree
`38377465053030919f7c807c217a75b8fd457636`, in a disposable overlay. Its 1,240
tracked files matched the frozen source; the sole extra source was a temporary
candidate-cache benchmark, SHA-256
`d270c863e4c809be69ec27579f6c63dff932027c5e6d52e435c2558d29cb32a1`.
The benchmark did not call a provider. This is a bounded candidate fixture,
not an end-user latency measurement or a real coding-task result.

All nine samples exited 0 and retained one candidate identity
`80029f78e6e1ca0fcde8e8e9dc24c112338898b1421673099dabbee117400aca`,
one files hash, module inventory, graph digest and normalized output digest.
Uncached/cold collection recorded one miss; warm collection recorded one hit.
Each cold/warm leaf contained two regular files including the marker, 702
bytes total. Candidate/source/module binding was revalidated despite the hit.

| Mode | Median collection seconds | Range seconds | Median allocated bytes | Median allocations |
| --- | ---: | --- | ---: | ---: |
| uncached | 2.856 | 1.747–3.072 | 8,556,744 | 50,923 |
| cold | 1.715 | 1.658–2.612 | 8,559,752 | 50,960 |
| warm | 2.537 | 1.677–2.813 | 8,530,552 | 50,902 |

| Mode | Median whole-process wall seconds | Max sampled process-tree RSS bytes | Max sampled cumulative CPU ms |
| --- | ---: | ---: | ---: |
| uncached | 8.164 | 292,687,872 | 3,859.4 |
| cold | 5.359 | 343,306,240 | 5,484.4 |
| warm | 11.279 | 344,899,584 | 4,718.8 |

Collection timings and wrapper timings measure different scopes; wrapper
samples include setup. Go/Rust suites were active on the host. Ranges overlap
broadly, allocations are effectively unchanged, and 100-ms sampling can miss
short-lived peaks. These observations establish exact reuse and bounded local
storage for this fixture, **not a speedup or memory improvement**. The cache
remains opt-in. No model-token, provider-cost or time-to-READY saving follows
from a parser hit.

Identity-audited analysis receipt SHA-256:
`ffc5f54da7aa93c6f558755bc8fa8731d3e60123df181f4498b5a7e582643437`.
Preliminary contaminated-overlay measurements and the first non-verbose
identity run were retained separately and were not substituted for this
receipt. Historical RI binaries and the clean frozen checkout were preserved.