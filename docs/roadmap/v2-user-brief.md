Da. După ce am refăcut auditul față de **Fabric-ul actual, `de61e6f`**, aș modifica direcția mai mult decât propusesem inițial.

**Nu aș face următorul milestone doar „better context”.**  
Aș construi o fundație unificată care leagă **codebase intelligence + caching + context compilation + decomposition + scheduling**.

Asta cred că este direcția în care Fabric devine cu adevărat bun, nu doar un harness cu multe features.

Programul activ chiar cere genul acesta de abordare: capability-first, iar după primul gate prioritățile sunt scheduling/parallel work, semantic/context routing, codebase intelligence, model allocation, verification și abia apoi JEV dacă există beneficiu măsurat. FABRIC_V1_PRODUCT_PROGRAM

# 1. Concluzia donor audit-ului

Donorii se împart acum destul de natural:

| Zonă | Ce aș folosi ca inspirație principală |
|---|---|
| Codebase intelligence | **SylphxAI/repomap + SCIP + Serena + Aider** |
| Decomposition | **Co-Coder + Agentless** |
| Parallel work / integration | **Pact + pi-subagent-tasks + CAID** |
| Review | **Alibaba OpenCodeReview** |
| Route/model efficiency | **Ultraswarm**, ulterior Agent-as-a-Router |
| Long horizon | **LongHorizon-Harness** |
| Caching / incremental work | **Bazel + Nix**, adaptate |
| Service architecture | **Symphony** |
| Runtime/sandbox | OpenHands / SWE-ReX, ulterior |
| Compaction | runtime-native + idei OpenHands/AutoCompact, nu memorie proprie Fabric |

Și sunt destul de multe lucruri pe care **nu** le-aș importa.

---

# 2. Cel mai interesant donor nou: `SylphxAI/repomap`

Aici mi-aș concentra mult atenția.

Repomap este Rust, read-only, local și construiește:

- tree-sitter AST;
- import graph;
- call graph;
- PageRank;
- Louvain communities;
- BM25;
- un mic model local de embeddings;
- call paths;
- impact analysis;
- tests potentially affected;
- incremental per-file indexing.

