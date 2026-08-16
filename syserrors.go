// @index Translation of operating system, filesystem, and network errors into typed trace categories.
package trace

import (
	"context"
	"errors"
	"io/fs"
	"os"
)

// @intent recognize timeouts reported by net.Error and similar types without importing net.
type timeoutReporter interface {
	Timeout() bool
}

// @intent recognize temporary transport failures reported by net.Error without importing net.
type temporaryReporter interface {
	Temporary() bool
}

// @intent give callers one entry point that turns stdlib failures into trace's typed categories.
// @domainRule the first matching category wins, checked from most specific to least.
// @domainRule an error that already carries a trace category is returned unchanged so
// existing classification is never downgraded.
// @domainRule the converted error carries no trace message, so the operating system string
// stays in the cause for debugging instead of becoming a client-facing HTTP message.
// @ensures returns nil for nil input and the original error when no category applies.
// ConvertSystemError converts an operating system, filesystem, or network error
// into the matching typed trace error.
//
// The original error stays reachable through errors.Is and errors.As. Errors that
// no rule matches are returned unchanged rather than forced into a category, so
// the caller can still wrap them itself.
func ConvertSystemError(err error) error {
	if err == nil {
		return nil
	}

	if alreadyTyped(err) {
		return err
	}

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return WrapNotFound(err)
	case errors.Is(err, fs.ErrExist):
		return WrapAlreadyExists(err)
	case errors.Is(err, fs.ErrPermission):
		return WrapAccessDenied(err)
	case errors.Is(err, context.Canceled):
		return Canceled(err)
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, os.ErrDeadlineExceeded):
		return Timeout(err)
	}

	var timeout timeoutReporter
	if errors.As(err, &timeout) && timeout.Timeout() {
		return Timeout(err)
	}

	if converted, ok := convertErrno(err); ok {
		return converted
	}

	var temporary temporaryReporter
	if errors.As(err, &temporary) && temporary.Temporary() {
		return ConnectionProblem(err)
	}

	return err
}

// @intent avoid re-classifying an error that already carries a trace category.
func alreadyTyped(err error) bool {
	return IsNotFound(err) ||
		IsAlreadyExists(err) ||
		IsBadParameter(err) ||
		IsNotImplemented(err) ||
		IsUnauthenticated(err) ||
		IsAccessDenied(err) ||
		IsConflict(err) ||
		IsConnectionProblem(err) ||
		IsLimitExceeded(err) ||
		IsTimeout(err) ||
		IsCanceled(err)
}
