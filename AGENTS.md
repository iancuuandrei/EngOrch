# Fabric agent instructions

## Scope and navigation

These instructions apply throughout the repository unless a deeper AGENTS.md
provides more specific guidance. Explicit user instructions take precedence.

Before changing behavior, read:

1. [README.md](README.md) for the product and documented entry points.
2. [docs/README.md](docs/README.md) to locate the authoritative area guide.
3. [CONTRIBUTING.md](CONTRIBUTING.md) for development and verification workflow.
4. The relevant specification or ADR before changing a documented contract.

| Responsibility | Authoritative location |
| --- | --- |
| Component structure and boundaries | docs/architecture/ |
| Required behavior and invariants | docs/specifications/ |
| Architectural decisions and rationale | docs/adr/ |
| Current user workflows | docs/guides/ |
| Command and interface reference | docs/reference/ |
| Scoped measurements and qualification | docs/evaluation/ and release records |
| Planned work and product priorities | docs/roadmap/ |
| Contribution and publication | CONTRIBUTING.md and docs/contributing/ |

Historical records describe their exact source, artifact or run. Do not infer
current behavior from them. Link to authoritative details instead of copying
changing rules into this file.

## Engineering workflow

Inspect relevant code, checks and documentation before editing. Establish the
affected contracts for substantial features or refactors, then make the
smallest coherent change that satisfies the task. Avoid speculative machinery.

Go orchestration code is under cmd/ and internal/; Rust repository intelligence
stays behind its documented interfaces. Preserve schemas and ownership across
these boundaries.

Preserve unrelated dirty changes, worktrees, journals and provider state.
Never reset or discard user work to obtain a clean checkout. Review the diff
and update documentation with externally visible behavior changes.

## Correctness and authority

Preserve external-effect ownership, retry authority, invocation/candidate/source/
run identity, immutable policies, durable evidence and bounded repair budgets.
Keep candidates isolated until their applicable verification and review pass.

Never repeat an external effect with an uncertain outcome to discover whether
it succeeded. UNKNOWN stays distinct from success and failure until admissible
evidence settles it. Diagnostics and telemetry cannot grant authority, approve
candidates or count as semantic results.

Expose the failed stage, a safe bounded diagnostic, retained identity and a
legal next action. Keep credentials, raw provider responses and prompts out of
public diagnostics. Prefer graceful degradation of optional capabilities and
measured resource admission over unexplained termination.

## Verification and evidence

Follow the pinned toolchains and checks in [CONTRIBUTING.md](CONTRIBUTING.md).
Format changed code. When tests or verification are requested, start with
discriminating checks for the changed surface; broaden only as justified.
Do not start repeated benchmarks merely to claim progress.

Report PASS, FAIL, NOT RUN, BLOCKED and unavailable usage honestly. Never claim
an unexecuted check, benchmark, provider run or qualification. Local tests do
not establish installed, hosted or release qualification. Do not weaken frozen
oracles or combine separate attempts into an unchanged-cohort claim. Cached
input is a subset of input; reasoning output is a subset of output.

## Documentation and publication

Follow [the documentation standard](docs/contributing/documentation-standard.md).
Preserve historical evaluation scope; specifications define required behavior
and do not prove implementation or qualification.

Branch, commit, automation identity and publication rules have one authoritative
home: [source-history.md](docs/contributing/source-history.md). Read it before
committing or publishing. Preserve immutable tags, assets and SHA-bound evidence.
Other public-history rewrites and destructive Git operations require explicit
user authorization and must satisfy that policy.

## Completion

Confirm the requested change, relevant executed checks, matching documentation
and preservation of unrelated work. Summarize what changed and the exact
verification performed without extending evidence beyond its demonstrated scope.
