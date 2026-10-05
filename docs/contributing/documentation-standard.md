# Documentation standard

Normative project convention, informed by the
[research synthesis](../research/documentation-practices-2026.md).

Documentation MUST change with the behavior it describes. Each fact has one
authoritative home: specifications define invariants, ADRs explain decisions,
source comments define API use, and guides explain tasks. Historical decisions
MUST NOT masquerade as current implementation status. The [documentation hub](../README.md)
is navigation, not a second specification or evidence record.

Use plain language, sentence-case headings, working relative links and the
simplest executable example first. Commands omit shell prompts. Separate local,
fixture, provider and hosted evidence. Unknown results are never zero or PASS.

Significant Go packages use `doc.go`; Rust crates use module-level rustdoc.
Exported APIs document caller-visible behavior, errors, effects, lifecycle and
concurrency where material. Implementation comments explain non-obvious reasons.
Critical journal, identity, transition and reconciliation functions document
their invariants. Examples SHOULD execute as tests. Do not add README files that
duplicate native package documentation.

Consequential decisions use numbered ADRs with status, date, context/problem,
decision, alternatives, rationale, consequences, compatibility, validation and
references. Status is PROPOSED, ACCEPTED, SUPERSEDED or REJECTED. Specifications
describe accepted contracts using MUST/MUST NOT/SHOULD/MAY; implementation status
is reported separately. A specification is not evidence that code exists.

Checks MUST catch broken local links, duplicate ADR IDs, invalid ADR statuses,
undocumented exported APIs and stale generated references when those references
exist. CLI/config examples must be checked against their actual parser once
implemented. Conceptual prose is authored, not generated.

## Navigation and evidence

Keep a clear path from the repository README to the documentation hub, current
task guides, contracts and exact evaluation records. Link a summary to its
authoritative detail instead of repeating evolving rules in several places.
When newer evidence changes a status, preserve historical reports and add a
dated note or current index entry. State the source, candidate, artifact or run
scope next to each qualification claim. Distinguish implementation, local test,
provider, hosted, installed-package and release evidence. An unmeasured value is
unavailable, not zero.

## Diagrams and visual consistency

Use GitHub-compatible Mermaid for compact system, lifecycle and decision
diagrams. Draw only relationships supported by the current source or an
explicitly labelled proposal. Show optional integrations and authority
boundaries; do not make a model, reviewer assertion or telemetry feed look like
an approval gate. Keep diagrams smaller than their explanatory prose and link
to the authoritative contract for detailed rules.

For project diagrams, use a restrained shared palette: deep navy for the core
flow, teal for accepted evidence and verified boundaries, slate for external
or optional components, and amber for blocked or conditional paths. Define
meaningful Mermaid `classDef` styles locally so the diagram renders in GitHub.
Color MUST reinforce, not replace, text labels, edge labels or shape. Prefer
short node text, consistent direction and explicit labels on non-obvious
relationships. Avoid decorative gradients, unsupported icons, dense crossings
and diagrams that imply unqualified performance or release status.

Render changed Mermaid blocks before publication; a fenced block and valid
Markdown do not establish valid diagram syntax. Avoid reserved syntax words
as node identifiers. Check every repository diagram when changing shared
diagram conventions, and report the renderer version and executed scope.

When a diagram describes a process, keep its decision points consistent with
the corresponding guide and source. When a branch or release diagram describes
repository history, link the source-history policy and label rolling snapshots
separately from immutable tags and packages.
