"""Deterministic, offline scheduling and evidence-stopping comparison.

This is a decision experiment, not a controller implementation or a formal
JEV calculation. It models finite synthetic DAG tasks; both policies dispatch
the same task set under the same worker bound. Safety and admission are not
modeled.
"""
import json
import random

SEED = 20261003
WORKERS = 2


def downstream(tasks, estimates):
    children = {t: [] for t in tasks}
    for t, spec in tasks.items():
        for dep in spec["deps"]:
            children[dep].append(t)
    memo = {}
    def visit(t):
        if t not in memo:
            memo[t] = estimates[t] + max([visit(c) for c in children[t]] or [0])
        return memo[t]
    return {t: visit(t) for t in tasks}


def unique_descendants(tasks):
    children = {t: [] for t in tasks}
    for t, spec in tasks.items():
        for dep in spec["deps"]:
            children[dep].append(t)
    def visit(t):
        seen, stack = set(), list(children[t])
        while stack:
            child = stack.pop()
            if child in seen:
                continue
            seen.add(child)
            stack.extend(children[child])
        return len(seen)
    return {t: visit(t) for t in tasks}


def simulate(tasks, estimates, actual, policy):
    remaining = set(tasks)
    done, active = set(), []
    now, dispatches, peak = 0, 0, 0
    starts = []
    lengths = downstream(tasks, estimates)
    descendants = unique_descendants(tasks)
    order = {task_id: index for index, task_id in enumerate(tasks)}
    def priority(task_id):
        if policy == "critical_path":
            return (-lengths[task_id], task_id)
        if policy == "production_unique_descendants":
            return (-descendants[task_id], order[task_id])
        return (0, task_id)
    while remaining or active:
        ready = [t for t in remaining if all(d in done for d in tasks[t]["deps"])]
        ready.sort(key=priority)
        while ready and len(active) < WORKERS:
            t = ready.pop(0)
            remaining.remove(t)
            active.append((now + actual[t], t))
            starts.append({"task": t, "time": now})
            dispatches += 1
            peak = max(peak, len(active))
            ready = [x for x in remaining if all(d in done for d in tasks[x]["deps"]) and x not in [a[1] for a in active]]
            ready.sort(key=priority)
        if not active:
            raise RuntimeError("stalled DAG")
        now = min(finish for finish, _ in active)
        completed = [t for finish, t in active if finish == now]
        done.update(completed)
        active = [(finish, t) for finish, t in active if finish != now]
    return {"makespan": now, "dispatch_count": dispatches, "peak_active": peak, "starts": starts,
            "outcome": "all_tasks_completed"}


def task(deps=()): return {"deps": list(deps)}


def synthetic_duration_error(rng):
    """Find a small valid seeded counterexample to critical-path priority."""
    for _ in range(10000):
        tasks = {
            "a": task(), "b": task(), "c": task(), "d": task(),
            "e": task(("a",)), "f": task(("b",)), "g": task(("c", "d")),
        }
        estimates = {k: rng.randint(1, 40) for k in tasks}
        actual = {k: rng.randint(1, 40) for k in tasks}
        base = simulate(tasks, estimates, actual, "production_unique_descendants")
        critical = simulate(tasks, estimates, actual, "critical_path")
        if critical["makespan"] > base["makespan"]:
            return tasks, estimates, actual, base, critical
    raise RuntimeError("fixed seed did not produce duration-error counterexample")


def evidence_case(name, observations):
    """Compare collect-all with a fixed expected-loss stopping rule.

    Each observation supplies an *assigned* expected loss reduction and cost,
    plus a seeded hidden decision effect. This is an executable heuristic
    comparison only: it does not estimate probabilities from production data
    and must not be called formal JEV.
    """
    all_cost = sum(x["cost"] for x in observations)
    selected = [x for x in observations if x["expected_loss_reduction"] > x["cost"]]
    cost = sum(x["cost"] for x in selected)
    expected_residual_loss = sum(x["expected_loss_reduction"] for x in observations if x not in selected)
    loss = sum(x["hidden_loss_if_skipped"] for x in observations if x not in selected and x["reveals_error"])
    return {"name": name, "policy": "stop when assigned expected loss reduction <= evidence cost",
            "all_evidence": {"evidence_cost": all_cost, "expected_decision_loss": 0, "decision_loss": 0, "dispatch_count": len(observations)},
            "stopping_policy": {"evidence_cost": cost, "expected_decision_loss": expected_residual_loss, "decision_loss": loss, "dispatch_count": len(selected),
                                "selected": [x["id"] for x in selected]},
            "outcome_equivalent": loss == 0,
            "note": "assigned values and hidden outcomes are deterministic synthetic inputs, not calibrated JEV"}


