package trace

import (
	"context"
	"errors"
	"time"
)

type contextKey string

const (
	traceIDKey     contextKey = "trace_id"
	traceFieldsKey contextKey = "trace_fields"
)

// ContextWithTraceID adds a trace ID to the context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext retrieves the trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ContextWithFields adds fields to the context for error enrichment
func ContextWithFields(ctx context.Context, fields map[string]any) context.Context {
	existing := FieldsFromContext(ctx)
	merged := make(map[string]any, len(existing)+len(fields))
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, traceFieldsKey, merged)
}

// ContextWithField adds a single field to the context
func ContextWithField(ctx context.Context, key string, value any) context.Context {
	fields := FieldsFromContext(ctx)
	newFields := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		newFields[k] = v
	}
	newFields[key] = value
	return context.WithValue(ctx, traceFieldsKey, newFields)
}

// FieldsFromContext retrieves fields from context.
// Returns a copy of the fields to prevent external mutation of the context value.
func FieldsFromContext(ctx context.Context) map[string]any {
	if v := ctx.Value(traceFieldsKey); v != nil {
		if m, ok := v.(map[string]any); ok {
			cp := make(map[string]any, len(m))
			for k, v := range m {
				cp[k] = v
			}
			return cp
		}
	}
	return nil
}

// WrapContext wraps an error with context information
func WrapContext(ctx context.Context, err error, msg ...string) error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, msg...)

	// Add trace ID if present
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		wrapped = WithField(wrapped, "trace_id", traceID)
	}

	// Add context fields
	if fields := FieldsFromContext(ctx); fields != nil {
		wrapped = WithFields(wrapped, fields)
	}

	return wrapped
}

// FromContext checks for context errors and wraps them appropriately.
// Uses context.Cause (Go 1.20+) to capture the cancellation cause when available,
// preserving the original reason for cancellation rather than just context.Canceled.
func FromContext(ctx context.Context) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}

	frame := captureFrame(2)

	// Prefer context.Cause over ctx.Err() — it carries the original cancellation reason
	inner := context.Cause(ctx)
	if inner == nil {
		inner = err
	}

	if errors.Is(err, context.Canceled) {
		return &CanceledError{
			TraceError: &TraceError{
				Err:    inner,
				Frames: Frames{frame},
				Fields: contextFieldsToMap(ctx),
			},
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &TimeoutError{
			TraceError: &TraceError{
				Err:    inner,
				Frames: Frames{frame},
				Fields: contextFieldsToMap(ctx),
			},
		}
	}

	return &TraceError{
		Err:    inner,
		Frames: Frames{frame},
		Fields: contextFieldsToMap(ctx),
	}
}

func contextFieldsToMap(ctx context.Context) map[string]any {
	fields := make(map[string]any)
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		fields["trace_id"] = traceID
	}
	if ctxFields := FieldsFromContext(ctx); ctxFields != nil {
		for k, v := range ctxFields {
			fields[k] = v
		}
	}
	return fields
}

// CanceledError represents a context cancellation error
type CanceledError struct {
	*TraceError
}

func (e *CanceledError) IsCanceled() bool    { return true }
func (e *CanceledError) HTTPStatusCode() int { return 499 } // Client Closed Request (nginx convention)
func (e *CanceledError) Error() string       { return e.TraceError.Error() }
func (e *CanceledError) Unwrap() error       { return e.TraceError }

// ErrorCanceled is an interface for canceled errors
type ErrorCanceled interface {
	error
	IsCanceled() bool
}

// IsCanceled checks if an error is a cancellation error
func IsCanceled(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorCanceled
	if errors.As(err, &e) && e.IsCanceled() {
		return true
	}
	return errors.Is(err, context.Canceled)
}

// IsDeadlineExceeded checks if an error is due to deadline exceeded
func IsDeadlineExceeded(err error) bool {
	if err == nil {
		return false
	}
	if IsTimeout(err) {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}

// Contextualizer wraps operations with context-aware error handling
type Contextualizer struct {
	ctx context.Context
}

// NewContextualizer creates a new Contextualizer
func NewContextualizer(ctx context.Context) *Contextualizer {
	return &Contextualizer{ctx: ctx}
}

// Wrap wraps an error with context information
func (c *Contextualizer) Wrap(err error, msg ...string) error {
	return WrapContext(c.ctx, err, msg...)
}

// Do executes a function and wraps any error with context
func (c *Contextualizer) Do(fn func() error) error {
	err := fn()
	if err != nil {
		return WrapContext(c.ctx, err)
	}
	return nil
}

// DoValue executes a function returning a value and wraps any error
func DoValue[T any](c *Contextualizer, fn func() (T, error)) (T, error) {
	value, err := fn()
	if err != nil {
		return value, WrapContext(c.ctx, err)
	}
	return value, nil
}

// CheckContext checks if context is still valid and returns error if not
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return FromContext(ctx)
	default:
		return nil
	}
}

// WrapIfContextDone wraps the error with context info if context is done
func WrapIfContextDone(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// Check if context is done
	select {
	case <-ctx.Done():
		// Context is done, add context error info
		ctxErr := FromContext(ctx)
		return Aggregate(err, ctxErr)
	default:
		return WrapContext(ctx, err)
	}
}

// DetachedContext returns a context that carries the parent's values
// but is not canceled when the parent is canceled (Go 1.21+).
// Useful for background cleanup or logging that should outlive the request.
func DetachedContext(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}

// OnCancel registers fn to run after the context is canceled (Go 1.21+).
// Returns a stop function that prevents fn from running if called before cancellation.
func (c *Contextualizer) OnCancel(fn func()) func() bool {
	return context.AfterFunc(c.ctx, fn)
}

// WithCancelCause returns a context with a CancelCauseFunc (Go 1.20+).
// The cause can later be retrieved via context.Cause(ctx).
func WithCancelCause(parent context.Context) (context.Context, context.CancelCauseFunc) {
	return context.WithCancelCause(parent)
}

// WithTimeoutCause returns a context that is canceled after the given duration
// with the specified cause error (Go 1.21+).
func WithTimeoutCause(parent context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parent, d, cause)
}
