// Package gitexec resolves Git through fixed conventional installation paths
// or an explicit absolute operator override. It never searches PATH, and a
// resolution failure is returned before any subprocess starts.
package gitexec
