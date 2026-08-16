<!-- generated-by: code-context-graph docs -->
# trace.go

> Core error wrapping, stack capture, and structured error inspection APIs.

## Functions

### resolve
- **Lines:** 31–46
- **Intent:** resolve the stored program counter into human-readable call-site data.
- **Ensures:**
  - returns zero values for a zero frame and an "unknown" function name
when symbol information is unavailable.

### Function
- **Lines:** 49–52
- **Intent:** expose the resolved function name for the recorded call site.
- **Calls:** resolve

### File
- **Lines:** 55–58
- **Intent:** expose the resolved file base name for the recorded call site.
- **Calls:** resolve

### Line
- **Lines:** 61–64
- **Intent:** expose the resolved line number for the recorded call site.
- **Calls:** resolve

### String
- **Lines:** 69–72
- **Intent:** render a single frame in a compact file-line-function format for debugging output.
- **Ensures:**
  - returns a string containing the file name, line number, and function name.
String returns a human-readable representation of the frame
- **Calls:** resolve

### MarshalJSON
- **Lines:** 77–84
- **Intent:** keep the frame's JSON wire shape stable while the in-memory layout stays a bare program counter.
- **Ensures:**
  - emits the {"function":...,"file":...,"line":...} object shape used by earlier versions.
MarshalJSON implements json.Marshaler.
- **Calls:** resolve

### String
- **Lines:** 93–110
- **Intent:** render the recorded stack trace in call-order for compact error messages.
- **Ensures:**
  - returns an empty string when no frames are present.
String returns a human-readable representation of all frames
- **Calls:** resolve, String

### Error
- **Lines:** 128–148
- **Intent:** produce the primary human-readable message for traced errors and their wrapped causes.
- **Ensures:**
  - includes frame and cause information when available.
Error implements the error interface
- **Calls:** Error

### Unwrap
- **Lines:** 153–155
- **Intent:** expose the wrapped cause so traced errors participate in standard Go error traversal.
- **Ensures:**
  - returns the original wrapped error.
Unwrap implements the errors.Unwrap interface for Go 1.13+

### Format
- **Lines:** 160–184
- **Intent:** support concise and verbose formatting styles for traced errors.
- **Ensures:**
  - %+v output includes stack frames and structured fields when present.
Format implements fmt.Formatter for customizable output
- **Calls:** Error, Error, Error, Error

### LogValue
- **Lines:** 190–208
- **Intent:** serialize traced errors into structured slog fields without losing cause or stack context.
- **Ensures:**
  - includes message, cause, trace, and fields only when those values are present.
LogValue implements slog.LogValuer for structured logging.
Schema: {"message":..., "cause":..., "trace":[{"file":..., "line":..., "func":...}], "fields":{...}}
- **Calls:** Error, framesToSerializable

### framesToSerializable
- **Lines:** 212–223
- **Intent:** convert recorded frames into a JSON-friendly structure for structured logging output.
- **Ensures:**
  - preserves file, line, and function data for each frame.
- **Calls:** resolve

### CaptureFrame
- **Lines:** 233–242
- **Intent:** capture the caller information that anchors trace output to a concrete source location.
- **Ensures:**
  - returns an empty frame when runtime caller information is unavailable.
CaptureFrame captures a single stack frame at the given skip level.

### Wrap
- **Lines:** 252–273
- **Intent:** preserve the original error while adding call-site debugging context.
- **Domain Rules:**
  - typed errors stay discoverable through the Err chain for errors.Is and errors.As.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
  - returns nil unchanged when no source error is provided.
Wrap wraps an error with stack trace information.
If err is nil, Wrap returns nil.
Wrap always creates a new TraceError, preserving the original error
(including typed errors like NotFoundError) in the Err field.
- **Calls:** CaptureFrame, findTraceError

### Wrapf
- **Lines:** 279–284
- **Intent:** preserve the original error while adding formatted debugging context.
- **Domain Rules:**
  - typed errors stay discoverable through the Err chain for errors.Is and errors.As.
