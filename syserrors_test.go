package trace_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/tae2089/trace/v2"
)

type reportingError struct {
	message   string
	timeout   bool
	temporary bool
}

func (e *reportingError) Error() string   { return e.message }
func (e *reportingError) Timeout() bool   { return e.timeout }
func (e *reportingError) Temporary() bool { return e.temporary }

func TestConvertSystemErrorClassifiesPortableCategories(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "absent.txt")
	_, openErr := os.Open(missingPath)
	if openErr == nil {
		t.Fatal("expected opening a missing file to fail")
	}

	tests := []struct {
		name  string
		err   error
		check func(error) bool
	}{
		{"not exist", openErr, trace.IsNotFound},
		{"exist", &fs.PathError{Op: "mkdir", Path: "/tmp/x", Err: fs.ErrExist}, trace.IsAlreadyExists},
		{"permission", &fs.PathError{Op: "open", Path: "/tmp/x", Err: fs.ErrPermission}, trace.IsAccessDenied},
		{"canceled", context.Canceled, trace.IsCanceled},
		{"deadline exceeded", context.DeadlineExceeded, trace.IsTimeout},
		{"os deadline exceeded", os.ErrDeadlineExceeded, trace.IsTimeout},
		{"reports timeout", &reportingError{message: "i/o timeout", timeout: true}, trace.IsTimeout},
		{"reports temporary", &reportingError{message: "try again", temporary: true}, trace.IsConnectionProblem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted := trace.ConvertSystemError(tt.err)
			if !tt.check(converted) {
				t.Fatalf("ConvertSystemError(%v) was not classified: %#v", tt.err, converted)
			}
			if !errors.Is(converted, tt.err) {
				t.Fatal("converted error must keep the original reachable through errors.Is")
			}
		})
	}
}

func TestConvertSystemErrorKeepsOSTextOutOfTheTraceMessage(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "private-key.pem")
	_, openErr := os.Open(secretPath)
	if openErr == nil {
		t.Fatal("expected opening a missing file to fail")
	}

	converted := trace.ConvertSystemError(openErr)

	if got := trace.UserMessage(converted); got != openErr.Error() {
		t.Fatalf("UserMessage should fall back to the cause, got %q", got)
	}
	var te *trace.TraceError
	if !errors.As(converted, &te) {
		t.Fatal("converted error should carry a TraceError")
	}
	if te.Message != "" {
		t.Fatalf("converted error must carry no trace message, got %q", te.Message)
	}
}

func TestConvertSystemErrorLeavesOtherErrorsAlone(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := trace.ConvertSystemError(nil); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("unclassifiable", func(t *testing.T) {
		original := errors.New("something else entirely")
		if got := trace.ConvertSystemError(original); got != original {
			t.Fatalf("expected the original error back, got %#v", got)
		}
	})

	t.Run("already typed", func(t *testing.T) {
		original := trace.BadParameter("invalid email")
		if got := trace.ConvertSystemError(original); got != original {
			t.Fatalf("an already typed error must not be reclassified, got %#v", got)
		}
	})

	t.Run("already typed wrapping a system error", func(t *testing.T) {
		original := trace.WrapAccessDenied(&fs.PathError{Op: "open", Path: "/x", Err: fs.ErrNotExist}, "denied")
		got := trace.ConvertSystemError(original)
		if !trace.IsAccessDenied(got) {
			t.Fatal("existing classification must win over the inner system error")
		}
	})
}

func TestCanceledWrapsCause(t *testing.T) {
	if got := trace.Canceled(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}

	cause := errors.New("client hung up")
	err := trace.Canceled(cause, "request canceled")

	if !trace.IsCanceled(err) {
		t.Fatal("expected a canceled error")
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected the cause to stay reachable")
	}
	if len(trace.GetFrames(err)) == 0 {
		t.Fatal("expected a captured frame")
	}
}

func TestWrapLimitExceededPreservesCauseAndFields(t *testing.T) {
	if got := trace.WrapLimitExceeded(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}

	cause := trace.WithField(trace.Wrap(errors.New("too many"), "inner"), "req_id", "r1")
	err := trace.WrapLimitExceeded(cause, "quota exceeded")

	if !trace.IsLimitExceeded(err) {
		t.Fatal("expected a limit exceeded error")
	}
	if !trace.IsRetryable(err) {
		t.Fatal("limit exceeded errors should be retryable")
	}
	if fields := trace.GetFields(err); fields["req_id"] != "r1" {
		t.Fatalf("expected inner fields to carry forward, got %v", fields)
	}
}
