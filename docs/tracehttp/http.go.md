<!-- generated-by: code-context-graph docs -->
# tracehttp/http.go

> HTTP adapters that translate trace errors into API responses, middleware behavior, and client-side classifications.

## Functions

### ToHTTPError
- **Lines:** 79–97
- **Intent:** classify an error chain into a validated client-safe HTTP representation.
- **Domain Rules:**
  - the outermost classifiable link in the chain decides the response, so an
outer AccessDenied wrap is never overridden by an inner NotFound.
  - invalid classifications become internal errors and every 5xx message is sanitized.
- **Ensures:**
  - returns the zero value for nil and never derives a message from an outer trace wrapper.
ToHTTPError returns the safe HTTP representation for err.
- **Calls:** Errors, classify, aggregateHTTPError, validate, validate, internalHTTPError, HTTPError

### classify
- **Lines:** 102–153
- **Intent:** classify exactly one error value without following its Unwrap chain.
- **Domain Rules:**
  - categories are tested in a fixed order so an error implementing several
behavior interfaces always produces the same status.
- **Calls:** linkMessage

### linkMessage
- **Lines:** 156–162
- **Intent:** read the message this specific error carries, never the message of an outer wrapper.

### aggregateHTTPError
- **Lines:** 166–177
- **Intent:** choose the safe representation with the highest child status for an aggregate failure.
- **Domain Rules:**
  - the most severe child wins rather than the first one traversed.
- **Calls:** Unwrap, ToHTTPError, internalHTTPError

### validate
- **Lines:** 180–191
- **Intent:** reject malformed classifications and keep server-side detail out of 5xx responses.
- **Calls:** internalHTTPError

### internalHTTPError
- **Lines:** 194–200
- **Intent:** centralize the generic fallback used whenever a failure cannot be exposed safely.

### GetHTTPStatusCode
- **Lines:** 205–210
- **Intent:** translate trace error categories into HTTP response codes at service boundaries.
- **Ensures:**
  - returns HTTP 200 for nil errors and HTTP 500 for unclassifiable errors.
GetHTTPStatusCode returns the HTTP status code for an error.
- **Calls:** ToHTTPError

### ErrorResponseFor
- **Lines:** 230–243
- **Intent:** build a framework-neutral response from the safe HTTP classification.
- **Domain Rules:**
  - request IDs are caller-supplied and trace fields or details are never copied from err.
- **Ensures:**
  - returns zero values for nil errors.
ErrorResponseFor returns the status and safe response body for err.
- **Calls:** ToHTTPError

### WriteError
- **Lines:** 249–258
- **Intent:** expose a simple entry point for converting domain errors into HTTP responses.
- **Side Effects:** writes status code, headers, and a JSON body to the response writer.
- **Ensures:**
  - returns without writing anything when the input error is nil.
WriteError writes a safe error response without logging.
- **Calls:** ErrorResponseFor

### ErrorMiddleware
- **Lines:** 268–281
- **Intent:** adapt error-returning handlers into standard net/http handlers.
- **Side Effects:** attaches method and path fields before writing the HTTP error response.
- **Ensures:**
  - renders handler errors through the logger-free safe response writer.
ErrorMiddleware converts an ErrorHandlerFunc to a standard http.HandlerFunc.
- **Calls:** WithFields, WriteError

### ReadErrorResponse
- **Lines:** 287–359
- **Intent:** restore typed trace semantics from the public ErrorResponse envelope without deserializing internal trace data.
- **Domain Rules:**
  - HTTP 2xx and 3xx statuses are not errors.
  - malformed, unknown, or status-mismatched responses fail closed without retaining the raw body.
ReadErrorResponse converts a public HTTP error response into a trace error.
- **Calls:** CaptureFrame, newReadErrorResponseTrace, newReadErrorResponseTrace, newReadErrorResponseTrace, newReadErrorResponseTrace, errorCodeMatchesStatus

### newReadErrorResponseTrace
- **Lines:** 362–372
- **Intent:** build the trace error that carries the safe message and upstream response metadata.

