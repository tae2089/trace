<!-- generated-by: code-context-graph docs -->
# context.go

> Context helpers for propagating trace IDs, fields, and cancellation causes into trace errors.

## Functions

### ContextWithTraceID
- **Lines:** 21–23
- **Intent:** attach a request-scoped trace identifier so downstream errors and logs can be correlated.
- **Mutates:** returns a derived context carrying the trace_id value.
ContextWithTraceID adds a trace ID to the context

### TraceIDFromContext
- **Lines:** 28–35
- **Intent:** retrieve the request-scoped trace identifier for correlation in logs and errors.
- **Ensures:**
  - returns an empty string when the context has no trace ID.
TraceIDFromContext retrieves the trace ID from context

### ContextWithFields
- **Lines:** 41–51
- **Intent:** accumulate request-scoped metadata that should be copied into later trace errors.
- **Domain Rules:**
  - new field values override existing keys from the parent context.
- **Mutates:** returns a derived context with a merged trace fields map.
ContextWithFields adds fields to the context for error enrichment
- **Calls:** FieldsFromContext

### ContextWithField
- **Lines:** 57–65
- **Intent:** add one request-scoped field so later trace wrapping can include it.
- **Domain Rules:**
  - the new value overrides an existing field with the same key.
- **Mutates:** returns a derived context with an updated trace fields map.
ContextWithField adds a single field to the context
- **Calls:** FieldsFromContext

### FieldsFromContext
- **Lines:** 71–82
- **Intent:** expose request-scoped trace metadata without leaking mutable context state.
- **Ensures:**
  - returns a defensive copy of stored fields when trace metadata exists.
FieldsFromContext retrieves fields from context.
Returns a copy of the fields to prevent external mutation of the context value.

### WrapContext
- **Lines:** 88–106
- **Intent:** combine ordinary failures with request trace metadata before they cross a boundary.
- **Domain Rules:**
  - trace_id and stored context fields are copied into the returned error when present.
- **Ensures:**
  - returns nil unchanged when no source error is provided.
WrapContext wraps an error with context information
- **Calls:** TraceIDFromContext, FieldsFromContext, Wrap, WithField, WithFields

### FromContext
- **Lines:** 115–154
- **Intent:** translate context cancellation and deadline signals into trace-aware error values.
- **Domain Rules:**
  - context.Cause is preferred so the original cancellation reason is preserved.
- **Ensures:**
  - captures the current call site and copies trace metadata from the context into the returned error.
  - returns nil when the context has not been canceled.
FromContext checks for context errors and wraps them appropriately.
Uses context.Cause (Go 1.20+) to capture the cancellation cause when available,
preserving the original reason for cancellation rather than just context.Canceled.
- **Calls:** contextFieldsToMap, contextFieldsToMap, contextFieldsToMap, Err, captureFrame

### contextFieldsToMap
- **Lines:** 158–169
- **Intent:** collect trace ID and contextual fields into one map ready to attach to a TraceError.
- **Ensures:**
  - returns a map containing trace metadata present on the context.
- **Calls:** TraceIDFromContext, FieldsFromContext

### IsCanceled
- **Lines:** 178–178
- **Intent:** advertise cancellation semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 182–182
- **Intent:** map cancellation failures to HTTP 499-style client-aborted responses.
HTTP 499 follows the nginx-style Client Closed Request convention.

### HTTPError
- **Lines:** 185–187
- **Intent:** expose only a fixed cancellation message to clients.

### Error
- **Lines:** 190–190
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 193–193
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsCanceled
- **Lines:** 205–214
- **Intent:** detect cancellation semantics anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsCanceled checks if an error is a cancellation error
- **Calls:** IsCanceled, As

### IsDeadlineExceeded
- **Lines:** 219–227
- **Intent:** detect deadline-expired failures across both context and trace timeout wrappers.
- **Ensures:**
  - returns false for nil errors.
IsDeadlineExceeded checks if an error is due to deadline exceeded
- **Calls:** IsTimeout

