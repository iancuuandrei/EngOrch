// Package taskcontext selects a small, deterministic, task-relevant text view
// from caller-admitted repository bytes. It has no filesystem, Git, runtime,
// provider, or candidate authority. Callers must first bind its Input.Scope and
// Files to an immutable source or a reobserved candidate.
//
// Empty Selector preserves exact legacy weighted selection. The experimental
// Selector "rrf-coverage-v1" fuses independent ordinal rankers with equal
// unweighted Reciprocal Rank Fusion (k=60, fixed provenance from
// Cormack/Clarke/Buettcher SIGIR 2009) and bounded diminishing-return
// coverage over visible excerpts. Pure selectors grant no authority.
package taskcontext
