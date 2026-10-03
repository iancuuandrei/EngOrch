# Committed Go module inventory v1

`internal/ri.CollectGoModuleInventory` observes regular committed `go.mod`,
`go.work`, and `vendor/modules.txt` files from one repository identity. It
preflights each manifest's size, applies per-file, total-byte, and manifest-count
bounds, and copies admitted content through the repository's single selected
source batch. Each observation records its committed Git blob identity and,
when content was read, its byte count and SHA-256. Parse failures, unsupported
file kinds, oversize inputs, and exhausted inventory budgets remain explicit;
they cannot silently produce complete coverage.

The parser dependency is pinned to `golang.org/x/mod v0.38.0` and is used only
for the official `modfile.Parse` and `modfile.ParseWork` syntax trees. The
upstream module's `LICENSE` is the Go Authors' BSD-style three-clause license;
the source distribution and license are available from the
[v0.38.0 module source](https://github.com/golang/mod/tree/v0.38.0) and the
package API is documented at
[pkg.go.dev/golang.org/x/mod/modfile](https://pkg.go.dev/golang.org/x/mod/modfile).
The module reference describes the declarations recorded here:
[Go Modules Reference](https://go.dev/ref/mod).

## What ownership means

`GoModuleOwnershipForPath` uses the unique deepest valid containing `go.mod`
to derive a declared module path plus the source file's directory. If a
containing nested manifest is malformed, unreadable, oversized, or otherwise
omitted, the answer is `ambiguous`; the helper never falls back to a parent
module. If no containing module is observed, it returns
`source_local_fallback`, preserving the pre-existing source-local identity
behavior. If inventory enumeration itself was truncated, every ownership
answer is ambiguous because an unseen nested module could change the result.

Sensitive and protected paths are rejected before they enter the selected
content batch. Sensitive or unsafe paths are redacted; protected and
unsupported-kind paths are path-only omissions. Omitted records carry no blob
identity, content hash, or parsed declarations. An omitted `go.mod` makes
ownership ambiguous because its root cannot safely be disclosed for matching.

The inventory reports `go.work use` and `replace` declarations as written and
resolves safe workspace/local-replacement paths only to repository-relative
lexical roots. It also records the exact `vendor/modules.txt` blob and digest. These
are source declarations and metadata presence, not evidence that a particular
Go invocation activated a workspace, local replacement, or vendor mode. The
helper does not invoke `go list`, `go env`, a compiler, module resolution, or
vendor execution. It does not establish active dependency selection, package
type resolution, or call resolution.

The serialized inventory digest binds the normalized observation to the
repository ID, commit, and tree. `ValidateGoModuleInventoryRecord` checks the
record's bounded structure and digest without filesystem access;
`ValidateGoModuleInventory` additionally checks the exact supplied immutable
repository identity. Neither validator re-runs a Go command or claims that
declarations are active build configuration.

Sensitive and protected manifests are excluded before content batching. Their
omissions contain no blob, hash or parsed declarations; sensitive/unsafe paths
are redacted. An omitted `go.mod` leaves ownership ambiguous. The CLI additionally
re-collects the committed inventory before accepting one embedded in a graph
specification, so a caller cannot establish ownership by changing declarations
and recomputing a record digest. This does not make trusted-local derived
records cryptographically authenticated.
