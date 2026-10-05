# Engineering skills

Reusable workflows live under `.agents/skills/`. Fabric captures the committed
catalog for new autonomous runs, filters discovery by role and loads bodies
only for selected tasks. See [agent context resolution](agent-context.md) for
scope, source identity, replay, bounds and the opt-out.

| Skill | Roles | Method |
| --- | --- | --- |
| [architecture](../../.agents/skills/architecture/SKILL.md) | planner, explorer, reviewer | Boundaries, interfaces and consequential tradeoffs |
| [codebase-trace](../../.agents/skills/codebase-trace/SKILL.md) | explorer, writer, fixer | Evidence-linked implementation trace |
| [decision-rationale](../../.agents/skills/decision-rationale/SKILL.md) | planner, explorer, reviewer | Recorded reasons, alternatives and hypotheses |
| [requirement-analysis](../../.agents/skills/requirement-analysis/SKILL.md) | planner, explorer | Consequential ambiguity and acceptance criteria |
| [blast-radius](../../.agents/skills/blast-radius/SKILL.md) | planner, explorer, fixer, reviewer | Consumers, regressions and uncovered scope |
| [tdd](../../.agents/skills/tdd/SKILL.md) | writer, fixer | Discriminating regression and correction workflow |
| [verification-design](../../.agents/skills/verification-design/SKILL.md) | planner, reviewer | Claim-to-evidence matrix and independent oracles |
| [code-review](../../.agents/skills/code-review/SKILL.md) | reviewer | Candidate and compatibility review |
| [repair](../../.agents/skills/repair/SKILL.md) | fixer | Bounded correction from failed-gate evidence |

The seven earlier portable procedures were relocated and their vague names
clarified; their historical [evaluation plan](../evaluation/procedures.md)
retains its original names and NOT RUN model scores. The two new workflows do
not inherit behavioral qualification from those records. Standard packaging
and Fabric prompt inclusion do not prove selection accuracy or provider quality.

No skill grants permissions, credentials, tool capabilities or effect authority.
Use host-native discovery only according to that host's documented behavior.
