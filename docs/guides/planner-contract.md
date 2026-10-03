# Explicit planner assignment

`planner_contract = "plan-v1"` opts a new run into an explicit planner-only assignment. The planner receives a canonical envelope containing its role, the planning instruction and the exact original objective. Read-only source access is expected; editing and verification belong to later roles. The wrapper does not grant tools or effects.

The empty setting retains the historical invocation bytes and identity. Unknown values fail closed. The chosen setting is part of immutable run configuration, and dispatch, replay, access settlement, scheduled identity reconstruction and usage accounting use the same planner invocation constructor.

The writer receives the original objective plus the approved plan through its existing implementer contract. Planner-specific instructions are not inserted into writer/fixer prompts. This clarifies role responsibility without changing the task requirements or writable scope.

A completed model invocation is not automatically an acceptable plan. The operator's exact plan gate still rejects responses that request capabilities instead of proposing the bounded implementation. Bootstrap goal authority permits the operator to approve a conforming plan locally; it does not permit scope expansion.

## Optional committed-source context

`fabric run --autonomous --planner-context source-bounded-v1` records a bounded committed-source manifest before the planner is dispatched. The manifest is part of the planner input and can be inspected with the run record. It binds the configured repository identity, commit and tree, the objective digest, and every selected excerpt.

The collector scans committed eligible paths, ranks objective path hints, and attempts at most 24 complete UTF-8 blobs of at most 32 KiB each. It selects at most 12 excerpts totaling 48 KiB. The record explicitly states that this is partial committed-source coverage, that no RI query was available, and that omitted paths do not establish absence or irrelevance.

The option is independent of role task context and does not create a workspace, candidate, RI binding, or provider effect. With no planner-context option, the legacy planner input and invocation identity remain unchanged.
