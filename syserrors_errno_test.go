//go:build unix || windows

package trace_test

import (
	"errors"
	"syscall"
	"testing"

	"github.com/tae2089/trace/v2"
)

func TestConvertSystemErrorClassifiesErrno(t *testing.T) {
	tests := []struct {
		name  string
		errno syscall.Errno
		check func(error) bool
	}{
		{"connection refused", syscall.ECONNREFUSED, trace.IsConnectionProblem},
		{"connection reset", syscall.ECONNRESET, trace.IsConnectionProblem},
		{"host unreachable", syscall.EHOSTUNREACH, trace.IsConnectionProblem},
		{"network unreachable", syscall.ENETUNREACH, trace.IsConnectionProblem},
		{"broken pipe", syscall.EPIPE, trace.IsConnectionProblem},
		{"timed out", syscall.ETIMEDOUT, trace.IsTimeout},
		{"too many open files", syscall.EMFILE, trace.IsLimitExceeded},
		{"file table overflow", syscall.ENFILE, trace.IsLimitExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted := trace.ConvertSystemError(tt.errno)
			if !tt.check(converted) {
				t.Fatalf("ConvertSystemError(%v) was not classified: %#v", tt.errno, converted)
			}
			if !errors.Is(converted, tt.errno) {
				t.Fatal("converted error must keep the errno reachable through errors.Is")
			}
			if !trace.IsRetryable(converted) {
				t.Fatal("transport and limit failures should be retryable")
			}
		})
	}
}
