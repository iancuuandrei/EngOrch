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
cost stays unknown. Selection is based on quality; it makes no cheaper-model,
speedup or billing claim from absent costs or profile estimates.

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

The implementation and fixtures qualify this mechanism only. Real matched
training/holdout trials and independent review are still required before any
measured routing improvement or release qualification is claimed.
