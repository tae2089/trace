// Package trace provides enhanced error handling with stack traces,
// structured logging support, and full compatibility with Go 1.20+ error handling.
// @index Core error wrapping, stack capture, and structured error inspection APIs.
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

// @intent describe one recorded call site so errors and logs can point back to their origin.
// Frame represents a single stack frame
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// @intent render a single frame in a compact file-line-function format for debugging output.
// @ensures returns a string containing the file name, line number, and function name.
// String returns a human-readable representation of the frame
func (f Frame) String() string {
	return fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Function)
}

// @intent represent an ordered stack trace that can be rendered or serialized with an error.
// Frames is a slice of stack frames
type Frames []Frame

// @intent render the recorded stack trace in call-order for compact error messages.
// @ensures returns an empty string when no frames are present.
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

// @intent serve as the canonical trace-aware error wrapper carrying cause, message, frames, and structured fields.
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

// @intent produce the primary human-readable message for traced errors and their wrapped causes.
// @ensures includes frame and cause information when available.
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

// @intent expose the wrapped cause so traced errors participate in standard Go error traversal.
// @ensures returns the original wrapped error.
// Unwrap implements the errors.Unwrap interface for Go 1.13+
func (e *TraceError) Unwrap() error {
	return e.Err
}

// @intent support concise and verbose formatting styles for traced errors.
// @ensures %+v output includes stack frames and structured fields when present.
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

// @intent serialize traced errors into structured slog fields without losing cause or stack context.
// @ensures includes message, cause, trace, and fields only when those values are present.
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

// @intent convert recorded frames into a JSON-friendly structure for structured logging output.
// @ensures preserves file, line, and function data for each frame.
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

// @intent capture the caller information that anchors trace output to a concrete source location.
// @intent let packages outside this module mint errors whose first frame points at their own caller.
// @ensures returns an empty frame when runtime caller information is unavailable.
// CaptureFrame captures a single stack frame at the given skip level.
//
// skip follows runtime.Caller: 0 is CaptureFrame itself, 1 is its immediate
// caller, and 2 is the caller of that function. Constructors that want the
// frame to point at their own caller pass 2.
func CaptureFrame(skip int) Frame {
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

// @intent preserve the original error while adding call-site debugging context.
// @domainRule typed errors stay discoverable through the Err chain for errors.Is and errors.As.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// @ensures returns nil unchanged when no source error is provided.
// Wrap wraps an error with stack trace information.
// If err is nil, Wrap returns nil.
// Wrap always creates a new TraceError, preserving the original error
// (including typed errors like NotFoundError) in the Err field.
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}

	frame := CaptureFrame(2)
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
	}
}

// @intent preserve the original error while adding formatted debugging context.
// @domainRule typed errors stay discoverable through the Err chain for errors.Is and errors.As.
// @ensures returns nil unchanged when no source error is provided.
// Wrapf wraps an error with stack trace and a formatted message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return wrapInternal(err, fmt.Sprintf(format, args...), CaptureFrame(2))
}

// @intent share the common TraceError wrapping path used by formatted and typed error helpers.
// @ensures prepends the supplied frame to any existing trace frames.
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
	}
}

// @intent enrich an error with structured diagnostics that can flow into logs and HTTP responses.
// @domainRule field attachment must not discard the wrapped error chain.
// @mutates adds key-value metadata to the returned TraceError fields map.
// @ensures returns nil unchanged when no source error is provided.
// WrapWithFields wraps an error with stack trace and structured fields
func WrapWithFields(err error, fields map[string]any, msg ...string) error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, msg...)
	if te, ok := wrapped.(*TraceError); ok && len(fields) > 0 {
		te.Fields = copyFields(fields)
	}
	return wrapped
}

// @intent create a fresh traceable application error at the current call site.
// @ensures records the current call site as the first trace frame.
// @ensures allocates structured fields storage lazily on first field attachment.
// New creates a new error with stack trace
func New(msg string) error {
	frame := CaptureFrame(2)
	return &TraceError{
		Message: msg,
		Frames:  Frames{frame},
	}
}

