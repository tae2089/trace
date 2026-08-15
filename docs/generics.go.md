<!-- generated-by: code-context-graph docs -->
# generics.go

> Generic Result and Pipeline helpers for composing trace-aware success and failure flows.

## Functions

### Ok
- **Lines:** 18–20
- **Intent:** wrap a successful value in the Result abstraction without adding trace overhead.
- **Ensures:**
  - returns a Result with no error.
Ok creates a successful Result

### Err
- **Lines:** 25–27
- **Intent:** convert a failure into a Result while preserving trace metadata for downstream composition.
- **Ensures:**
  - wraps the provided error with trace context before storing it.
Err creates a failed Result
- **Calls:** Wrap

### ErrMsg
- **Lines:** 32–41
- **Intent:** create a failed Result directly from an application message when no underlying error exists.
- **Ensures:**
  - records the current call site as the first trace frame.
ErrMsg creates a failed Result with a message
- **Calls:** CaptureFrame

### IsOk
- **Lines:** 46–48
- **Intent:** let callers branch on success without unpacking the Result payload.
- **Ensures:**
  - returns true only when the Result holds no error.
IsOk returns true if the Result contains a value

### IsErr
- **Lines:** 53–55
- **Intent:** let callers branch on failure without unpacking the Result payload.
- **Ensures:**
  - returns true only when the Result holds an error.
IsErr returns true if the Result contains an error

### Unwrap
- **Lines:** 60–65
- **Intent:** extract the successful value in contexts where failure should abort immediately.
- **Domain Rules:**
  - panics with the stored error when the Result is failed.
Unwrap returns the value or panics if there's an error

### UnwrapOr
- **Lines:** 70–75
- **Intent:** provide a fallback value when the Result is failed.
- **Ensures:**
  - returns the stored value on success and the provided default on failure.
UnwrapOr returns the value or the provided default

### UnwrapOrElse
- **Lines:** 80–85
- **Intent:** derive a fallback value from the failure reason when the Result is failed.
- **Ensures:**
  - calls the fallback function only when the Result holds an error.
UnwrapOrElse returns the value or calls the function to get a default

### Value
- **Lines:** 90–92
- **Intent:** bridge Result back into Go's conventional value-plus-error calling style.
- **Ensures:**
  - returns the stored value and stored error unchanged.
Value returns the value and error separately (Go-style)

### Error
- **Lines:** 97–99
- **Intent:** expose the failure component directly for interoperability with Go error handling.
- **Ensures:**
  - returns nil for successful results.
Error returns the error if present

### Map
- **Lines:** 104–109
- **Intent:** transform successful results while leaving failures untouched.
- **Ensures:**
  - preserves the existing error without calling fn when the Result is failed.
Map transforms the value if present

### MapErr
- **Lines:** 114–119
- **Intent:** rewrite failure values without changing successful payloads.
- **Ensures:**
  - leaves successful results unchanged.
MapErr transforms the error if present

### FlatMap
- **Lines:** 124–129
- **Intent:** chain operations that already return Result values without manual error branching.
- **Ensures:**
  - skips fn and preserves the existing error when the Result is failed.
FlatMap chains Result-returning operations

### Try
- **Lines:** 134–146
- **Intent:** lift ordinary Go return pairs into a Result for composable error handling.
- **Ensures:**
  - wraps non-nil errors with trace context before storing them.
Try wraps a function call that returns (T, error) into a Result
- **Calls:** CaptureFrame

### Must
- **Lines:** 151–153
- **Intent:** provide concise success-only access in contexts where failure should panic.
- **Domain Rules:**
  - panics when the Result contains an error.
Must unwraps a Result, panicking on error
- **Calls:** Unwrap

### MustValue
- **Lines:** 158–163
- **Intent:** collapse a Go-style value-plus-error pair when failure should panic immediately.
- **Domain Rules:**
  - panics with a wrapped trace error when err is non-nil.
MustValue unwraps a (value, error) pair, panicking on error
- **Calls:** Wrap

### Collect
- **Lines:** 168–185
- **Intent:** combine many Result values into one collection while preserving all failures.
- **Domain Rules:**
  - any failed input produces an aggregated error instead of a partial success.
Collect collects multiple Results into a single Result containing a slice
- **Calls:** Aggregate

### As
- **Lines:** 190–196
- **Intent:** perform typed error extraction without repeating target boilerplate at call sites.
- **Ensures:**
  - returns the zero value of T and false when no matching error is found.
As is a generic version of errors.As
- **Calls:** As

### NewPipeline
- **Lines:** 208–210
- **Intent:** start a chain of trace-aware transformations from an initial value.
- **Ensures:**
  - returns a pipeline with no initial error.
NewPipeline creates a new pipeline with an initial value

### Then
- **Lines:** 215–224
- **Intent:** apply the next transformation only while the pipeline remains successful.
- **Ensures:**
  - wraps new step failures with trace context before storing them.
Then executes the function if no error has occurred
- **Calls:** Wrap

### ThenDo
- **Lines:** 229–238
- **Intent:** run a side-effecting validation or action without changing the pipeline value.
- **Ensures:**
  - wraps new step failures with trace context before storing them.
ThenDo executes a function that doesn't modify the value
- **Calls:** Wrap

### Result
- **Lines:** 243–245
- **Intent:** expose the pipeline state as a conventional Go value-plus-error pair.
- **Ensures:**
  - returns the current pipeline value and error unchanged.
Result returns the final value and error

### ToResult
- **Lines:** 250–252
- **Intent:** convert pipeline state into the Result abstraction for further composition.
- **Ensures:**
  - returns a Result containing the current pipeline value and error.
ToResult converts the pipeline to a Result

### Recover
- **Lines:** 257–263
- **Intent:** give failed pipelines a chance to replace their error with a recovery value or new error.
- **Ensures:**
  - calls fn only when the pipeline is currently failed.
Recover attempts to recover from an error

### RecoverWith
- **Lines:** 268–274
- **Intent:** replace a failed pipeline with a caller-provided default value.
- **Ensures:**
  - clears the stored error when recovery is applied.
RecoverWith recovers with a default value

### TransformPipeline
- **Lines:** 279–288
- **Intent:** continue a pipeline while changing the value type and preserving trace-aware failure handling.
- **Ensures:**
  - carries forward existing pipeline errors without invoking fn.
TransformPipeline transforms between different types
- **Calls:** Wrap

## Classes

### Result
- **Lines:** 10–13
- **Intent:** model success and failure as one value so callers can compose operations without losing trace context.
Result represents either a value or an error (similar to Rust's Result)

### Pipeline
- **Lines:** 200–203
- **Intent:** model stepwise transformations that should stop automatically after the first failure.
Pipeline allows chaining operations that may fail
