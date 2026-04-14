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
	"time"

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
	err := serviceLayer()
	if err == nil {
		t.Fatal("expected error")
	}

	// Should have multiple frames (serviceLayer + repository + dbQuery = 3)
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

func serviceLayer() error {
	return trace.Wrap(repository(), "service layer")
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
			err:        trace.NotFound(fmt.Sprintf("user %s not found", "abc123")),
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

// Example: Error message with arrow-style newlines
func TestErrorMessageArrowFormat(t *testing.T) {
	// Create a layered error
	rootErr := errors.New("sql: no rows in result set")
	repoErr := trace.Wrap(rootErr, "user 123 not found in database")
	serviceErr := trace.Wrap(repoErr, "service: failed to get user 123")

	// Error message should be formatted with arrows and newlines
	errMsg := serviceErr.Error()

	// Should contain newlines
	if !strings.Contains(errMsg, "\n") {
		t.Errorf("error message should contain newlines, got: %s", errMsg)
	}

	// Should contain arrow separator
	if !strings.Contains(errMsg, "→") {
		t.Errorf("error message should contain arrow (→), got: %s", errMsg)
	}

	// Verify the format: each layer on its own line
	lines := strings.Split(errMsg, "\n")
	// At minimum, we should have 3 lines (service, repo, root)
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d: %s", len(lines), errMsg)
	}
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
		return nil, trace.Wrapf(err, "service: failed to get user %s", userID)
	}
	return user, nil
}

