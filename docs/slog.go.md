<!-- generated-by: code-context-graph docs -->
# slog.go

> slog integration for serializing trace errors and automatically expanding error attributes in log records.

## Functions

### SlogError
- **Lines:** 14–36
- **Intent:** convert trace errors into a structured slog group that preserves message, cause, trace, and fields.
- **Domain Rules:**
  - plain errors degrade to a minimal message-only structure instead of losing log compatibility.
- **Ensures:**
  - returns an empty slog.Attr when no error is provided.
SlogError returns slog attributes for an error.
Schema matches TraceError.LogValue: {"message":..., "cause":..., "trace":[...], "fields":{...}}
- **Calls:** attrsToAny, errorCause, framesToSerializable, findTraceError

### attrsToAny
- **Lines:** 40–46
- **Intent:** adapt slog attribute slices to the variadic shape required by slog.Group.
- **Ensures:**
  - preserves attribute order in the returned slice.

### SlogErrorValue
- **Lines:** 51–61
- **Intent:** expose the trace-aware slog representation as a Value for callers building custom attributes.
- **Ensures:**
  - returns an empty string value when no error is provided.
SlogErrorValue returns a slog.Value for an error
- **Calls:** findTraceError

### errorCause
- **Lines:** 65–70
- **Intent:** extract the printable cause payload used in trace-aware slog serialization.
- **Ensures:**
  - returns nil when no error cause exists.

### NewErrorHandler
- **Lines:** 82–87
- **Intent:** wrap slog handlers so error attributes are consistently expanded into trace-aware structure.
- **Domain Rules:**
  - nil handlers fall back to slog.DiscardHandler for safe optional logging.
- **Ensures:**
  - always returns a non-nil ErrorHandler.
NewErrorHandler creates a new ErrorHandler wrapping the given handler

### Enabled
- **Lines:** 91–93
- **Intent:** defer level checks to the wrapped handler so logging enablement stays consistent.
- **Ensures:**
  - returns the wrapped handler's enabled decision unchanged.
- **Calls:** Enabled

### Handle
- **Lines:** 98–115
- **Intent:** rewrite incoming slog records so embedded error attributes use the trace logging schema.
- **Side Effects:** creates a replacement slog.Record and forwards it to the wrapped handler.
- **Ensures:**
  - non-error attributes are preserved verbatim.
- **Calls:** SlogError, Handle

### WithAttrs
- **Lines:** 120–122
- **Intent:** preserve trace-aware error rewriting when callers derive a handler with additional attributes.
- **Ensures:**
  - returns another ErrorHandler wrapping the derived handler.
WithAttrs returns a new handler with the given attributes
- **Calls:** WithAttrs

### WithGroup
- **Lines:** 127–129
- **Intent:** preserve trace-aware error rewriting when callers derive a grouped handler.
- **Ensures:**
  - returns another ErrorHandler wrapping the grouped handler.
WithGroup returns a new handler with the given group name
- **Calls:** WithGroup

### logAtLevel
- **Lines:** 133–143
- **Intent:** centralize level-specific error logging so all severities share the same trace serialization path.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** SlogError

### LogError
- **Lines:** 147–149
- **Intent:** log an error at error level using the trace serialization schema.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** logAtLevel

### LogWarn
- **Lines:** 153–155
- **Intent:** log an error at warn level using the trace serialization schema.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** logAtLevel

### LogDebug
- **Lines:** 159–161
- **Intent:** log an error at debug level using the trace serialization schema.
- **Ensures:**
  - does nothing when the input error is nil.
- **Calls:** logAtLevel

## Classes

### ErrorHandler
- **Lines:** 74–76
- **Intent:** adapt slog handlers so error attributes are rewritten into the package's trace-aware schema.
ErrorHandler wraps an slog.Handler to automatically extract trace information
