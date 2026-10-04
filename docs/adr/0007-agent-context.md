# 0007: Bind repository guidance and workflows to role invocations

Status: ACCEPTED
Date: 2026-10-05

## Context and problem

Portable procedures were disconnected from Fabric's runtime roles, while
repository guidance had no explicit scope-aware input binding. Injecting every
workflow into every prompt would increase context and blur working methods
with executable authority.

## Decision and rationale

Capture committed AGENTS.md and standard skill files once as immutable run
input. Resolve parent guidance by task/write/changed-file scope. Advertise
role-filtered metadata and include only selected workflow bodies. Validate
task skill references against the bound catalog. Existing invocation identities
cover the resulting prompt; replay uses retained content, not current files.

Go role contracts and effect gates remain authoritative. Skills do not grant
capabilities. No recovery controller, instruction database or provider-specific
skill registration is introduced.

## Alternatives and consequences

Live filesystem reads during replay would allow instruction drift and are
rejected. Copying role contracts into Markdown would duplicate authority.
Native host discovery alone would not consistently bind inputs across routes.
The bounded retained inventory costs journal storage; metadata-only discovery
keeps unselected bodies out of prompts. No efficiency gain is claimed without
measurement.

## Compatibility and validation

The optional creation field preserves old record identities when absent.
New CLI autonomous runs enable capture by default with an explicit opt-out.
Skill references are a bounded optional task field; legacy execution rejects
nonempty selections. Started-task revisions preserve references. Qualification
must distinguish structural/replay checks from actual provider selection and
end-to-end behavior.

## References

- [Agent context guide](../guides/agent-context.md)
- [Authority](0002-authority.md)
- [Agent Skills specification](https://agentskills.io/specification)
