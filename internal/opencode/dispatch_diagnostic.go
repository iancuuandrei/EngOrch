package opencode

import (
	"context"
	"errors"
)

// DispatchDiagnostic is explanatory telemetry, never dispatch or retry authority.
// It carries no provider text, URL, request, response, or credentials.
type DispatchDiagnostic struct {
	Stage string `json:"stage"`
	Code  string `json:"code"`
}

type dispatchStageError struct {
	diagnostic DispatchDiagnostic
	cause      error
}

// Error reports bounded dispatch labels without exposing the underlying cause.
func (e *dispatchStageError) Error() string {
	return "OpenCode dispatch failed at " + e.diagnostic.Stage + " (" + e.diagnostic.Code + ")"
}

// Unwrap preserves the cause for errors.Is and errors.As without printing it.
func (e *dispatchStageError) Unwrap() error { return e.cause }

func dispatchFailure(stage string, cause error) error {
	code := "validation_or_runtime_failure"
	if errors.Is(cause, context.DeadlineExceeded) {
		code = "deadline_exceeded"
	} else if errors.Is(cause, context.Canceled) {
		code = "cancelled"
	} else if _, ok := TransportFailureFromError(cause); ok {
		code = "transport_failure"
	}
	return &dispatchStageError{DispatchDiagnostic{stage, code}, cause}
}

// DispatchDiagnosticFromError extracts only internally assigned, bounded labels.
func DispatchDiagnosticFromError(cause error) *DispatchDiagnostic {
	var failure *dispatchStageError
	if !errors.As(cause, &failure) {
		return nil
	}
	diagnostic := failure.diagnostic
	return &diagnostic
}
