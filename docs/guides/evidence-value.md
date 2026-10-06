# Finite evidence value inspection

## Current scope

`fabric evidence-value` compares a bounded set of proposed information purchases
against a declared decision model and resource constraints. It emits an
**advisory recommendation**, or stops when no known positive estimated value
remains. It does not dispatch an action, modify policies, update budgets, append
journal events, change a working context or accept a candidate.

`fabric evidence-acquire` additionally applies the model to one optional bounded
explorer source selection. Autonomous research stopping in a live agent,
provider actions and empirical calibration remain pending.
JEV controls acquisition proposals; working context controls retention;
Fabric's existing controller owns authority and engineering acceptance.

## Inspect one exact snapshot

```powershell
fabric evidence-value RUN --template > C:\private-evidence\evidence-request.json
```

The template binds `run_id`, `source_id`, `candidate_id` and `journal_head` from
one replay-validated snapshot. Populate its `resources` and `actions` arrays with
estimates appropriate to the task, then inspect it:

```powershell
fabric evidence-value RUN C:\private-evidence\evidence-request.json
```

A changed journal head or source/candidate/run binding rejects the request.
Generate a fresh template after controller progress; never rewrite an existing
invocation or retry an uncertain effect to refresh this report. The result
includes the exact canonical request hash and snapshot binding. Archive the
request alongside the result if reconstructing its model later matters.

UNKNOWN removes the recommendation and reports `effect_requires_reconciliation`.
READY removes it because the accepted checkpoint needs no additional research.
An inactive lifecycle also removes it. These are inspection diagnostics, not
new admission, retry, settlement or acceptance permissions.

## Acquire selected source evidence

For a run with bounded task context and a resolved workspace/candidate, wrap a
populated `evidence-value --template` model in this request shape:

```json
{
  "model": {"version":1,"binding":{},"resources":[],"actions":[]},
  "queries": {"read-local-source":"Explain parser.go record boundary behavior"}
}
```

The empty binding/model above shows the shape only: use the exact generated
binding and declare the `read-local-source` action and its costs as below.
This command supports only `source_read` actions. Supply exactly one bounded
query per action identity; other kinds and missing/extra targets are rejected.
Then run:

```powershell
fabric evidence-acquire RUN C:\private-evidence\acquisition.json
fabric inspect RUN
```

The controller records `evidence.context-decided` with the complete model,
queries, request hash, decision identity and recomputable selection. A positive
known JEV enters existing explorer context admission: candidate capture, read
lease, sensitive-path filtering, existing file/read/selection bounds and
source/candidate-bound journal evidence. New admissions include
`evidence_decision_id`; preparation and replay recheck the admitted selected
query, controller stop state and exact decision-bound candidate. A concurrent
candidate or stop transition cannot admit a context under the old decision.
It may reuse a matching admitted
context. A subsequent explorer invocation for that exact query receives the
selected evidence; the command itself does not dispatch that invocation.
Existing optional RI localization may assist ordinary source selection.

No positive eligible action means decision evidence only, with no source
acquisition. UNKNOWN, READY or inactive lifecycle suppress both acquisition and
decision writes. Replay rejects stale producing prefixes, modified targets,
selection or decision hashes. Historical runs retain absent-field serialization;
no new execution policy is enabled automatically.

The request becomes stale after controller progress, including its own decision
append. If acquisition subsequently fails, the retained decision does not prove
that evidence was acquired. Inspect it, then prepare a fresh bound request for
any further read. Optional source reads cannot authorize retrying a provider
effect. Supplied costs remain estimates; existing read limits are enforced, but
this command does not claim observed token/money spending or calibrated prices.
Queries and source excerpts are private run evidence; avoid putting credentials
in them or publishing the raw journal.

## Model a finite information action

Allowed action kinds are `source_read`, `graph_query`, `targeted_test`, `coverage`,
`spawn_explorer`, `stronger_model` and `patch_candidate`. Names identify proposal
classes only; they are not tool calls or effect intents. Later execution still
requires every ordinary controller gate and exact action-specific admission.

For example, add this resource and action to the bound template:

