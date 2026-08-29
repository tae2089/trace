<!-- generated-by: code-context-graph docs -->
# trace.go

## Functions

### New
- **Lines:** 19–24

New returns a traced error with the given message and origin call site.
- **Intent:** create a traced error while keeping the public surface limited to ordinary error values.
- **Calls:** New, callerPC

### Errorf
- **Lines:** 29–36

Errorf returns a traced formatted error and preserves fmt.Errorf %w traversal.
- **Intent:** create a formatted traced error and preserve fmt.Errorf %w traversal semantics.
- **Calls:** Errorf, callerPC, unwrapCause

### Wrap
- **Lines:** 41–54

Wrap returns err with one additional context message and propagation call site.
- **Intent:** add one propagation context and call site without changing the wrapped error's meaning.
- **Calls:** callerPC

### Wrapf
- **Lines:** 59–72

Wrapf returns err with one formatted context message and propagation call site.
- **Intent:** add one formatted propagation context and call site around an existing cause.
- **Calls:** callerPC

### Error
- **Lines:** 83–88
- **Intent:** render the ordinary error message without exposing source locations.
- **Calls:** Error, Error

### Unwrap
- **Lines:** 91–93
- **Intent:** expose only the semantic cause used by standard errors traversal.

### Format
- **Lines:** 96–111
- **Intent:** keep ordinary fmt output clean while reserving call-site traces for %+v.
- **Calls:** Error, Error, Error, Error, writeDebug

### callerPC
- **Lines:** 114–120
- **Intent:** capture the external caller's program counter without resolving symbols on the hot path.

### writeDebug
- **Lines:** 129–134
- **Intent:** format the full human-readable trace tree for detailed debugging.
- **Calls:** writeErrorDebug

### writeErrorDebug
- **Lines:** 137–181
- **Intent:** walk standard single-cause and multi-cause error trees while preserving branch order.
- **Calls:** Error, Error, Error, Unwrap, Unwrap, writeErrorDebug, writeErrorDebug, writeTraceDebug, enter, isNonComparable, writeIndent, writeIndent, writeIndent, writeIndent

### writeTraceDebug
- **Lines:** 184–208
- **Intent:** print one Trace frame and continue into the recorded cause or leaf error.
- **Calls:** writeErrorDebug, writeErrorDebug, resolvePC, writeIndent, writeIndent

### unwrapCause
- **Lines:** 211–218
- **Intent:** keep Errorf as an origin when no %w cause exists while preserving %w chains when present.

### enter
- **Lines:** 221–240
- **Intent:** add comparable errors to the active path and count non-comparable path depth.
- **Calls:** isNonComparable

### isNonComparable
- **Lines:** 243–245
- **Intent:** identify values that cannot be used for precise identity-based cycle tracking.

### resolvePC
- **Lines:** 248–263
- **Intent:** resolve a stored program counter only when detailed formatting needs it.

### writeIndent
- **Lines:** 266–268
- **Intent:** keep nested debug output readable without making the format a machine API.

## Classes

### traceError
- **Lines:** 75–80
- **Intent:** carry one immutable trace frame plus the display error and traversable cause.

### debugState
- **Lines:** 123–126
- **Intent:** track the current traversal path so cyclic error trees do not loop forever.
