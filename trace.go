// Package trace provides enhanced error handling with stack traces,
// structured logging support, and full compatibility with Go 1.20+ error handling.
package trace

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Frame represents a single stack frame
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// String returns a human-readable representation of the frame
func (f Frame) String() string {
	return fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Function)
}

// Frames is a slice of stack frames
type Frames []Frame

// String returns a human-readable representation of all frames
func (fs Frames) String() string {
	if len(fs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("[")
	for i, f := range fs {
		if i > 0 {
			b.WriteString(" <- ")
		}
		b.WriteString(f.File)
		b.WriteString(":")
		b.WriteString(strconv.Itoa(f.Line))
	}
	b.WriteString("]")
	return b.String()
}

// TraceError is the core error type that captures stack traces
type TraceError struct {
	// Original error being wrapped
	Err error
	// Message is additional context
	Message string
	// Frames contains the stack trace
	Frames Frames
	// Fields contains structured data for logging
	Fields map[string]any
}

// Error implements the error interface
func (e *TraceError) Error() string {
	var b strings.Builder

	if len(e.Frames) > 0 {
		b.WriteString(e.Frames.String())
		b.WriteString(" ")
	}

	if e.Message != "" {
		b.WriteString(e.Message)
		if e.Err != nil {
			b.WriteString("\n→ ")
		}
	}

	if e.Err != nil {
		b.WriteString(e.Err.Error())
	}

	return b.String()
}

// Unwrap implements the errors.Unwrap interface for Go 1.13+
func (e *TraceError) Unwrap() error {
	return e.Err
}

// Format implements fmt.Formatter for customizable output
func (e *TraceError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// Detailed format with full stack trace
			io.WriteString(s, e.Error())
			io.WriteString(s, "\nStack trace:\n")
			for _, f := range e.Frames {
				fmt.Fprintf(s, "  %s\n", f.String())
			}
			if len(e.Fields) > 0 {
				io.WriteString(s, "Fields:\n")
				for k, v := range e.Fields {
					fmt.Fprintf(s, "  %s: %v\n", k, v)
				}
			}
			return
		}
		io.WriteString(s, e.Error())
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// LogValue implements slog.LogValuer for structured logging.
// Schema: {"message":..., "cause":..., "trace":[{"file":..., "line":..., "func":...}], "fields":{...}}
func (e *TraceError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 4)

	attrs = append(attrs, slog.String("message", e.Message))

	if e.Err != nil {
		attrs = append(attrs, slog.String("cause", e.Err.Error()))
	}

	if len(e.Frames) > 0 {
		attrs = append(attrs, slog.Any("trace", framesToSerializable(e.Frames)))
	}

	if len(e.Fields) > 0 {
		attrs = append(attrs, slog.Any("fields", e.Fields))
	}

	return slog.GroupValue(attrs...)
}

func framesToSerializable(frames Frames) []map[string]any {
	result := make([]map[string]any, len(frames))
	for i, f := range frames {
		result[i] = map[string]any{
			"file": f.File,
			"line": f.Line,
			"func": f.Function,
		}
	}
	return result
}

// captureFrame captures a single stack frame at the given skip level
func captureFrame(skip int) Frame {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return Frame{}
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = filepath.Base(fn.Name())
	}

	return Frame{
		Function: funcName,
		File:     filepath.Base(file),
		Line:     line,
	}
}

// Wrap wraps an error with stack trace information.
// If err is nil, Wrap returns nil.
// Wrap always creates a new TraceError, preserving the original error
// (including typed errors like NotFoundError) in the Err field.
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}

	frame := captureFrame(2)
	var message string
	if len(msg) > 0 {
		message = msg[0]
	}

	var existingFrames Frames
	var te *TraceError
	if errors.As(err, &te) {
		existingFrames = te.Frames
	}

	return &TraceError{
		Err:     err,
		Message: message,
		Frames:  append(Frames{frame}, existingFrames...),
		Fields:  make(map[string]any),
	}
}

// Wrapf wraps an error with stack trace and a formatted message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return wrapInternal(err, fmt.Sprintf(format, args...), captureFrame(2))
}

func wrapInternal(err error, msg string, frame Frame) error {
	var existingFrames Frames
	var te *TraceError
	if errors.As(err, &te) {
		existingFrames = te.Frames
	}

	return &TraceError{
		Err:     err,
		Message: msg,
		Frames:  append(Frames{frame}, existingFrames...),
		Fields:  make(map[string]any),
	}
}

// WrapWithFields wraps an error with stack trace and structured fields
func WrapWithFields(err error, fields map[string]any, msg ...string) error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, msg...)
	if te, ok := wrapped.(*TraceError); ok {
		for k, v := range fields {
			te.Fields[k] = v
		}
	}
	return wrapped
}

// New creates a new error with stack trace
func New(msg string) error {
	frame := captureFrame(2)
	return &TraceError{
		Message: msg,
		Frames:  Frames{frame},
		Fields:  make(map[string]any),
	}
}

// Errorf creates a new error with formatted message and stack trace
func Errorf(format string, args ...any) error {
	frame := captureFrame(2)
	return &TraceError{
		Message: fmt.Sprintf(format, args...),
		Frames:  Frames{frame},
		Fields:  make(map[string]any),
	}
}

// formatMessage formats message and args similar to fmt.Sprintf
func formatMessage(msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 {
		return ""
	}

	if len(msgAndArgs) == 1 {
		if s, ok := msgAndArgs[0].(string); ok {
			return s
		}
		return fmt.Sprint(msgAndArgs[0])
	}

	if format, ok := msgAndArgs[0].(string); ok {
		return fmt.Sprintf(format, msgAndArgs[1:]...)
	}

	return fmt.Sprint(msgAndArgs...)
}