Exact genul de informație pe care Fabric ar trebui să o poată furniza plannerului. [GitHub](https://github.com/SylphxAI/repomap?utm_source=chatgpt.com)

Dar **nu aș pune repomap lângă RI și să avem două sisteme**.

Fabric are deja:

```text
tgrep
+
SCIP
+
Graphify structural facts
+
candidate lexical overlays
```

Mai degrabă extragem mecanismele care ne lipsesc:

```text
resolved imports
call relationships
module/community detection
centrality
change impact
test reachability
incremental parse cache
hybrid ranking
```

și le construim peste infrastructura RI existentă.

Asta ar produce ceva mai interesant decât repomap însuși, pentru că Fabric poate lega fiecare fact de:

```text
source identity
candidate identity
producer
coverage
freshness
provenance
```

---

# 3. Serena: API-ul semantic este foarte bine gândit

Serena oferă agentului operații semantice simple:

```text
find_symbol
symbol_overview
find_references
find_declaration
find_implementations
type_hierarchy
```

în loc să îl lase să facă `grep` + read file la infinit. [GitHub](https://github.com/oraios/serena?utm_source=chatgpt.com)

Aș copia **vocabularul mental**, nu implementarea Serena.

Fabric ar putea oferi:

```text
ri.symbol
ri.references
ri.callers
ri.callees
ri.implementations
ri.impact
ri.tests
ri.module
ri.path
```

implementate de RI-ul nostru.

Asta reduce enorm tool churn-ul agentului.

---

# 4. Aider rămâne interesant pentru „orientation pack”

Aider RepoMap nu este suficient pentru Fabric ca backend, dar ideea este foarte bună:

> în loc să dai modelului repository-ul, îi dai un map compact, rank-uit, în limita unui token budget.

Fabric ar trebui să facă asta cu date mult mai bune decât Aider:

```text
RepositoryEvidence
    ↓
task/query
    ↓
Context Compiler
    ↓
bounded orientation pack
```

Planner-ul ar putea primi:

```text
Likely modules
Relevant symbols
Relevant files
Call relationships
Shared dependencies
Generated-code relations
Candidate changes
Tests likely affected
Coverage gaps
```

Nu 50.000 de linii de source.

---

# 5. Aici apare primul flaw concret din Fabric actual

În `de61e6f`, bounded task context este disponibil pentru:

```go
explorer
writer
fixer
reviewer
```

dar **nu pentru planner**.

Asta este invers față de unde avem probabil cel mai mare leverage.

Planner-ul este cel care decide:

```text
ce trebuie investigat
ce depinde de ce
ce se poate paraleliza
cine deține ce
```

Dacă el ia decizia cu information slabă și apoi exploratorii primesc context excelent, deja am construit DAG-ul greșit.

Generated-code failure-ul nostru este un exemplu perfect:

```text
Int64 change ─┐
              ├── shared generator/template
Uint64 change ┘
```

planner-ul le-a tratat ca independente.

Scheduler-ul nu putea repara asta.

---

# 6. Co-Coder schimbă puțin cum văd parallelism-ul

Aici donor audit-ul mi-a schimbat cel mai mult opinia.

Co-Coder tratează codebase-ul și taskurile ca un graph și încearcă să evite distribuirea naivă a taskurilor:

- identifică hub-uri structurale;
- izolează/shared dependencies;
- folosește community detection;
- construiește task dependencies;
- scheduler-ul ține cont de această structură.

Autorii raportează atât creșteri de task-pass cât și speed/cost improvements, dar evident trebuie replicate în Fabric, nu luate ca adevăr pentru produsul nostru. [GitHub](https://github.com/Flitternie/CoCoder)

Ideea care contează este:

> **Parallelize low-coupling clusters; serialize shared hubs.**

Mult mai bun decât:

> „fișiere diferite = taskuri independente”.

---

# 7. Pact ne arată cum ar trebui să evolueze writer parallelism

Fabric acum poate avea doi writers pentru ownership disjoint și apoi un aggregate file effect.

Este bun pentru schimbări relativ independente.

Dar pentru taskuri mari aș merge spre:

```text
task cluster A → worktree A
task cluster B → worktree B
task cluster C → worktree C

                    ↓

             Integration Candidate

                    ↓
         combined verification/review
```

Exact aici Pact este foarte interesant:

- worktree real per agent;
- integration branch;
- merge sequencing după risc;
- structural merge pentru anumite manifest files;
- merge failure nu distruge restul batch-ului;
- fiecare merge poate fi test-gated;
- conflict resolution poate fi agentic, dar este acceptată numai dacă testele trec. [GitHub](https://github.com/zekariasasaminew/pact?utm_source=chatgpt.com)

Fabric poate face asta mai strict datorită candidate/effect identities.

---

# 8. `pi-subagent-tasks`: o idee importantă pentru resource scheduling

Nu DSL-ul lor mă interesează.

Ce merită copiat conceptual este faptul că concurrency poate fi limitat simultan după mai multe resurse:

```text
total workers
AND provider capacity
AND model capacity
AND maybe toolchain/CPU capacity
```

Fabric ar putea ajunge la:

```text
ResourceVector {
    total_agents
    model_slots
    provider_slots
    writer_slots
    cpu_weight
    memory_weight
    verification_slots
}
```

Scheduler-ul admite taskul numai dacă vectorul încape.

Asta este mult mai realist decât un simplu:

```text
MaxParallel = 4
```

---

# 9. OpenCodeReview este cel mai bun donor pentru Review v2

Alibaba OpenCodeReview are o combinație foarte sănătoasă:

```text
deterministic filtering
        ↓
semantic file grouping
        ↓
bounded parallel reviews
        ↓
structured findings
```

Semantic grouping-ul folosește metadata despre fișiere și limitează grupurile; dacă gruparea LLM eșuează, revine la one-file-per-group. [GitHub](https://github.com/alibaba/open-code-review/blob/main/pages/src/content/docs/en/architecture.md?utm_source=chatgpt.com)

Mai are și review checkpoints pentru a evita rereview-ul întregului diff pe fiecare schimbare. [GitHub](https://github.com/alibaba/open-code-review/blob/main/examples/github_actions/README.md?utm_source=chatgpt.com)

Pentru Fabric:

```text
changed candidate
      ↓
impact graph
      ↓
review groups
 ┌────┼─────┐
 R1   R2    R3
 └────┼─────┘
      ↓
cross-cutting reviewer
```

Aș prefera asta eventual față de un singur reviewer care primește tot.

---

# 10. LongHorizon-Harness are probabil ideea corectă pentru taskuri de 5–20 ore

Nu aș încerca să ținem aceeași conversație agentică în viață la nesfârșit.

LongHorizon face:

```text
goal
 ↓
fresh context
 ↓
one bounded action
 ↓
independent verification
 ↓
verified checkpoint
 ↓
fresh context
 ↓
next step
```

Numai progresul verificat intră în state-ul trusted. [GitHub](https://github.com/AMAP-ML/LongHorizon-Harness?utm_source=chatgpt.com)

Asta se potrivește foarte bine cu Fabric.

Fabric deja are avantajul că ține:

```text
objective
candidate
task DAG
effects
verification
review
observations
```

în afara conversației LLM.

Prin urmare:

> **Fabric should externalize intelligence state, not conversation state.**

Și asta înseamnă că putem folosi contexte fresh agresiv.

---

# 11. OpenHands ne confirmă că nu trebuie să inventăm „Fabric Memory”

OpenHands are un sistem explicit de condensers care produce views reduse ale history-ului. [GitHub](https://github.com/OpenHands/docs/blob/main/sdk/arch/sdk.mdx?utm_source=chatgpt.com)

E util ca inspirație, dar eu aș menține regula veche Fabric:

```text
LLM summary != truth
```

Truth:

```text
journal
candidate
task graph
RI
verification
effects
```

Un summary poate ajuta modelul, dar nu devine state authority.

---

# 12. Symphony — păstrăm separarea

Symphony are o separare foarte clară între:

```text
tracker
orchestrator
workspace
agent runner
```

și un singur orchestrator authority pentru dispatch/reconciliation. [GitHub](https://github.com/openai/symphony/blob/main/SPEC.md?utm_source=chatgpt.com)

Fabric deja este mai strict decât Symphony la external effects.

Nu aș copia retry/backoff semantics de acolo, dar separarea conceptuală este bună.

---

# 13. Ce facem cu cachingul

Aici aș merge mult mai departe decât am spus inițial.

## Nu un „cache subsystem”

Ci:

> **content-addressed computation everywhere deterministic work is expensive.**

### Level 1 — parsed-file cache

```text
FileFactsKey =
H(
    parser_schema,
    parser_version,
    language,
    file_sha256
)
```

Produce:

```text
symbols
imports
calls
type relations
generator markers
tests
structural facts
```

Dacă fișierul nu s-a schimbat, **zero reparse**.

---

### Level 2 — repository graph

```text
RepoView =
BaseSnapshot
+ CandidateDelta
```

Nu reconstruiești graph-ul pentru fiecare candidate.

Exact modelul:

```text
immutable base
+
small overlay
```

pe care Fabric îl are deja lexical.

Îl generalizăm.

---

### Level 3 — compiled context cache

```text
ContextKey =
H(
    candidate_id,
    task_id,
    role,
    query_hash,
    repo_view_id,
    selector_version,
    budget
)
```

Dacă e identic:

```text
ContextManifest
```

poate fi reutilizat exact.

---

### Level 4 — deterministic verification cache

Inspirat de Bazel/Nix.

Conceptual:

```text
ActionKey =
H(
    executable,
    argv,
    cwd identity,
    environment policy,
    toolchain,
    complete input closure
)
```

→ exact output receipt.

Bazel își fundamentează caching-ul tot pe digests de command/input root/platform etc., iar Nix pe derivations/content-addressing. [GitHub](https://github.com/bazelbuild/remote-apis/blob/main/build/bazel/remote/execution/v2/remote_execution.pb.go)

Dar asta trebuie să fie **opt-in per check**.

Nu vreau:

```text
"go test passed earlier lol"
```

Vreau:

```text
exact same action
+
exact same input closure
=
reusable deterministic observation
```

Final release gates pot rămâne obligatoriu fresh.

---

# 14. Asta ne permite o arhitectură Fabric mult mai coerentă

Eu aș redefini conceptual Fabric în 7 plane-uri:

```text
                 HIGH-LEVEL OBJECTIVE
                         │
                         ▼
┌────────────────────────────────────────────┐
│ 1. INCREMENTAL ENGINEERING INTELLIGENCE    │
│                                            │
│ lexical │ SCIP │ structural │ call graph   │
│ imports │ generators │ tests │ impact      │
│                                            │
│       content-addressed + candidate delta  │
└─────────────────────┬──────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│ 2. CONTEXT COMPILER                        │
│                                            │
│ planner orientation                        │
│ explorer evidence                          │
│ writer implementation context              │
│ reviewer impact context                    │
└─────────────────────┬──────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│ 3. ARCHITECT / PLANNER                     │
│                                            │
│ objective → dependency DAG                 │
│ graph evidence → ownership / coupling      │
└─────────────────────┬──────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│ 4. SCHEDULER                               │
│                                            │
│ critical path + work conserving            │
│ resource vectors                           │
│ cohesion-aware parallelism                 │
└─────────────────────┬──────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│ 5. EXECUTION / INTEGRATION                 │
│                                            │
│ isolated writers / cluster worktrees       │
│ serial governed integration                │
└─────────────────────┬──────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│ 6. VERIFY / REVIEW / REPAIR                │
│                                            │
│ targeted checks │ impact tests             │
│ parallel review │ repair loop              │
└─────────────────────┬──────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────┐
│ 7. DURABLE CHECKPOINT                      │
│                                            │
│ accepted state → fresh-context continuation│
└────────────────────────────────────────────┘
```

Asta mi se pare acum **mult mai bun decât să continuăm să adăugăm funcții individuale**.

---

# 15. Model routing devine natural după asta

Ultraswarm are deja route-history/calibration/preflight ideas.

Fabric ar trebui să păstreze pentru fiecare task:

```text
task class
role
context size
model
effort
duration
input tokens
output tokens
repairs
verification outcome
review outcome
```

Ulterior:

```text
P(success | task, role, model)
cost
latency
```

și abia atunci faci adaptive routing.

Nu trebuie direct RL.

Un LightGBM/GBDT sau chiar priors empirice bine calibrate ar fi suficient inițial.

---

# 16. JEV intră abia aici

Atunci JEV devine foarte interesant.

Exemplu:

```text
planner uncertainty = 0.42

possible actions:
A: start writer
B: spend 8k tokens on another explorer
C: perform impact query for ~20ms
```

Avem istoric care estimează:

```text
E[success | information]
E[cost]
E[delay]
```

și putem aproxima:

\[
VOI(a) =
E[U \mid \text{after evidence }a]
-
E[U \mid \text{now}]
-
C(a)
\]

Atunci:

```text
if VOI(search) > 0:
    search
else:
    execute
```

Asta este JEV cu sens.

Nu „introducem JEV pentru că matematica sună bine”.

---

# 17. Ce aș elimina din roadmap

După audit, eu **nu** aș prioritiza:

| Idee | Verdict |
|---|---|
| generic vector DB | NU |
| Neo4j/graph DB pentru RI | NU |
| universal agent DSL | NU |
| 50–80 roluri specializate | NU |
| consensus/voting swarm | NU |
| Fabric-owned sandbox | NU |
| Fabric-owned conversation summarizer | NU |
| generic LLM response cache | NU |
| alt provider gateway gigantic | NU |
| full OpenHands dependency | NU |
| generic web dashboard acum | NU |
| JEV scheduler imediat | NU |
| distributed cache | NU |

---

# 18. Noul roadmap pe care l-aș adopta

Aș face următoarele milestones:

### **v1.1 — Engineering Intelligence**

Acesta trebuie să fie următorul.

Construim:

```text
incremental code graph
+
content-addressed per-file facts
+
base + candidate overlay
+
planner context
+
Context Compiler
+
impact/dependency/generator/test relations
```

Testăm întâi retrieval separat, apoi task outcome.

Ținta end-to-end:

```text
4/6 → 6/6
+
godotenv
+
generated-code ownership task
```

sau cel puțin eliminarea demonstrabilă a unei clase importante de BLOCKED.

---

### **v1.2 — Topology-Aware Parallel Engineering**

Pe inteligența codebase-ului:

```text
task partitioning
dependency graph
hub/shared-file detection
community/cohesion
cluster worktrees
governed serial integration
```

A/B:

```text
current writers
vs
new topology-aware parallelism
```

---

### **v1.3 — Content-Addressed Efficiency**

Generalizăm cache-ul:

```text
RI cache
context cache
derived facts
selected deterministic verification
```

Măsurăm:

```text
cold latency
warm latency
CPU
RAM
disk
model tokens
time-to-READY
```

---

### **v1.4 — Adaptive Model Allocation**

După ce avem suficiente traces reale:

```text
cheap model
medium model
frontier model
```

în funcție de task/role/risk.

---

### **v1.5 — Long-Horizon Autonomous Engineering**

```text
checkpoint
fresh-context round
verification
resume
```

și qualification pentru runtime-native compaction.

---

# 19. Primul milestone l-aș mai rafina puțin

Numele pe care l-aș da:

> **Fabric Engineering Intelligence v1**

Nu „Context Intelligence”.

Pentru că include:

```text
Repository Facts
        +
Incremental Graph
        +
Context Compiler
        +
Planner Evidence
```

iar **nu** scheduling încă.

Primul experiment poate fi foarte simplu:

```text
de61e6f baseline
        vs
de61e6f + Engineering Intelligence
```

aceleași:

```text
GPT-6 Luna High
runtime
objectives
repair budget
verification
held-out gates
```

pe:

```text
difflib
atomic
godotenv
generated-code
humanize
afero
```

Dacă nu îmbunătățește rezultatul, arhitectura nu intră în main.

---

# 20. Și asta cred că este identitatea corectă a Fabric

Nu:

> another multi-agent coding framework.

Și nici:

> a giant deterministic bureaucracy around an LLM.

Ci:

> **Fabric is an evidence-driven engineering execution layer that gives strong models an incremental understanding of a codebase, decomposes work according to real software dependencies, executes it safely in parallel, and accepts results only through executable evidence.**

Modelul trebuie să rămână partea inteligentă.

Fabric ar trebui să îi ofere:

**better information → better decomposition → efficient execution → strong verification → cheap reuse of deterministic work.**

După auditul ăsta, asta este direcția pe care aș fixa-o. Donorii cei mai importanți pentru următoarea etapă sunt, în ordinea valorii tehnice pe care o văd acum: **Sylphx repomap / SCIP / Serena pentru intelligence, Co-Coder pentru decomposition, Pact pentru integration, OpenCodeReview pentru review, LongHorizon pentru durată mare, Bazel/Nix pentru caching și Ultraswarm pentru routing/accounting.** Celelalte rămân referințe sau infrastructură de etapă ulterioară.