```json
{
  "resources": [
    {"name":"wall_ms","limit":"100","used":"0","price":"1","gradient_squares":"1"}
  ],
  "actions": [{
    "id":"read-local-source",
    "kind":"source_read",
    "reliability":{"ordinal":2,"successes":0,"failures":0},
    "outcomes":[
      {"probability":"0.5","utilities":["1","0"]},
      {"probability":"0.5","utilities":["0","1"]}
    ],
    "costs":{"wall_ms":"10"}
  }]
}
```

This illustrative model has two possible observations and two subsequent
engineering decisions. Its finite information value is 0.5; supplied reliability
is 0.5; normalized shadow cost is 0.1; estimated JEV is approximately 0.15.
Those are model assumptions, not measured task success or provider prices.
This example declares no monetary constraint or monetary estimate, and makes
no monetary-cost claim. Every supplied cost must have a matching resource
constraint; undeclared costs are rejected rather than silently omitted.
If money is an active constraint, its unavailable cost makes the action
ineligible even when the shadow price is zero.

Numeric estimates use bounded strings because Fabric's canonical v1 JSON does
not permit floating-point numbers. Version, ordinal and observation counts stay
integers. Null cost is unavailable; a supplied numeric zero is a known estimate
of zero. Malformed, nonfinite, negative or excessive estimates are rejected.

## Calculation and resources

| Quantity | Meaning |
| --- | --- |
| Information value | Expected best decision utility after observing an outcome minus best expected utility beforehand |
| Reliability | Beta posterior mean with a fixed symmetric logistic ordinal prior and declared binary attribution counts |
| Shadow cost | Sum of price times action cost divided by the matching declared budget |
| Estimated JEV | Reliability times information value minus normalized shadow cost |

Every action must describe the same prior decision state: identical decision
dimensions and marginal expected utilities. Utilities/probabilities are bounded
to [0,1]; probabilities must sum to one. An action whose observation cannot
improve the decision has zero information value. Select the greatest strictly
positive JEV among resource-eligible actions; ties use lexical action identity.
Unknown active costs and exhausted/insufficient budgets exclude actions before
selection. An empty frontier stops safely.

Use distinct resource axes: `uncached_input_tokens`, `cached_input_tokens`,
`ordinary_output_tokens`, `reasoning_output_tokens`, `wall_ms`,
`money_minor_units`, `local_compute_ms`, `human_interventions`, `runtime_slots`.
Uncached input equals input minus cached input; ordinary output equals output
minus reasoning output. Do not charge those subsets again as total input/output.
If a category is unavailable, retain null rather than inventing a zero.
The report identifies unknown active costs. Supplied costs without a declared
constraint are rejected, whether known or unavailable.
Its scalar shadow cost is not a monetary-cost report.

Bounds: 64 KiB request, 16 actions, nine constraints, 16 outcomes per action,
16 decision utilities per outcome, 64 characters per numeric string, numerical
estimates at most 1e12, budgets at least one unit and one million total binary
reliability observations. Pure calculation never spends the declared budget.

## Calibration and adaptive prices

The ordinal prior is `p = logistic(ordinal - 2)`, with ordinal 0–4 and two
equivalent observations: `alpha=2p`, `beta=2(1-p)`. Declared successes/failures
update the posterior mean. These initial priors are not empirically calibrated.
Counts need objective downstream attribution before a consumer may treat them
as learned reliability; this CLI explicitly labels its input model unvalidated.

The pure `UpdatePrice` primitive uses projected scalar AdaGrad:
normalize observed usage by its explicitly supplied observation allocation,
subtract one, accumulate squared gradients, update price by gradient divided by
the square root of accumulated squared gradients, and project to nonnegative.
Zero gradient leaves state unchanged. It does not alter used resources or limit.
Prices may start at zero; hard constraints and unknown-cost exclusion still
apply. Runtime observation ingestion and persisted automatic updates are pending.

This independently implemented mechanism draws on
[finite value-of-information decision analysis](https://arxiv.org/abs/1703.08994)
and [AdaGrad](https://www.jmlr.org/papers/v12/duchi11a.html).
No donor source or serving/training stack is imported. See the
[normative contract](../specifications/evidence-value.md) and
[executed checks](../evaluation/evidence-value.md).
