// Package taskcontext selects a small, deterministic, task-relevant text view
// from caller-admitted repository bytes. It has no filesystem, Git, runtime,
// provider, or candidate authority. Callers must first bind its Input.Scope and
// Files to an immutable source or a reobserved candidate.
package taskcontext
