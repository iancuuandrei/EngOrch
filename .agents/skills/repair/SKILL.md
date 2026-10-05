---
name: repair
description: Repair a candidate after a concrete failed check or requested review change. Use when supplied failure evidence and declared write ownership bound a correction.
metadata:
  fabric.roles: "fixer"
---

# Repair

Start from the supplied failed-gate evidence and current candidate. Distinguish
a source defect from an unavailable dependency or unresolved provider effect.
Find the smallest correction within the declared write paths; preserve public
behavior outside the requested change. Address the actual finding without
weakening tests, acceptance criteria or the role's output contract.

Use exact current file hashes and checked edits. Identify the regression that
would distinguish this failure from nearby valid behavior. Report only checks
that have recorded observations. Do not widen scope, reset evidence, repeat
uncertain effects or exceed the controller's repair budget.
