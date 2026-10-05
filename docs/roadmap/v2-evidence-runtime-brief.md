/goal

# Fabric v2.0 — Evidence-Driven Autonomous Software Engineering Runtime

## Accepted architectural steering — 2026-10-05

The user's [Adaptive Working Context decision](adaptive-working-context.md)
extends this brief. Implement the opt-in bounded per-agent editable projection,
one role end-to-end, deterministic resume/stale/corruption/bound evidence,
independent review and matched A/B/C evaluation. JEV controls acquisition;
agents control retention; Fabric controls authority. Develop independently from
the CLM paper, preserve existing licensing/provenance, and defer unnecessary
fixed-compaction expansion until the experiment is complete. This steering
does not narrow any original v2 requirement or permit hidden benchmark exposure.

## Mission

Take Fabric from its current v1.1.x state to a genuinely differentiated v2.0 autonomous software-engineering runtime.

Fabric must not become a larger collection of orchestration mechanisms.

The v2 objective is to make the existing system substantially better at the five decisions that determine autonomous software-engineering quality:

1. Where is the relevant code and evidence?
2. What work should be done, in what dependency structure, and what can safely run in parallel?
3. Which model/tool/action is worth invoking next?
4. Why did a candidate fail, and what is the cheapest correct way to repair it?
5. When is there enough evidence to stop investigating and act?

Fabric already has a strong control plane. v2 must add genuine engineering intelligence on top of it without weakening its deterministic evidence boundary.

The desired end state is:

    strong interchangeable models
                +
    repository / program intelligence
                +
    evidence-valued context acquisition
                +
    topology-aware engineering scheduling
                +
    structured diagnosis and repair
                +
    empirically calibrated model allocation
                +
    exact verification and independent review
                +
    durable recoverable execution

Do not optimize for feature count.

Optimize for:

    task success
    reliability
    time-to-accepted-candidate
    evidence quality
    useful parallelism
    efficient model/computation use
    recoverability
    comprehensibility of decisions

The governing product principle is:

    models perform cognition;
    Fabric owns engineering state, evidence, authority and execution policy.

And:

    fail closed on authority;
    fail soft on optional capability;
    preserve useful engineering state on failure.

---

# 0. Start from actual current state

Before modifying anything, inspect the actual `dev` HEAD, working tree, current source-history policy, release state, evaluation records and capability documentation.

Do not rely blindly on historical roadmap entries.

At the time this goal was written, the important state was approximately:

- v1.1.0 is an immutable published Windows amd64 release.
- current development has progressed through v1.1.4.
- main is a rolling major/minor source snapshot while dev retains incremental history.
- v1.1 introduced resilient autonomous execution:
  - typed outcomes;
  - graceful optional-capability fallback;
  - bounded semantic corrections;
  - candidate-bound replan;
  - progress-aware execution;
  - memory-aware future admission;
  - verification preflight;
  - safe cache degradation;
  - retained semantic results when optional accounting metadata is unavailable.
- v1.1.1 added continuation after exact sealed read-only capacity rejection.
- v1.1.2 strengthened writer/fixer behavior preservation.
- later v1.1 development recorded accepted candidates for all eight canonical task identities:
  - one original cohort remained 7/8 because one planner was refused by provider capacity;
  - one separately authorized fresh successor passed the remaining task;
  - therefore this is 8/8 accepted task coverage, not a claim that the original frozen cohort was 8/8.
- v1.1.4 added source-bound AGENTS.md and Agent Skills context.
- that new agent-context behavior is not automatically qualified by earlier cohorts.

Re-read the actual repository and correct this summary if development has moved.

Preserve historical evidence exactly.

Never rewrite an old BLOCKED/UNKNOWN/FAIL result as PASS merely because a later version succeeds.

---

# 1. Explicitly retire obsolete planning documents

`docs/research/donor-gap-audit.md` describes an old EngOrch/Fabric baseline.

Most of its important gaps are now implemented, superseded or explicitly rejected.

Mark it clearly as SUPERSEDED and point current development to the newer donor/evidence documents.

Do the same for any roadmap statement whose “missing capability” is already present in current source.

Do not delete useful historical research.

Separate:

    historical research
    current architecture
    current evidence
    future hypotheses

A reader must not mistake an old donor gap for current product state.

Current donor research should be consolidated around:

- engineering intelligence;
- topology / review / repair;
- cache / routing / long-horizon execution;
- OSS reuse decisions;
- exact provenance and licensing.

---

# 2. Non-negotiable Fabric invariants

All v2 work must preserve these.

## Authority and identity

Never weaken:

- repository identity;
- source commit/tree identity;
- candidate identity;
- exact task/graph identity;
- effect intent/receipt ordering;
- file preimage/hash validation;
- safe path confinement;
- explicit write ownership;
- credential boundaries;
- runtime/model identity;
- candidate-bound verification;
- candidate-bound review;
- journal integrity;
- replay identity.

UNKNOWN remains UNKNOWN.

An uncertain external effect is never automatically resent.

A different invocation after a known terminal failure must receive a new identity.

No algorithmic score, LLM judgment, graph relation, learned probability or JEV result grants write authority.

## Optional intelligence is advisory

The following can influence selection, prioritization or scheduling but never authority:

- BM25;
- embeddings;
- PPR;
- graph centrality;
- community detection;
- SBFL;
- model confidence;
- learned rankings;
- JEV;
- routing estimates;
- review grouping;
- skills;
- LLM-generated task decomposition.

The controller remains authoritative.

## Graceful degradation

Optional capability failure should degrade where safe.

Examples:

    semantic RI unavailable
        -> lexical/source context

    isolated parallelism unavailable
        -> serial execution

    derived cache unavailable/corrupt
        -> recompute

    optional compaction unavailable
        -> continue without compaction

But:

    identity mismatch
    unsafe filesystem state
    unresolved effect
    corrupted authoritative journal
    unauthorized mutation

must remain hard stops.

---

# 3. General mathematical policy: no coefficient soup

Fabric v2 must not accumulate arbitrary expressions such as:

    score =
        0.37 * BM25
      + 0.21 * PPR
      + 0.14 * graph_degree
      + ...

unless those coefficients are learned or derived from explicitly documented evidence.

Use this preference order:

    1. hard constraints
    2. exact statistics
    3. ordinal evidence classes
    4. rank aggregation
    5. lexicographic optimization
    6. Bayesian empirical estimates
    7. Value-of-Information-derived decisions
    8. learned models only after sufficient data

Any free parameter must have one of these origins:

- mathematical definition;
- confidence level;
- explicit user/product constraint;
- observed resource budget;
- posterior estimate;
- optimization dual variable / shadow price;
- empirical calibration;
- documented ordinal prior which is later replaceable by evidence.

Do not tune constants until benchmarks “look good”.

Do not introduce LambdaMART, GNNs, graph transformers, reinforcement learning or large learned policies merely because they are sophisticated.

Complexity requires measured incremental value.

---

# 4. Build a common Evidence Value Controller

This should become one of Fabric v2's central intellectual contributions.

