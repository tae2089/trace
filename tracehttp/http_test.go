package tracehttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tae2089/trace/v2"
	"github.com/tae2089/trace/v2/tracehttp"
)

type customHTTPError struct {
	response tracehttp.HTTPError
}

type failingResponseWriter struct {
	header http.Header
	err    error
	status int
}

func (w *failingResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func (e customHTTPError) Error() string {
	return "custom HTTP error"
}

func (e customHTTPError) HTTPError() tracehttp.HTTPError {
	return e.response
}

func TestToHTTPErrorPreservesTypedPublicMessageThroughWraps(t *testing.T) {
	err := trace.NotFound("user not found")
	err = trace.Wrap(err, "query users table")
	err = trace.Wrap(err, "handle request")

	got := tracehttp.ToHTTPError(err)
	want := tracehttp.HTTPError{
		Status:  http.StatusNotFound,
		Code:    tracehttp.CodeNotFound,
		Message: "user not found",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
	}
}

func TestToHTTPErrorHidesInternalError(t *testing.T) {
	const secret = "password=do-not-expose"
	err := trace.Wrap(errors.New(secret), "query users table")

	got := tracehttp.ToHTTPError(err)

	if got != (tracehttp.HTTPError{
		Status:  http.StatusInternalServerError,
		Code:    tracehttp.CodeInternal,
		Message: "internal server error",
	}) {
		t.Fatalf("ToHTTPError() = %#v", got)
	}
	if strings.Contains(got.Message, secret) {
		t.Fatalf("internal message leaked secret %q", secret)
	}
}

func TestToHTTPErrorMapsBuiltInCategories(t *testing.T) {
	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		err  error
		want tracehttp.HTTPError
	}{
		{
			name: "bad request",
			err:  trace.BadParameter("invalid email"),
			want: tracehttp.HTTPError{Status: http.StatusBadRequest, Code: tracehttp.CodeBadRequest, Message: "invalid email"},
		},
		{
			name: "unauthenticated",
			err:  trace.Unauthenticated("expired bearer token"),
			want: tracehttp.HTTPError{Status: http.StatusUnauthorized, Code: tracehttp.CodeUnauthenticated, Message: "authentication required"},
		},
		{
			name: "access denied",
			err:  trace.AccessDenied("missing admin role"),
			want: tracehttp.HTTPError{Status: http.StatusForbidden, Code: tracehttp.CodeAccessDenied, Message: "access denied"},
		},
		{
			name: "not found",
			err:  trace.NotFound("user not found"),
			want: tracehttp.HTTPError{Status: http.StatusNotFound, Code: tracehttp.CodeNotFound, Message: "user not found"},
		},
		{
			name: "already exists",
			err:  trace.AlreadyExists("email already registered"),
			want: tracehttp.HTTPError{Status: http.StatusConflict, Code: tracehttp.CodeAlreadyExists, Message: "email already registered"},
		},
		{
			name: "conflict",
			err:  trace.Conflict("version mismatch"),
			want: tracehttp.HTTPError{Status: http.StatusConflict, Code: tracehttp.CodeConflict, Message: "version mismatch"},
		},
		{
			name: "limit exceeded",
			err:  trace.LimitExceeded("quota exceeded"),
			want: tracehttp.HTTPError{Status: http.StatusTooManyRequests, Code: tracehttp.CodeLimitExceeded, Message: "quota exceeded"},
		},
		{
			name: "canceled",
			err:  trace.FromContext(canceledContext),
			want: tracehttp.HTTPError{Status: 499, Code: tracehttp.CodeCanceled, Message: "request canceled"},
		},
		{
			name: "not implemented",
			err:  trace.NotImplemented("admin export is unfinished"),
			want: tracehttp.HTTPError{Status: http.StatusNotImplemented, Code: tracehttp.CodeNotImplemented, Message: "internal server error"},
		},
		{
			name: "unavailable",
			err:  trace.ConnectionProblem(errors.New("dial tcp database.internal"), "database unavailable"),
			want: tracehttp.HTTPError{Status: http.StatusServiceUnavailable, Code: tracehttp.CodeUnavailable, Message: "internal server error"},
		},
		{
			name: "timeout",
			err:  trace.Timeout(errors.New("upstream deadline"), "inventory request timed out"),
			want: tracehttp.HTTPError{Status: http.StatusGatewayTimeout, Code: tracehttp.CodeTimeout, Message: "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tracehttp.ToHTTPError(tt.err)
			if got != tt.want {
				t.Fatalf("ToHTTPError() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestToHTTPErrorSupportsValidatedCustomProviders(t *testing.T) {
	t.Run("wrapped custom provider", func(t *testing.T) {
		err := trace.Wrap(customHTTPError{response: tracehttp.HTTPError{
			Status:  http.StatusTeapot,
			Code:    tracehttp.ErrorCode("teapot"),
			Message: "short and stout",
		}}, "internal context")

		got := tracehttp.ToHTTPError(err)
		want := tracehttp.HTTPError{
			Status:  http.StatusTeapot,
			Code:    tracehttp.ErrorCode("teapot"),
			Message: "short and stout",
		}
		if got != want {
			t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
		}
	})

	tests := []struct {
		name     string
		provided tracehttp.HTTPError
		want     tracehttp.HTTPError
	}{
		{
			name:     "status below error range",
			provided: tracehttp.HTTPError{Status: http.StatusOK, Code: tracehttp.ErrorCode("success"), Message: "unsafe"},
			want:     tracehttp.HTTPError{Status: http.StatusInternalServerError, Code: tracehttp.CodeInternal, Message: "internal server error"},
		},
		{
			name:     "empty code",
			provided: tracehttp.HTTPError{Status: http.StatusTeapot, Message: "unsafe"},
			want:     tracehttp.HTTPError{Status: http.StatusInternalServerError, Code: tracehttp.CodeInternal, Message: "internal server error"},
		},
		{
			name:     "custom 5xx message",
			provided: tracehttp.HTTPError{Status: http.StatusServiceUnavailable, Code: tracehttp.ErrorCode("dependency_down"), Message: "redis password=secret"},
			want:     tracehttp.HTTPError{Status: http.StatusServiceUnavailable, Code: tracehttp.ErrorCode("dependency_down"), Message: "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tracehttp.ToHTTPError(customHTTPError{response: tt.provided})
			if got != tt.want {
				t.Fatalf("ToHTTPError() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestErrorResponseForBuildsSafeNestedBody(t *testing.T) {
	err := trace.WithFields(trace.NotFound("user not found"), map[string]any{
		"trace_id": "trace-must-not-leak",
		"details":  map[string]any{"table": "users"},
	})
	err = trace.Wrap(err, "query users table")

	status, got := tracehttp.ErrorResponseFor(err, "01KREQUEST")
	want := tracehttp.ErrorResponse{
		Error: tracehttp.ErrorBody{
			Code:      tracehttp.CodeNotFound,
			Message:   "user not found",
			RequestID: "01KREQUEST",
		},
	}
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("response = %#v, want %#v", got, want)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	const wantJSON = `{"error":{"code":"not_found","message":"user not found","request_id":"01KREQUEST"}}`
	if string(encoded) != wantJSON {
		t.Fatalf("JSON = %s, want %s", encoded, wantJSON)
	}
}

func TestErrorResponseForOmitsEmptyRequestID(t *testing.T) {
	status, got := tracehttp.ErrorResponseFor(trace.BadParameter("invalid email"), "")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	const wantJSON = `{"error":{"code":"bad_request","message":"invalid email"}}`
	if string(encoded) != wantJSON {
		t.Fatalf("JSON = %s, want %s", encoded, wantJSON)
	}
}

func TestWriteErrorWritesSafeResponseWithoutLogging(t *testing.T) {
	var logs strings.Builder
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	recorder := httptest.NewRecorder()
	err := tracehttp.WriteError(recorder, trace.NotFound("user not found"), "01KREQUEST")
	if err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	const wantBody = `{"error":{"code":"not_found","message":"user not found","request_id":"01KREQUEST"}}` + "\n"
	if recorder.Body.String() != wantBody {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), wantBody)
	}
	if logs.Len() != 0 {
		t.Fatalf("WriteError emitted logs: %s", logs.String())
	}
}

func TestWriteErrorHandlesNilAndWriteFailure(t *testing.T) {
	t.Run("nil is a no-op", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		if err := tracehttp.WriteError(recorder, nil, "01KREQUEST"); err != nil {
			t.Fatalf("WriteError(nil) error = %v", err)
		}
		if recorder.Body.Len() != 0 {
			t.Fatalf("WriteError(nil) body = %q", recorder.Body.String())
		}
	})

	t.Run("write failure is returned", func(t *testing.T) {
		wantErr := errors.New("client disconnected")
		writer := &failingResponseWriter{
			header: make(http.Header),
			err:    wantErr,
		}

		gotErr := tracehttp.WriteError(writer, trace.NotFound("user not found"), "")
		if !errors.Is(gotErr, wantErr) {
			t.Fatalf("WriteError() error = %v, want %v", gotErr, wantErr)
		}
		if writer.status != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", writer.status, http.StatusNotFound)
		}
	})
}

func TestFromHTTPResponseDistinguishesAuthenticationAndAuthorization(t *testing.T) {
	tests := []struct {
		name                string
		status              int
		wantCode            tracehttp.ErrorCode
		wantUnauthenticated bool
		wantAccessDenied    bool
	}{
		{
			name:                "unauthorized",
			status:              http.StatusUnauthorized,
			wantCode:            tracehttp.CodeUnauthenticated,
			wantUnauthenticated: true,
		},
		{
			name:             "forbidden",
			status:           http.StatusForbidden,
			wantCode:         tracehttp.CodeAccessDenied,
			wantAccessDenied: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &http.Response{
				StatusCode: tt.status,
				Status:     http.StatusText(tt.status),
			}

			err := tracehttp.FromHTTPResponse(response, []byte("upstream diagnostic"))
			if trace.IsUnauthenticated(err) != tt.wantUnauthenticated {
				t.Fatalf("IsUnauthenticated() = %v", trace.IsUnauthenticated(err))
			}
			if trace.IsAccessDenied(err) != tt.wantAccessDenied {
				t.Fatalf("IsAccessDenied() = %v", trace.IsAccessDenied(err))
			}
			if got := tracehttp.ToHTTPError(err).Code; got != tt.wantCode {
				t.Fatalf("code = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

func TestToHTTPErrorPreservesAggregateStatusPriority(t *testing.T) {
	err := trace.Aggregate(
		trace.BadParameter("invalid email"),
		trace.Timeout(errors.New("upstream deadline"), "inventory timed out"),
	)

	got := tracehttp.ToHTTPError(err)
	want := tracehttp.HTTPError{
		Status:  http.StatusGatewayTimeout,
		Code:    tracehttp.CodeTimeout,
		Message: "internal server error",
	}
	if got != want {
		t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
	}
}

func TestReadErrorResponseReturnsNilForNonErrorStatus(t *testing.T) {
	for _, status := range []int{
		http.StatusOK,
		http.StatusNoContent,
		http.StatusMovedPermanently,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			err := tracehttp.ReadErrorResponse(status, []byte("not JSON"))
			if err != nil {
				t.Fatalf("ReadErrorResponse() error = %v, want nil", err)
			}
		})
	}
}

func TestReadErrorResponseRoundTripsBuiltInCodes(t *testing.T) {
	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name      string
		sourceErr error
		matches   func(error) bool
	}{
		{name: "bad request", sourceErr: trace.BadParameter("invalid email"), matches: trace.IsBadParameter},
		{name: "unauthenticated", sourceErr: trace.Unauthenticated("expired token"), matches: trace.IsUnauthenticated},
		{name: "access denied", sourceErr: trace.AccessDenied("missing role"), matches: trace.IsAccessDenied},
		{name: "not found", sourceErr: trace.NotFound("user not found"), matches: trace.IsNotFound},
		{name: "already exists", sourceErr: trace.AlreadyExists("email exists"), matches: trace.IsAlreadyExists},
		{name: "conflict", sourceErr: trace.Conflict("version mismatch"), matches: trace.IsConflict},
		{name: "limit exceeded", sourceErr: trace.LimitExceeded("quota exceeded"), matches: trace.IsLimitExceeded},
		{name: "canceled", sourceErr: trace.FromContext(canceledContext), matches: trace.IsCanceled},
		{name: "not implemented", sourceErr: trace.NotImplemented("export unavailable"), matches: trace.IsNotImplemented},
		{
			name:      "unavailable",
			sourceErr: trace.ConnectionProblem(errors.New("dial failed"), "database unavailable"),
			matches:   trace.IsConnectionProblem,
		},
		{
			name:      "timeout",
			sourceErr: trace.Timeout(errors.New("deadline"), "inventory timed out"),
			matches:   trace.IsTimeout,
		},
		{
			name:      "internal",
			sourceErr: errors.New("database password must remain hidden"),
			matches: func(err error) bool {
				return tracehttp.ToHTTPError(err).Code == tracehttp.CodeInternal
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, response := tracehttp.ErrorResponseFor(tt.sourceErr, "01KREQUEST")
			body, err := json.Marshal(response)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			gotErr := tracehttp.ReadErrorResponse(status, body)
			if gotErr == nil {
				t.Fatal("ReadErrorResponse() error = nil")
			}
			if !tt.matches(gotErr) {
				t.Fatalf("ReadErrorResponse() type = %T, predicate did not match", gotErr)
			}
			if got, want := tracehttp.ToHTTPError(gotErr), tracehttp.ToHTTPError(tt.sourceErr); got != want {
				t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
			}

			fields := trace.GetFields(gotErr)
			if got := fields["status_code"]; got != status {
				t.Fatalf("status_code = %#v, want %d", got, status)
			}
			if got := fields["request_id"]; got != "01KREQUEST" {
				t.Fatalf("request_id = %#v, want %q", got, "01KREQUEST")
			}
		})
	}
}

func TestReadErrorResponseKeepsOnlySafeEnvelopeFields(t *testing.T) {
	const secret = "database password=do-not-copy"
	body := []byte(`{
		"error": {
			"code": "not_found",
			"message": "user not found",
			"request_id": "01KREQUEST",
			"cause": "` + secret + `",
			"details": {"table": "users"},
			"trace": [{"file": "internal/repo.go"}]
		},
		"traces": [{"file": "internal/service.go"}],
		"fields": {"trace_id": "must-not-copy"}
	}`)

	err := tracehttp.ReadErrorResponse(http.StatusNotFound, body)
	if !trace.IsNotFound(err) {
		t.Fatalf("ReadErrorResponse() type = %T, want not found", err)
	}
	wantFields := map[string]any{
		"status_code": http.StatusNotFound,
		"request_id":  "01KREQUEST",
	}
	if got := trace.GetFields(err); !reflect.DeepEqual(got, wantFields) {
		t.Fatalf("GetFields() = %#v, want %#v", got, wantFields)
	}
	for _, output := range []string{err.Error(), trace.DebugReport(err)} {
		if strings.Contains(output, secret) ||
			strings.Contains(output, "internal/repo.go") ||
			strings.Contains(output, "must-not-copy") {
			t.Fatalf("internal response field leaked in %q", output)
		}
	}
}

func TestReadErrorResponseNormalizesSensitiveMessages(t *testing.T) {
	const secret = "token=do-not-expose"
	tests := []struct {
		name    string
		status  int
		code    tracehttp.ErrorCode
		want    string
		matches func(error) bool
	}{
		{
			name:    "unauthenticated",
			status:  http.StatusUnauthorized,
			code:    tracehttp.CodeUnauthenticated,
			want:    "authentication required",
			matches: trace.IsUnauthenticated,
		},
		{
			name:    "access denied",
			status:  http.StatusForbidden,
			code:    tracehttp.CodeAccessDenied,
			want:    "access denied",
			matches: trace.IsAccessDenied,
		},
		{
			name:    "canceled",
			status:  499,
			code:    tracehttp.CodeCanceled,
			want:    "request canceled",
			matches: trace.IsCanceled,
		},
		{
			name:    "not implemented",
			status:  http.StatusNotImplemented,
			code:    tracehttp.CodeNotImplemented,
			want:    "internal server error",
			matches: trace.IsNotImplemented,
		},
		{
			name:    "unavailable",
			status:  http.StatusServiceUnavailable,
			code:    tracehttp.CodeUnavailable,
			want:    "internal server error",
			matches: trace.IsConnectionProblem,
		},
		{
			name:    "timeout",
			status:  http.StatusGatewayTimeout,
			code:    tracehttp.CodeTimeout,
			want:    "internal server error",
			matches: trace.IsTimeout,
		},
		{
			name:   "internal",
			status: http.StatusInternalServerError,
			code:   tracehttp.CodeInternal,
			want:   "internal server error",
			matches: func(err error) bool {
				return tracehttp.ToHTTPError(err).Code == tracehttp.CodeInternal
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tracehttp.ErrorResponse{
				Error: tracehttp.ErrorBody{Code: tt.code, Message: secret},
			})
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			gotErr := tracehttp.ReadErrorResponse(tt.status, body)
			if !tt.matches(gotErr) {
				t.Fatalf("ReadErrorResponse() type = %T, predicate did not match", gotErr)
			}
			if got := trace.UserMessage(gotErr); got != tt.want {
				t.Fatalf("UserMessage() = %q, want %q", got, tt.want)
			}
			if strings.Contains(gotErr.Error(), secret) ||
				strings.Contains(trace.DebugReport(gotErr), secret) {
				t.Fatalf("sensitive message leaked from %q", gotErr)
			}
		})
	}
}

func TestReadErrorResponseFailsClosed(t *testing.T) {
	const secret = "password=do-not-expose"
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "empty body", status: http.StatusBadRequest},
		{name: "malformed JSON", status: http.StatusBadRequest, body: `{` + secret},
		{
			name:   "missing code",
			status: http.StatusBadRequest,
			body:   `{"error":{"message":"` + secret + `"}}`,
		},
		{
			name:   "missing message",
			status: http.StatusBadRequest,
			body:   `{"error":{"code":"bad_request"}}`,
		},
		{
			name:   "unknown code",
			status: http.StatusTeapot,
			body:   `{"error":{"code":"teapot","message":"` + secret + `","request_id":"01KREQUEST"}}`,
		},
		{
			name:   "code and status mismatch",
			status: http.StatusBadRequest,
			body:   `{"error":{"code":"not_found","message":"` + secret + `"}}`,
		},
		{
			name:   "trailing JSON",
			status: http.StatusBadRequest,
			body:   `{"error":{"code":"bad_request","message":"safe"}}{"secret":"` + secret + `"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tracehttp.ReadErrorResponse(tt.status, []byte(tt.body))
			if err == nil {
				t.Fatal("ReadErrorResponse() error = nil")
			}
			want := tracehttp.HTTPError{
				Status:  http.StatusInternalServerError,
				Code:    tracehttp.CodeInternal,
				Message: "internal server error",
			}
			if got := tracehttp.ToHTTPError(err); got != want {
				t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
			}
			if got := trace.UserMessage(err); got != "invalid error response" {
				t.Fatalf("UserMessage() = %q, want %q", got, "invalid error response")
			}
			if strings.Contains(err.Error(), secret) ||
				strings.Contains(trace.DebugReport(err), secret) {
				t.Fatalf("raw response leaked from %q", err)
			}
		})
	}
}

// Example: HTTP middleware
func TestHTTPMiddleware(t *testing.T) {
	handler := tracehttp.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		return trace.NotFound("resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var resp tracehttp.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error.Code != tracehttp.CodeNotFound {
		t.Error("response error code mismatch")
	}
}

// P3: WrapHTTPError preserves frames/fields
func TestWrapHTTPErrorPreservesFramesAndFields(t *testing.T) {
	inner := trace.Wrap(errors.New("root"), "inner")
	inner = trace.WithField(inner, "req_id", "r1")
	innerFrameCount := len(trace.GetFrames(inner))

	wrapped := tracehttp.WrapHTTPError(inner, 503, "service down")
	wrappedFrames := trace.GetFrames(wrapped)
	wrappedFields := trace.GetFields(wrapped)

	if len(wrappedFrames) < innerFrameCount+1 {
		t.Errorf("WrapHTTPError should accumulate frames: got %d, inner had %d", len(wrappedFrames), innerFrameCount)
	}
	if wrappedFields["req_id"] != "r1" {
		t.Error("WrapHTTPError should preserve inner fields")
	}
	if wrappedFields["http_status"] != 503 {
		t.Error("WrapHTTPError should set http_status field")
	}
}

func TestToHTTPErrorHidesFilesystemPathsFromConvertedSystemErrors(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "private-key.pem")
	_, openErr := os.Open(secretPath)
	if openErr == nil {
		t.Fatal("expected opening a missing file to fail")
	}

	converted := trace.ConvertSystemError(openErr)
	if !trace.IsNotFound(converted) {
		t.Fatalf("expected a not found error, got %#v", converted)
	}

	got := tracehttp.ToHTTPError(converted)

	if got.Status != http.StatusNotFound || got.Code != tracehttp.CodeNotFound {
		t.Fatalf("ToHTTPError() = %#v, want 404/not_found", got)
	}
	if strings.Contains(got.Message, secretPath) || strings.Contains(got.Message, "private-key.pem") {
		t.Fatalf("client message leaked a filesystem path: %q", got.Message)
	}
	if got.Message != string(tracehttp.CodeNotFound) {
		t.Fatalf("expected the generic code fallback, got %q", got.Message)
	}
}

// WithField/WithFields must not drop the explicit status override carried by
// WrapHTTPError's wrapper (github.com/tae2089/trace/v2 TraceErrorReplacer hook).
func TestWithFieldKeepsStatusOverride(t *testing.T) {
	base := errors.New("teapot broke")
	wrapped := tracehttp.WrapHTTPError(base, http.StatusTeapot)

	withField := trace.WithField(wrapped, "req_id", "r1")
	if got := tracehttp.GetHTTPStatusCode(withField); got != http.StatusTeapot {
		t.Errorf("WithField dropped the status override: got %d, want %d", got, http.StatusTeapot)
	}
	if fields := trace.GetFields(withField); fields["req_id"] != "r1" {
		t.Errorf("field not attached: %v", fields)
	}

	withFields := trace.WithFields(wrapped, map[string]any{"a": 1, "b": 2})
	if got := tracehttp.GetHTTPStatusCode(withFields); got != http.StatusTeapot {
		t.Errorf("WithFields dropped the status override: got %d, want %d", got, http.StatusTeapot)
	}

	// The original must stay untouched (immutability contract).
	if fields := trace.GetFields(wrapped); fields["req_id"] != nil {
		t.Errorf("original error mutated: %v", fields)
	}
}
