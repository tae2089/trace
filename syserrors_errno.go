//go:build unix || windows

// @index Errno-based classification of network and resource-exhaustion failures on unix and windows.
package trace

import (
	"errors"
	"syscall"
)

// @intent map the syscall errno values that carry a clear trace category.
// @domainRule refused, reset, unreachable, and broken-pipe errnos are transport failures.
// @domainRule descriptor-table exhaustion is a limit, not a transport failure.
// @ensures reports false when no errno rule applies, leaving the caller's fallbacks in charge.
func convertErrno(err error) (error, bool) {
	switch {
	case errors.Is(err, syscall.ECONNREFUSED),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ECONNABORTED),
		errors.Is(err, syscall.EHOSTUNREACH),
		errors.Is(err, syscall.ENETUNREACH),
		errors.Is(err, syscall.ENETDOWN),
		errors.Is(err, syscall.EPIPE):
		return ConnectionProblem(err), true
	case errors.Is(err, syscall.ETIMEDOUT):
		return Timeout(err), true
	case errors.Is(err, syscall.EMFILE), errors.Is(err, syscall.ENFILE):
		return WrapLimitExceeded(err), true
	}
	return nil, false
}