The same problem currently appears repeatedly:

- should we inspect more source?
- should we calculate a graph query?
- should we run SBFL?
- should we ask another explorer?
- should we invoke a stronger model?
- should we run a targeted test?
- should we generate another patch?
- should we stop researching and act?

Model all of these as actions which consume resources and may improve the engineering decision.

Let the current evidence state be D.

For an information-producing action q:

    VOI(q)
      =
      E_y [
          max_a E[U(a) | D, y]
      ]
      -
      max_a E[U(a) | D]
      -
      C(q)

Equivalent internal naming such as JEV / Engineering Value of Information is acceptable, but define the term precisely.

The controller should ask:

    what is the expected improvement in engineering decision quality
    from obtaining this information,
    minus the real opportunity/resource cost?

If:

    JEV(q) <= 0

do not purchase that evidence.

If several actions have positive JEV, prefer the highest admissible action.

After observing its result, update the evidence state and recompute.

This creates an anytime loop:

    current evidence
        ->
    estimate candidate information actions
        ->
    execute highest positive-JEV action
        ->
    update posterior/evidence
        ->
    repeat
        ->
    no positive JEV
        ->
    act

The controller must always obey hard safety and authority constraints before JEV optimization.

---

# 5. Resource prices must be dynamic shadow prices, not hand-tuned lambdas

Where engineering utility trades quality against resources, do not manually choose arbitrary weights.

Instead formulate:

    maximize expected engineering quality Q

subject to constraints such as:

    E[token use] <= token budget
    E[monetary spend] <= spend budget
    E[latency] <= latency budget
    E[human intervention] <= intervention budget
    E[concurrency] <= resource capacity

The constrained problem creates Lagrange multipliers / shadow prices.

For resource j:

    lambda_j(t+1)
      =
      projection_positive(
          lambda_j(t)
          +
          eta_t * (observed_cost_j - budget_j)
      )

Normalize resource usage by the corresponding budget before updating.

Do not hand-tune eta indefinitely.

Prefer an adaptive step-size algorithm such as AdaGrad/AdaHedge-style adaptation derived from observed gradient magnitude/regret.

Interpretation:

- if Fabric repeatedly exceeds latency budget, latency becomes more expensive;
- if token budget is consistently underused, token shadow price decreases;
- if provider capacity is scarce, the provider-slot price rises;
- if the user explicitly requests maximum-quality mode, resource budgets can be relaxed.

Then define action cost in JEV as:

    C(q)
      =
      lambda_tokens   * token_cost(q)
      +
      lambda_latency  * latency(q)
      +
      lambda_money    * monetary_cost(q)
      +
      lambda_compute  * local_compute(q)
      +
      lambda_human    * expected_intervention(q)

Unknown cost remains unknown.

Never convert unknown monetary cost into zero.

The Evidence Value Controller and model router should share this internal resource economy rather than maintaining unrelated manually weighted formulas.

---

# 6. Dynamic evidence trust without arbitrary weights

For every evidence source maintain an empirical reliability model.

Examples:

- lexical retrieval;
- exact identifier matching;
- SCIP definitions;
- SCIP references;
- PPR;
- SBFL;
- stack traces;
- generator relations;
- change proximity;
- review findings;
- runtime diagnostics.

Start with documented ordinal priors rather than arbitrary continuous tuning.

For example:

    E0 = no useful relation
    E1 = weak supporting evidence
    E2 = moderate supporting evidence
    E3 = direct evidence
    E4 = near-conclusive direct evidence

If a numeric prior is required, map the ordinal scale through a fixed symmetric transform such as logistic log-odds.

Do not claim those initial values are measured.

Over time replace priors with empirical posteriors.

For binary reliability events a simple starting model is:

    r_i ~ Beta(alpha_i, beta_i)

with updates based on objectively observable downstream evidence.

Do not use vague labels such as “the signal seemed useful”.

Define measurable attribution, for example:

- accepted repair target appeared in top-K localization;
- selected source contained the final modified symbol;
- evidence discriminated failing from passing behavior;
- source changed the preferred action and the resulting action passed acceptance.

Where attribution is ambiguous, record it as ambiguous rather than force an update.

For a current task, estimate marginal decision value:

    MV_i
      =
      V(D)
      -
      V(D without evidence source i)

Dynamic evidence weighting may then combine:

    historical reliability
    +
    current marginal value

If a continuous normalized expert distribution is useful, one possible form is:

    w_i
      proportional to
      prior_reliability_i * exp(MV_i / temperature)

but production use must avoid a manually tuned temperature.

Prefer adaptive exponential-weights algorithms such as AdaHedge/Squint-style adaptation if this mechanism is promoted.

Before enough observations exist, use equal rank aggregation or ordinal rules.

---

# 7. Engineering Context Intelligence v2

Do not build another vector database or another repository graph store.

Extend the existing Rust RI and SCIP/structural evidence.

The context system should answer:

    what is the smallest bounded evidence set that gives the agent the highest expected engineering value?

## 7.1 Candidate generation

Produce candidates from independent evidence sources:

- exact path/identifier matches;
- current lexical/tgrep/BM25 retrieval;
- SCIP definitions;
- SCIP references;
- explicit IMPLEMENTS relationships;
- package/module relations;
- generator relations;
- test relations;
- changed candidate paths;
- failing verification evidence;
- stack traces where available;
- current task ownership;
- graph neighbors.

Preserve each source separately.

Do not collapse everything into one opaque score.

## 7.2 Hierarchical localization

Use progressive narrowing:

    repository
      ->
    files/modules
      ->
    symbols
      ->
    precise spans

Do not give line-level context merely because it is available.

A file-level localization stage should narrow the search first.

A symbol stage should narrow again.

Only then select relevant spans/body/context.

## 7.3 Independent rankers + RRF baseline

Use multiple independent rankers.

Examples:

    lexical rank
    exact-symbol rank
    SCIP/reference rank
    graph-proximity rank
    SBFL rank when applicable
    changed-code rank

The safe initial fusion algorithm should be Reciprocal Rank Fusion:

    RRF(d)
      =
      sum_r 1 / (k + rank_r(d))

Use the standard/justified fixed k or evaluate a small predeclared set; do not manually assign ranker weights.

Weighted RRF may be introduced only after the Evidence Value Controller has meaningful source-specific reliability/marginal-value evidence.

## 7.4 Personalized PageRank experiment

Over the already admitted graph, evaluate Personalized PageRank:

    pi
      =
      (1 - alpha) * s
      +
      alpha * P^T * pi

Seed distribution s should come from observed task evidence:

- objective identifiers;
- failing tests;
- changed files;
- stack frames;
- review findings;
- explicit task paths.

alpha must be a documented standard graph-diffusion parameter or selected through fixed evaluation rather than tuned per benchmark.

PPR is a selector, not authority.

Compare PPR treatment against the current selector at identical file/byte/token budgets.

Do not promote it merely because graph rankings look reasonable.

## 7.5 Heat-kernel diffusion experiment

Optionally test:

    h_t = exp(-tL) s

