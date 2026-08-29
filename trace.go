// Package trace records where Go errors originate and how they propagate.
package trace

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const maxNonComparableDebugDepth = 64

// New returns a traced error with the given message and origin call site.
//
// @intent create a traced error while keeping the public surface limited to ordinary error values.
func New(message string) error {
	return &traceError{
		err: errors.New(message),
		pc:  callerPC(),
	}
}

// Errorf returns a traced formatted error and preserves fmt.Errorf %w traversal.
//
// @intent create a formatted traced error and preserve fmt.Errorf %w traversal semantics.
func Errorf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	return &traceError{
		err:   err,
		cause: unwrapCause(err),
		pc:    callerPC(),
	}
}

// Wrap returns err with one additional context message and propagation call site.
//
// @intent add one propagation context and call site without changing the wrapped error's meaning.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	if message == "" {
		return err
	}
	return &traceError{
		message: message,
		err:     err,
		cause:   err,
		pc:      callerPC(),
	}
}

// Wrapf returns err with one formatted context message and propagation call site.
//
// @intent add one formatted propagation context and call site around an existing cause.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	if format == "" {
		return err
	}
	return &traceError{
		message: fmt.Sprintf(format, args...),
		err:     err,
		cause:   err,
		pc:      callerPC(),
	}
}

// @intent carry one immutable trace frame plus the display error and traversable cause.
type traceError struct {
	message string
	err     error
	cause   error
	pc      uintptr
}

// @intent render the ordinary error message without exposing source locations.
func (e *traceError) Error() string {
	if e.message == "" {
		return e.err.Error()
	}
	return e.message + ": " + e.err.Error()
}

// @intent expose only the semantic cause used by standard errors traversal.
func (e *traceError) Unwrap() error {
	return e.cause
}

// @intent keep ordinary fmt output clean while reserving call-site traces for %+v.
func (e *traceError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			writeDebug(s, e)
			return
		}
		io.WriteString(s, e.Error())
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	default:
		fmt.Fprintf(s, "%%!%c(%T=%s)", verb, e, e.Error())
	}
}

// @intent capture the external caller's program counter without resolving symbols on the hot path.
func callerPC() uintptr {
	var pcs [1]uintptr
	if runtime.Callers(3, pcs[:]) == 0 {
		return 0
	}
	return pcs[0]
}

// @intent track the current traversal path so cyclic error trees do not loop forever.
type debugState struct {
	seen               map[error]struct{}
	nonComparableDepth int
}

// @intent format the full human-readable trace tree for detailed debugging.
func writeDebug(w io.Writer, err error) {
	state := debugState{
		seen: map[error]struct{}{},
	}
	writeErrorDebug(w, err, 0, &state)
}

// @intent walk standard single-cause and multi-cause error trees while preserving branch order.
func writeErrorDebug(w io.Writer, err error, depth int, state *debugState) {
	if err == nil {
		return
	}
	if isNonComparable(err) && state.nonComparableDepth >= maxNonComparableDebugDepth {
		writeIndent(w, depth)
		io.WriteString(w, "depth limit reached: ")
		io.WriteString(w, err.Error())
		io.WriteString(w, "\n")
		return
	}

	entered, leave := state.enter(err)
	if !entered {
		writeIndent(w, depth)
		io.WriteString(w, "cycle detected: ")
		io.WriteString(w, err.Error())
		io.WriteString(w, "\n")
		return
	}
	defer leave()

	if traced, ok := err.(*traceError); ok {
		writeTraceDebug(w, traced, depth, state)
		return
	}

	writeIndent(w, depth)
	io.WriteString(w, "error: ")
	io.WriteString(w, err.Error())
	io.WriteString(w, "\n")

	switch unwrapped := err.(type) {
	case interface{ Unwrap() []error }:
		for i, cause := range unwrapped.Unwrap() {
			writeIndent(w, depth+1)
			io.WriteString(w, "branch ")
			io.WriteString(w, strconv.Itoa(i+1))
			io.WriteString(w, ":\n")
			writeErrorDebug(w, cause, depth+2, state)
		}
	case interface{ Unwrap() error }:
		writeErrorDebug(w, unwrapped.Unwrap(), depth+1, state)
	}
}

// @intent print one Trace frame and continue into the recorded cause or leaf error.
func writeTraceDebug(w io.Writer, err *traceError, depth int, state *debugState) {
	writeIndent(w, depth)
	if err.message == "" {
		io.WriteString(w, "trace: origin")
	} else {
		io.WriteString(w, "trace: ")
		io.WriteString(w, err.message)
	}
	io.WriteString(w, "\n")

	function, file, line := resolvePC(err.pc)
	writeIndent(w, depth+1)
	io.WriteString(w, file)
	io.WriteString(w, ":")
	io.WriteString(w, strconv.Itoa(line))
	io.WriteString(w, " ")
	io.WriteString(w, function)
	io.WriteString(w, "\n")

	if err.cause != nil {
		writeErrorDebug(w, err.cause, depth+1, state)
		return
	}
	writeErrorDebug(w, err.err, depth+1, state)
}

// @intent keep Errorf as an origin when no %w cause exists while preserving %w chains when present.
func unwrapCause(err error) error {
	switch err.(type) {
	case interface{ Unwrap() error }, interface{ Unwrap() []error }:
		return err
	default:
		return nil
	}
}

// @intent add comparable errors to the active path and count non-comparable path depth.
func (s *debugState) enter(err error) (bool, func()) {
	if err == nil {
		return true, func() {}
	}

	if isNonComparable(err) {
		s.nonComparableDepth++
		return true, func() {
			s.nonComparableDepth--
		}
	}

	if _, ok := s.seen[err]; ok {
		return false, func() {}
	}
	s.seen[err] = struct{}{}
	return true, func() {
		delete(s.seen, err)
	}
}

// @intent identify values that cannot be used for precise identity-based cycle tracking.
func isNonComparable(err error) bool {
	return err != nil && !reflect.TypeOf(err).Comparable()
}

// @intent resolve a stored program counter only when detailed formatting needs it.
func resolvePC(pc uintptr) (function, file string, line int) {
	if pc == 0 {
		return "unknown", "unknown", 0
	}

	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	function = frame.Function
	if function == "" {
		function = "unknown"
	}
	file = frame.File
	if file == "" {
		file = "unknown"
	}
	return function, file, frame.Line
}

// @intent keep nested debug output readable without making the format a machine API.
func writeIndent(w io.Writer, depth int) {
	io.WriteString(w, strings.Repeat("  ", depth))
}
