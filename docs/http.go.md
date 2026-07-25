<!-- generated-by: code-context-graph docs -->
# http.go

> HTTP adapters that translate trace errors into API responses, middleware behavior, and client-side classifications.

## Functions

### ToHTTPError
- **Lines:** 63–81
- **Intent:** classify an error chain into a validated client-safe HTTP representation.
- **Domain Rules:**
  - invalid classifications become internal errors and every 5xx message is sanitized.
- **Ensures:**
  - returns the zero value for nil and never derives a message from an outer trace wrapper.
ToHTTPError returns the safe HTTP representation for err.
- **Calls:** As, internalHTTPError, internalHTTPError

### internalHTTPError
- **Lines:** 84–90
- **Intent:** centralize the generic fallback used whenever a failure cannot be exposed safely.

### ErrorResponseFor
- **Lines:** 110–123
- **Intent:** build a framework-neutral response from the safe HTTP classification.
- **Domain Rules:**
  - request IDs are caller-supplied and trace fields or details are never copied from err.
- **Ensures:**
  - returns zero values for nil errors.
ErrorResponseFor returns the status and safe response body for err.
- **Calls:** ToHTTPError

### WriteError
- **Lines:** 129–138
- **Intent:** expose a simple entry point for converting domain errors into HTTP responses.
- **Side Effects:** writes status code, headers, and a JSON body to the response writer.
- **Ensures:**
  - returns without writing anything when the input error is nil.
WriteError writes a safe error response without logging.
- **Calls:** ErrorResponseFor

### WriteErrorWithLogger
- **Lines:** 144–162
- **Intent:** retain the legacy opt-in logging adapter while directing new callers to application-owned logging.
WriteErrorWithLogger writes an error response and logs it.
- **Calls:** ToHTTPError, WriteError, Error, Error, SlogError

### ErrorMiddleware
- **Lines:** 171–184
- **Intent:** adapt error-returning handlers into standard net/http handlers.
- **Ensures:**
  - renders handler errors through the logger-free safe response writer.
ErrorMiddleware converts an ErrorHandlerFunc to a standard http.HandlerFunc
- **Calls:** WriteError, WithFields

### ErrorMiddlewareWithLogger
- **Lines:** 192–204
- **Intent:** normalize handler failures into traced HTTP responses enriched with request metadata.
- **Side Effects:** attaches method and path fields before writing the HTTP error response.
- **Ensures:**
  - successful handlers pass through without writing an additional error response.
ErrorMiddlewareWithLogger converts an ErrorHandlerFunc to http.HandlerFunc with logging.
- **Calls:** WriteErrorWithLogger, WithFields

### RecoverMiddleware
- **Lines:** 212–251
- **Intent:** prevent panics from escaping the HTTP boundary and turn them into traceable server errors.
- **Domain Rules:**
  - recovered panics always return HTTP 500 with a generic client-facing message.
- **Side Effects:** recovers panics, logs structured diagnostics, and writes a fallback JSON response.
RecoverMiddleware recovers from panics and converts them to errors.
- **Calls:** Error, Error, SlogError, Wrap, Errorf, WithFields

### FromHTTPResponse
- **Lines:** 257–300
- **Intent:** classify non-success upstream HTTP responses into trace error categories.
- **Domain Rules:**
  - 2xx responses are not errors, while known status codes map to typed trace errors.
- **Ensures:**
  - records the current call site as the first trace frame and stores upstream status metadata in error fields.
FromHTTPResponse creates an appropriate error from an HTTP response
- **Calls:** captureFrame

### IsHTTPError
- **Lines:** 305–307
- **Intent:** test whether an error chain resolves to a specific HTTP status mapping.
- **Ensures:**
  - delegates status resolution to GetHTTPStatusCode.
IsHTTPError checks if an error corresponds to a specific HTTP status code
- **Calls:** GetHTTPStatusCode

### WrapHTTPError
- **Lines:** 313–333
- **Intent:** override or attach explicit HTTP status semantics to an existing error chain.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Mutates:** adds http_status metadata to the returned traced wrapper.
WrapHTTPError wraps an error with HTTP status code information
- **Calls:** wrapTypedInternal, captureFrame

### HTTPStatusCode
- **Lines:** 342–342
- **Intent:** expose the explicit status override carried by this internal HTTP error wrapper.

### Error
- **Lines:** 345–345
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 348–348
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### NewClient
- **Lines:** 359–364
- **Intent:** provide an HTTP client wrapper that returns trace-classified transport failures.
- **Ensures:**
  - falls back to http.DefaultClient when no custom client is supplied.
NewClient creates a new trace-aware HTTP client

### Do
- **Lines:** 370–380
- **Intent:** classify outbound HTTP transport failures into timeout or connection problem errors.
- **Domain Rules:**
  - deadline exceeded maps to Timeout and other transport failures map to ConnectionProblem.
- **Side Effects:** executes the underlying HTTP request through the wrapped client.
Do executes the request and wraps any errors with trace information
- **Calls:** ConnectionProblem, Timeout, Do

## Classes

### HTTPError
- **Lines:** 46–50
- **Intent:** carry only validated status, semantic code, and client-safe message across an HTTP boundary.
HTTPError is the safe client-facing representation of an error.

### ErrorBody
- **Lines:** 94–98
- **Intent:** group safe client fields under one stable JSON error envelope.
ErrorBody is the stable machine-readable body nested under the error key.

### ErrorResponse
- **Lines:** 102–104
- **Intent:** define the client-facing error payload returned by HTTP helpers in this package.
ErrorResponse represents a structured JSON error response.

### httpStatusError
- **Lines:** 336–339
- **Intent:** carry an explicit HTTP status override for errors that do not map to one of the standard typed categories.

### Client
- **Lines:** 352–354
- **Intent:** wrap http.Client so transport failures come back as trace-classified errors.
Client is an HTTP client that wraps errors with trace information

## Types

### ErrorCode
- **Lines:** 15–15
- **Intent:** give clients a stable machine-readable classification independent of HTTP status text.
ErrorCode is a stable machine-readable HTTP error classification.

### HTTPErrorProvider
- **Lines:** 54–57
- **Intent:** let application-owned error types opt into explicit safe HTTP classification.
HTTPErrorProvider lets custom errors define a safe HTTP representation.

### ErrorHandlerFunc
- **Lines:** 166–166
- **Intent:** let HTTP handlers return errors directly so middleware can centralize response rendering.
ErrorHandlerFunc is a function that handles HTTP requests and may return an error
