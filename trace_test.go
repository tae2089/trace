package trace_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tae2089/trace"
)

// Example: Basic wrapping
func TestBasicWrap(t *testing.T) {
	originalErr := errors.New("database connection failed")
	wrapped := trace.Wrap(originalErr, "failed to fetch user")

	if wrapped == nil {
		t.Fatal("expected wrapped error")
	}

	// Should contain both messages
	if !strings.Contains(wrapped.Error(), "failed to fetch user") {
		t.Errorf("error should contain message: %v", wrapped)
	}

	// Original error should be extractable
	if !errors.Is(wrapped, originalErr) {
		t.Error("errors.Is should match original error")
	}
}

// Example: Stack trace accumulation
func TestStackTraceAccumulation(t *testing.T) {
	err := repository()
	if err == nil {
		t.Fatal("expected error")
	}

	// Should have multiple frames
	frames := trace.GetFrames(err)
	if len(frames) < 3 {
		t.Errorf("expected at least 3 frames, got %d", len(frames))
	}

	// Check verbose output
	verbose := fmt.Sprintf("%+v", err)
	if !strings.Contains(verbose, "Stack trace") {
		t.Error("verbose format should contain stack trace")
	}
}

func repository() error {
	return trace.Wrap(dbQuery(), "repository layer")
}

func dbQuery() error {
	return trace.Wrap(errors.New("connection refused"), "db query failed")
}

// Example: Typed errors
func TestTypedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		checkFunc  func(error) bool
		statusCode int
	}{
		{
			name:       "NotFound",
			err:        trace.NotFound("user %s not found", "abc123"),
			checkFunc:  trace.IsNotFound,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "AlreadyExists",
			err:        trace.AlreadyExists("user already exists"),
			checkFunc:  trace.IsAlreadyExists,
			statusCode: http.StatusConflict,
		},
		{
			name:       "BadParameter",
			err:        trace.BadParameter("invalid email format"),
			checkFunc:  trace.IsBadParameter,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "AccessDenied",
			err:        trace.AccessDenied("insufficient permissions"),
			checkFunc:  trace.IsAccessDenied,
			statusCode: http.StatusForbidden,
		},
		{
			name:       "LimitExceeded",
			err:        trace.LimitExceeded("rate limit exceeded"),
			checkFunc:  trace.IsLimitExceeded,
			statusCode: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.checkFunc(tt.err) {
				t.Errorf("%s check failed", tt.name)
			}
			if code := trace.GetHTTPStatusCode(tt.err); code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, code)
			}
		})
	}
}

// Example: Wrapping typed errors preserves type
func TestWrapPreservesType(t *testing.T) {
	original := trace.NotFound("user not found")
	wrapped := trace.Wrap(original, "service layer")
	wrapped = trace.Wrap(wrapped, "handler layer")

	if !trace.IsNotFound(wrapped) {
		t.Error("wrapped error should still be NotFound")
	}
}

// Example: With fields for structured logging
func TestWithFields(t *testing.T) {
	err := trace.NotFound("user not found")
	err = trace.WithFields(err, map[string]any{
		"user_id":    "abc123",
		"request_id": "req-456",
	})

	fields := trace.GetFields(err)
	if fields["user_id"] != "abc123" {
		t.Error("field not preserved")
	}
}

// Example: slog integration
func TestSlogIntegration(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	err := trace.Wrap(errors.New("db error"), "failed to fetch user")
	err = trace.WithField(err, "user_id", "123")

	logger.Error("operation failed", trace.SlogError(err))

	output := buf.String()
	if !strings.Contains(output, "db error") {
		t.Error("log should contain error message")
	}
}

// Example: Aggregate errors (Go 1.20+)
func TestAggregateErrors(t *testing.T) {
	err1 := trace.NotFound("user not found")
	err2 := trace.BadParameter("invalid email")
	err3 := trace.AccessDenied("no permission")

	combined := trace.Aggregate(err1, err2, err3)

	// Should match all error types
	if !trace.IsNotFound(combined) {
		t.Error("should match NotFound")
	}
	if !trace.IsBadParameter(combined) {
		t.Error("should match BadParameter")
	}
	if !trace.IsAccessDenied(combined) {
		t.Error("should match AccessDenied")
	}
}

// Example: HTTP middleware
func TestHTTPMiddleware(t *testing.T) {
	handler := trace.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		return trace.NotFound("resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var resp trace.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Code != http.StatusNotFound {
		t.Error("response code mismatch")
	}
}

// Example: Context integration
func TestContextIntegration(t *testing.T) {
	ctx := context.Background()
	ctx = trace.ContextWithTraceID(ctx, "trace-123")
	ctx = trace.ContextWithField(ctx, "user_id", "user-456")

	err := trace.WrapContext(ctx, errors.New("db error"), "query failed")

	fields := trace.GetFields(err)
	if fields["trace_id"] != "trace-123" {
		t.Error("trace_id not preserved")
	}
	if fields["user_id"] != "user-456" {
		t.Error("user_id not preserved")
	}
}

// Example: Context cancellation
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := trace.FromContext(ctx)
	if !trace.IsCanceled(err) {
		t.Error("should be canceled error")
	}
}

