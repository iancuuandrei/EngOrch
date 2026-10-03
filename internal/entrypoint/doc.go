// Package entrypoint documents the shared native CLI process bootstrap.
package entrypoint

// This package owns the small signal, telemetry, working-directory, CLI and
// error-reporting sequence shared by the fabric and harness binaries. Each
// command keeps its own stderr prefix and exit behavior while delegating the
// identical lifecycle to Run so future process handling stays in one place.