where L is the graph Laplacian.

Treat this as an alternate locality model.

Do not ship both PPR and heat diffusion unless the second adds measured value.

## 7.6 Budgeted context selection

After candidates are ranked, do not simply select independent top-K items.

Use a bounded coverage formulation.

Let a_ic represent whether/how strongly evidence item i covers concept c.

A clean initial objective is:

    F(S)
      =
      sum_c [
        1 - product_(i in S)(1 - a_ic)
      ]

subject to:

    sum_(i in S) cost(i) <= B

where B is the exact role context budget.

This naturally creates diminishing returns and reduces the need for a separate arbitrary redundancy penalty.

Use lazy greedy / gain-per-token selection:

    choose item maximizing
        marginal_gain(item | selected)
        /
        token_or_byte_cost(item)

until the budget is exhausted or marginal/JEV value becomes non-positive.

The concepts c should themselves be evidence-bound:

- requested symbols;
- failing tests;
- relevant APIs;
- generators;
- dependencies;
- review findings;
- expected evidence from the accepted task graph.

Do not invent thousands of latent concepts merely to make the formula sophisticated.

## 7.7 Learned ranking only if justified

Do not implement LambdaMART initially.

Accumulate real traces.

Only reconsider a learned ranker after enough independent tasks exist.

As a rough precondition, seek hundreds to thousands of clean localization episodes with stable labels and held-out evaluation.

Compare any learned ranker to the simple RRF/submodular baseline.

If the improvement is marginal, retain the simpler baseline.

---

# 8. Repair Intelligence — highest-priority new product capability

Fabric already detects failure well.

v2 must become excellent at understanding and closing failures.

Replace prose-only repair with a structured pipeline:

    verification/reviewer/analyzer failure
        ->
    structured Finding
        ->
    localization
        ->
    validated/re-anchored RepairSpec
        ->
    strategy selection
        ->
    candidate repair
        ->
    closure oracle
        ->
    finding closed or diagnosed

Take inspiration from Alibaba OpenCodeReview, Agentless, AutoCodeRover and related repair systems without importing their authority assumptions.

## 8.1 Structured Finding

Evolve the current minimal finding into a richer bounded schema.

Conceptually:

    Finding {
        id
        candidate_id

        source:
            reviewer
            native_test
            static_analyzer
            security
            performance
            mutation
            other_typed_source

        rule_id?
        category
        severity

        path
        symbol?
        range?

        existing_hash?
        existing_code?

        message
        suggested_change?

        evidence_ids[]

        closure {
            type
            exact_oracle_identity
        }
    }

Do not require every producer to populate every optional field.

Preserve source-specific truth.

A reviewer suggestion is not the same thing as a failing executable oracle.

## 8.2 Finding anchoring and validation

Before repair:

- verify candidate identity;
- verify path;
- re-resolve symbol/range if candidate changed;
- verify existing hash/code where supplied;
- mark stale findings explicitly;
- reject ambiguous relocation rather than silently guessing.

If exact re-anchoring is impossible, use Engineering Intelligence to produce bounded candidates and require an explicit repair localization decision.

## 8.3 Fault localization

Combine deterministic and graph evidence.

### Spectrum-based fault localization

Where test coverage is available, calculate Ochiai:

    suspiciousness(e)
      =
      failed_tests_executing_e
      /
      sqrt(
        total_failed_tests
        *
        (
          failed_tests_executing_e
          +
          passing_tests_executing_e
        )
      )

Keep raw values.

### Stack / failure evidence

Rank directly observed stack frames, failing assertions and exact diagnostics separately.

### Graph propagation

Seed PPR from failing/changed/reviewer evidence and propagate suspicion through admitted dependency/reference relations.

### Fusion

Initially combine rankings using RRF.

Do not introduce manually tuned linear coefficients.

Later dynamic weighting may come from the Evidence Value Controller.

## 8.4 RepairSpec

A Finding should be converted into a repair-specific object:

    RepairSpec {
        finding_ids
        candidate_id
        localization_candidates
        allowed_scope
        expected_behavior
        prohibited_behavior_changes
        suggested_strategy
        closure_oracles
        evidence
    }

The RepairSpec cannot widen file ownership by itself.

If additional ownership is necessary, go through the existing candidate-bound replan mechanism.

## 8.5 Strategy router

Choose among:

    deterministic transformation
    localized LLM repair
    frontier-model escalation
    bounded multi-candidate tournament

Selection should depend on evidence and JEV, not a fixed “always call the strongest fixer” policy.

Examples:

    exact mechanical rename
        -> deterministic transform

    single localized syntax/API change
        -> cheap/local writer

    ambiguous semantic compatibility failure
        -> stronger fixer

    multiple plausible implementations with strong cheap oracles
        -> patch tournament

## 8.6 Deterministic transformation engine

Do not invent another transformation DSL.

Evaluate thin integrations around mature engines where appropriate:

- ast-grep;
- Semgrep autofix;
- OpenRewrite where language/domain fits;
- Coccinelle where relevant;
- native compiler/formatter/code-mod tools.

Fabric owns:

- exact candidate binding;
- transformation specification;
- scope;
- resulting diff;
- verification.

The external transformation engine only produces deterministic bytes.

Do not add a dependency unless a real repair class benefits.

## 8.7 Patch tournament

For difficult findings, allow several bounded candidate repairs when verification is sufficiently cheap/discriminative.

Do not blindly run full verification on every candidate.

Use multi-fidelity racing:

    syntax/parse/type sanity
        ->
    directly failing test
        ->
    targeted impacted tests
        ->
    mutation/property oracle when relevant
        ->
    broader native checks
        ->
    review

Remove dominated candidates early.

Use Pareto dominance over observable dimensions such as:

- number/severity of findings closed;
- targeted checks passed;
- blast radius;
- diff size;
- mutation score;
- full verification status.

Do not collapse those dimensions into an arbitrary scalar score unless a resource-constrained decision actually requires it.

An optional Determinantal Point Process may later be evaluated for selecting diverse hypotheses:

    L_ij = q_i q_j S_ij

and maximize determinant over a subset.

Do not make DPP a default before proving that redundant candidate generation is a real bottleneck.

## 8.8 Closure must be exact

A finding is closed only by its declared closure oracle.

Examples:

    native test
    analyzer rule
    exact reviewer re-check
    invariant/property test
    mutation discriminator
    performance threshold

“Fixer says fixed” is not closure.

---

# 9. Test and dynamic-program intelligence

Verification must evolve from a binary final gate into useful repair evidence while preserving full final acceptance.

## Targeted test selection

Use:

- changed symbols/files;
- explicit test relationships;
- SCIP/structural references;
- package/module ownership;
- historical failure observations;

to propose a targeted verification subset for early repair racing.

This is optimization only.

Final acceptance still runs the configured required checks.

## SBFL support

Where supported by the language/tooling, collect coverage sufficiently to compute fault-localization statistics.

Keep producer/toolchain identity.

Do not pretend coverage is complete when it is not.

## Mutation testing

