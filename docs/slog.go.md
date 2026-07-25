<!-- generated-by: code-context-graph docs -->
# slog.go

> slog integration for serializing trace errors and automatically expanding error attributes in log records.

## Functions

### SlogError
- **Lines:** 15–38
- **Intent:** convert trace errors into a structured slog group that preserves message, cause, trace, and fields.
- **Domain Rules:**
  - plain errors degrade to a minimal message-only structure instead of losing log compatibility.
- **Ensures:**
  - returns an empty slog.Attr when no error is provided.
SlogError returns slog attributes for an error.
Schema matches TraceError.LogValue: {"message":..., "cause":..., "trace":[...], "fields":{...}}
- **Calls:** As, attrsToAny, errorCause, framesToSerializable

### attrsToAny
- **Lines:** 42–48
- **Intent:** adapt slog attribute slices to the variadic shape required by slog.Group.
- **Ensures:**
  - preserves attribute order in the returned slice.

### SlogErrorValue
- **Lines:** 53–64
- **Intent:** expose the trace-aware slog representation as a Value for callers building custom attributes.
- **Ensures:**
  - returns an empty string value when no error is provided.
SlogErrorValue returns a slog.Value for an error
- **Calls:** As

### errorCause
- **Lines:** 68–73
- **Intent:** extract the printable cause payload used in trace-aware slog serialization.
- **Ensures:**
  - returns nil when no error cause exists.

### NewErrorHandler
- **Lines:** 85–90
- **Intent:** wrap slog handlers so error attributes are consistently expanded into trace-aware structure.
- **Domain Rules:**
  - nil handlers fall back to slog.DiscardHandler for safe optional logging.
- **Ensures:**
  - always returns a non-nil ErrorHandler.
NewErrorHandler creates a new ErrorHandler wrapping the given handler

### Enabled
- **Lines:** 94–96
- **Intent:** defer level checks to the wrapped handler so logging enablement stays consistent.
- **Ensures:**
  - returns the wrapped handler's enabled decision unchanged.
- **Calls:** Enabled

### Handle
- **Lines:** 101–118
- **Intent:** rewrite incoming slog records so embedded error attributes use the trace logging schema.
- **Side Effects:** creates a replacement slog.Record and forwards it to the wrapped handler.
- **Ensures:**
  - non-error attributes are preserved verbatim.
- **Calls:** SlogError, Handle

### WithAttrs
- **Lines:** 123–125
- **Intent:** preserve trace-aware error rewriting when callers derive a handler with additional attributes.
- **Ensures:**
  - returns another ErrorHandler wrapping the derived handler.
WithAttrs returns a new handler with the given attributes
- **Calls:** WithAttrs

### WithGroup
- **Lines:** 130–132
- **Intent:** preserve trace-aware error rewriting when callers derive a grouped handler.
- **Ensures:**
  - returns another ErrorHandler wrapping the grouped handler.
WithGroup returns a new handler with the given group name
- **Calls:** WithGroup

### logAtLevel
- **Lines:** 136–146
- **Intent:** centralize level-specific error logging so all severities share the same trace serialization path.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** SlogError

### LogError
- **Lines:** 150–152
- **Intent:** log an error at error level using the trace serialization schema.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** logAtLevel

### LogWarn
- **Lines:** 156–158
- **Intent:** log an error at warn level using the trace serialization schema.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** logAtLevel

### LogDebug
- **Lines:** 162–164
- **Intent:** log an error at debug level using the trace serialization schema.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** logAtLevel

## Classes

### ErrorHandler
- **Lines:** 77–79
- **Intent:** adapt slog handlers so error attributes are rewritten into the package's trace-aware schema.
ErrorHandler wraps an slog.Handler to automatically extract trace information
