package trace

import (
	"context"
	"errors"
	"log/slog"
)

// SlogError returns slog attributes for an error
// Usage: slog.Error("operation failed", trace.SlogError(err))
func SlogError(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	var te *TraceError
	if errors.As(err, &te) {
		return slog.Group("error",
			slog.String("message", te.Message),
			slog.Any("cause", errorCause(te.Err)),
			slog.Any("trace", framesToSlogValue(te.Frames)),
			slog.Any("fields", te.Fields),
		)
	}

	return slog.Group("error",
		slog.String("message", err.Error()),
	)
}

// SlogErrorValue returns a slog.Value for an error
func SlogErrorValue(err error) slog.Value {
	if err == nil {
		return slog.StringValue("")
	}

	var te *TraceError
	if errors.As(err, &te) {
		return te.LogValue()
	}

	return slog.StringValue(err.Error())
}

func errorCause(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}

func framesToSlogValue(frames Frames) []any {
	if len(frames) == 0 {
		return nil
	}

	result := make([]any, len(frames))
	for i, f := range frames {
		result[i] = slog.Group("",
			slog.String("file", f.File),
			slog.Int("line", f.Line),
			slog.String("func", f.Function),
		)
	}
	return result
}

// ErrorHandler wraps an slog.Handler to automatically extract trace information
type ErrorHandler struct {
	slog.Handler
}

// NewErrorHandler creates a new ErrorHandler wrapping the given handler
func NewErrorHandler(h slog.Handler) *ErrorHandler {
	return &ErrorHandler{Handler: h}
}

func (h *ErrorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

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

// WithAttrs returns a new handler with the given attributes
func (h *ErrorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ErrorHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new handler with the given group name
func (h *ErrorHandler) WithGroup(name string) slog.Handler {
	return &ErrorHandler{Handler: h.Handler.WithGroup(name)}
}

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

func LogError(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	logAtLevel(ctx, logger, slog.LevelError, msg, err, attrs...)
}

func LogWarn(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	logAtLevel(ctx, logger, slog.LevelWarn, msg, err, attrs...)
}

func LogDebug(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	logAtLevel(ctx, logger, slog.LevelDebug, msg, err, attrs...)
}