For tasks where the agent writes/changes tests, bounded mutation testing is especially valuable.

Mutation score:

    M
      =
      killed_mutants
      /
      non_equivalent_mutants

Do not make a universal mutation threshold a global gate.

Use mutation analysis when JEV indicates that it is valuable, especially when:

- new tests were produced by the same agent that wrote the implementation;
- the bug is a boundary/parser/state-machine problem;
- native tests pass too easily;
- review suspects insufficient regression coverage.

Keep mutation tool and operator identity exact.

## Property/fuzz testing

For parsers, serializers, numeric conversions, state machines and protocol boundaries, allow bounded property-based or fuzz oracles where they provide cheaper discrimination than another model call.

Again:

    evidence source
    not authority to widen scope.

---

# 10. Topology-Aware Engineering Scheduler

Fabric already has task graphs, resource vectors, parallel explorers, isolated writers and serial integration.

v2 must make task decomposition and scheduling genuinely structure-aware.

The scheduler should optimize execution of already valid tasks; it must not invent authority.

## 10.1 Model the engineering system as a heterogeneous hypergraph

Use nodes representing relevant entities:

    files
    symbols
    packages/modules
    tests
    generators
    task nodes

Use typed relationships:

    IMPORTS
    CALLS / REFERENCES
    IMPLEMENTS
    GENERATES
    TESTS
    SHARED_FIXTURE
    SHARED_CONFIG
    OWNED_BY_TASK
    DEPENDS_ON_TASK

Generator families and similar n-ary relations should be representable as hyperedges, not forced into arbitrary pairwise approximations.

## 10.2 Separate dependency, coupling and ownership

This is a fundamental invariant:

    dependency != coupling != ownership

Dependency determines legal readiness.

Coupling predicts whether concurrent work is likely to interfere.

Ownership determines write authority.

A task can have disjoint write ownership but still be strongly coupled through a shared API/generator/config.

Do not equate “different files” with “independent work”.

## 10.3 Coupling tiers before numeric weights

Prefer ordinal structural classes:

    C4 — hard coupling
        write overlap
        same generated family requiring shared mutation
        same mutable owner

    C3 — strong structural coupling
        direct API/call dependency
        shared critical config
        shared state contract

    C2 — moderate coupling
        same package/module
        shared test fixture
        high reference overlap

    C1 — weak coupling
        graph proximity
        shared import neighborhood

    C0 — no observed coupling

Use hard rules for C4.

Use lower tiers as optimization information.

If numeric edge strength is needed, prefer values derived from the graph:

    Jaccard neighborhood overlap
    normalized common-neighbor counts
    cosine over relation incidence
    pointwise mutual information
    observed historical conflict probability

rather than arbitrary per-edge constants.

## 10.4 Lexicographic scheduling optimization

Do not define:

    0.4 conflict + 0.2 latency + 0.3 balance ...

Use lexicographic optimization.

First enforce hard constraints:

- dependency readiness;
- non-overlapping legal ownership;
- generator ownership;
- candidate compatibility;
- provider/model/runtime capacity;
- RAM/CPU limits;
- verification slots.

Then optimize in order.

A reasonable hierarchy:

    1. minimize illegal/unsafe solution count -> hard zero
    2. minimize strong coupling cuts / predicted integration conflict
    3. minimize critical-path makespan
    4. minimize idle resource time
    5. improve load balance

Use repeated CP-SAT/MILP solves or lexicographic objectives:

    solve objective 1
    freeze optimum
    solve objective 2
    freeze optimum
    ...

For small task graphs, prefer near-exact CP-SAT.

For larger graphs, evaluate multilevel partitioning/coarsening with deterministic constraint repair.

Do not introduce a huge solver dependency if current task graphs are tiny enough for a small internal search.

Measure.

## 10.5 Graph structure features

Evaluate:

- connected components;
- articulation points;
- bridge edges;
- degree;
- betweenness centrality;
- k-core;
- conductance;
- optional Leiden/Louvain communities.

Treat community detection as a proposal.

A central hub/generator may naturally produce:

    shared hub first
        ->
    independent leaves concurrently

rather than starting all leaves immediately.

## 10.6 Receding-horizon scheduling

Do not assume the first plan remains optimal forever.

After meaningful terminal observations:

    G_t -> G_(t+1)

recompute the remainder using current evidence:

- completed tasks;
- failed tasks;
- discovered coupling;
- changed resource availability;
- new ownership requirements;
- measured durations;
- model/provider availability.

Use Model Predictive Control semantics:

    solve current horizon
    execute admitted next wave
    observe
    re-solve remainder

Never rewrite completed task history.

Never use this mechanism to evade repair budgets or UNKNOWN settlement.

## 10.7 Dynamic idle-worker assignment

Borrow the good CAID/Pact idea:

when capacity becomes free, assign another already-validated ready task.

Do not ask a manager model to invent new legal work merely because a worker is idle.

Only tasks already valid under the accepted graph/replan can be scheduled.

## 10.8 Integration

Keep isolated worktrees where parallel implementation requires attribution.

Integration remains explicit and serially controlled.

Do not adopt:

    git merge -X theirs
    “prefer one branch”
    automatic semantic conflict resolution

without a fresh candidate, verification and review.

---

# 11. Evidence-Calibrated Model Allocation

Fabric has already demonstrated multiple routes/models.

v2 should turn that into a real engineering-economic advantage.

The goal is not “use cheap models”.

The goal is:

    use the least expensive admissible model
    whose evidence supports the required reliability,
    and escalate only when expected engineering value justifies it.

## 11.1 Start simple: Beta-Bernoulli route reliability

For route/model/role/task-class buckets:

    p ~ Beta(alpha, beta)

PASS/accepted outcomes update alpha.

Semantically failed exercised outcomes update beta under a carefully defined label policy.

UNKNOWN/provider-capacity/no-exercise must not become FAIL automatically.

Maintain enough dimensions to be meaningful without making every bucket empty.

Potential hierarchy:

    global route prior
        ->
    role
        ->
    language
        ->
    coarse task class

Only split buckets when sample support is sufficient.

## 11.2 Reliability-constrained routing

Prefer formulation:

    minimize expected resource cost

subject to:

    lower_confidence_bound(
        P(success | context, model)
    )
    >= required_quality

instead of:

    quality - 0.3 cost - 0.2 latency

The required quality threshold should be derived from a baseline policy.

For example:

    Q_min
      =
      LCB(frontier_baseline)
      -
      explicitly allowed quality degradation epsilon

epsilon is a product constraint, not a learned coefficient.

High-risk tasks may set epsilon = 0.

## 11.3 Escalation ladder

A safe initial policy can be:

    efficient worker
        ->
    stronger worker
        ->
    frontier fixer/reviewer

Escalate on evidence such as:

- previous semantic failure;
- ambiguous localization;
- repeated review finding;
- high coupling;
- high blast radius;
- low posterior confidence;
- JEV of stronger model positive and greater than alternatives.

Do not hard-code one provider/model forever.

## 11.4 Contextual bandit only after sufficient traces