- **Ensures:**
  - returns nil unchanged when no source error is provided.
Wrapf wraps an error with stack trace and a formatted message.
- **Calls:** CaptureFrame, wrapInternal

### wrapInternal
- **Lines:** 288–299
- **Intent:** share the common TraceError wrapping path used by formatted and typed error helpers.
- **Ensures:**
  - prepends the supplied frame to any existing trace frames.
- **Calls:** findTraceError

### WrapWithFields
- **Lines:** 306–316
- **Intent:** enrich an error with structured diagnostics that can flow into logs and HTTP responses.
- **Domain Rules:**
  - field attachment must not discard the wrapped error chain.
- **Mutates:** adds key-value metadata to the returned TraceError fields map.
- **Ensures:**
  - returns nil unchanged when no source error is provided.
WrapWithFields wraps an error with stack trace and structured fields
- **Calls:** Wrap, copyFields

### New
- **Lines:** 322–328
- **Intent:** create a fresh traceable application error at the current call site.
- **Ensures:**
  - records the current call site as the first trace frame.
  - allocates structured fields storage lazily on first field attachment.
New creates a new error with stack trace
- **Calls:** CaptureFrame

### Errorf
- **Lines:** 334–340
- **Intent:** create a traceable error from formatted application context.
- **Ensures:**
  - records the current call site as the first trace frame.
  - allocates structured fields storage lazily on first field attachment.
Errorf creates a new error with formatted message and stack trace
- **Calls:** CaptureFrame

### chainAs
- **Lines:** 346–370
- **Intent:** walk an error tree with plain type assertions so hot paths avoid the reflection cost of errors.As.
- **Domain Rules:**
  - matches errors.As semantics: direct match first, then a custom As(any) bool hook,
then single-cause and aggregate traversal in order.
- **Ensures:**
  - returns false without touching target when no match exists in the tree.
- **Calls:** As, Unwrap, Unwrap, chainAs

### findTraceError
- **Lines:** 376–402
- **Intent:** locate the nearest TraceError without the reflection cost of errors.As.
- **Domain Rules:**
  - kept non-generic: the generic chainAs pays a dictionary-based type-assertion
cost per link that this monomorphic loop avoids on the construction hot path.
- **Ensures:**
  - returns nil when the chain carries no TraceError.
- **Calls:** As, Unwrap, Unwrap, findTraceError

### formatMessage
- **Lines:** 407–424
- **Intent:** normalize mixed message-and-arguments inputs into one human-readable error message.
- **Ensures:**
  - returns an empty string when no message arguments are supplied.
formatMessage formats message and args similar to fmt.Sprintf

### GetFrames
- **Lines:** 430–437
- **Intent:** expose captured stack frames for diagnostics without leaking mutable internal state.
- **Ensures:**
  - returns a defensive copy of the stored frames when trace data exists.
GetFrames extracts frames from an error if available.
Returns a copy of the frames to prevent external mutation.
- **Calls:** findTraceError

### GetFields
- **Lines:** 443–448
- **Intent:** expose structured trace metadata for logging, transport, or inspection layers.
- **Ensures:**
  - returns a defensive copy of the stored fields when trace data exists.
GetFields extracts fields from an error if available.
Returns a copy of the fields to prevent external mutation.
- **Calls:** findTraceError, copyFields

### WithField
- **Lines:** 456–474
- **Intent:** attach one diagnostic attribute without mutating the original error instance.
- **Domain Rules:**
  - built-in trace wrapper types are preserved when replacing the inner TraceError.
- **Mutates:** adds or replaces a single field on the returned error copy.
- **Ensures:**
  - returns nil unchanged when no source error is provided.
WithField adds a field to the error for structured logging.
It returns a new wrapper error with the field added, preserving the original error immutably.
- **Calls:** Wrap, findTraceError, cloneTraceError, replaceTraceError, copyFields

### WithFields
- **Lines:** 480–500
- **Intent:** attach multiple diagnostic attributes without mutating the original error instance.
- **Domain Rules:**
  - later field values override earlier values for the same key.
