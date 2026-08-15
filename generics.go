// @index Generic Result and Pipeline helpers for composing trace-aware success and failure flows.
package trace

import (
	"errors"
)

// @intent model success and failure as one value so callers can compose operations without losing trace context.
// Result represents either a value or an error (similar to Rust's Result)
type Result[T any] struct {
	value T
	err   error
}

// @intent wrap a successful value in the Result abstraction without adding trace overhead.
// @ensures returns a Result with no error.
// Ok creates a successful Result
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// @intent convert a failure into a Result while preserving trace metadata for downstream composition.
// @ensures wraps the provided error with trace context before storing it.
// Err creates a failed Result
func Err[T any](err error) Result[T] {
	return Result[T]{err: Wrap(err)}
}

// @intent create a failed Result directly from an application message when no underlying error exists.
// @ensures records the current call site as the first trace frame.
// ErrMsg creates a failed Result with a message
func ErrMsg[T any](msg string) Result[T] {
	frame := CaptureFrame(2)
	return Result[T]{
		err: &TraceError{
			Message: msg,
			Frames:  Frames{frame},
		},
	}
}

// @intent let callers branch on success without unpacking the Result payload.
// @ensures returns true only when the Result holds no error.
// IsOk returns true if the Result contains a value
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// @intent let callers branch on failure without unpacking the Result payload.
// @ensures returns true only when the Result holds an error.
// IsErr returns true if the Result contains an error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// @intent extract the successful value in contexts where failure should abort immediately.
// @domainRule panics with the stored error when the Result is failed.
// Unwrap returns the value or panics if there's an error
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

// @intent provide a fallback value when the Result is failed.
// @ensures returns the stored value on success and the provided default on failure.
// UnwrapOr returns the value or the provided default
func (r Result[T]) UnwrapOr(defaultVal T) T {
	if r.err != nil {
		return defaultVal
	}
	return r.value
}

// @intent derive a fallback value from the failure reason when the Result is failed.
// @ensures calls the fallback function only when the Result holds an error.
// UnwrapOrElse returns the value or calls the function to get a default
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	if r.err != nil {
		return fn(r.err)
	}
	return r.value
}

// @intent bridge Result back into Go's conventional value-plus-error calling style.
// @ensures returns the stored value and stored error unchanged.
// Value returns the value and error separately (Go-style)
func (r Result[T]) Value() (T, error) {
	return r.value, r.err
}

// @intent expose the failure component directly for interoperability with Go error handling.
// @ensures returns nil for successful results.
// Error returns the error if present
func (r Result[T]) Error() error {
	return r.err
}

// @intent transform successful results while leaving failures untouched.
// @ensures preserves the existing error without calling fn when the Result is failed.
// Map transforms the value if present
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return Result[U]{value: fn(r.value)}
}

// @intent rewrite failure values without changing successful payloads.
// @ensures leaves successful results unchanged.
// MapErr transforms the error if present
func MapErr[T any](r Result[T], fn func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return Result[T]{err: fn(r.err)}
}

// @intent chain operations that already return Result values without manual error branching.
// @ensures skips fn and preserves the existing error when the Result is failed.
// FlatMap chains Result-returning operations
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return fn(r.value)
}

// @intent lift ordinary Go return pairs into a Result for composable error handling.
// @ensures wraps non-nil errors with trace context before storing them.
// Try wraps a function call that returns (T, error) into a Result
func Try[T any](value T, err error) Result[T] {
	if err != nil {
		frame := CaptureFrame(2)
		return Result[T]{
			err: &TraceError{
				Err:    err,
				Frames: Frames{frame},
			},
		}
	}
	return Result[T]{value: value}
}

// @intent provide concise success-only access in contexts where failure should panic.
// @domainRule panics when the Result contains an error.
// Must unwraps a Result, panicking on error
func Must[T any](r Result[T]) T {
	return r.Unwrap()
}

