// Package engineeringplan defines the bounded, inspectable task graph used by
// an autonomous engineering run. It is deliberately independent of execution
// and journaling: callers bind an accepted graph to their own durable run.
package engineeringplan