Later, evaluate contextual LinUCB/Thompson Sampling.

Features may include:

    role
    language
    task type
    graph size
    coupling
    number of owned paths
    context size
    repair count
    previous failure classes
    verification cost
    localization uncertainty

For a linear contextual model:

    mu_hat_a(x) = x^T theta_hat_a

    sigma_a(x)
      =
      sqrt(
        x^T A_a^-1 x
      )

Confidence exploration parameter beta should derive from:

- selected confidence level;
- observation noise assumptions;
- sample count/dimension;

not manual tuning.

Use a safety baseline.

If the candidate route's quality lower confidence bound is insufficient, use the baseline/frontier route.

## 11.5 JEV-driven model escalation

Treat a stronger model call as another information/action option:

    JEV(stronger_model_call)
      =
      expected improvement in eventual accepted candidate
      -
      shadow-priced resource cost

Compare it with:

    run targeted test
    calculate graph evidence
    run SBFL
    call another explorer
    generate deterministic transform

A stronger model should not automatically win merely because it is stronger.

This is where Fabric can become materially more efficient.

## 11.6 Do not use full RL yet

No PPO/reward-model training.

No opaque learned policy controlling authority.

Only revisit RL when the repository has a genuinely large corpus of clean, independently evaluated engineering episodes and simple contextual methods demonstrably plateau.

---

# 12. Stop research intelligently

Fabric should avoid both:

    insufficient exploration

and:

    five agents reading the same files

Use JEV as a stopping rule.

At every research stage estimate available evidence actions.

If:

    max_q JEV(q) <= 0

stop acquiring evidence.

Proceed with the best admissible engineering action.

This decision and its evidence should be inspectable.

For example:

    candidate evidence actions:
      PPR expansion       +0.04
      targeted test       +0.19
      Muse explorer       -0.01
      Sol explorer        -0.07

    selected:
      targeted test

After receiving the test result:

    recompute.

The displayed values must be labelled estimates and carry their provenance/model version.

---

# 13. Cross-language Engineering Intelligence

Fabric v2 should not remain effectively Go-only while claiming to be a general software-engineering runtime.

Do not write four language servers/parsers inside Fabric.

SCIP remains the preferred semantic interchange.

Add/qualify external producers incrementally.

Priority languages:

    Go
    Rust
    TypeScript / JavaScript
    Python

For each language, qualify at least:

- producer binary/version/arguments;
- source input hashes;
- path and position encoding;
- definitions;
- references;
- applicable relationships;
- incomplete-coverage behavior;
- false-absence regression.

Structural/tree-sitter evidence remains a weaker fallback.

Do not claim semantic absence from incomplete producers.

Add small Serena-inspired read-only vocabulary where it improves agent ergonomics:

    symbols_overview
    find_symbol
    references
    implementations
    declaration

but implement it over existing immutable evidence.

Do not integrate Serena's mutable runtime/LSP server into Fabric core.

---

# 14. Agent context / Agent Skills qualification

The v1.1.4 agent-context system is architecturally sensible but requires actual task-quality qualification.

Preserve:

- source-bound committed AGENTS.md;
- scope inheritance;
- exact bundle identity;
- task/role skill filtering;
- explicit selected skills;
- advisory-only authority.

Do not grow this into:

    skill marketplace
    generic workflow language
    dynamic permission system
    another planner DSL

Run a matched evaluation:

    no agent-context
    vs
    scoped AGENTS.md + selected skills

on tasks where repository-local guidance should matter.

Measure:

- task acceptance;
- tool calls;
- repairs;
- input/cached/uncached tokens;
- latency;
- instruction violations.

Promote defaults based on evidence, not novelty.

---

# 15. Deterministic incrementality and caching

Fabric already has bounded caches.

Do not build a giant cache subsystem merely because Bazel/Nix are sophisticated.

Profile first.

## Keep

Content-address deterministic expensive facts such as:

- parse facts;
- immutable graph facts;
- compiled context;
- candidate review facts;
- deterministic format observations.

## Generic action cache

A broader deterministic verification/action cache may be implemented only if profiling shows repeated verification setup/execution is a meaningful contributor to time-to-READY.

An action identity must include complete relevant closure:

    candidate/source manifest
    executable/toolchain identity
    argv
    environment
    cwd/path mapping
    timeout/policy
    declared inputs
    expected outputs
    platform-relevant state

A hit returns a derived artifact/observation.

It must never create:

    a new READY verification pass
    review approval
    permission
    settled external effect

Final release acceptance remains fresh.

Corrupt safe derived cache entries should be discarded/recomputed.

Unsafe roots or identity substitutions remain DENY.

---

# 16. Runtime integration policy

Do not turn Fabric into a provider/runtime framework for its own sake.

Existing:

    Codex
    OpenCode
    direct bounded providers

are enough for the v2 core.

## ACP

A thin ACP adapter is strategically useful, but it is not automatically a v2 release blocker.

Implement it only if there is a concrete runtime that we want to support through ACP.

Use the official SDK/protocol where feasible.

Translate:

    session
    cancellation
    permissions
    updates

into existing Fabric state.

ACP must not grant direct file/effect authority.

## OpenHands / SWE-ReX

Remote/cloud worker execution is post-core work unless a concrete v2 acceptance requirement needs it.

If evaluated, use them as optional execution environments under Fabric authority.

Do not build an owned Docker/Kubernetes/sandbox platform.

## gitoxide

Do not add gix because the donor ledger calls it “required”.

First profile current Git CLI object-read overhead.

Only introduce a Rust object-read seam if measured startup/object-read cost materially matters.

Keep worktree/effect mutations on real Git unless equivalence is proven.

---

# 17. Remote / multi-repository engineering is not a v2 core requirement

Do not delay v2.0 merely to implement distributed multi-repository autonomy.

Cross-repo SCIP/intelligence and remote workers remain valid future directions.

They should become v2.x work unless an actual release-critical task proves they are required.

The same applies to:

- organization-wide graph stores;
- distributed caches;
- cloud orchestration;
- large-scale SWE-bench infrastructure.

---

# 18. UX must remain simpler than the architecture

Fabric may become mathematically sophisticated internally.

The user should not have to understand PPR, CP-SAT, JEV or Beta posteriors to use it.

A normal task should still look approximately like:

    fabric run --autonomous "objective"

and:

    fabric inspect RUN

Important user-visible states remain small and explicit:

    READY
    PAUSED
    NEEDS_ATTENTION
    NEEDS_REPLAN
    UNKNOWN / NEEDS_RECONCILIATION
    UNSAFE

Do not leak internal failure taxonomies unnecessarily.

Every non-READY state must provide:

- what happened;
- candidate retained?;
- what evidence exists;
- whether continuation is safe;
- next legal action.

Do not resurrect generic `execution_blocked` where a typed outcome exists.

---

# 19. Decision explanation

For important adaptive decisions, retain enough structured evidence to answer:

    Why did Fabric read this file?
    Why did these tasks run in parallel?
    Why did this task remain serial?
    Why did it choose Muse instead of Sol?
    Why did it call another explorer?
    Why did it stop exploring?
    Why did it choose this repair candidate?
    Why was this finding considered closed?

