// @index slog integration for serializing trace errors and automatically expanding error attributes in log records.
package trace

import (
	"context"
	"log/slog"
)

// @intent convert trace errors into a structured slog group that preserves message, cause, trace, and fields.
// @domainRule plain errors degrade to a minimal message-only structure instead of losing log compatibility.
// @ensures returns an empty slog.Attr when no error is provided.
// SlogError returns slog attributes for an error.
// Schema matches TraceError.LogValue: {"message":..., "cause":..., "trace":[...], "fields":{...}}
func SlogError(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	if te := findTraceError(err); te != nil {
		attrs := []slog.Attr{
			slog.String("message", te.Message),
			slog.Any("cause", errorCause(te.Err)),
		}
		if len(te.Frames) > 0 {
			attrs = append(attrs, slog.Any("trace", framesToSerializable(te.Frames)))
		}
		if len(te.Fields) > 0 {
			attrs = append(attrs, slog.Any("fields", te.Fields))
		}
		return slog.Group("error", attrsToAny(attrs)...)
	}

	return slog.Group("error",
		slog.String("message", err.Error()),
	)
}

// @intent adapt slog attribute slices to the variadic shape required by slog.Group.
// @ensures preserves attribute order in the returned slice.
func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, a := range attrs {
		result[i] = a
	}
	return result
}

// @intent expose the trace-aware slog representation as a Value for callers building custom attributes.
// @ensures returns an empty string value when no error is provided.
// SlogErrorValue returns a slog.Value for an error
func SlogErrorValue(err error) slog.Value {
	if err == nil {
		return slog.StringValue("")
	}

	if te := findTraceError(err); te != nil {
		return te.LogValue()
	}

	return slog.StringValue(err.Error())
}

// @intent extract the printable cause payload used in trace-aware slog serialization.
// @ensures returns nil when no error cause exists.
func errorCause(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}

// @intent adapt slog handlers so error attributes are rewritten into the package's trace-aware schema.
// ErrorHandler wraps an slog.Handler to automatically extract trace information
type ErrorHandler struct {
	slog.Handler
}

// @intent wrap slog handlers so error attributes are consistently expanded into trace-aware structure.
// @domainRule nil handlers fall back to slog.DiscardHandler for safe optional logging.
// @ensures always returns a non-nil ErrorHandler.
// NewErrorHandler creates a new ErrorHandler wrapping the given handler
func NewErrorHandler(h slog.Handler) *ErrorHandler {
	if h == nil {
		h = slog.DiscardHandler
	}
	return &ErrorHandler{Handler: h}
}

// @intent defer level checks to the wrapped handler so logging enablement stays consistent.
// @ensures returns the wrapped handler's enabled decision unchanged.
func (h *ErrorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// @intent rewrite incoming slog records so embedded error attributes use the trace logging schema.
// @sideEffect creates a replacement slog.Record and forwards it to the wrapped handler.
// @ensures non-error attributes are preserved verbatim.
func (h *ErrorHandler) Handle(ctx context.Context, r slog.Record) error {
	// Create a new record with enhanced error attributes
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	r.Attrs(func(a slog.Attr) bool {
		// Check if this attribute is an error
		if a.Key == "error" || a.Key == "err" {
			if err, ok := a.Value.Any().(error); ok {
				newRecord.AddAttrs(SlogError(err))
				return true
			}
		}
		newRecord.AddAttrs(a)
		return true
	})

	return h.Handler.Handle(ctx, newRecord)
}

// @intent preserve trace-aware error rewriting when callers derive a handler with additional attributes.
// @ensures returns another ErrorHandler wrapping the derived handler.
// WithAttrs returns a new handler with the given attributes
func (h *ErrorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ErrorHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// @intent preserve trace-aware error rewriting when callers derive a grouped handler.
// @ensures returns another ErrorHandler wrapping the grouped handler.
// WithGroup returns a new handler with the given group name
func (h *ErrorHandler) WithGroup(name string) slog.Handler {
	return &ErrorHandler{Handler: h.Handler.WithGroup(name)}
}

// @intent centralize level-specific error logging so all severities share the same trace serialization path.
// @ensures does nothing when the input error is nil.
func logAtLevel(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, err error, attrs ...slog.Attr) {
	if err == nil {
		return
	}
	allAttrs := make([]any, 0, len(attrs)+1)
	allAttrs = append(allAttrs, SlogError(err))
	for _, a := range attrs {
		allAttrs = append(allAttrs, a)
	}
	logger.Log(ctx, level, msg, allAttrs...)
}

// @intent log an error at error level using the trace serialization schema.
// @ensures does nothing when the input error is nil.
func LogError(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	logAtLevel(ctx, logger, slog.LevelError, msg, err, attrs...)
}

// @intent log an error at warn level using the trace serialization schema.
// @ensures does nothing when the input error is nil.
func LogWarn(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	logAtLevel(ctx, logger, slog.LevelWarn, msg, err, attrs...)
}

// @intent log an error at debug level using the trace serialization schema.
// @ensures does nothing when the input error is nil.
func LogDebug(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	logAtLevel(ctx, logger, slog.LevelDebug, msg, err, attrs...)
}
