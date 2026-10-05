# Fabric contributor instructions

## Product objective

Fabric builds useful agent infrastructure. Prioritize the documented journey:
clone, install/build, configure a model, initialize, perform a real coding task,
plan/delegate/execute/review, and inspect durable results. Ship capabilities;
harden demonstrated failures. Self-hosting is dogfooding, not mandatory authoring.
Direct Codex/human product edits are allowed.

Freeze unnecessary controller/replay/settlement expansion. Infrastructure repairs
must address data corruption, duplicate external effects, or a demonstrated
normal-use blocker. Bound investigations and prefer the smallest useful repair.
Do not start repeated benchmarks or governance-only work packages.

## Branches and commits

- Work on `dev` and preserve its complete incremental history. New development
  commit titles begin with the current patch version, e.g. `v1.1.3: explain ...`.
- `main` contains one rolling snapshot per major/minor line, e.g. `v1.0`, `v1.1`.
  The historical v0.0.0/v0.0.1 bootstrap checkpoints remain separate.
- Each trusted push to `dev` updates the matching `main` snapshot to the exact
  dev tree. For the same major/minor line, replace only the latest checkpoint,
  preserving its title, sole parent, original author date and committer date.
- A newer major/minor line creates one new checkpoint with the previous main
  snapshot as its parent. Older-line promotions are refused.
- All new commits use author AND committer
  `automation:fabric-v1-sol-supervisor <automation@fabric.invalid>`.
  Do not add human coauthors or `Co-authored-by` trailers.
- Never fast-forward `dev` onto `main`, merge the development lineage into main,
  or add ordinary fix/docs/merge commits to main. Keep every dev commit intact.
- The `automation:fabric-v1-sol-supervisor` workflow performs rolling snapshot
  updates from exact trusted current dev SHAs, using a main force-with-lease.
  The user has explicitly authorized these updates on every dev push; do not
  request confirmation again for each update.
- Review changes and perform applicable requested verification before publishing
  dev. Milestone source snapshots are separate from installed/release qualification.
  Published releases still require their applicable review, acceptance and
  Sonar/security gates.
- If explicitly asked to create a PR targeting main, use a separate automation
  bot-backed publication procedure and
  a snapshot branch; do not submit the full dev lineage. Attach created PRs to
  the Codex task. The rolling synchronization job itself does not create PRs.
- Preserve published tags, release assets and SHA-bound evidence. Never move
  them to match a rolling checkpoint. The main commit SHA changes as its tree
  changes; its preserved date does not mean the source tree is unchanged.
- Other public history restructures require explicit authorization, an exact
  remote backup first, and `--force-with-lease` with the expected old ref.

The authorized October 2026 consolidation leaves `v0.0.0`, `v0.0.1`, `v1.0`,
and `v1.1` on main. The v1.1 snapshot includes subsequent v1.1.1-v1.1.3 fixes
and these rules, with the original v1.1.0 checkpoint date. It does not replace
that already published tag or package. The full development history is on dev.

## Correctness and evidence

Preserve external-effect ownership, invocation/candidate/source identity,
immutable policies and bounded repair budgets. Never resend an uncertain effect.
UNKNOWN stays UNKNOWN until admissible evidence settles it. Diagnostics cannot
grant retry authority or count as semantic results.

Failures should expose the stage, a safe bounded diagnostic, retained run/result
identity and a legal next action. Keep credentials, raw provider responses and
prompts out of public diagnostics. Prefer graceful optional-capability degradation
and measured resource admission over unexplained abrupt termination.

Do not weaken oracles, alter frozen acceptance criteria, pool attempts into an
unchanged-cohort claim, or count controller generations as product completion.
Report PASS, FAIL, NOT RUN, BLOCKED and unavailable usage honestly. Cached input
is a subset of input; reasoning output is a subset of output.

## Development workflow

Read the current README and relevant guides/evaluation records before changing
behavior. Go code is under `cmd/` and `internal/`; repository intelligence and
release tooling have documented interfaces. Preserve schemas and ownership
contracts across those boundaries. Update user documentation with behavior changes.

Run/add tests when requested by the user; prefer discriminating regressions and
targeted checks to repeated full-suite/benchmark runs. Formatting and static
inspection are routine checks. Never report unexecuted qualification.

Preserve unrelated dirty changes, journals, worktrees and provider state. Use
an isolated checkout when needed. Do not reset dirty checkouts or delete evidence.

Use subagents only when explicitly requested or authorized for delegation; when
used, select GPT 6 Luna with High reasoning. Give writers explicit ownership
and warn that other agents share the codebase. Muse Spark through OpenCode is
preferred when usable; an uncertain route failure is not permission to repeat
effects or claim success.