Do not store giant natural-language rationales as authority.

Persist compact structured decision evidence:

    considered options
    constraints
    estimates/posteriors
    selected action
    policy/version
    observed result

Explanation is derived from that state.

---

# 20. Observability and data collection for future learning

Fabric v2 needs a clean engineering episode dataset, produced naturally by real usage.

Record bounded structured observations where already available:

    objective/task class
    language
    role
    graph topology summary
    candidate size
    context source/ranker provenance
    selected evidence
    model/profile
    tool calls
    repairs
    verification result
    review result
    held-out result where evaluation uses one
    latency
    input tokens
    cached input
    output
    reasoning
    known cost
    unknown cost marker
    resource observations

Do not record secrets or unnecessary source/prompts in telemetry.

Canonical journals remain the truth.

Derived analytical tables can be rebuilt.

Do not create a mysterious mutable ML feature store.

---

# 21. Evaluation methodology

This is essential.

Do not benchmark-chase.

Do not repeatedly rerun until a task happens to PASS and then call that the score.

Every experiment must have a written hypothesis and frozen treatment before provider dispatch.

## Existing canonical engineering tasks

Keep the existing task identities and unchanged held-out oracles.

Historical results remain immutable.

The previous 8/8 accepted-task coverage with one capacity successor is useful evidence but is not a pristine single-cohort 8/8 claim.

## Capacity-successor policy

Before final v2 evaluation, explicitly define treatment of a provider refusal occurring before semantic result/effect.

For example:

A task may receive one separately identified fresh successor only if:

- the original invocation is proven terminal;
- no result exists;
- no uncertain effect exists;
- the failure is classified as provider capacity/service refusal;
- the original evaluation row remains failed/refused;
- the successor is reported separately.

Never relabel the original cohort.

This prevents infrastructure availability from being confused with model/task failure without hiding the raw score.

## Fresh canonical final cohort

On the exact final v2 release candidate:

run one fresh unchanged canonical cohort.

Report the raw cohort result.

Target:

    all canonical task identities accepted
    with unchanged native + review + held-out contracts

If one row is a pre-effect provider refusal, apply only the predeclared successor policy above and still report the original raw cohort.

No pooling across arbitrary versions.

## Fresh canaries

Add several held-out tasks not used to tune:

- at least one Rust task;
- at least one TypeScript/JavaScript task;
- at least one Python task;
- at least one nontrivial repair task;
- at least one generated/shared-owner task;
- at least one task with meaningful topology/parallelism potential.

Do not add canaries whose answers become embedded in prompts or skills.

---

# 22. Specific subsystem experiments required before default adoption

## Repair Intelligence experiment

Compare current repair versus structured Finding/RepairSpec treatment on frozen repair-requiring tasks.

Measure:

    accepted candidate rate
    findings closed
    repair attempts
    model invocations
    targeted/full verification executions
    uncached tokens
    time-to-READY

Require no reduction in acceptance rigor.

## Context selector experiment

Compare:

    current selector
    vs
    hierarchical RRF/PPR/submodular treatment

under the same source/task/model/context byte/token ceiling.

Record:

    selected files/symbols/spans
    target coverage
    tool calls
    downstream acceptance
    latency
    token types

Do not promote PPR/heat diffusion if it only improves synthetic ranking fixtures.

## Topology scheduler experiment

Use at least three classes:

    independent leaves
    shared hub/API
    shared generator

First run deterministic/fake-runtime scheduling against the exact same graph.

Then run real model tasks where useful.

Compare serial and topology treatments with:

    identical objective
    source pin
    model configuration
    acceptance
    resource policy

Measure:

    writer overlap
    critical path
    integration conflicts
    repairs
    total wall time
    tokens
    candidate acceptance

A parallel result that changes the semantic plan too much cannot establish causal speedup; report it honestly.

## Model routing experiment

Compare:

    frontier baseline

against:

    evidence-calibrated worker policy

Potential treatment:

    Sol/frontier planner where judgment is high
    Muse efficient explorers/writers where confidence permits
    Luna reviewer where qualified
    Sol/frontier fixer on escalation

Use exact unchanged tasks.

Primary metric:

    accepted-task quality

Then:

    latency
    tokens
    repairs
    escalations
    known cost

The desired result is frontier-level engineering acceptance with less frontier computation.

Do not claim savings from missing price information.

## JEV experiment

Begin offline where possible using frozen episode traces.

Evaluate whether a JEV policy would have:

- skipped redundant exploration;
- preferred a useful deterministic test over another model call;
- escalated earlier on ambiguous failures;
- stopped when additional evidence had low value.

Then run a small predeclared live treatment.

Compare against:

    always acquire all configured evidence
    fixed current policy
    always frontier model

Measure regret/quality/resource use separately.

---

# 23. Algorithms which are explicitly deferred until data justifies them

Do not implement by default:

- LambdaMART;
- neural ranking;
- GNN fault localization;
- graph transformers;
- learned community detection;
- RL scheduling;
- RL model routing;
- model-output semantic cache;
- giant embedding database.

Reconsider only when the repository contains a sufficiently large clean episode dataset and simple methods have demonstrably plateaued.

A practical trigger for learned ranking/routing should be at least hundreds of clean independently evaluated examples, preferably 500–1000+ relevant episodes.

Even then:

    learned method must beat the simple baseline materially on held-out tasks.

Otherwise retain the simpler system.

---

# 24. Donor-specific remaining work

Treat the old donor audit as research input, not a to-do checklist.

The only donor mechanisms with high remaining product value are approximately:

## Alibaba OpenCodeReview / Agentless / AutoCodeRover

Use for:

    structured findings
    grouping/localization
    repair specifications
    patch competition
    exact closure

This is highest priority.

## CoCoder / Pact / pi-subagent-tasks / CAID

Use for:

    coupling-aware decomposition
    resource-bounded waves
    idle-ready-task dispatch
    isolated implementation
    ordered integration

Do not copy unsafe lock stealing, shared-write authority or automatic “theirs” merge behavior.

## Aider / repomap

Use only for experiments around:

    compact orientation
    personalized graph ranking
    context budget allocation

Do not import mutable mtime/path caches or another graph/index runtime.

## ACRouter / Ultraswarm

Use for:

    empirical route calibration
    role-level model economics
    history-informed route selection

Do not adopt guessed monetary prices as observed cost.

## Bazel / Nix

Use only for:

    deterministic input closure
    content-addressed derived observations
    reproducibility reasoning

Do not build generic remote cache infrastructure.

## Serena

Use only for:

    ergonomic bounded symbol query vocabulary

over Fabric evidence.

## LongHorizon / OpenHands condenser

The key concepts are mostly already absorbed:

    durable engineering state
    fresh context
    compaction as derived view

Do not build a second memory subsystem.

## Symphony

Most valuable separation is already represented.

No broad Symphony port is required.

---

# 25. Work sequencing

Do not implement every subsystem concurrently.