- **Mutates:** merges the provided fields into the returned error copy.
- **Ensures:**
  - returns nil unchanged when no source error is provided.
- **Calls:** Wrap, findTraceError, cloneTraceError, replaceTraceError, copyFields, copyFields

### cloneTraceError
- **Lines:** 504–511
- **Intent:** duplicate a TraceError so field updates can preserve immutable error semantics.
- **Ensures:**
  - returns a copy with duplicated frames and fields.
- **Calls:** copyFields

### replaceTraceError
- **Lines:** 527–586
- **Intent:** swap the inner TraceError while preserving known wrapper types around it.
- **Domain Rules:**
  - built-in typed wrappers are recreated so errors.Is and errors.As continue to work.

### copyFields
- **Lines:** 590–596
- **Intent:** defensively copy structured error fields before mutation or external exposure.
- **Ensures:**
  - returns a new map containing every existing field.

### DebugReport
- **Lines:** 602–614
- **Intent:** render a full developer-facing report for nested and aggregated error chains.
- **Domain Rules:**
  - aggregate errors must include every branch in the rendered report.
- **Ensures:**
  - returns an empty string when no error is provided.
DebugReport returns a detailed report of the error chain
- **Calls:** debugReportWalk

### debugReportWalk
- **Lines:** 618–653
- **Intent:** recursively expand an error tree into the developer-facing debug report.
- **Ensures:**
  - traverses both single-cause chains and aggregate branches.
- **Calls:** Error, Unwrap, Unwrap, debugReportWalk, debugReportWalk

### UserMessage
- **Lines:** 659–674
- **Intent:** extract the safest high-level message to show outside debugging channels.
- **Domain Rules:**
  - prefer explicit TraceError messages before falling back to wrapped causes.
- **Ensures:**
  - returns an empty string when no error is provided.
UserMessage returns a user-friendly error message without stack traces
- **Calls:** Error, findTraceError, UserMessage

### Errors
- **Lines:** 682–686
- **Intent:** iterate every error reachable from wrapped and aggregated trace errors.
- **Domain Rules:**
  - aggregate branches are traversed recursively, not flattened into a single message.
- **Ensures:**
  - yields each reachable error once per traversal path until the consumer stops.
Errors returns an iterator over the error chain (Go 1.23+).
It yields each error in the chain by following Unwrap() error and
recursively traversing Unwrap() []error (e.g., AggregateError).
- **Calls:** errorsWalk

### errorsWalk
- **Lines:** 690–706
- **Intent:** recursively traverse every reachable error in a chain or aggregate until the consumer stops.
- **Ensures:**
  - returns false as soon as the yield function asks traversal to stop.
- **Calls:** Unwrap, Unwrap, errorsWalk, errorsWalk

## Classes

### Frame
- **Lines:** 24–26
- **Intent:** describe one recorded call site so errors and logs can point back to their origin.
- **Domain Rules:**
  - only the program counter is stored; symbol resolution is deferred to render
time because errors are created far more often than they are printed.
Frame represents a single stack frame. It records the call site as a
program counter and resolves the function name, file, and line lazily.

### TraceError
- **Lines:** 114–123
- **Intent:** serve as the canonical trace-aware error wrapper carrying cause, message, frames, and structured fields.
TraceError is the core error type that captures stack traces

## Types

### Frames
- **Lines:** 88–88
- **Intent:** represent an ordered stack trace that can be rendered or serialized with an error.
Frames is a slice of stack frames

### TraceErrorReplacer
- **Lines:** 521–523
- **Intent:** let custom wrapper types survive field updates instead of being peeled away.
- **Domain Rules:**
  - the wrapper must return ok=false when original is not its direct inner TraceError.
TraceErrorReplacer lets wrapper types outside this package survive WithField
and WithFields. When those functions rebuild an error chain they replace the
inner *TraceError; wrappers this package does not know are otherwise dropped.
A wrapper that implements this method is asked to rebuild itself around
replacement and report ok=true, or ok=false if original is not its inner
TraceError.