// @intent create a traceable error from formatted application context.
// @ensures records the current call site as the first trace frame.
// @ensures allocates structured fields storage lazily on first field attachment.
// Errorf creates a new error with formatted message and stack trace
func Errorf(format string, args ...any) error {
	frame := CaptureFrame(2)
	return &TraceError{
		Message: fmt.Sprintf(format, args...),
		Frames:  Frames{frame},
	}
}

// @intent normalize mixed message-and-arguments inputs into one human-readable error message.
// @ensures returns an empty string when no message arguments are supplied.
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

// @intent expose captured stack frames for diagnostics without leaking mutable internal state.
// @ensures returns a defensive copy of the stored frames when trace data exists.
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

// @intent expose structured trace metadata for logging, transport, or inspection layers.
// @ensures returns a defensive copy of the stored fields when trace data exists.
// GetFields extracts fields from an error if available.
// Returns a copy of the fields to prevent external mutation.
func GetFields(err error) map[string]any {
	var te *TraceError
	if errors.As(err, &te) {
		return copyFields(te.Fields)
	}
	return nil
}

// @intent attach one diagnostic attribute without mutating the original error instance.
// @domainRule built-in trace wrapper types are preserved when replacing the inner TraceError.
// @mutates adds or replaces a single field on the returned error copy.
// @ensures returns nil unchanged when no source error is provided.
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
		wte.Fields = map[string]any{key: value}
	}
	return wrapped
}

// @intent attach multiple diagnostic attributes without mutating the original error instance.
// @domainRule later field values override earlier values for the same key.
// @mutates merges the provided fields into the returned error copy.
// @ensures returns nil unchanged when no source error is provided.
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
	if wte, ok := wrapped.(*TraceError); ok && len(fields) > 0 {
		wte.Fields = copyFields(fields)
	}
	return wrapped
}

// @intent duplicate a TraceError so field updates can preserve immutable error semantics.
// @ensures returns a copy with duplicated frames and fields.
func cloneTraceError(te *TraceError) *TraceError {
	return &TraceError{
		Err:     te.Err,
		Message: te.Message,
		Frames:  append(Frames(nil), te.Frames...),
		Fields:  copyFields(te.Fields),
	}
}

// @intent swap the inner TraceError while preserving known wrapper types around it.
// @domainRule built-in typed wrappers are recreated so errors.Is and errors.As continue to work.
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
	case *UnauthenticatedError:
		if e.TraceError == original {
			return &UnauthenticatedError{TraceError: replacement}
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
	}

	return replacement
}

// @intent defensively copy structured error fields before mutation or external exposure.
// @ensures returns a new map containing every existing field.
func copyFields(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src)+1)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// @intent render a full developer-facing report for nested and aggregated error chains.
// @domainRule aggregate errors must include every branch in the rendered report.
// @ensures returns an empty string when no error is provided.
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

// @intent recursively expand an error tree into the developer-facing debug report.
// @ensures traverses both single-cause chains and aggregate branches.
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

// @intent extract the safest high-level message to show outside debugging channels.
// @domainRule prefer explicit TraceError messages before falling back to wrapped causes.
// @ensures returns an empty string when no error is provided.
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

// @intent iterate every error reachable from wrapped and aggregated trace errors.
// @domainRule aggregate branches are traversed recursively, not flattened into a single message.
// @ensures yields each reachable error once per traversal path until the consumer stops.
// Errors returns an iterator over the error chain (Go 1.23+).
// It yields each error in the chain by following Unwrap() error and
// recursively traversing Unwrap() []error (e.g., AggregateError).
func Errors(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		errorsWalk(err, yield)
	}
}

// @intent recursively traverse every reachable error in a chain or aggregate until the consumer stops.
// @ensures returns false as soon as the yield function asks traversal to stop.
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