// GetFrames extracts frames from an error if available.
// Returns a copy of the frames to prevent external mutation.
func GetFrames(err error) Frames {
	var te *TraceError
	if errors.As(err, &te) {
		cp := make(Frames, len(te.Frames))
		copy(cp, te.Frames)
		return cp
	}
	return nil
}

// GetFields extracts fields from an error if available.
// Returns a copy of the fields to prevent external mutation.
func GetFields(err error) map[string]any {
	var te *TraceError
	if errors.As(err, &te) {
		return copyFields(te.Fields)
	}
	return nil
}

// WithField adds a field to the error for structured logging.
// It returns a new wrapper error with the field added, preserving the original error immutably.
func WithField(err error, key string, value any) error {
	if err == nil {
		return nil
	}

	var te *TraceError
	if errors.As(err, &te) {
		newFields := copyFields(te.Fields)
		newFields[key] = value
		clone := cloneTraceError(te)
		clone.Fields = newFields
		return replaceTraceError(err, te, clone)
	}

	wrapped := Wrap(err)
	if wte, ok := wrapped.(*TraceError); ok {
		wte.Fields[key] = value
	}
	return wrapped
}

func WithFields(err error, fields map[string]any) error {
	if err == nil {
		return nil
	}

	var te *TraceError
	if errors.As(err, &te) {
		newFields := copyFields(te.Fields)
		for k, v := range fields {
			newFields[k] = v
		}
		clone := cloneTraceError(te)
		clone.Fields = newFields
		return replaceTraceError(err, te, clone)
	}

	wrapped := Wrap(err)
	if wte, ok := wrapped.(*TraceError); ok {
		for k, v := range fields {
			wte.Fields[k] = v
		}
	}
	return wrapped
}

func cloneTraceError(te *TraceError) *TraceError {
	return &TraceError{
		Err:     te.Err,
		Message: te.Message,
		Frames:  append(Frames(nil), te.Frames...),
		Fields:  copyFields(te.Fields),
	}
}

func replaceTraceError(err error, original *TraceError, replacement *TraceError) error {
	if err == original {
		return replacement
	}

	switch e := err.(type) {
	case *NotFoundError:
		if e.TraceError == original {
			return &NotFoundError{TraceError: replacement}
		}
	case *AlreadyExistsError:
		if e.TraceError == original {
			return &AlreadyExistsError{TraceError: replacement}
		}
	case *BadParameterError:
		if e.TraceError == original {
			return &BadParameterError{TraceError: replacement}
		}
	case *NotImplementedError:
		if e.TraceError == original {
			return &NotImplementedError{TraceError: replacement}
		}
	case *AccessDeniedError:
		if e.TraceError == original {
			return &AccessDeniedError{TraceError: replacement}
		}
	case *ConflictError:
		if e.TraceError == original {
			return &ConflictError{TraceError: replacement}
		}
	case *ConnectionProblemError:
		if e.TraceError == original {
			return &ConnectionProblemError{TraceError: replacement}
		}
	case *LimitExceededError:
		if e.TraceError == original {
			return &LimitExceededError{TraceError: replacement}
		}
	case *TimeoutError:
		if e.TraceError == original {
			return &TimeoutError{TraceError: replacement}
		}
	case *CanceledError:
		if e.TraceError == original {
			return &CanceledError{TraceError: replacement}
		}
	case *httpStatusError:
		if e.TraceError == original {
			return &httpStatusError{TraceError: replacement, statusCode: e.statusCode}
		}
	}

	return replacement
}

func copyFields(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src)+1)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// DebugReport returns a detailed report of the error chain
func DebugReport(err error) string {
	if err == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("Error Report:\n")
	b.WriteString("=============\n\n")

	debugReportWalk(&b, err, 0)

	return b.String()
}

func debugReportWalk(b *strings.Builder, err error, depth int) {
	if err == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(b, "%s[%d] %T\n", indent, depth, err)

	if te, ok := err.(*TraceError); ok {
		if te.Message != "" {
			fmt.Fprintf(b, "%s    Message: %s\n", indent, te.Message)
		}
		for _, f := range te.Frames {
			fmt.Fprintf(b, "%s    at %s\n", indent, f.String())
		}
		if len(te.Fields) > 0 {
			fmt.Fprintf(b, "%s    Fields:\n", indent)
			for k, v := range te.Fields {
				fmt.Fprintf(b, "%s      %s: %v\n", indent, k, v)
			}
		}
	} else {
		fmt.Fprintf(b, "%s    %s\n", indent, err.Error())
	}

	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		for _, child := range multi.Unwrap() {
			debugReportWalk(b, child, depth+1)
		}
		return
	}

	if next := errors.Unwrap(err); next != nil {
		debugReportWalk(b, next, depth+1)
	}
}

// UserMessage returns a user-friendly error message without stack traces
func UserMessage(err error) string {
	if err == nil {
		return ""
	}

	var te *TraceError
	if errors.As(err, &te) {
		if te.Message != "" {
			return te.Message
		}
		if te.Err != nil {
			return UserMessage(te.Err)
		}
	}

	return err.Error()
}

// Errors returns an iterator over the error chain (Go 1.23+).
// It yields each error in the chain by following Unwrap() error and
// recursively traversing Unwrap() []error (e.g., AggregateError).
func Errors(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		errorsWalk(err, yield)
	}
}

func errorsWalk(err error, yield func(error) bool) bool {
	if err == nil {
		return true
	}
	if !yield(err) {
		return false
	}
	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		for _, child := range multi.Unwrap() {
			if !errorsWalk(child, yield) {
				return false
			}
		}
		return true
	}
	return errorsWalk(errors.Unwrap(err), yield)
}
