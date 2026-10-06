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

## v1.1.25 advisory reuse for later roles

An optional finite acquisition now also informs later bounded task-context
selection. Selected source paths from a successful decision-linked explorer
admission act as advisory hints for subsequent writer, fixer and reviewer
selection, within the existing read, file and byte bounds. Only the exact
current source, candidate and selected query contribute; stale, wrong-role,
missing, malformed and unavailable entries are ignored. A reused historical
context without a decision link can match one validated explicit decision for
the same identities and query. This is operator-triggered acquisition only:
automatic EVC policy and measured quality or token benefit remain pending.
Runs without EVC decisions behave as before.

## v1.1.26 automatic acquisition before the initial writer

A new serial graph run can opt into one automatic finite acquisition before
the first writer's mandatory context:

```powershell
fabric run --autonomous --evidence-policy C:\private-evidence\evc-policy.json "Implement the scoped change"
fabric run --autonomous --evidence-policy C:\private-evidence\evc-policy.json --file goal.md
fabric run --autonomous --inspect-plan --evidence-policy C:\private-evidence\evc-policy.json
```

The policy file is version 1 with an empty binding, a finite `WireRequest`
model and exactly one bounded query per `source_read` action. The CLI validates
it before creating the run; incompatible parallel or isolated writers are
rejected rather than silently dropped. The frozen template is filled with the
live run, source, candidate and journal head at decision time, evaluated for
the highest eligible positive JEV, acquired through the existing explorer
bounds, then followed by ordinary mandatory writer context. A no-positive or
unknown-cost outcome still proceeds with writer context. The single tagged
decision is inspectable, never repeated on resume, and never claims empirical
utility, calibration or efficiency. Live quality and efficiency benefit remain
**NOT RUN**; v2 remains planned, not complete. Resume never changes the frozen
policy.

Usable version 1 template with an empty binding, bounded numeric-string
resource/action and an exact query map:

```json
{
  "version": 1,
  "model": {
    "version": 1,
    "binding": {"run_id": "", "source_id": "", "candidate_id": "", "journal_head": ""},
    "resources": [
      {"name": "local_compute_ms", "limit": "100", "used": "0", "price": "1", "gradient_squares": "1"}
    ],
    "actions": [
      {
        "id": "inspect-source",
        "kind": "source_read",
        "reliability": {"ordinal": 2, "successes": 0, "failures": 0},
        "outcomes": [
          {"probability": "0.5", "utilities": ["1", "0"]},
          {"probability": "0.5", "utilities": ["0", "1"]}
        ],
        "costs": {"local_compute_ms": "1"}
      }
    ]
  },
  "queries": {"inspect-source": "Explain file.txt base implementation"}
}
```

The first hard acquisition failure after the decision append exposes the
retained decision identity and its safe diagnostic without retrying the
uncertain source read. A subsequent resume observes the retained tagged
decision, skips automatic acquisition and admits ordinary required writer
context.

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
apply. `evidence-feedback` now connects this primitive to receipt-verified typed
token observations. Automatic persisted resource policy remains pending.

### Apply observed token feedback

```powershell
fabric evidence-feedback RUN C:\private-evidence\feedback-request.json
```

The request contains `model` (a populated snapshot-bound `evidence-value`
template), `allocations` (a numeric-string per-observation allocation for every
declared resource), and `consumed` (initially an empty object). For example,
`"allocations":{"uncached_input_tokens":"10000"}` is valid only when that is
the model's sole resource. Allocations and limits are explicit operator inputs;
the command does not learn an optimal budget or monetary conversion.

The report includes an exact usage-basis hash, token observations and
`feedback.request`, which carries the updated model and per-resource cursor.
Save that complete request for the next update; refresh its model binding after
controller progress. Use its `model` for the next `evidence-value` or wrap it
with source queries for `evidence-acquire`. This is optional feedback, not a
default autonomous policy.

Only journal-present, receipt-matched, completed Codex invocations with no
pending tool calls and clean OBSERVED typed accounting contribute. Costs use
the accounting delta, split into uncached/cached input and ordinary/reasoning
output. Each axis/invocation is applied once. An unchanged repeated cursor makes
no price or usage update; unknown axes remain unconsumed. Limits are preserved.
The baseline `used`, prices and accumulated gradients remain caller-supplied
advisory state. They are not validated cumulative accounting or controller
budget authority; resetting a cursor can repeat advisory arithmetic and does
not refund or authorize any real expenditure.

Money, wall time, local compute, interventions and slots are not derived from
token counts. A skipped resource retains its declared baseline, including a
baseline zero; that is not an observed zero cost. Missing typed accounting or
non-Codex routes do not establish free work. This projection is limited to 128
invocations, nine axes and a 64 KiB request; exceeding the limit rejects this
optional operation without changing the run. Invalid/unknown/duplicate cursors
and nonfinite allocations are rejected. Fake invocations have no provider
observation. The command is read-only and never dispatches or settles effects.

This independently implemented mechanism draws on
[finite value-of-information decision analysis](https://arxiv.org/abs/1703.08994)
and [AdaGrad](https://www.jmlr.org/papers/v12/duchi11a.html).
No donor source or serving/training stack is imported. See the
[normative contract](../specifications/evidence-value.md) and
[executed checks](../evaluation/evidence-value.md).
