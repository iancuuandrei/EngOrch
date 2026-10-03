# Configured model routing

Fabric can select an already configured profile for a role. The selection is
mechanical and remains bound to the configuration, runtime invocation, and
access route. It never grants an additional dispatch or resolves an UNKNOWN
provider outcome.

## Tested TOML shape

`model_policy` is optional. Its nested fields use CamelCase because they are
Go fields without TOML tags. JSON names such as `default_profile` are rejected.

```toml
[model_policy]
Version = 1

[[model_policy.Profiles]]
Name = "explorer-economy"
Runtime = "fake"
Provider = "deterministic"
Model = "explorer-economy"
Effort = "low"

[[model_policy.Profiles]]
Name = "explorer-standard"
Runtime = "fake"
Provider = "deterministic"
Model = "explorer-standard"
Effort = "medium"

[[model_policy.Profiles]]
Name = "explorer-strong"
Runtime = "fake"
Provider = "deterministic"
Model = "explorer-strong"
Effort = "high"

[model_policy.Rules.explorer]
DefaultProfile = "explorer-standard"
CheapProfile = "explorer-economy"
CheapContextBytes = 2048
EscalatedProfile = "explorer-strong"
ContextEscalationTokens = 1000000000
ContextEscalationBytes = 16384
FailureEscalationCount = 1
```

The example uses fake routes for its parser regression. Replace every profile
with an existing admitted profile for that role. A default profile must exactly
match the role's configured route; alternatives keep its runtime and provider.
Only the explorer role can use a cheap profile.

## Current inputs

Fabric currently routes from the role, the exact final UTF-8 input byte count,
and accepted prior failures for the same task or current candidate. Small
read-only explorer inputs may choose the cheap profile. Large input or the
configured accepted-failure count can select the escalated profile. UNKNOWN
outcomes are not counted as failures.

Semantic complexity, risk, and uncertainty currently remain `medium`
placeholders. The policy has no measured quality, cost, or latency improvement
claim.