### errorCodeMatchesStatus
- **Lines:** 375–402
- **Intent:** validate the public code and HTTP status as one coherent wire contract.

### FromHTTPResponse
- **Lines:** 412–451
- **Intent:** classify an upstream response for developer-facing debugging rather than client rendering.
- **Domain Rules:**
  - 2xx responses are not errors, while known status codes map to typed trace errors.
- **Ensures:**
  - records the current call site as the first trace frame and stores upstream status metadata in error fields.
FromHTTPResponse creates an appropriate error from an HTTP response.
- **Calls:** CaptureFrame

### IsHTTPError
- **Lines:** 456–458
- **Intent:** test whether an error chain resolves to a specific HTTP status mapping.
- **Ensures:**
  - delegates status resolution to GetHTTPStatusCode.
IsHTTPError checks if an error corresponds to a specific HTTP status code.
- **Calls:** GetHTTPStatusCode

### WrapHTTPError
- **Lines:** 466–491
- **Intent:** override or attach explicit HTTP status semantics to an existing error chain.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Mutates:** adds http_status metadata to the returned traced wrapper.
- **Ensures:**
  - the returned error accumulates the caller frame on top of the source frames and
carries the source fields forward, so GetFrames and GetFields stay useful after wrapping.
WrapHTTPError wraps an error with HTTP status code information.
- **Calls:** CaptureFrame, GetFrames, GetFields

### HTTPError
- **Lines:** 500–506
- **Intent:** expose the explicit status override through the single HTTP extension point.
- **Calls:** UserMessage

### Error
- **Lines:** 509–509
- **Intent:** delegate user-facing string rendering to the wrapped error.
- **Calls:** Error

### Unwrap
- **Lines:** 512–512
- **Intent:** expose the wrapped error to standard Go error traversal.

### NewClient
- **Lines:** 523–528
- **Intent:** provide an HTTP client wrapper that returns trace-classified transport failures.
- **Ensures:**
  - falls back to http.DefaultClient when no custom client is supplied.
NewClient creates a new trace-aware HTTP client.

### Do
- **Lines:** 534–543
- **Intent:** classify outbound HTTP transport failures into timeout or connection problem errors.
- **Domain Rules:**
  - deadline exceeded maps to Timeout and other transport failures map to ConnectionProblem.
- **Side Effects:** executes the underlying HTTP request through the wrapped client.
Do executes the request and wraps any errors with trace information.
- **Calls:** ConnectionProblem, Timeout, Do

## Classes

### HTTPError
- **Lines:** 56–60
- **Intent:** carry only validated status, semantic code, and client-safe message across an HTTP boundary.
HTTPError is the safe client-facing representation of an error.

### ErrorBody
- **Lines:** 214–218
- **Intent:** group safe client fields under one stable JSON error envelope.
ErrorBody is the stable machine-readable body nested under the error key.

### ErrorResponse
- **Lines:** 222–224
- **Intent:** define the client-facing error payload returned by HTTP helpers in this package.
ErrorResponse represents a structured JSON error response.

### statusError
- **Lines:** 494–497
- **Intent:** carry an explicit HTTP status override for errors that do not map to one of the standard typed categories.

### Client
- **Lines:** 516–518
- **Intent:** wrap http.Client so transport failures come back as trace-classified errors.
Client is an HTTP client that wraps errors with trace information.

## Types

### ErrorCode
- **Lines:** 21–21
- **Intent:** give clients a stable machine-readable classification independent of HTTP status text.
ErrorCode is a stable machine-readable HTTP error classification.

### HTTPErrorProvider
- **Lines:** 68–71
- **Intent:** let application-owned error types opt into explicit safe HTTP classification.
HTTPErrorProvider lets custom errors define a safe HTTP representation.

### ErrorHandlerFunc
- **Lines:** 262–262
- **Intent:** let HTTP handlers return errors directly so middleware can centralize response rendering.
ErrorHandlerFunc is a function that handles HTTP requests and may return an error.
