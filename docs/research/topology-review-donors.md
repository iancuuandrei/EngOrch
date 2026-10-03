# Multi-agent topology donor review

**Scope.** Read-only source review of five public repositories at the exact shallow-clone `HEAD` listed below. No donor code was run. This note records implementation mechanisms visible in those source snapshots; it does not qualify donor performance or Fabric behavior. Paths below are repository-relative. Links point to the reviewed commit, not a moving branch.

## Snapshot and license identity

| Donor | Reviewed source | Commit | License evidence at that commit |
|---|---|---|---|
| CoCoder | [Flitternie/CoCoder](https://github.com/Flitternie/CoCoder/tree/3ea363b9e2266260fa27c0df5c238c47ba41755e) | `3ea363b9e2266260fa27c0df5c238c47ba41755e` | Root `LICENSE`: Apache-2.0 |
| Pact | [zekariasasaminew/pact](https://github.com/zekariasasaminew/pact/tree/e45ca982be566e6138526899e6a7482d561ac42f) | `e45ca982be566e6138526899e6a7482d561ac42f` | Root `LICENSE`: MIT |
| Alibaba Open Code Review | [alibaba/open-code-review](https://github.com/alibaba/open-code-review/tree/a758d9cbfb689937c7857ad64b2dd66adb58c0c2) | `a758d9cbfb689937c7857ad64b2dd66adb58c0c2` | Root `LICENSE`: Apache-2.0 |
| pi-subagent-tasks | [harms-haus/pi-subagent-tasks](https://github.com/harms-haus/pi-subagent-tasks/tree/2bae805c1a0bdd97e699e2c9601fb4e6624f6f53) | `2bae805c1a0bdd97e699e2c9601fb4e6624f6f53` | Root `LICENSE`: MIT |
| CAID | [JiayiGeng/CAID](https://github.com/JiayiGeng/CAID/tree/f364d9e95727c2e2ba3dbf23f2d6de52c5f3d5fa) | `f364d9e95727c2e2ba3dbf23f2d6de52c5f3d5fa` | No root `LICENSE` file in this checkout; `pyproject.toml` declares MIT. Treat as metadata, not an independently verified license text. |

The requested CAID identity is JiayiGeng/CAID (“Centralized Asynchronous Isolated Delegation”), not another repository with “CAID” in its name. CoCoder’s checked repository is Flitternie/CoCoder; its README points readers at `CoCoder-Agent/CoCoder`, so do not silently substitute that upstream identity for the pin reviewed here.

## Findings by donor

### Agentless — hierarchical localization and patch validation

Reviewed [OpenAutoCoder/Agentless at `5ce5888b9f149beaace393957a55ea8ee46c9f71`](https://github.com/OpenAutoCoder/Agentless/tree/5ce5888b9f149beaace393957a55ea8ee46c9f71).
Its [README](https://github.com/OpenAutoCoder/Agentless/blob/5ce5888b9f149beaace393957a55ea8ee46c9f71/README.md)
describes file, class/function, then edit-location localization, candidate patch
sampling, and regression/reproduction tests for selection. The pinned
[license](https://github.com/OpenAutoCoder/Agentless/blob/5ce5888b9f149beaace393957a55ea8ee46c9f71/LICENSE)
is MIT. No donor code was executed or copied.

Fabric can use progressively narrower retrieval to spend context on relevant
symbols and retain distinct proposed patch identities. Test-based selection
must preserve the original acceptance contract, exact candidate binding and
independent review. Generated reproduction tests are supplementary evidence;
they cannot replace held-out acceptance or authorize repeated provider effects.
Qualify retrieval first against file/symbol fixtures, then compare equal-budget
coding tasks. Donor benchmark claims are not Fabric qualification.

### CoCoder — structural partitioning plus a dependency-aware file queue

**Mechanism.** `code_team/cohesionbase/tools/partition_into_groups.py` builds weighted file relationships from a repository information base (RIB), labels structural roles, clusters core files, lifts some independent siblings into separate groups, optionally merges small groups, then records groups and file ownership. Its `role_grouping` and `lift_independent` implementations are in `code_team/cohesionbase/partition/post_processing.py`; `merge_small_groups` only accepts a candidate merge if its zero-communication dependency makespan simulation does not increase. The composite tool writes each RIB file’s dependencies into a file-backed task list and starts a group agent with the files assigned to that group.

`code_team/cohesionbase/tools/shared_task_list.py` keeps per-file `pending/ready/in_progress/completed` state and changes readiness when dependencies complete. Claims are idempotent for the same agent and prevent a different agent from claiming an in-progress task. However, the module says any registered agent may claim any ready task: the computed `owner` is metadata, not an enforced write boundary. Its lock waits five seconds, then unlinks the lock and recreates it as a presumed stale lock without proving the original holder is dead. That is a concrete concurrency hazard under a slow but live holder.

**Transfer to Fabric.** Use the useful separation between (a) graph dependencies that determine readiness and (b) a task’s write owner. Retain the accepted graph as the authority for paths; never treat a clustering label, RIB edge, or advisory task owner as write permission. If introducing an optional partition heuristic, compare its suggested cohorts against a conservative dependency/coupling baseline and keep all file effects under Fabric’s existing candidate-bound aggregate/effect path. Do not port CoCoder’s lock-steal behavior.

**Qualification ideas.** Use graphs with a hub, independent siblings, transitive dependency, cycles, and a shared initializer/config file. Assert deterministic partitions; assert no two concurrently executable tasks have overlapping path ownership; verify dynamically that a newly ready dependent is admitted only after all declared prerequisites finish. Inject a slow live lock holder and prove no second claimant steals its lock. These are proposed Fabric checks, not evidence that the donor passes them.

Primary code: [partition tool](https://github.com/Flitternie/CoCoder/blob/3ea363b9e2266260fa27c0df5c238c47ba41755e/code_team/cohesionbase/tools/partition_into_groups.py), [post-processing](https://github.com/Flitternie/CoCoder/blob/3ea363b9e2266260fa27c0df5c238c47ba41755e/code_team/cohesionbase/partition/post_processing.py), [shared task state](https://github.com/Flitternie/CoCoder/blob/3ea363b9e2266260fa27c0df5c238c47ba41755e/code_team/cohesionbase/tools/shared_task_list.py).

### Pact — one-wave implementation with bounded resources and a serial integration gate

**Mechanism.** At this pin, `pact run` plans one wave of units. `crates/pact-core/src/run.rs` validates unique unit names, a maximum count, and disjoint declared files; its prompt requires briefs to list every file each unit will write and puts project-wide checks in one final verification list. Unit/file separation and effort balance remain planner-produced proposals, not proof of semantic independence. The MVP source comments explicitly say dependent waves are not yet executed by `run`.

The optional ACP path hosts lanes as sessions in one process and one shared tree; the alternative process runtime launches separate agent processes. The `README.md` and config code expose a maximum concurrency, minimum free-memory threshold, per-lane memory reserve, and launch staggering. This is a concrete resource-vector example: admission considers total concurrency and available-memory/reserve constraints rather than only “number of tasks.” It is not a guarantee that the memory estimate predicts a model’s actual peak.

The source establishes untouched-tree verification baselines before the lanes, waits for the wave, commits the combined tree, then runs verification. When verification needs repair, `run.rs` starts a single repair lane against the combined tree; repairs are serialized. The separate worktree command supports per-task branches and a serialized merge queue; `pact-vcs` includes moving-base checks, merge conflict handling, and per-merge vs final-only test-gate modes. The README currently labels the project experimental and not actively developed. Its benchmark numbers are a single author-reported 39-file task and should not be treated as portable performance evidence.

**Transfer to Fabric.** The strongest fit is the small-wave pattern: strict preflight for distinct scopes, immutable candidate binding for every member, bounded multidimensional admission, and exactly one existing aggregate effect after every member has a terminal recorded result. Keep review/native gates after the aggregate. Reuse resource accounting concepts only if Fabric can observe the relevant resources; otherwise expose a conservative cap rather than presenting a configured reserve as measured memory.

**Unsafe/advisory boundary.** Planner briefs and disjoint file lists cannot demonstrate that tasks do not share APIs, assumptions, or test fixtures. A model-proposed partition must not create authority. Do not import Pact’s shared-write implementation as-is into Fabric’s journaled effect model: shared filesystem writes weaken attribution and make member-level failures difficult to distinguish. Its worktree merge-helper and auto-resolution are also not proof that a conflict resolution is correct.

**Qualification ideas.** Include two actually independent and one subtly coupled plan; reject a shared-file collision before dispatch; verify resource admission is all-or-nothing across total/provider/model or memory dimensions; assert baseline checks run before any writer; require one aggregate effect and fresh combined verification/review; inject one failed or unknown member and prove no aggregate/retry. Test both `Each` and `Final` gate semantics where applicable and preserve exact source/plan/member identities through restart. These are Fabric test proposals, not donor test results.

Primary code: [plan validation and run orchestration](https://github.com/zekariasasaminew/pact/blob/e45ca982be566e6138526899e6a7482d561ac42f/crates/pact-core/src/run.rs), [VCS integration](https://github.com/zekariasasaminew/pact/blob/e45ca982be566e6138526899e6a7482d561ac42f/crates/pact-vcs/src/lib.rs), [README and documented resource/benchmark claims](https://github.com/zekariasasaminew/pact/blob/e45ca982be566e6138526899e6a7482d561ac42f/README.md).

### pi-subagent-tasks — durable task DAG, multiple concurrency dimensions, and ordered worktree composition

**Mechanism.** `src/run-tasks.ts` is the create/resume boundary; it validates dependencies and persists pool state. `src/scheduler.ts` resolves tasks whose dependencies are done, asks the compose cursor for currently ready atom demands, acquires resource slots before launching, and routes terminal agent results back through a completion path. `src/pools.ts` implements all-or-nothing AND admission for total, provider, and provider/model limits. The scheduler retains admission tokens to prevent late completion events from settling a later retry admission.

The explicit modes matter: `main-cwd` and `shared-worktree` can have concurrent writes in the same directory; the README says only `task-worktree` provides task isolation. In task-worktree mode, a task’s worktree is created lazily from the pool’s current HEAD after dependencies finish. `src/merge.ts` serializes task integrations through one FIFO merge worker (fast-forward, then rebase/merge conflict handling); thus independent tasks can run concurrently while branch integration remains serialized. The nested compose DSL also offers parallel/sequential atoms and a `gateLoop`; the latter parses a structured gate verdict and can rerun work after rejection. Pool state and task live events are persisted, and resume reconciles pool/worktree state.

**Transfer to Fabric.** This is the clearest scheduler reference for a DAG with explicit task lifecycle, durable progress, separately bounded resource pools, lazy dependent worktrees, and a serial merge/effect stage. Fabric can adapt the conceptual separation—parallel read/model work, serial candidate-bound filesystem effect, then gates—without importing its process/session semantics. A small state diagram and model-specific resource caps may help operations.

**Unsafe/advisory boundary.** A gate-loop reviewer verdict is generated by an agent; it is workflow input, not independent or native correctness evidence. The README’s own warning about shared-directory write races is material. Task prompts are shared with each atom, while file ownership comes from user-declared task specs; neither is a trusted enforcement boundary on its own. Preserve Fabric’s stricter UNKNOWN behavior rather than assuming a resumed process can infer a missing terminal result.

**Qualification ideas.** Test dependency admission, worktree base after parent merge, all-or-nothing resource caps, queued serial merges, duplicate/late completion tokens, crash/resume with a completed task and one unknown member, and review rejection not producing READY absent fresh native evidence. Exercise the three modes to verify the two shared-directory modes are never mistaken for isolated ownership. These are suggested Fabric tests, not donor run results.

Primary code: [scheduler](https://github.com/harms-haus/pi-subagent-tasks/blob/2bae805c1a0bdd97e699e2c9601fb4e6624f6f53/src/scheduler.ts), [resource pools](https://github.com/harms-haus/pi-subagent-tasks/blob/2bae805c1a0bdd97e699e2c9601fb4e6624f6f53/src/pools.ts), [merge worker](https://github.com/harms-haus/pi-subagent-tasks/blob/2bae805c1a0bdd97e699e2c9601fb4e6624f6f53/src/merge.ts), [create/resume integration](https://github.com/harms-haus/pi-subagent-tasks/blob/2bae805c1a0bdd97e699e2c9601fb4e6624f6f53/src/run-tasks.ts), [gate-loop parsing](https://github.com/harms-haus/pi-subagent-tasks/blob/2bae805c1a0bdd97e699e2c9601fb4e6624f6f53/src/gateloop.ts).

### Alibaba Open Code Review — review-oriented change grouping and bounded concurrent review

**Mechanism.** This repository’s grouping is for review, not implementation ownership. `internal/agent/grouping.go` chooses a deterministic partition strategy for small changes or asks an LLM to group larger diff sets; it falls back to per-file groups when the LLM grouping fails and applies a per-group token budget. `internal/agent/agent.go` dispatches Plan+Main review subtasks concurrently under `MaxConcurrency` (default 8), with a concurrent timeout and optional aggregate token budget. Comments are collected and filtered separately. Grouping uses changed files/diffs, and group-level comments remain review suggestions.

**Transfer to Fabric.** The useful idea is to split review context into bounded related groups, with deterministic fallback and explicit coverage/budget reporting. This can be a review-context optimization only; it must not decide implementation write owners or convert “no findings” into READY. Keep a mapping from every grouped item back to the original candidate paths and ensure every required path was reviewed.

**Unsafe/advisory boundary.** LLM-produced semantic groups and comments are advisory. A grouping failure’s per-file fallback improves coverage but may change cost and latency. A concurrent review budget can stop partway; any Fabric use must surface incomplete coverage instead of passing the gate. This repository’s own distinction between comment suggestions and native checks should remain explicit.

Primary code: [grouping and fallback](https://github.com/alibaba/open-code-review/blob/a758d9cbfb689937c7857ad64b2dd66adb58c0c2/internal/agent/grouping.go), [bounded review dispatch](https://github.com/alibaba/open-code-review/blob/a758d9cbfb689937c7857ad64b2dd66adb58c0c2/internal/agent/agent.go), [review finding data model](https://github.com/alibaba/open-code-review/blob/a758d9cbfb689937c7857ad64b2dd66adb58c0c2/internal/model/review.go).

### CAID — manager-delegated worktree lanes with round-based reassignment

**Mechanism.** `run_infer.py` and `core/subagent.py` run multiple engineer conversations concurrently. The manager’s `DelegationPlan` separates first-round tasks from remaining tasks. `core/manager.py` groups first-round tasks by engineer, records the common base commit, and creates one Git worktree per engineer. As a runner completes, the async pool can hand an idle engineer a remaining task; the manager records task/cost/timing information and can run background exploration while engineers work. A final manager review follows the engineer rounds.

**Transfer to Fabric.** CAID demonstrates the product value of assigning queued independent tasks to newly idle capacity and tracking per-role costs. Fabric’s accepted dependency graph and immutable repair ceiling should remain the source of legal task readiness and scope; use dynamic dispatch only among already validated ready nodes. Separate task-level outputs from a manager’s narrative summary and keep the exact model/native receipts.

**Unsafe/advisory boundary.** Delegation plans and later assignments are model-produced JSON; the inspected code builds them from event extraction and fallback heuristics rather than using a typed, graph-authoritative write-scope validator. In the Commit0 path, `manager.py` includes a `git merge -X theirs` fallback when an engineer has no rounds left. This can silently favor one branch in conflicting files and should not be copied into an evidence-bound integration path. Final manager review is not independent native verification. The repository has no root license text in the pin, so licensing should be verified before any code reuse.

**Qualification ideas.** Test manager assignment only from recorded, dependency-ready tasks; reject unknown IDs, duplicate assignment, already-running/finished engineer, and out-of-ceiling paths. Inject a merge conflict and ensure no “theirs” resolution is accepted without fresh verification and reviewer evidence. Verify async completion order does not change candidate identity, scopes, or the final result. These are proposed Fabric tests.

Primary code: [delegation result conversion](https://github.com/JiayiGeng/CAID/blob/f364d9e95727c2e2ba3dbf23f2d6de52c5f3d5fa/core/utils.py), [per-engineer worktree setup and merge policy](https://github.com/JiayiGeng/CAID/blob/f364d9e95727c2e2ba3dbf23f2d6de52c5f3d5fa/core/manager.py), [parallel runners and dynamic task assignment](https://github.com/JiayiGeng/CAID/blob/f364d9e95727c2e2ba3dbf23f2d6de52c5f3d5fa/core/subagent.py), [Commit0 task decomposition](https://github.com/JiayiGeng/CAID/blob/f364d9e95727c2e2ba3dbf23f2d6de52c5f3d5fa/tasks/commit0.py).

## Cross-donor synthesis for Fabric

The transferable unit is not “an agent” but a **validated ready task** with a frozen scope, candidate, and invocation identity. The code references support these complementary design ideas:

1. **Dependencies and coupling are separate inputs.** CoCoder’s RIB/community structure can suggest which files are related; Pact and CAID show planner-generated task briefs; pi has an explicit dependency DAG. Use such signals to propose task boundaries, but validate dependency readiness and path ownership deterministically before dispatch.
2. **Use bounded concurrent model work, not concurrent uncontrolled effects.** Pact’s lane/resource admission, pi’s all-or-nothing resource pools, OCR’s bounded review workers, and CAID’s per-engineer accounting illustrate independent ceilings. In Fabric, task proposals can run in parallel against one frozen candidate; effect integration should preserve the existing serial aggregate and fail closed for missing/unknown member results.
3. **Choose isolation to match the effect model.** Shared worktrees reduce branch-merge work but introduce write races (explicitly documented by pi and Pact). Per-task worktrees isolate writes but require ordered integration and base/head checks. Fabric’s candidate/effect journal needs to own this decision; neither a shared checkout nor a clean-looking diff is sufficient attribution.
4. **Review and native verification are distinct gates.** OCR and gate-loop systems make review a structured step, but model review remains advisory. A READY decision should require exact-candidate native checks and the required review result, with complete coverage and no unresolved/UNKNOWN work.
5. **Qualify effects, not just decomposition.** Compare serial and parallel runs on the same pinned repository/task, with separate immutable candidates and identical native/held-out checks. Record plan/task/member identities, scopes, resource admission, effect count, review/verification results, wall time, and incomplete or blocked outcomes. A timing comparison is interpretable only when task, model/effort, candidate/evidence requirements, environment, and gate coverage match; these donors do not establish Fabric speedup.

**Evidence boundary.** The mechanisms above were inspected in source at the listed pins. No donor program or test suite was executed during this review, and no Fabric run is claimed by this note. Any donor README or paper performance claim remains author-reported unless independently reproduced with a comparable workload.
