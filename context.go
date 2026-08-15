// @index Context helpers for propagating trace IDs, fields, and cancellation causes into trace errors.
package trace

import (
	"context"
	"errors"
	"time"
)

// @intent isolate trace-specific context values from unrelated context keys.
type contextKey string

const (
	traceIDKey     contextKey = "trace_id"
	traceFieldsKey contextKey = "trace_fields"
)

// @intent attach a request-scoped trace identifier so downstream errors and logs can be correlated.
// @mutates returns a derived context carrying the trace_id value.
// ContextWithTraceID adds a trace ID to the context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// @intent retrieve the request-scoped trace identifier for correlation in logs and errors.
// @ensures returns an empty string when the context has no trace ID.
// TraceIDFromContext retrieves the trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// @intent accumulate request-scoped metadata that should be copied into later trace errors.
// @domainRule new field values override existing keys from the parent context.
// @mutates returns a derived context with a merged trace fields map.
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

// @intent add one request-scoped field so later trace wrapping can include it.
// @domainRule the new value overrides an existing field with the same key.
// @mutates returns a derived context with an updated trace fields map.
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

// @intent expose request-scoped trace metadata without leaking mutable context state.
// @ensures returns a defensive copy of stored fields when trace metadata exists.
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

// @intent combine ordinary failures with request trace metadata before they cross a boundary.
// @domainRule trace_id and stored context fields are copied into the returned error when present.
// @ensures returns nil unchanged when no source error is provided.
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

// @intent translate context cancellation and deadline signals into trace-aware error values.
// @domainRule context.Cause is preferred so the original cancellation reason is preserved.
// @ensures captures the current call site and copies trace metadata from the context into the returned error.
// @ensures returns nil when the context has not been canceled.
// FromContext checks for context errors and wraps them appropriately.
// Uses context.Cause (Go 1.20+) to capture the cancellation cause when available,
// preserving the original reason for cancellation rather than just context.Canceled.
func FromContext(ctx context.Context) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}

	frame := CaptureFrame(2)

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

// @intent collect trace ID and contextual fields into one map ready to attach to a TraceError.
// @ensures returns a map containing trace metadata present on the context.
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

// @intent represent context cancellation as a typed trace error that still carries the original cause.
// CanceledError represents a context cancellation error
type CanceledError struct {
	*TraceError
}

// @intent advertise cancellation semantics for behavior-based error checks.
func (e *CanceledError) IsCanceled() bool { return true }

// @intent wrap an existing cancellation cause as a typed trace error.
// @domainRule returns nil unchanged when the source error is nil.
// @ensures records the current call site as the first trace frame.
// Canceled wraps err as a CanceledError.
func Canceled(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	return &CanceledError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), CaptureFrame(2)),
	}
}

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *CanceledError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *CanceledError) Unwrap() error { return e.TraceError }

// @intent let callers recognize cancellation semantics through behavior rather than concrete types.
// ErrorCanceled is an interface for canceled errors
type ErrorCanceled interface {
	error
	IsCanceled() bool
}

// @intent detect cancellation semantics anywhere in an error chain.
// @ensures returns false for nil errors.
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

// @intent detect deadline-expired failures across both context and trace timeout wrappers.
// @ensures returns false for nil errors.
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

// @intent bundle one context's trace metadata and cancellation hooks for reuse across multiple operations.
// Contextualizer wraps operations with context-aware error handling
type Contextualizer struct {
	ctx context.Context
}

// @intent create a helper that consistently applies one context's trace metadata across multiple operations.
// @ensures returns a Contextualizer bound to the provided context.
// NewContextualizer creates a new Contextualizer
func NewContextualizer(ctx context.Context) *Contextualizer {
	return &Contextualizer{ctx: ctx}
}

// @intent wrap an operation failure with the contextualizer's stored trace metadata.
// @ensures delegates to WrapContext using the contextualizer's context.
// Wrap wraps an error with context information
func (c *Contextualizer) Wrap(err error, msg ...string) error {
	return WrapContext(c.ctx, err, msg...)
}

// @intent run a function and automatically enrich any resulting error with the contextualizer's trace metadata.
// @ensures returns nil when the function succeeds.
// Do executes a function and wraps any error with context
func (c *Contextualizer) Do(fn func() error) error {
	err := fn()
	if err != nil {
		return WrapContext(c.ctx, err)
	}
	return nil
}

// @intent run a value-returning function and preserve its result while enriching any error with trace metadata.
// @ensures returns the function's value unchanged alongside a wrapped error when the call fails.
// DoValue executes a function returning a value and wraps any error
func DoValue[T any](c *Contextualizer, fn func() (T, error)) (T, error) {
	value, err := fn()
	if err != nil {
		return value, WrapContext(c.ctx, err)
	}
	return value, nil
}

// @intent provide a cheap guard for aborting work when the context is already done.
// @ensures returns nil while the context is still usable and a traced context error otherwise.
// CheckContext checks if context is still valid and returns error if not
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return FromContext(ctx)
	default:
		return nil
	}
}

// @intent preserve both the operation failure and any concurrent context cancellation signal.
// @domainRule when the context is done, the returned error aggregates the original error with the context-derived failure.
// @ensures returns nil unchanged when no source error is provided.
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

// @intent keep request metadata available for background work that must outlive request cancellation.
// @domainRule inherited values are preserved while cancellation is detached from the parent.
// DetachedContext returns a context that carries the parent's values
// but is not canceled when the parent is canceled (Go 1.21+).
// Useful for background cleanup or logging that should outlive the request.
func DetachedContext(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}

// @intent register cleanup or follow-up work that should run when the contextualizer's context is canceled.
// @sideEffect schedules a callback with the underlying context cancellation machinery.
// @ensures returns a stop function that can prevent the callback before cancellation.
// OnCancel registers fn to run after the context is canceled (Go 1.21+).
// Returns a stop function that prevents fn from running if called before cancellation.
func (c *Contextualizer) OnCancel(fn func()) func() bool {
	return context.AfterFunc(c.ctx, fn)
}

// @intent expose cancel-cause semantics through the trace package API so callers can preserve shutdown reasons.
// @ensures returned contexts can later surface their cause via context.Cause.
// WithCancelCause returns a context with a CancelCauseFunc (Go 1.20+).
// The cause can later be retrieved via context.Cause(ctx).
func WithCancelCause(parent context.Context) (context.Context, context.CancelCauseFunc) {
	return context.WithCancelCause(parent)
}

// @intent create a timeout that preserves an explicit business cause for later error wrapping.
// @ensures the returned context is canceled after the deadline with the supplied cause.
// WithTimeoutCause returns a context that is canceled after the given duration
// with the specified cause error (Go 1.21+).
func WithTimeoutCause(parent context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parent, d, cause)
}