// Example: Result type (generic)
func TestResultType(t *testing.T) {
	// Success case
	result := trace.Ok(42)
	if !result.IsOk() {
		t.Error("should be ok")
	}
	if result.Unwrap() != 42 {
		t.Error("value mismatch")
	}

	// Error case
	errResult := trace.Err[int](errors.New("failed"))
	if !errResult.IsErr() {
		t.Error("should be error")
	}
	if errResult.UnwrapOr(0) != 0 {
		t.Error("should return default")
	}
}

// Example: Try function
func TestTry(t *testing.T) {
	// Simulating a function that returns (value, error)
	getValue := func(success bool) (string, error) {
		if success {
			return "hello", nil
		}
		return "", errors.New("failed")
	}

	// Success
	result := trace.Try(getValue(true))
	if v, err := result.Value(); err != nil || v != "hello" {
		t.Error("try should succeed")
	}

	// Failure
	result = trace.Try(getValue(false))
	if result.IsOk() {
		t.Error("try should fail")
	}
}

// Example: Pipeline
func TestPipeline(t *testing.T) {
	result, err := trace.NewPipeline(10).
		Then(func(n int) (int, error) {
			return n * 2, nil
		}).
		Then(func(n int) (int, error) {
			return n + 5, nil
		}).
		Result()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 25 {
		t.Errorf("expected 25, got %d", result)
	}
}

// Example: Pipeline with error
func TestPipelineWithError(t *testing.T) {
	_, err := trace.NewPipeline(10).
		Then(func(n int) (int, error) {
			return 0, errors.New("step 1 failed")
		}).
		Then(func(n int) (int, error) {
			// This should not execute
			t.Error("should not reach here")
			return n, nil
		}).
		Result()

	if err == nil {
		t.Error("expected error")
	}
}

// Example: Debug report
func TestDebugReport(t *testing.T) {
	err := deepError()
	report := trace.DebugReport(err)

	if !strings.Contains(report, "Error Report") {
		t.Error("should contain header")
	}
	if !strings.Contains(report, "TraceError") {
		t.Error("should contain type info")
	}
}

func deepError() error {
	return level1()
}

func level1() error {
	return trace.Wrap(level2(), "level 1")
}

func level2() error {
	return trace.Wrap(level3(), "level 2")
}

func level3() error {
	return trace.Wrap(errors.New("root cause"), "level 3")
}

// Example: User-friendly message
func TestUserMessage(t *testing.T) {
	err := trace.Wrap(
		trace.Wrap(errors.New("SQLSTATE 23505"), "database error"),
		"failed to create user",
	)

	msg := trace.UserMessage(err)
	if msg != "failed to create user" {
		t.Errorf("unexpected message: %s", msg)
	}
}

// Example: Real-world usage pattern
func TestRealWorldUsage(t *testing.T) {
	// Simulate a layered application
	userID := "nonexistent"

	err := handleGetUser(userID)

	// Handler can check error type and respond appropriately
	if trace.IsNotFound(err) {
		// Return 404 to client
		t.Log("User not found - returning 404")
	}

	// For logging, get full trace
	t.Logf("Debug report:\n%s", trace.DebugReport(err))
}

func handleGetUser(userID string) error {
	user, err := serviceGetUser(userID)
	if err != nil {
		return trace.Wrap(err, "handler: get user failed")
	}
	_ = user
	return nil
}

func serviceGetUser(userID string) (*User, error) {
	user, err := repoFindUser(userID)
	if err != nil {
		return nil, trace.Wrap(err, "service: failed to get user %s", userID)
	}
	return user, nil
}

func repoFindUser(userID string) (*User, error) {
	// Simulate database query
	err := sql.ErrNoRows
	if err == sql.ErrNoRows {
		return nil, trace.WrapNotFound(err, "user %s not found in database", userID)
	}
	return &User{ID: userID}, nil
}

type User struct {
	ID string
}

// Benchmark
func BenchmarkWrap(b *testing.B) {
	err := errors.New("original error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trace.Wrap(err, "wrapped")
	}
}

func BenchmarkWrapChain(b *testing.B) {
	err := errors.New("original error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := trace.Wrap(err, "level 1")
		e = trace.Wrap(e, "level 2")
		e = trace.Wrap(e, "level 3")
		_ = e
	}
}

func BenchmarkIsNotFound(b *testing.B) {
	err := trace.NotFound("not found")
	err = trace.Wrap(err, "wrapped")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trace.IsNotFound(err)
	}
}

// Example output
func ExampleWrap() {
	err := errors.New("connection refused")
	wrapped := trace.Wrap(err, "failed to connect to database")
	fmt.Println(wrapped)
}

func ExampleNotFound() {
	err := trace.NotFound("user %s not found", "alice")
	if trace.IsNotFound(err) {
		fmt.Println("User not found!")
	}
	// Output: User not found!
}

func ExampleResult() {
	result := trace.Try(os.Open("nonexistent.txt"))
	value := result.UnwrapOr(nil)
	if value == nil {
		fmt.Println("File not found, using default")
	}
	// Output: File not found, using default
}
