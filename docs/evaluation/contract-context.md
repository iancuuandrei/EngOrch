# Contract-context planner evidence

`go-contract-context-v1` is an opt-in planner evidence mode. It requires the
same pinned local RI executable and SHA-256 binding as the other Go planner
context modes. Its immutable admission record is written before planner
dispatch; it does not create a provider, generator, workspace, or write effect.

The compiler consumes an already admitted Go graph, its complete admitted Go
source bytes, optional declared module inventory, and optional literal
generation metadata. It does not open files, invoke RI, execute generators,
call a provider, or resolve Go types.

It emits `engorch.ri.go-contract-context.v1`: partial, deterministic evidence
with a source ID, optional candidate ID, graph/producer/module/generation
digests, exact source SHA-256 values, byte spans, excerpt hashes, omissions and
a canonical digest. Missing evidence never establishes absence, safe scope, or
semantic correctness. Syntactic calls remain `UNRESOLVED`; definitions,
references and implementations are outside this compiler.

## Prompt budget

`fabric run --autonomous --planner-context go-contract-context-v1` uses this
view in place of the broad v2 source and generation prompt evidence. It also
requires `--planner-context-ri-executable` and its SHA-256 binding. The
admission record retains the graph, source digests, selected excerpts and
generation metadata so replay can revalidate their identities, ranges,
hashes, ownership bindings and compact-context digest without invoking RI.
For this mode it also retains the complete admitted Go and generation compiler
inputs as local-inspection-only string records; they are never included in the
planner prompt. If the complete record would exceed 768 KiB, admission retains
the longest relevance-ranked parsed prefix with explicit
`contract_record_budget` omissions. If no prefix can be admitted, it records
an unavailable result rather than presenting incomplete source evidence as a
complete context.
Generation discovery metadata remains durable, while its transient source
bytes are retained only for literal generator bindings selected by the record.
This avoids duplicating unrelated discovery files in the journal and does not
add any generation material to the model prompt.

- at most 12 KiB of declaration/test excerpts;
- at most 4 KiB of generation owner/template/tool/output excerpts;
- at most 12 excerpts across six paths, with a 2 KiB excerpt ceiling;
- at most 32 source-bound graph relations.

Selection starts with exact objective identifiers, then one-hop graph impact,
syntax-associated tests/examples, and explicit `GENERATED_BY` chains. Declared
module ownership is shown only when the supplied inventory gives one unique
owner. Generation template evidence is selected only from a literal discovered
binding; the compiler does not infer an inactive template relationship.

The current v2 experiment increased observed uncached input without improving
fixed acceptance. This mode therefore needs a fresh, same-pin A/B trial:
primary outcome is held-out accepted task completion; token accounting and
prompt bytes are secondary measurements. No prior run is resumed for that
comparison.
