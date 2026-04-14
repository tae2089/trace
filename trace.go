// Package trace provides enhanced error handling with stack traces,
// structured logging support, and full compatibility with Go 1.20+ error handling.
package trace

import (
	"errors"
	"fmt"
	"io"
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

// LogValue implements slog.LogValuer for structured logging
func (e *TraceError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 4+len(e.Fields))

	attrs = append(attrs, slog.String("message", e.Message))

	if e.Err != nil {
		attrs = append(attrs, slog.String("cause", e.Err.Error()))
	}

	if len(e.Frames) > 0 {
		frameAttrs := make([]any, 0, len(e.Frames)*3)
		for i, f := range e.Frames {
			frameAttrs = append(frameAttrs,
				slog.Group(strconv.Itoa(i),
					slog.String("file", f.File),
					slog.Int("line", f.Line),
					slog.String("func", f.Function),
				),
			)
		}
		attrs = append(attrs, slog.Any("trace", frameAttrs))
	}

	for k, v := range e.Fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	return slog.GroupValue(attrs...)
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
		return &TraceError{
			Err:     err,
			Message: "",
			Frames:  append(Frames(nil), te.Frames...),
			Fields:  newFields,
		}
	}

	wrapped := Wrap(err)
	if wte, ok := wrapped.(*TraceError); ok {
		wte.Fields[key] = value
	}
	return wrapped
}

// WithFields adds multiple fields to the error.
// It returns a new wrapper error with the fields added, preserving the original error immutably.
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
		return &TraceError{
			Err:     err,
			Message: "",
			Frames:  append(Frames(nil), te.Frames...),
			Fields:  newFields,
		}
	}

	wrapped := Wrap(err)
	if wte, ok := wrapped.(*TraceError); ok {
		for k, v := range fields {
			wte.Fields[k] = v
		}
	}
	return wrapped
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

	// Walk the error chain
	depth := 0
	for e := err; e != nil; e = errors.Unwrap(e) {
		indent := strings.Repeat("  ", depth)
		fmt.Fprintf(&b, "%s[%d] %T\n", indent, depth, e)

		if te, ok := e.(*TraceError); ok {
			if te.Message != "" {
				fmt.Fprintf(&b, "%s    Message: %s\n", indent, te.Message)
			}
			for _, f := range te.Frames {
				fmt.Fprintf(&b, "%s    at %s\n", indent, f.String())
			}
			if len(te.Fields) > 0 {
				fmt.Fprintf(&b, "%s    Fields:\n", indent)
				for k, v := range te.Fields {
					fmt.Fprintf(&b, "%s      %s: %v\n", indent, k, v)
				}
			}
		} else {
			fmt.Fprintf(&b, "%s    %s\n", indent, e.Error())
		}

		depth++
	}

	return b.String()
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