Use the following order because later stages can consume evidence produced by earlier ones.

## Phase A — consolidate baseline

- inspect current v1.1.x state;
- mark stale donor/roadmap docs;
- remove documented false gaps;
- ensure v1.1 resilience behavior remains covered;
- qualify current agent-context behavior;
- establish clean profiling hooks and structured episode fields.

No new major subsystem before baseline is internally coherent.

## Phase B — Repair Intelligence

Implement:

    Finding
    localization
    RepairSpec
    closure oracle
    repair strategy routing
    targeted test integration

Use RRF and deterministic evidence first.

Add mutation/property testing only where appropriate.

This should be the first major capability because it attacks current autonomous failure directly.

## Phase C — Evidence Value Controller

Implement:

    evidence action representation
    resource-cost representation
    ordinal priors
    simple Bayesian reliability
    JEV calculation
    research stopping
    structured decision evidence

Initially use conservative approximate JEV.

It is acceptable for the first implementation to support only a finite set of evidence actions.

Do not build a general probabilistic programming framework.

## Phase D — Context Intelligence

Implement/evaluate:

    hierarchical localization
    independent rankers
    RRF
    PPR experiment
    budgeted submodular/gain-per-token selection

Integrate EVC/JEV for optional expensive evidence.

## Phase E — Topology scheduler

Implement/evaluate:

    typed coupling
    hypergraph representation where justified
    lexicographic CP-SAT/finite optimization
    resource-aware waves
    receding-horizon reoptimization
    dynamic ready-task assignment

Preserve current isolated writer/effect architecture.

## Phase F — Model economics

Start with:

    Bayesian route reliability
    quality-constrained cheapest admissible route
    JEV-driven escalation

Only then test contextual bandit routing.

## Phase G — broader language/test intelligence

Qualify:

    Rust
    TypeScript/JavaScript
    Python

Add external SCIP producers and useful test/mutation/property integrations.

## Phase H — measured efficiency

Profile before adding:

    action cache
    gix
    extra concurrency
    additional persistent indexes

Integrate only measured wins.

## Phase I — release closure

Freeze features.

No more architecture additions.

Run final evaluation, packaging, documentation and installed acceptance.

---

# 26. Performance engineering rules

Do not do blind optimization.

Profile real workloads.

Measure at least:

    wall time by stage
    CPU where meaningful
    allocation volume
    process RSS / peak memory where trustworthy
    RI startup/query time
    graph construction
    candidate capture/hash time
    context compilation
    journal replay
    tool queue wait
    verification setup
    verification execution
    provider wait
    typed token usage

Use:

    ordinary successful task
    repair task
    topology/parallel task
    larger repository task

Before integrating an optimization, state:

    observed hotspot
    proposed mechanism
    expected benefit
    preservation invariants
    measurement method

If improvement is not useful, remove/archive the candidate.

Do not keep optimization code because it was expensive to write.

---

# 27. Resilience / fault-injection matrix must remain a first-class release gate

v2 must prove that ordinary capability failures do not destroy viable runs.

Include deterministic tests for at least:

    missing optional RI
        -> bounded fallback

    semantic producer incomplete
        -> explicit PARTIAL/UNKNOWN, no false absence

    cache permission failure
        -> recompute

    safe derived cache corruption
        -> quarantine/recompute

    compaction unsupported
        -> continue

    optional review-impact context unavailable
        -> safe lower-context review

    isolated writers unavailable
        -> serial fallback

    memory pressure
        -> reduce future admission

    legal tool burst
        -> queued/bounded execution

    planner malformed output
        -> bounded correction with new invocation identity

    unapplied writer malformed proposal
        -> bounded proposal correction

    stale/missing anchor
        -> bounded correction, never implicit whole-file rewrite

    reviewer malformed verdict
        -> bounded correction

    legitimate ownership expansion
        -> candidate-bound replan

    repair budget exhausted
        -> retained NEEDS_ATTENTION

    capacity refusal before semantic result
        -> exact sealed failure plus permitted fresh successor policy

    optional usage/compaction metadata unavailable
        -> retain semantic result where policy allows

    required verification executable unavailable
        -> pre-provider failure

    UNKNOWN external effect
        -> reconcile without resend

    candidate/source/path substitution
        -> DENY

    authoritative journal corruption
        -> DENY/UNSAFE

No scenario may silently convert uncertainty into success.

---

# 28. v2 release definition

Do not tag v2.0 merely because all planned source code exists.

v2.0 requires product evidence.

## Source qualification

Exact final commit:

- clean repository;
- complete fresh Go test suite with test-result caching disabled where appropriate;
- applicable race/concurrency regressions;
- complete Rust workspace tests;
- SCIP/importer/RI integration tests;
- all PowerShell harnesses;
- formatting/vet/static checks;
- SonarCloud;
- GitGuardian/security;
- independent final review;
- legacy journal replay/compatibility tests.

No hidden local uncommitted qualification dependency.

## Resilience qualification

Complete the fault-injection matrix above.

## Repair Intelligence qualification

At least several real repair-requiring tasks must demonstrate:

    finding
    localization
    RepairSpec
    repaired candidate
    exact closure

with unchanged final acceptance.

## Topology qualification

At least one genuinely independent multi-writer case and one shared-hub/generator case must demonstrate correct scheduling/integration.

A speed claim requires a comparable measurement.

No speed claim is required merely to release if topology improves correctness/decomposition; document exactly what was proved.

## Context qualification

The promoted selector must demonstrate real downstream usefulness or at minimum no quality regression with meaningful context/resource benefit.

If PPR/heat/submodular treatment does not win, retain the simpler selector.

## Model-allocation qualification

The v2 release may ship conservative calibration even if no cheaper route becomes default.

Do not fake an adaptive-routing win.

If mixed-model routing is promoted, prove it on held-out tasks against a fixed high-quality baseline.

## Cross-language qualification

Before broadly presenting Fabric as language-general, execute real accepted tasks on at least:

    Go
    Rust
    TypeScript/JavaScript
    Python

using appropriate native project checks and candidate-bound review.

If this is not completed, state the supported/evaluated language set explicitly rather than blocking truth behind generic marketing.

## Canonical cohort

Run a fresh exact-final-candidate canonical cohort.

Report raw outcome.

Seek acceptance of every task identity.

Use only the predeclared provider-capacity successor policy for terminal pre-effect service refusals.

Never pool arbitrary historical successes.

## Installed acceptance

Build real installable artifacts from exact final source.

At minimum:

    Windows amd64
    Linux amd64

should receive actual installation and real-task qualification if feasible for v2.

If Linux live execution cannot be obtained, do not claim it.

macOS can remain explicitly unqualified unless real infrastructure is available.

## Reproducibility

Build release artifacts independently at least twice where the platform permits and compare exact bytes/hashes.

Publish:

    release metadata
    checksums
    qualification receipt
    source/tag identity
    known platform limits

Published v2.0.0 tag and release artifacts are immutable.

---

# 29. Product demonstration for v2

Prepare one concise end-to-end demonstration that shows why Fabric exists.

