# Scheduling and evidence decision experiment

This is a deterministic, offline comparison of scheduling heuristics and a
separate synthetic evidence-stopping rule. It does not change production
scheduling and is not a formal JEV calculation.

## Run it

From this directory, run:

```text
python experiment.py > PATH_OUTSIDE_THE_CHECKOUT/receipt.json
```

The script uses seed `20261003` and two workers, prints the JSON receipt to
standard output, and does not write files. Use shell redirection to save the
receipt outside the checkout. The receipt contains only the experiment inputs
and results, not machine-specific paths.

## What it models

The production baseline is based on the current code:

- `internal/taskpool/dag.go`, `Graph.Ready`, orders ready tasks by descending
  count of unique downstream dependents, preserving declaration order for
  ties.
- `internal/taskscheduler/scheduler.go`, `orderedCandidates`, consumes that
  order, applies dynamic-agent FIFO filtering, then prioritizes parked tasks.

The experiment models static, nonparked DAG tasks only. Therefore it matches
the task-pool ready priority for those tasks, but does not model parked-task
preference, dynamic-agent FIFO, controller admission, uncertainty, or external
effects. The comparison policy uses longest estimated downstream duration;
the duration-error example gives estimates and actual durations separately.
The lexical comparator is only a standalone component comparison, not the
production graph-provider priority.

The three-node implementation/verification/review shape and the parallel
implementation test shape are synthetic inputs. Their durations are modeled,
not measured runtime values. The fixed-seed counterexample is also synthetic.
No external receipt or local absolute path is required to run the script.

The evidence-stopping cases assign expected loss reduction, evidence cost, and
hidden outcomes directly. They show one safe stopping example and one negative
case where stopping skips a decision-relevant test. These values are not
calibrated production data and must not be described as observed benefit or
formal JEV.

## Result interpretation

The retained shapes tie under critical-path and actual nonparked ready-task
priority. In the fixed-seed duration-error case, critical-path priority has a
longer modeled makespan. Evidence stopping reduces modeled cost in one case
with equivalent outcome, but incurs decision loss in the negative case. These
results provide no basis to adopt either heuristic in production.
