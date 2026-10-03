# Explicit fixer model allocation

`init` keeps the existing configuration bytes when both fixer flags are
omitted. To select an independent model for the existing fixer role, supply
either or both fixer overrides and a complete user-owned v2 access policy:

```powershell
fabric init --codex C:\tools\codex.exe --model gpt-6-luna `
  --fixer-model gpt-6-sol --fixer-effort high --access-config access.json
```

The fixer model and effort are independent of writer overrides. If only one
fixer flag is supplied, the other value comes from the base `--model` or
`--effort` option; it never borrows `--writer-model` or `--writer-effort`. Init
requires `--access-config` because configuring a fixer upgrades the generated
configuration to v2. The file is strict JSON for the existing `config.Access`
schema, limited to 32 KiB. Duplicate or unknown fields and trailing data are
rejected. Init validates the complete role and budget mapping before creating
runtime state or writing `harness.toml`.

For a Codex subscription route, an access file has this shape; choose the
repository class, auth mode, token ceilings, and concurrency for your own
policy:

```json
{
  "class": "PRIVATE",
  "limits": {"tokens": 300000, "cost_micro_usd": null, "concurrency": 3},
  "profiles": [{
    "version": 1,
    "name": "codex-session",
    "kind": "subscription",
    "runtime": "codex-app-server",
    "provider": "openai",
    "credential_ref": "",
    "auth_mode": "session",
    "repository_classes": ["PRIVATE"]
  }],
  "roles": {
    "planner": "codex-session",
    "explorer": "codex-session",
    "writer": "codex-session",
    "fixer": "codex-session",
    "reviewer": "codex-session"
  },
  "invocations": {
    "planner": {"tokens": 30000, "cost_micro_usd": null},
    "explorer": {"tokens": 30000, "cost_micro_usd": null},
    "writer": {"tokens": 80000, "cost_micro_usd": null},
    "fixer": {"tokens": 80000, "cost_micro_usd": null},
    "reviewer": {"tokens": 30000, "cost_micro_usd": null}
  }
}
```

These example values are operator choices, not generated defaults or measured
limits. Subscription cost remains unknown; `cost_micro_usd: null` does not mean
zero cost. The model name also does not establish capability or quality.

These ceilings are deliberately illustrative, not a qualified happy-path
budget. Input includes cached input, and a runtime invocation can consume
tokens across multiple tool iterations. The completed v1.0.9 serial examples
used 387,706 input tokens for Humanize and 1,338,255 for numeric-text across
their role invocations. Choose invocation and total ceilings from your own
workload evidence; the example's 300,000 total does not cover those runs.
See `../evaluation/v1.0.9-isolated-writers.md` for the measured token types.

The controller decides whether a repair requires the fixer role. These flags do
not add repair attempts, reset a run budget, or change the failure-escalation
policy. Without fixer flags, init leaves the fixer profile absent and preserves
the existing legacy configuration output.