### NewContextualizer
- **Lines:** 238–240
- **Intent:** create a helper that consistently applies one context's trace metadata across multiple operations.
- **Ensures:**
  - returns a Contextualizer bound to the provided context.
NewContextualizer creates a new Contextualizer

### Wrap
- **Lines:** 245–247
- **Intent:** wrap an operation failure with the contextualizer's stored trace metadata.
- **Ensures:**
  - delegates to WrapContext using the contextualizer's context.
Wrap wraps an error with context information
- **Calls:** WrapContext

### Do
- **Lines:** 252–258
- **Intent:** run a function and automatically enrich any resulting error with the contextualizer's trace metadata.
- **Ensures:**
  - returns nil when the function succeeds.
Do executes a function and wraps any error with context
- **Calls:** WrapContext

### DoValue
- **Lines:** 263–269
- **Intent:** run a value-returning function and preserve its result while enriching any error with trace metadata.
- **Ensures:**
  - returns the function's value unchanged alongside a wrapped error when the call fails.
DoValue executes a function returning a value and wraps any error
- **Calls:** WrapContext

### CheckContext
- **Lines:** 274–281
- **Intent:** provide a cheap guard for aborting work when the context is already done.
- **Ensures:**
  - returns nil while the context is still usable and a traced context error otherwise.
CheckContext checks if context is still valid and returns error if not
- **Calls:** FromContext

### WrapIfContextDone
- **Lines:** 287–301
- **Intent:** preserve both the operation failure and any concurrent context cancellation signal.
- **Domain Rules:**
  - when the context is done, the returned error aggregates the original error with the context-derived failure.
- **Ensures:**
  - returns nil unchanged when no source error is provided.
WrapIfContextDone wraps the error with context info if context is done
- **Calls:** WrapContext, FromContext, Aggregate

### DetachedContext
- **Lines:** 308–310
- **Intent:** keep request metadata available for background work that must outlive request cancellation.
- **Domain Rules:**
  - inherited values are preserved while cancellation is detached from the parent.
DetachedContext returns a context that carries the parent's values
but is not canceled when the parent is canceled (Go 1.21+).
Useful for background cleanup or logging that should outlive the request.

### OnCancel
- **Lines:** 317–319
- **Intent:** register cleanup or follow-up work that should run when the contextualizer's context is canceled.
- **Side Effects:** schedules a callback with the underlying context cancellation machinery.
- **Ensures:**
  - returns a stop function that can prevent the callback before cancellation.
OnCancel registers fn to run after the context is canceled (Go 1.21+).
Returns a stop function that prevents fn from running if called before cancellation.

### WithCancelCause
- **Lines:** 325–327
- **Intent:** expose cancel-cause semantics through the trace package API so callers can preserve shutdown reasons.
- **Ensures:**
  - returned contexts can later surface their cause via context.Cause.
WithCancelCause returns a context with a CancelCauseFunc (Go 1.20+).
The cause can later be retrieved via context.Cause(ctx).
- **Calls:** WithCancelCause

### WithTimeoutCause
- **Lines:** 333–335
- **Intent:** create a timeout that preserves an explicit business cause for later error wrapping.
- **Ensures:**
  - the returned context is canceled after the deadline with the supplied cause.
WithTimeoutCause returns a context that is canceled after the given duration
with the specified cause error (Go 1.21+).
- **Calls:** WithTimeoutCause

## Classes

### CanceledError
- **Lines:** 173–175
- **Intent:** represent context cancellation as a typed trace error that still carries the original cause.
CanceledError represents a context cancellation error

### Contextualizer
- **Lines:** 231–233
- **Intent:** bundle one context's trace metadata and cancellation hooks for reuse across multiple operations.
Contextualizer wraps operations with context-aware error handling

## Types

### contextKey
- **Lines:** 11–11
- **Intent:** isolate trace-specific context values from unrelated context keys.

### ErrorCanceled
- **Lines:** 197–200
- **Intent:** let callers recognize cancellation semantics through behavior rather than concrete types.
ErrorCanceled is an interface for canceled errors
