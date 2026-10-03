# Explicit planner assignment

`planner_contract = "plan-v1"` opts a new run into an explicit planner-only assignment. The planner receives a canonical envelope containing its role, the planning instruction and the exact original objective. Read-only source access is expected; editing and verification belong to later roles. The wrapper does not grant tools or effects.

The empty setting retains the historical invocation bytes and identity. Unknown values fail closed. The chosen setting is part of immutable run configuration, and dispatch, replay, access settlement, scheduled identity reconstruction and usage accounting use the same planner invocation constructor.

The writer receives the original objective plus the approved plan through its existing implementer contract. Planner-specific instructions are not inserted into writer/fixer prompts. This clarifies role responsibility without changing the task requirements or writable scope.

A completed model invocation is not automatically an acceptable plan. The operator's exact plan gate still rejects responses that request capabilities instead of proposing the bounded implementation. Bootstrap goal authority permits the operator to approve a conforming plan locally; it does not permit scope expansion.

## Optional committed-source context

`fabric run --autonomous --planner-context source-bounded-v1` records a bounded committed-source manifest before the planner is dispatched. The manifest is part of the planner input and can be inspected with the run record. It binds the configured repository identity, commit and tree, the objective digest, and every selected excerpt.

The collector scans committed eligible paths, ranks objective path hints, and attempts at most 24 complete UTF-8 blobs of at most 32 KiB each. It selects at most 12 excerpts totaling 48 KiB. The record explicitly states that this is partial committed-source coverage, that no RI query was available, and that omitted paths do not establish absence or irrelevance.

The option is independent of role task context and does not create a workspace, candidate, RI binding, or provider effect. With no planner-context option, the legacy planner input and invocation identity remain unchanged.

### Opt-in Go source graph

`go-source-context-v1` and `go-source-context-v2` are separate opt-in modes for repositories with committed Go source. Invoke either with the matching `--planner-context` value and explicit `--planner-context-ri-executable PATH --planner-context-ri-executable-sha256 SHA256` flags. Both require the same absolute clean parser path and 64 lowercase hexadecimal digest; the selected path and digest are immutable run inputs, and admission verifies executable bytes before planner dispatch. No executable is discovered from PATH or environment variables. The v1 planner evidence contract remains unchanged; v2 is a separately versioned controller contract. With no mode selected, legacy behavior remains unchanged.

Admission uses the pinned parser to build bounded evidence from committed Go files before planning. The resulting graph/context is partial and source-bound; missing, omitted, or unresolved files do not prove absence or correctness. This mode leaves the existing `source-bounded-v1` behavior unchanged and remains independent of role context, prompt ordering, and scheduler settings.

The collector attempts at most 24 files, admitting complete UTF-8 sources of
at most 1 MiB each and 8 MiB in total. Git `cat-file --batch-command` checks
blob sizes before requesting bodies, so over-budget objects are skipped
without reading their contents. Parser-request and 512 KiB graph budgets may
omit additional whole files; counts and reasons remain in the durable record.
Package identities are explicitly source-local namespaces, including test
associations. They do not establish module ownership or resolve imports.

Both modes use context schema v2 with bounded syntax-based caller hints and
separated contract excerpts. The `go-source-context-v1` recipe allows 48 KiB
of Go excerpts; `go-source-context-v2` allows 24 KiB plus at most 8 KiB of
committed generation owner/tool/template/build files. Generation context admits
only whole files whose content hash can be checked from persisted JSON; sources
that do not fit are explicitly omitted while ownership paths remain visible.
V2 records retain
literal directive and output bindings, exact supporting source digests and
discovery omissions; the prompt receives a compact ownership view. Discovery
uses a narrow literal command/Makefile/embed recipe and never executes it.
Missing bindings or omitted sources remain partial evidence. Calls remain unresolved; hinted callers are not type-checked
references. The planner receives this compact view, not the complete graph.
The graph, source digests, context and objective remain bound in the journal
for inspection and replay. Parsing does not execute generators or grant write
scope, approval or verification. No context cache is used in this first mode.