The demo should not be a trivial typo.

Choose a task with:

- multiple affected files;
- meaningful repository context;
- at least one dependency or generator/test relation;
- opportunity for targeted verification;
- at least one repair/review finding if naturally occurring;
- model/evidence decisions that can be inspected.

Show:

    objective
      ->
    selected engineering context
      ->
    task/dependency plan
      ->
    model/evidence decisions
      ->
    isolated implementation
      ->
    verification
      ->
    structured finding if needed
      ->
    repair
      ->
    review
      ->
    READY

Then show:

    fabric inspect
    fabric diff
    fabric usage
    fabric checkpoint

The demonstration should explain decisions from structured evidence rather than a giant transcript.

---

# 30. Public documentation quality

Before v2:

README must answer immediately:

    What is Fabric?
    Why would I use it instead of a coding agent directly?
    What does it do during a task?
    What does READY mean?
    What are its real limitations?
    How do I install and run it?

Do not make the README a research paper.

Keep advanced mathematics in architecture/research docs.

Create a concise architecture diagram showing:

    objective
      ->
    evidence/context
      ->
    task graph/scheduler
      ->
    model execution
      ->
    candidate
      ->
    verification/review
      ->
    repair intelligence
      ->
    READY

Document the Evidence Value Controller separately with equations and assumptions.

Clearly distinguish:

    measured
    configured
    estimated
    inferred
    unknown

Remove stale version references across current docs.

Historical evidence remains historical.

---

# 31. Explicit non-goals before v2.0

Do not add any of the following unless a demonstrated product blocker invalidates this decision:

- G0/self-hosting promotion machinery;
- infrastructure built merely so Fabric can more perfectly build Fabric;
- generic vector database;
- Neo4j/graph database as repository-intelligence core;
- generic agent DSL;
- generic workflow engine;
- generic provider gateway;
- dozens of provider adapters;
- autonomous browser/cloud platform;
- custom Docker/Kubernetes sandbox platform;
- generic semantic conversation memory;
- cross-run model-output cache;
- approval cache;
- swarm/voting theater;
- large learned routing/retrieval models without sufficient data;
- RL;
- GNN/graph transformer merely for prestige;
- multi-repository autonomy as a v2 blocker;
- UI/dashboard work before the CLI/runtime behavior is complete.

---

# 32. Engineering discipline

For every substantial capability:

1. Identify the real failure/opportunity.
2. Inspect current Fabric implementation.
3. Inspect relevant donor/source research where useful.
4. Prefer:
       dependency
       ->
       thin adapter
       ->
       bounded port
       ->
       bounded reimplementation.
5. Define the invariant boundary.
6. Add deterministic focused tests.
7. Run a real task/measurement when the claim concerns model behavior or performance.
8. Obtain independent review.
9. Fix Sonar/security issues in the same development increment.
10. Integrate only after evidence supports the capability.
11. Remove experimental code that does not demonstrate useful benefit.

Do not create process artifacts that do not improve product implementation or evidence.

Direct human/Codex edits remain allowed.

Fabric self-hosting is optional dogfooding.

---

# 33. Source history and publication

Respect the repository's current source-history policy.

Development belongs on `dev` with full incremental history.

The rolling `main` major/minor snapshot is synchronized according to the existing workflow/policy.

Do not manually fast-forward or merge dev lineage into main contrary to that policy.

Do not move existing published tags.

Release tags are immutable source identities.

Use the configured automation author/committer identity for automated development commits:

    automation:fabric-v1-sol-supervisor <automation@fabric.invalid>

Do not add human co-author trailers unless repository policy changes explicitly.

---

# 34. What “done” means

Fabric v2 is not done because the repository contains sophisticated algorithms.

It is done when the final system demonstrates this behavior:

A user gives Fabric a nontrivial software-engineering objective.

Fabric understands enough of the repository to identify relevant evidence without reading everything.

It gathers additional evidence only while the expected value of that information justifies the cost.

It decomposes the objective into tasks while distinguishing dependency, coupling and ownership.

It safely parallelizes genuinely independent work and serializes shared hubs or generators.

It chooses models according to empirically grounded reliability and resource constraints rather than always using the most expensive model.

It produces candidate changes under exact ownership.

When verification or review fails, it creates a structured diagnosis, localizes the likely cause, chooses an appropriate repair strategy and proves closure of the finding.

It preserves useful work through recoverable failures.

It never resends an uncertain external effect.

It does not claim success from a model assertion.

The final candidate reaches READY only when the exact configured verification and review gates pass against that candidate.

The entire engineering process is inspectable and reproducible enough to explain:

    what Fabric knew,
    what it did,
    why it chose that action,
    what it cost,
    what failed,
    what was repaired,
    and why the final candidate was accepted.

That is Fabric v2.0.

---

# 35. Final architectural target

The desired conceptual pipeline is:

                         OBJECTIVE
                            |
                            v
                  Evidence Value Controller
                  /         |          \
           cheap evidence   |       expensive evidence
                  \         |          /
                   dynamic JEV acquisition
                            |
                            v
             Hierarchical Repository Localization
          lexical + exact + SCIP + graph + failure evidence
                            |
                            v
                   Budgeted Context Set
                            |
                            v
                  Engineering Hypergraph
                            |
                  dependency / coupling /
                       ownership split
                            |
                            v
             Lexicographic Resource Scheduler
                    /                 \
              task wave A          task wave B
                    \                 /
                     model allocator
                    /   |      |      \
                 Muse  Luna   Sol   frontier
                    \   |      |      /
                       candidate
                            |
                            v
                   Native Verification
                            |
                +-----------+-----------+
                |                       |
              PASS                    FAILURE
                |                       |
                v                       v
              Review             Structured Finding
                |                       |
          +-----+-----+           Fault Localization
          |           |                 |
       APPROVE     FINDING              v
          |           |              RepairSpec
          |           +-----------------+
          |                             |
          |                     Strategy Selection
          |                    /        |         \
          |             deterministic localized   tournament
          |                    \        |         /
          |                     candidate repair
          |                             |
          |                      closure oracle
          |                             |
          +-----------------------------+
                        |
                        v
                      READY

JEV may also operate inside repair/model/context decisions.

Hard safety boundaries remain outside and above this optimization:

    identity
    authority
    effects
    UNKNOWN
    ownership
    verification truth

No adaptive mechanism may override them.

---

# 36. Final requirement to the implementing agent

Do not merely produce a new design document for this goal.

Carry the work through actual implementation and evidence in bounded increments.

At each point, challenge whether the proposed mechanism still makes sense given the current repository and measured results.

If a simpler mechanism performs as well, choose the simpler mechanism.

If a sophisticated mechanism fails to produce measured value, remove or defer it.

If new real failures reveal that this goal's design is wrong, update the design based on evidence rather than forcing the original idea.

The objective is not to implement every equation above.

The objective is to build the best practical Fabric v2 system that the evidence supports.

The resulting v2 should feel substantially more intelligent than v1.1 not because it contains more infrastructure, but because it makes better engineering decisions with less wasted work.
