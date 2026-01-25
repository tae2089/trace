package trace

import (
	"errors"
)

// Result represents either a value or an error (similar to Rust's Result)
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a failed Result
func Err[T any](err error) Result[T] {
	return Result[T]{err: Wrap(err)}
}

// ErrMsg creates a failed Result with a message
func ErrMsg[T any](msg string) Result[T] {
	frame := captureFrame(2)
	return Result[T]{
		err: &TraceError{
			Message: msg,
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// IsOk returns true if the Result contains a value
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result contains an error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns the value or panics if there's an error
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

// UnwrapOr returns the value or the provided default
func (r Result[T]) UnwrapOr(defaultVal T) T {
	if r.err != nil {
		return defaultVal
	}
	return r.value
}

// UnwrapOrElse returns the value or calls the function to get a default
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	if r.err != nil {
		return fn(r.err)
	}
	return r.value
}

// Value returns the value and error separately (Go-style)
func (r Result[T]) Value() (T, error) {
	return r.value, r.err
}

// Error returns the error if present
func (r Result[T]) Error() error {
	return r.err
}

// Map transforms the value if present
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return Result[U]{value: fn(r.value)}
}

// MapErr transforms the error if present
func MapErr[T any](r Result[T], fn func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return Result[T]{err: fn(r.err)}
}

// FlatMap chains Result-returning operations
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	return fn(r.value)
}

// Try wraps a function call that returns (T, error) into a Result
func Try[T any](value T, err error) Result[T] {
	if err != nil {
		frame := captureFrame(2)
		return Result[T]{
			err: &TraceError{
				Err:    err,
				Frames: Frames{frame},
				Fields: make(map[string]any),
			},
		}
	}
	return Result[T]{value: value}
}

// Must unwraps a Result, panicking on error
func Must[T any](r Result[T]) T {
	return r.Unwrap()
}

// MustValue unwraps a (value, error) pair, panicking on error
func MustValue[T any](value T, err error) T {
	if err != nil {
		panic(Wrap(err))
	}
	return value
}

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

// As is a generic version of errors.As
func As[T error](err error) (T, bool) {
	var target T
	if errors.As(err, &target) {
		return target, true
	}
	return target, false
}

// TypedWrap wraps an error and returns a specific typed error
func TypedWrap[T interface {
	error
	*E
}, E any](err error, create func(*TraceError) T, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}

	frame := captureFrame(2)
	te := &TraceError{
		Err:     err,
		Message: formatMessage(msgAndArgs...),
		Frames:  Frames{frame},
		Fields:  make(map[string]any),
	}

	return create(te)
}

// Pipeline allows chaining operations that may fail
type Pipeline[T any] struct {
	value T
	err   error
}

// NewPipeline creates a new pipeline with an initial value
func NewPipeline[T any](value T) *Pipeline[T] {
	return &Pipeline[T]{value: value}
}

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

// Result returns the final value and error
func (p *Pipeline[T]) Result() (T, error) {
	return p.value, p.err
}

// ToResult converts the pipeline to a Result
func (p *Pipeline[T]) ToResult() Result[T] {
	return Result[T]{value: p.value, err: p.err}
}

// Recover attempts to recover from an error
func (p *Pipeline[T]) Recover(fn func(error) (T, error)) *Pipeline[T] {
	if p.err == nil {
		return p
	}
	p.value, p.err = fn(p.err)
	return p
}

// RecoverWith recovers with a default value
func (p *Pipeline[T]) RecoverWith(defaultVal T) *Pipeline[T] {
	if p.err != nil {
		p.value = defaultVal
		p.err = nil
	}
	return p
}

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
