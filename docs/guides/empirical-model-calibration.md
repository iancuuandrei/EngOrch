# Empirical fixer allocation

This optional feature compares complete task outcomes under two explicitly
configured fixer profiles. It retains static escalation and conservative
fallback. It has no qualified default model change or generalization claim.

Use `fabric calibrate-models CALIBRATION_JSON` to inspect a recommendation
without dispatching a model. Input is strict JSON, at most 128 KiB, with bounded
rows and profiles. Duplicate/unknown JSON fields and trailing content reject.
The artifact declares a family, baseline/candidate profiles, a minimum per arm,
an exact objective-hash scope, and disjoint training and holdout rows.

Each row identifies its source, task, run and objective, assigned fixer profile,
all observed fixer profiles, fixer invocation count, whole-task outcome, repairs
and shared non-fixer policy digest. Match source/task/objective and the shared
policy digest across arms. The shared policy must include the frozen runtime,
other role routes, task objective/source/acceptance, budgets and execution options,
excluding only the varied fixer profile. Digests identify supplied observations;
they do not attest that an operator's claims are true. Preserve the underlying
frozen run evidence and validate its provenance before using the artifact.

UNKNOWN, NOT_EXERCISED, profile drift, unmatched arms, insufficient rows and
missing improvement retain the baseline. Repairs are part of the whole-task
outcome. A completed individual model invocation is not a semantic PASS.
Training and holdout acceptance rates must both strictly improve before a
candidate is recommended. An operator-declared minimum of one is a bounded
observation, not statistical confidence. Historically seen tasks withheld from
fitting must not be described as globally unseen.

Typed input, cached input, output and reasoning are separate optional fields.
Cached input is a subset of input; reasoning is a subset of output. Missing
cost stays unknown. Version 1 selection is based on quality; it makes no
cheaper-model, speedup or billing claim from absent costs or profile estimates.

## Version 2 reliability-constrained cheaper fixer (opt-in)

Calibration `version` 2 is an explicitly opted-in immutable variant for
recommending a strictly cheaper fixer. It reuses the exact version 1 row
domain: bounded disjoint matched whole-task training/holdout rows with shared
source/task/objective/shared-policy bindings, the same profile and objective
validation, and the same conservative UNKNOWN, NOT_EXERCISED, drift,
unmatched-arm and minimum-per-arm fallbacks. Version 1 semantics, reasons and
digests are unchanged; existing evidence stays legacy compatible.

Version 2 adds two frozen integer parts-per-million parameters, both required
including an explicit zero epsilon:

- `epsilon_ppm` in `[0, 1000000]`: explicitly allowed quality degradation.
- `quality_floor_ppm` in `[0, 1000000]`: absolute candidate quality floor.

Posterior and bound: uniform Beta(1,1) prior. Each ACCEPTED whole-task outcome
that exercised the fixer adds one to alpha; each exercised semantic FAILED
adds one to beta. UNKNOWN, NOT_EXERCISED and drifted rows never update the
posterior; they force fallback and are never treated as FAIL or dropped. The
main selection drift exclusion (any observed fixer profile differing from the
exact assignment) and the fallback diagnostic drift exclusion are identical:
early-fallback diagnostics use the same filtered posterior counts, while the
canonical counts retain every row. The
per-arm per-cohort lower bound is the deterministic 5% lower quantile (fixed
95% confidence) of that posterior, quantified in integer ppm with conservative
floor rounding and deterministic replay. This is a conditional Bayesian bound
under the stated prior and within-cohort exchangeability, not a frequentist
proof, and it makes no generalization claim beyond the measured cohorts. The
implementation uses a bounded binary search over the integer-parameter
binomial form of the regularized incomplete beta function with no new
dependencies.

Requirement: for each cohort separately, `required = max(0,
baseline_lower_ppm - epsilon_ppm)`. The candidate lower bound must meet both
`required` and `quality_floor_ppm` in training AND holdout, in addition to the
operator minimum-per-arm threshold. The candidate does not need strict
observed improvement when it is admissibly cheaper; version 1 strict
improvement is unchanged.

Cost: every row in both arms and cohorts must carry a complete measured
whole-task `cost_micro_usd`. Missing cost falls back as unknown; profile cost
estimates are never used. Cached-input and reasoning subset accounting is
preserved by row validation. Costs are canonical 53-bit integers: noncanonical
values reject before any sum. All cost sums use overflow-checked int64
arithmetic; overflow falls back conservatively. Admitted calibrations hold at
most 64 rows, so 64 maximum-canonical costs cannot overflow int64; the checked
sum is defense-in-depth for helper robustness. The candidate mean cost must
be strictly lower than the baseline mean in BOTH cohorts; exact ties keep the
baseline. Output-only cost means with `costs_complete: false` are unknown,
not a measured zero; a true measured zero remains `0` with `costs_complete:
true`.

Routing and replay reuse the existing immutable configuration, digest and
evidence seam: `calibrate-models` reports the version 2 recommendation,
`init` embeds the frozen artifact, calibrated routing substitutes the
default-configured fixer only when the objective is in scope and static
risk, prior-failure and context escalation did not trigger, and replay
rederives the exact decision from frozen configuration and objective. Static
escalation still wins. Version 2 output may include output-only
posterior/quality/cost diagnostics; they never grant authority beyond the
frozen policy decision. There is no default promotion and no
measurement, quality or efficiency claim beyond the fixtures.

Embed the complete policy once during initialization:

```powershell
fabric init --codex C:\tools\codex.exe --model gpt-6-luna --effort high `
  --fixer-model gpt-6-luna --fixer-effort high --access-config access.json `
  --model-policy model-policy.json
```

The JSON contains the existing model policy profiles/rules and optional embedded
calibration. Configuration validation binds its baseline and candidate to the
fixer rule and configured runtime/provider.

The official matched-task runner supports this treatment with `-ModelPolicyPath`
on both `Prepare` and `Evaluate`. It requires the explicit fixer/access options,
binds the absolute policy path and file hash/size in the prepared receipt,
rechecks the file before initialization, and verifies that inspected creation
config contains the same policy. Omitting the option keeps the runner's prior
init arguments and receipt shape.

Keep the static escalation profile separate from the empirical candidate:
the policy has a configured baseline, calibrated alternative and static frontier
route. Static escalation thresholds remain explicit operator choices.

The resulting `harness.toml` contains
the complete policy; later execution and replay never reread the input file.
Without the new flag, initialization preserves its existing behavior.

Calibrated routing requires explicit decision-evidence version 2. The immutable
objective scope applies only to the fixer. Existing high-context/prior-failure
static escalation takes priority; other roles retain their existing selection.
Persisted evidence binds the frozen config, objective hash, calibration digest,
selected profile, reason and bounded cohort counts. Replay rederives that choice
from the frozen configuration and original objective rather than live history.

Real training/holdout trials exercised the mechanism, but UNKNOWN outcomes
left the empirical candidate unselected. An independently reviewed unseen
Wordwrap task passed with the baseline fixer and the recorded
`calibration-objective-out-of-scope` fallback. This demonstrates conservative
fallback, without a routing improvement, price estimate or default promotion.
See the [current product checkpoints](../roadmap/v2-product-plan.md) for the
source identities, acceptance evidence and remaining release gates.
