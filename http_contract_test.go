package trace_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/tae2089/trace"
)

type customHTTPError struct {
	response trace.HTTPError
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

func (e customHTTPError) HTTPError() trace.HTTPError {
	return e.response
}

func TestToHTTPErrorPreservesTypedPublicMessageThroughWraps(t *testing.T) {
	err := trace.NotFound("user not found")
	err = trace.Wrap(err, "query users table")
	err = trace.Wrap(err, "handle request")

	got := trace.ToHTTPError(err)
	want := trace.HTTPError{
		Status:  http.StatusNotFound,
		Code:    trace.CodeNotFound,
		Message: "user not found",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
	}
}

func TestToHTTPErrorHidesInternalError(t *testing.T) {
	const secret = "password=do-not-expose"
	err := trace.Wrap(errors.New(secret), "query users table")

	got := trace.ToHTTPError(err)

	if got != (trace.HTTPError{
		Status:  http.StatusInternalServerError,
		Code:    trace.CodeInternal,
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
		want trace.HTTPError
	}{
		{
			name: "bad request",
			err:  trace.BadParameter("invalid email"),
			want: trace.HTTPError{Status: http.StatusBadRequest, Code: trace.CodeBadRequest, Message: "invalid email"},
		},
		{
			name: "unauthenticated",
			err:  trace.Unauthenticated("expired bearer token"),
			want: trace.HTTPError{Status: http.StatusUnauthorized, Code: trace.CodeUnauthenticated, Message: "authentication required"},
		},
		{
			name: "access denied",
			err:  trace.AccessDenied("missing admin role"),
			want: trace.HTTPError{Status: http.StatusForbidden, Code: trace.CodeAccessDenied, Message: "access denied"},
		},
		{
			name: "not found",
			err:  trace.NotFound("user not found"),
			want: trace.HTTPError{Status: http.StatusNotFound, Code: trace.CodeNotFound, Message: "user not found"},
		},
		{
			name: "already exists",
			err:  trace.AlreadyExists("email already registered"),
			want: trace.HTTPError{Status: http.StatusConflict, Code: trace.CodeAlreadyExists, Message: "email already registered"},
		},
		{
			name: "conflict",
			err:  trace.Conflict("version mismatch"),
			want: trace.HTTPError{Status: http.StatusConflict, Code: trace.CodeConflict, Message: "version mismatch"},
		},
		{
			name: "limit exceeded",
			err:  trace.LimitExceeded("quota exceeded"),
			want: trace.HTTPError{Status: http.StatusTooManyRequests, Code: trace.CodeLimitExceeded, Message: "quota exceeded"},
		},
		{
			name: "canceled",
			err:  trace.FromContext(canceledContext),
			want: trace.HTTPError{Status: 499, Code: trace.CodeCanceled, Message: "request canceled"},
		},
		{
			name: "not implemented",
			err:  trace.NotImplemented("admin export is unfinished"),
			want: trace.HTTPError{Status: http.StatusNotImplemented, Code: trace.CodeNotImplemented, Message: "internal server error"},
		},
		{
			name: "unavailable",
			err:  trace.ConnectionProblem(errors.New("dial tcp database.internal"), "database unavailable"),
			want: trace.HTTPError{Status: http.StatusServiceUnavailable, Code: trace.CodeUnavailable, Message: "internal server error"},
		},
		{
			name: "timeout",
			err:  trace.Timeout(errors.New("upstream deadline"), "inventory request timed out"),
			want: trace.HTTPError{Status: http.StatusGatewayTimeout, Code: trace.CodeTimeout, Message: "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trace.ToHTTPError(tt.err)
			if got != tt.want {
				t.Fatalf("ToHTTPError() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestToHTTPErrorSupportsValidatedCustomProviders(t *testing.T) {
	t.Run("wrapped custom provider", func(t *testing.T) {
		err := trace.Wrap(customHTTPError{response: trace.HTTPError{
			Status:  http.StatusTeapot,
			Code:    trace.ErrorCode("teapot"),
			Message: "short and stout",
		}}, "internal context")

		got := trace.ToHTTPError(err)
		want := trace.HTTPError{
			Status:  http.StatusTeapot,
			Code:    trace.ErrorCode("teapot"),
			Message: "short and stout",
		}
		if got != want {
			t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
		}
	})

	tests := []struct {
		name     string
		provided trace.HTTPError
		want     trace.HTTPError
	}{
		{
			name:     "status below error range",
			provided: trace.HTTPError{Status: http.StatusOK, Code: trace.ErrorCode("success"), Message: "unsafe"},
			want:     trace.HTTPError{Status: http.StatusInternalServerError, Code: trace.CodeInternal, Message: "internal server error"},
		},
		{
			name:     "empty code",
			provided: trace.HTTPError{Status: http.StatusTeapot, Message: "unsafe"},
			want:     trace.HTTPError{Status: http.StatusInternalServerError, Code: trace.CodeInternal, Message: "internal server error"},
		},
		{
			name:     "custom 5xx message",
			provided: trace.HTTPError{Status: http.StatusServiceUnavailable, Code: trace.ErrorCode("dependency_down"), Message: "redis password=secret"},
			want:     trace.HTTPError{Status: http.StatusServiceUnavailable, Code: trace.ErrorCode("dependency_down"), Message: "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trace.ToHTTPError(customHTTPError{response: tt.provided})
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

	status, got := trace.ErrorResponseFor(err, "01KREQUEST")
	want := trace.ErrorResponse{
		Error: trace.ErrorBody{
			Code:      trace.CodeNotFound,
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
	status, got := trace.ErrorResponseFor(trace.BadParameter("invalid email"), "")
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
	err := trace.WriteError(recorder, trace.NotFound("user not found"), "01KREQUEST")
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
		if err := trace.WriteError(recorder, nil, "01KREQUEST"); err != nil {
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

		gotErr := trace.WriteError(writer, trace.NotFound("user not found"), "")
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
		wantCode            trace.ErrorCode
		wantUnauthenticated bool
		wantAccessDenied    bool
	}{
		{
			name:                "unauthorized",
			status:              http.StatusUnauthorized,
			wantCode:            trace.CodeUnauthenticated,
			wantUnauthenticated: true,
		},
		{
			name:             "forbidden",
			status:           http.StatusForbidden,
			wantCode:         trace.CodeAccessDenied,
			wantAccessDenied: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &http.Response{
				StatusCode: tt.status,
				Status:     http.StatusText(tt.status),
			}

			err := trace.FromHTTPResponse(response, []byte("upstream diagnostic"))
			if trace.IsUnauthenticated(err) != tt.wantUnauthenticated {
				t.Fatalf("IsUnauthenticated() = %v", trace.IsUnauthenticated(err))
			}
			if trace.IsAccessDenied(err) != tt.wantAccessDenied {
				t.Fatalf("IsAccessDenied() = %v", trace.IsAccessDenied(err))
			}
			if got := trace.ToHTTPError(err).Code; got != tt.wantCode {
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

	got := trace.ToHTTPError(err)
	want := trace.HTTPError{
		Status:  http.StatusGatewayTimeout,
		Code:    trace.CodeTimeout,
		Message: "internal server error",
	}
	if got != want {
		t.Fatalf("ToHTTPError() = %#v, want %#v", got, want)
	}
}
