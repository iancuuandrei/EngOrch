---
name: code-review
description: Review a concrete candidate against its objective, existing behavior and executed checks. Use for independent review after implementation or repair.
metadata:
  fabric.roles: "reviewer"
---

# Code review

Trace each changed contract to callers, persisted readers and failure paths.
Compare the candidate with the base; check unrelated behavior for regressions.
Prioritize concrete correctness issues and cite the affected path and mechanism.
Separate recorded passing checks from coverage assumptions and proposed checks.
Report unresolved evidence as uncertainty rather than a successful result.

Use the supplied verdict schema and exact candidate/verification identities.
This workflow does not execute tests, edit files, approve effects or grant retry
authority. The controller admits the verdict against its unchanged gates.