func repoFindUser(userID string) (*User, error) {
	// Simulate database query
	err := sql.ErrNoRows
	if err == sql.ErrNoRows {
		return nil, trace.WrapNotFound(err, fmt.Sprintf("user %s not found in database", userID))
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
	for range b.N {
		_ = trace.Wrap(err, "wrapped")
	}
}

func BenchmarkWrapChain(b *testing.B) {
	err := errors.New("original error")
	b.ResetTimer()
	for range b.N {
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
	for range b.N {
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
	err := trace.NotFound(fmt.Sprintf("user %s not found", "alice"))
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

// P1: slog JSON schema — trace should serialize as []map[string]any
func TestSlogTraceJSONSchema(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	err := trace.Wrap(errors.New("db error"), "query failed")
	err = trace.WithField(err, "user_id", "u1")
	logger.Error("op failed", trace.SlogError(err))

	var parsed map[string]any
	if e := json.Unmarshal([]byte(buf.String()), &parsed); e != nil {
		t.Fatalf("invalid JSON: %v", e)
	}

	errObj, ok := parsed["error"].(map[string]any)
	if !ok {
		t.Fatal("missing error object")
	}

	if errObj["message"] != "query failed" {
		t.Errorf("unexpected message: %v", errObj["message"])
	}
	if errObj["cause"] != "db error" {
		t.Errorf("unexpected cause: %v", errObj["cause"])
	}

	traceArr, ok := errObj["trace"].([]any)
	if !ok || len(traceArr) == 0 {
		t.Fatal("trace should be a non-empty array")
	}
	frame0, ok := traceArr[0].(map[string]any)
	if !ok {
		t.Fatal("trace element should be an object")
	}
	for _, key := range []string{"file", "line", "func"} {
		if _, exists := frame0[key]; !exists {
			t.Errorf("trace frame missing key %q", key)
		}
	}

	fieldsObj, ok := errObj["fields"].(map[string]any)
	if !ok {
		t.Fatal("fields should be an object")
	}
	if fieldsObj["user_id"] != "u1" {
		t.Errorf("field user_id not found: %v", fieldsObj)
	}
}

// P2: WithField/WithFields preserves typed wrapper
func TestWithFieldPreservesTypedWrapper(t *testing.T) {
	err := trace.NotFound("user not found")
	err = trace.WithField(err, "user_id", "abc")
	err = trace.WithFields(err, map[string]any{"tenant": "t1"})

	if !trace.IsNotFound(err) {
		t.Error("should still be NotFound after WithField/WithFields")
	}

	fields := trace.GetFields(err)
	if fields["user_id"] != "abc" {
		t.Error("user_id field missing")
	}
	if fields["tenant"] != "t1" {
		t.Error("tenant field missing")
	}
}

// P3: WrapNotFound preserves existing frames from inner TraceError
func TestWrapTypedPreservesFrames(t *testing.T) {
	inner := trace.Wrap(errors.New("root"), "inner layer")
	innerFrames := trace.GetFrames(inner)

	wrapped := trace.WrapNotFound(inner, "not found")
	wrappedFrames := trace.GetFrames(wrapped)

	if len(wrappedFrames) < len(innerFrames)+1 {
		t.Errorf("WrapNotFound should accumulate frames: got %d, inner had %d", len(wrappedFrames), len(innerFrames))
	}

	if !trace.IsNotFound(wrapped) {
		t.Error("should be NotFound")
	}
}

// P3: WrapNotFound preserves existing fields from inner TraceError
func TestWrapTypedPreservesFields(t *testing.T) {
	inner := trace.Wrap(errors.New("root"), "inner")
	inner = trace.WithField(inner, "key1", "val1")

	wrapped := trace.WrapNotFound(inner, "not found")
	fields := trace.GetFields(wrapped)

	if fields["key1"] != "val1" {
		t.Errorf("WrapNotFound should preserve inner fields, got: %v", fields)
	}
}

// P3: WrapHTTPError preserves frames/fields
func TestWrapHTTPErrorPreservesFramesAndFields(t *testing.T) {
	inner := trace.Wrap(errors.New("root"), "inner")
	inner = trace.WithField(inner, "req_id", "r1")
	innerFrameCount := len(trace.GetFrames(inner))

	wrapped := trace.WrapHTTPError(inner, 503, "service down")
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

// P4: DebugReport walks AggregateError children
func TestDebugReportAggregateError(t *testing.T) {
	err1 := trace.NotFound("item 1 missing")
	err2 := trace.BadParameter("invalid input")
	agg := trace.Aggregate(err1, err2)

	report := trace.DebugReport(agg)

	if !strings.Contains(report, "item 1 missing") {
		t.Error("report should contain first child message")
	}
	if !strings.Contains(report, "invalid input") {
		t.Error("report should contain second child message")
	}
	if !strings.Contains(report, "AggregateError") {
		t.Error("report should contain AggregateError type")
	}
}

// P5: ConnectionProblem(nil) returns nil
func TestConnectionProblemNilGuard(t *testing.T) {
	if err := trace.ConnectionProblem(nil, "should be nil"); err != nil {
		t.Errorf("ConnectionProblem(nil) should return nil, got: %v", err)
	}
}

// P5: Timeout(nil) returns nil
func TestTimeoutNilGuard(t *testing.T) {
	if err := trace.Timeout(nil, "should be nil"); err != nil {
		t.Errorf("Timeout(nil) should return nil, got: %v", err)
	}
}

func TestCollectAllOk(t *testing.T) {
	r := trace.Collect(trace.Ok(1), trace.Ok(2), trace.Ok(3))
	if r.IsErr() {
		t.Fatal("expected Ok")
	}
	vals := r.Unwrap()
	if len(vals) != 3 || vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Errorf("unexpected values: %v", vals)
	}
}

func TestCollectWithErrors(t *testing.T) {
	r := trace.Collect(
		trace.Ok(1),
		trace.Err[int](errors.New("fail1")),
		trace.Err[int](errors.New("fail2")),
	)
	if r.IsOk() {
		t.Fatal("expected Err")
	}
	errStr := r.Error().Error()
	if !strings.Contains(errStr, "fail1") || !strings.Contains(errStr, "fail2") {
		t.Errorf("aggregate should contain both errors: %s", errStr)
	}
}

func TestMapErr(t *testing.T) {
	r := trace.Err[int](errors.New("original"))
	mapped := trace.MapErr(r, func(err error) error {
		return fmt.Errorf("wrapped: %w", err)
	})
	if mapped.IsOk() {
		t.Fatal("expected Err")
	}
	if !strings.Contains(mapped.Error().Error(), "wrapped") {
		t.Errorf("error should be mapped: %v", mapped.Error())
	}
}

func TestMapErrOnOk(t *testing.T) {
	r := trace.Ok(42)
	mapped := trace.MapErr(r, func(err error) error {
		return fmt.Errorf("should not be called")
	})
	if !mapped.IsOk() || mapped.Unwrap() != 42 {
		t.Error("MapErr on Ok should pass through")
	}
}

func TestErrMsg(t *testing.T) {
	r := trace.ErrMsg[string]("something failed")
	if r.IsOk() {
		t.Fatal("expected Err")
	}
	if !strings.Contains(r.Error().Error(), "something failed") {
		t.Errorf("unexpected error: %v", r.Error())
	}
}

func TestFlatMap(t *testing.T) {
	r := trace.Ok(10)
	result := trace.FlatMap(r, func(v int) trace.Result[string] {
		return trace.Ok(fmt.Sprintf("val=%d", v))
	})
	if !result.IsOk() || result.Unwrap() != "val=10" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestFlatMapPropagatesError(t *testing.T) {
	r := trace.Err[int](errors.New("initial"))
	result := trace.FlatMap(r, func(v int) trace.Result[string] {
		t.Fatal("should not be called")
		return trace.Ok("")
	})
	if result.IsOk() {
		t.Error("FlatMap should propagate error")
	}
}

func TestPipelineThenDo(t *testing.T) {
	var sideEffect int
	result, err := trace.NewPipeline(5).
		ThenDo(func(v int) error {
			sideEffect = v * 10
			return nil
		}).
		Result()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("ThenDo should not modify value, got %d", result)
	}
	if sideEffect != 50 {
		t.Errorf("side effect not executed, got %d", sideEffect)
	}
}

func TestPipelineThenDoError(t *testing.T) {
	_, err := trace.NewPipeline(5).
		ThenDo(func(v int) error {
			return errors.New("side effect failed")
		}).
		Then(func(v int) (int, error) {
			t.Fatal("should not execute after ThenDo error")
			return v, nil
		}).
		Result()

	if err == nil {
		t.Error("expected error from ThenDo")
	}
}

func TestPipelineRecover(t *testing.T) {
	result, err := trace.NewPipeline(0).
		Then(func(v int) (int, error) {
			return 0, errors.New("oops")
		}).
		Recover(func(err error) (int, error) {
			return 99, nil
		}).
		Result()

	if err != nil {
		t.Errorf("expected recovery, got error: %v", err)
	}
	if result != 99 {
		t.Errorf("expected 99, got %d", result)
	}
}

func TestPipelineRecoverWith(t *testing.T) {
	result, err := trace.NewPipeline(0).
		Then(func(v int) (int, error) {
			return 0, errors.New("oops")
		}).
		RecoverWith(42).
		Result()

	if err != nil {
		t.Errorf("expected recovery, got error: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestPipelineRecoverWithNoError(t *testing.T) {
	result, err := trace.NewPipeline(10).
		RecoverWith(42).
		Result()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 10 {
		t.Errorf("RecoverWith should not change value when no error, got %d", result)
	}
}

func TestTransformPipeline(t *testing.T) {
	p := trace.NewPipeline(42)
	result := trace.TransformPipeline(p, func(v int) (string, error) {
		return fmt.Sprintf("num=%d", v), nil
	})

	val, err := result.Result()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "num=42" {
		t.Errorf("expected 'num=42', got %q", val)
	}
}

func TestTransformPipelinePropagatesError(t *testing.T) {
	p := trace.NewPipeline(0).Then(func(v int) (int, error) {
		return 0, errors.New("upstream fail")
	})

	result := trace.TransformPipeline(p, func(v int) (string, error) {
		t.Fatal("should not be called")
		return "", nil
	})

	_, err := result.Result()
	if err == nil {
		t.Error("expected error propagation")
	}
}

// F1: context.Cause preserves cancellation cause
func TestFromContextWithCancelCause(t *testing.T) {
	causeErr := errors.New("shutdown requested")
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(causeErr)

	err := trace.FromContext(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if !trace.IsCanceled(err) {
		t.Error("should be canceled")
	}
	if !strings.Contains(err.Error(), "shutdown requested") {
		t.Errorf("cause should be preserved, got: %s", err.Error())
	}
}

func TestFromContextWithCancelCauseNilFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := trace.FromContext(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if !trace.IsCanceled(err) {
		t.Error("should be canceled")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("should fall back to ctx.Err(), got: %s", err.Error())
	}
}

// F2: Errors iterator
func TestErrorsIteratorSingleChain(t *testing.T) {
	root := errors.New("root")
	wrapped := trace.Wrap(root, "layer1")
	wrapped = trace.Wrap(wrapped, "layer2")

	var count int
	for range trace.Errors(wrapped) {
		count++
	}
	if count < 3 {
		t.Errorf("expected at least 3 errors in chain, got %d", count)
	}
}

func TestErrorsIteratorAggregate(t *testing.T) {
	err1 := trace.NotFound("a")
	err2 := trace.BadParameter("b")
	agg := trace.Aggregate(err1, err2)

	var messages []string
	for e := range trace.Errors(agg) {
		messages = append(messages, e.Error())
	}
	found := strings.Join(messages, " | ")
	if !strings.Contains(found, "a") || !strings.Contains(found, "b") {
		t.Errorf("should traverse all children, got: %s", found)
	}
}

func TestErrorsIteratorNil(t *testing.T) {
	var count int
	for range trace.Errors(nil) {
		count++
	}
	if count != 0 {
		t.Error("nil error should yield nothing")
	}
}

func TestErrorsIteratorEarlyBreak(t *testing.T) {
	root := errors.New("root")
	wrapped := trace.Wrap(root, "l1")
	wrapped = trace.Wrap(wrapped, "l2")

	var count int
	for range trace.Errors(wrapped) {
		count++
		if count == 1 {
			break
		}
	}
	if count != 1 {
		t.Error("early break should stop iteration")
	}
}

// F3: DetachedContext
func TestDetachedContext(t *testing.T) {
	parent := context.Background()
	parent = trace.ContextWithTraceID(parent, "trace-abc")
	parent = trace.ContextWithField(parent, "env", "test")

	detached := trace.DetachedContext(parent)

	if trace.TraceIDFromContext(detached) != "trace-abc" {
		t.Error("detached context should preserve values")
	}
	fields := trace.FieldsFromContext(detached)
	if fields["env"] != "test" {
		t.Error("detached context should preserve fields")
	}
}

func TestDetachedContextNotCanceled(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	detached := trace.DetachedContext(parent)
	cancel()

	select {
	case <-detached.Done():
		t.Error("detached context should NOT be canceled when parent is canceled")
	default:
	}
}

// F4: OnCancel
func TestOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := trace.NewContextualizer(ctx)

	called := make(chan struct{})
	c.OnCancel(func() {
		close(called)
	})

	cancel()

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Error("OnCancel callback should have been called")
	}
}

func TestOnCancelStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := trace.NewContextualizer(ctx)

	var called bool
	stop := c.OnCancel(func() {
		called = true
	})

	if !stop() {
		t.Error("stop should return true before cancellation")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)

	if called {
		t.Error("callback should not run after stop()")
	}
}

// F5: WithCancelCause / WithTimeoutCause
func TestWithCancelCause(t *testing.T) {
	cause := errors.New("manual shutdown")
	ctx, cancel := trace.WithCancelCause(context.Background())
	cancel(cause)

	<-ctx.Done()
	if context.Cause(ctx) != cause {
		t.Errorf("expected cause %v, got %v", cause, context.Cause(ctx))
	}
}

func TestWithTimeoutCause(t *testing.T) {
	cause := errors.New("slow query")
	ctx, cancel := trace.WithTimeoutCause(context.Background(), 10*time.Millisecond, cause)
	defer cancel()

	<-ctx.Done()
	if context.Cause(ctx) != cause {
		t.Errorf("expected cause %v, got %v", cause, context.Cause(ctx))
	}
}

// F6: NewErrorHandler nil fallback uses DiscardHandler
func TestNewErrorHandlerNilUsesDiscard(t *testing.T) {
	h := trace.NewErrorHandler(nil)
	if h == nil {
		t.Fatal("should return non-nil handler")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
	}

	err := h.Handle(context.Background(), slog.NewRecord(
		time.Now(), slog.LevelError, "test", 0,
	))
	if err != nil {
		t.Errorf("DiscardHandler should not error: %v", err)
	}
}

// F1+F5 integration: WithCancelCause + FromContext preserves cause
func TestFromContextWithTimeoutCause(t *testing.T) {
	cause := errors.New("db timeout")
	ctx, cancel := trace.WithTimeoutCause(context.Background(), 10*time.Millisecond, cause)
	defer cancel()

	<-ctx.Done()
	err := trace.FromContext(ctx)

	if !trace.IsTimeout(err) && !trace.IsCanceled(err) {
		t.Error("should be timeout or canceled")
	}
	if !strings.Contains(err.Error(), "db timeout") {
		t.Errorf("cause should be preserved, got: %s", err.Error())
	}
}