// @intent collapse a Go-style value-plus-error pair when failure should panic immediately.
// @domainRule panics with a wrapped trace error when err is non-nil.
// MustValue unwraps a (value, error) pair, panicking on error
func MustValue[T any](value T, err error) T {
	if err != nil {
		panic(Wrap(err))
	}
	return value
}

// @intent combine many Result values into one collection while preserving all failures.
// @domainRule any failed input produces an aggregated error instead of a partial success.
// Collect collects multiple Results into a single Result containing a slice
func Collect[T any](results ...Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))
	var errs []error

	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			values = append(values, r.value)
		}
	}

	if len(errs) > 0 {
		return Result[[]T]{err: Aggregate(errs...)}
	}

	return Result[[]T]{value: values}
}

// @intent perform typed error extraction without repeating target boilerplate at call sites.
// @ensures returns the zero value of T and false when no matching error is found.
// As is a generic version of errors.As
func As[T error](err error) (T, bool) {
	var target T
	if errors.As(err, &target) {
		return target, true
	}
	return target, false
}

// @intent model stepwise transformations that should stop automatically after the first failure.
// Pipeline allows chaining operations that may fail
type Pipeline[T any] struct {
	value T
	err   error
}

// @intent start a chain of trace-aware transformations from an initial value.
// @ensures returns a pipeline with no initial error.
// NewPipeline creates a new pipeline with an initial value
func NewPipeline[T any](value T) *Pipeline[T] {
	return &Pipeline[T]{value: value}
}

// @intent apply the next transformation only while the pipeline remains successful.
// @ensures wraps new step failures with trace context before storing them.
// Then executes the function if no error has occurred
func (p *Pipeline[T]) Then(fn func(T) (T, error)) *Pipeline[T] {
	if p.err != nil {
		return p
	}
	p.value, p.err = fn(p.value)
	if p.err != nil {
		p.err = Wrap(p.err)
	}
	return p
}

// @intent run a side-effecting validation or action without changing the pipeline value.
// @ensures wraps new step failures with trace context before storing them.
// ThenDo executes a function that doesn't modify the value
func (p *Pipeline[T]) ThenDo(fn func(T) error) *Pipeline[T] {
	if p.err != nil {
		return p
	}
	p.err = fn(p.value)
	if p.err != nil {
		p.err = Wrap(p.err)
	}
	return p
}

// @intent expose the pipeline state as a conventional Go value-plus-error pair.
// @ensures returns the current pipeline value and error unchanged.
// Result returns the final value and error
func (p *Pipeline[T]) Result() (T, error) {
	return p.value, p.err
}

// @intent convert pipeline state into the Result abstraction for further composition.
// @ensures returns a Result containing the current pipeline value and error.
// ToResult converts the pipeline to a Result
func (p *Pipeline[T]) ToResult() Result[T] {
	return Result[T]{value: p.value, err: p.err}
}

// @intent give failed pipelines a chance to replace their error with a recovery value or new error.
// @ensures calls fn only when the pipeline is currently failed.
// Recover attempts to recover from an error
func (p *Pipeline[T]) Recover(fn func(error) (T, error)) *Pipeline[T] {
	if p.err == nil {
		return p
	}
	p.value, p.err = fn(p.err)
	return p
}

// @intent replace a failed pipeline with a caller-provided default value.
// @ensures clears the stored error when recovery is applied.
// RecoverWith recovers with a default value
func (p *Pipeline[T]) RecoverWith(defaultVal T) *Pipeline[T] {
	if p.err != nil {
		p.value = defaultVal
		p.err = nil
	}
	return p
}

// @intent continue a pipeline while changing the value type and preserving trace-aware failure handling.
// @ensures carries forward existing pipeline errors without invoking fn.
// TransformPipeline transforms between different types
func TransformPipeline[T, U any](p *Pipeline[T], fn func(T) (U, error)) *Pipeline[U] {
	if p.err != nil {
		return &Pipeline[U]{err: p.err}
	}
	value, err := fn(p.value)
	if err != nil {
		err = Wrap(err)
	}
	return &Pipeline[U]{value: value, err: err}
}
