# Deterministic Go format observation

`fabric observe-format RUN MANIFEST_JSON CACHE_DIR` records one local artifact
for a fixed in-process `go/format.Source` operation. It is an opt-in diagnostic
for exact candidate bytes. It never runs a configured check, appends a journal
event, consumes a repair, creates a review input, or changes `READY`.

`MANIFEST_JSON` is canonical JSON with exactly this shape:

```json
{"paths":["pkg/example.go"],"version":1}
```

Paths are sorted, unique, repository-relative eligible `.go` files. The command
requires a replay-valid run with a confirmed workspace candidate. It holds a
read lease and rejects source drift. At most 24 files of 32 KiB each are read;
a larger or partial file is rejected rather than partially formatted.

The action key binds the exact candidate and workspace identities, every input
path/hash/size, the formatter engine version, hash of the running Fabric
executable before and after the observation, Go runtime version, platform, and
the fixed no-environment/virtual-cwd contract. It has no argv, glob, directory
walk, ambient environment, network, subprocess, or write to the candidate.

The result is `unchanged`, `formatted`, or `format_error`, never PASS or FAIL.
The cache stores canonical metadata plus a raw payload containing the formatted
bytes. Output includes the metadata and payload paths when publication
succeeds, so an operator can inspect the actual bounded result. During local
cache-lock contention, the computed observation is returned with
`cache_state: "contended"` and no artifact paths are advertised.

`CACHE_DIR` is an explicit absolute clean non-root directory outside the
workspace, repository, and Git metadata. Cache entries are keyed by the action
identity, validated on read, and published atomically. Corrupt entries are
misses. The cache holds at most 64 artifacts and 48 MiB including metadata,
payloads, and reserved publication space; metadata is capped at 16 KiB and
formatted payloads at 768 KiB. The formatter also rejects any individual output
greater than 32 KiB, so expansion cannot bypass the input bounds.

A cache hit describes a prior local formatting artifact for the exact action.
It is not fresh verification evidence. `fabric verify` and final release gates
remain fresh and do not consult this cache. Generic commands such as `go test`
are not eligible for this reuse contract because Fabric does not have their
complete dependency, toolchain, environment, time, network, or side-effect
closure.