def main():
    rng = random.Random(SEED)
    # The first uses the shape of a three-node implementation/verification/
    # review graph. The second mirrors the repository's parallel-implementation
    # test shape. Durations are modeled values, not observed runtime data.
    cases = [
        ("three_node_parsebytes_task_shape", {"parsebytes-separators": task(), "native-verification": task(("parsebytes-separators",)), "native-review": task(("native-verification",))},
         {"parsebytes-separators": 1800, "native-verification": 300, "native-review": 300},
         {"parsebytes-separators": 1800, "native-verification": 300, "native-review": 300},
         "synthetic three-node implementation/verification/review shape; no local receipt dependency"),
        ("repository_parallel_implementation_test_shape", {"research": task(), "impl-api": task(("research",)), "impl-tests": task(("research",)), "verify": task(("impl-api", "impl-tests")), "review": task(("verify",))},
         {"research": 10, "impl-api": 10, "impl-tests": 10, "verify": 10, "review": 10},
         {"research": 10, "impl-api": 10, "impl-tests": 10, "verify": 10, "review": 10},
         "shape from internal/engineeringplan/parallel_implementation_test.go; no external runtime outcome claimed"),
        ("synthetic_critical_path_benefit", {"alpha": task(), "bravo": task(), "zulu": task(), "child": task(("zulu",))},
         {"alpha": 1, "bravo": 1, "zulu": 10, "child": 10}, {"alpha": 1, "bravo": 1, "zulu": 10, "child": 10},
         "same shape and expected 21 vs 20 result as engineeringplan schedule test"),
    ]
    results = []
    for name, tasks, estimates, actual, note in cases:
        production = simulate(tasks, estimates, actual, "production_unique_descendants")
        critical = simulate(tasks, estimates, actual, "critical_path")
        lexical = simulate(tasks, estimates, actual, "lexical_component_only")
        results.append({"name": name, "tasks": tasks, "estimates": estimates, "actual": actual,
                        "production_unique_descendants": production, "critical_path": critical,
                        "lexical_component_only": lexical,
                        "makespan_delta_critical_minus_production": critical["makespan"] - production["makespan"], "note": note})
    tasks, estimates, actual, production, critical = synthetic_duration_error(rng)
    lexical = simulate(tasks, estimates, actual, "lexical_component_only")
    results.append({"name": "synthetic_duration_error_counterexample", "tasks": tasks, "estimates": estimates, "actual": actual,
                    "production_unique_descendants": production, "critical_path": critical, "lexical_component_only": lexical,
                    "makespan_delta_critical_minus_production": critical["makespan"] - production["makespan"],
                    "note": "fixed-seed estimated-duration error negative case"})
    evidence = [
        evidence_case("safe_stop", [
            {"id":"compile","cost":5,"expected_loss_reduction":30,"reveals_error":True,"hidden_loss_if_skipped":30},
            {"id":"duplicate_read","cost":9,"expected_loss_reduction":2,"reveals_error":False,"hidden_loss_if_skipped":0},
        ]),
        evidence_case("unsafe_stop_negative", [
            {"id":"test","cost":5,"expected_loss_reduction":20,"reveals_error":True,"hidden_loss_if_skipped":20},
            {"id":"rare_regression","cost":9,"expected_loss_reduction":2,"reveals_error":True,"hidden_loss_if_skipped":100},
        ]),
    ]
    out = {"schema_version":1,"seed":SEED,"workers":WORKERS,
           "definitions":{"production_unique_descendants":"taskpool.Graph.Ready priority: most unique downstream dependents, then declaration order. The experiment models only static, nonparked tasks; taskscheduler.orderedCandidates additionally prioritizes parked work and enforces dynamic-agent FIFO, neither of which is modeled here.", "critical_path":"longest estimated downstream duration among ready tasks", "lexical_component_only":"engineeringplan Lexical comparison component; not claimed as production graphprovider priority", "evidence_stopping":"synthetic assigned expected-loss reduction versus evidence cost; not formal JEV"},
           "results":results,"evidence_stopping":evidence,
           "conclusion":"Against the actual nonparked taskpool unique-descendant priority, critical-path ties the retained graph shapes and the exact-duration synthetic case, then loses under a fixed-seed duration-error case. Lexical output remains a standalone component comparison only. Evidence stopping saves cost in one synthetic case but creates decision loss in another. Neither policy is adopted."}
    raw = json.dumps(out, indent=2, sort_keys=True) + "\n"
    print(raw, end="")

if __name__ == "__main__": main()
