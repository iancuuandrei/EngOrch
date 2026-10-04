# Structured repair matched pilot

## Frozen scope

Compare current repair with the opt-in structured Finding/RepairSpec treatment
on the historical v1 godotenv task, upstream source
`ddf83eb33bbb136f62617a409142b74b91dbcff3`. This public task has been exercised
previously; it is neither an unseen benchmark nor the frozen v2 cohort.
The existing v1 manifest objective, native checks and evaluator remain unchanged.

Use one fresh prepared clone per arm, the same product binary and Codex runtime,
GPT 6 Luna High for all roles, serial concurrency and the existing two-repair
ceiling. The treatment uses `--repair-intelligence`; the control omits it.
This flag also enables explicit reviewer rechecks. Report that bundled policy
difference rather than attribute effects to one component or to imported
coverage. No automatic coverage acquisition is part of this trial.

Record accepted candidate/external acceptance, repair attempts, exact closed
findings where available, model invocations, verification executions, typed
token categories and elapsed time. Unavailable metrics remain unavailable.
Every arm is dispatched once; uncertain effects are retained without resend.
If no fixer invocation occurs, mark repair quality NOT_EXERCISED. No successor
attempt is authorized by this experiment. Freeze identities and bounds before
dispatch, retain private raw journals, and publish only bounded summaries.

Status: **NOT QUALIFIED**; repair quality **NOT_EXERCISED**. Both fresh arms
stopped after workspace registration, before any fixer or candidate verification.
This record does not qualify default adoption or v2.

## Runner integration evidence

The v1.1.18 runner adds explicit Native-only `-RepairIntelligence`, preserves
default argument order and legacy request omission, and checks requested and
observed policy types/versions rather than coercing strings or booleans.
Configured structured review also requires its bundled recheck policy.
The new repair-treatment regression and existing NativeRunArgs/AgentContext
regressions: **PASS**. Documentation check: **PASS** (0.559 s). These checks
perform no provider calls and are separate from the live pilot outcome.
Independent GPT 6 Luna High static review: **APPROVE** for the runner integration;
the reviewer did not execute checks or providers.

## Observed pilot

Exact product source: `c47d0ef21b3115596b575e739d81dbe2d25c8852`; binary SHA-256:
`b7a560bde5a699f87a2d8987d0a7f927f914b3a353557ebf757c4e8f66173c52`.
[Machine-readable summary](repair-intelligence-pilot-20261005.json) binds runtime,
runner and candidate-copy hashes, fresh run identities and measured usage.
Private raw evidence is retained under
`D:\fabric-ci2-tools\repair-intelligence-pair-v1118-20261005`.

Both arms passed upstream preflight and produced the expected historical
held-out baseline failure. Each then completed exactly one receipt-matched
planner invocation. Git registered the intended worktree but could not observe
its Windows working directory: control path 263 UTF-16 units, treatment 267.
The run outcome remains `IMPLEMENTING / effect_requires_reconciliation`, with
workspace outcome **UNKNOWN**. Control reconciliation observed the same path
failure and did not settle it. No writer/fixer was invoked; candidate checks,
external acceptance and finding closure were not executed. No effects were
resent and no successor attempt is pooled into this pilot.

| Arm | Input | Cached input | Uncached input | Output | Reasoning output | Evaluation task wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Control | 83,041 | 60,160 | 22,881 | 1,330 | 494 | 40.524 s |
| Treatment | 96,867 | 73,216 | 23,651 | 1,356 | 414 | 40.077 s |

These are matched completed planner observations, not successful repair costs.
Cached input is part of input; reasoning is part of output. Cost and time to
READY are unavailable. Ordered execution and incomplete tasks establish no
causal benefit. A preliminary local startup failed before Fabric initialization
because its candidate-copy argument was missing; that zero-provider preflight
attempt is retained separately and is not another model attempt.

## Demonstrated failure hardening

v1.1.19 checks the Windows workspace destination before planner dispatch and
before new workspace intent/registration. It exposes a safe shorter-checkout
diagnostic. Request validation and old pending journals remain replayable;
UNKNOWN is not settled by a path diagnosis. This prevents spending model tokens
and creating an uncertain workspace for this demonstrated unsupported path.
It does not remove the underlying Git/platform path limit or qualify repair.

Targeted worktree, autonomous preflight, failure classification, isolated
materialization and documentation regressions: **PASS** (worktree 9.013 s,
control 4.415 s, documentation 0.793 s). `go vet` for worktree/control and
`git diff --check`: **PASS**. Independent GPT 6 Luna High static review:
**APPROVE**; the reviewer did not run providers or tests. Regression coverage
includes UTF-16 boundaries, manual approved workspace startup, replayable
historical requests and UNKNOWN precedence.

Next experiment: use a short checkout root and declare a separate fresh cohort
before dispatch. Retain both original UNKNOWN effects and this failed pilot;
do not retry their workspace effects or pool subsequent outcomes with these arms.
