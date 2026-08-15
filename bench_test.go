package trace_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tae2089/trace/v2"
)

// Benchmarks for the error-construction hot path. Errors are created far more
// often than they are rendered, so construction cost dominates in practice.

var (
	errSink   error
	boolSink  bool
	strSink   string
	frameSink trace.Frame
)

func BenchmarkFmtErrorfWrap(b *testing.B) {
	base := errors.New("base")
	b.ReportAllocs()
	for b.Loop() {
		errSink = fmt.Errorf("layer: %w", base)
	}
}

func BenchmarkWrap(b *testing.B) {
	base := errors.New("base")
	b.ReportAllocs()
	for b.Loop() {
		errSink = trace.Wrap(base, "layer")
	}
}

func BenchmarkWrapDepth10(b *testing.B) {
	base := errors.New("base")
	b.ReportAllocs()
	for b.Loop() {
		err := error(base)
		for range 10 {
			err = trace.Wrap(err, "layer")
		}
		errSink = err
	}
}

func BenchmarkNotFound(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		errSink = trace.NotFound("user not found")
	}
}

func BenchmarkWrapNotFound(b *testing.B) {
	base := errors.New("no rows")
	b.ReportAllocs()
	for b.Loop() {
		errSink = trace.WrapNotFound(base, "user not found")
	}
}

func BenchmarkCaptureFrame(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		frameSink = trace.CaptureFrame(1)
	}
}

func benchDeepError(n int) error {
	err := error(trace.NotFound("missing"))
	for range n {
		err = trace.Wrap(err, "layer")
	}
	return err
}

func BenchmarkIsNotFoundDepth10(b *testing.B) {
	err := benchDeepError(10)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		boolSink = trace.IsNotFound(err)
	}
}

func BenchmarkErrorStringDepth5(b *testing.B) {
	err := benchDeepError(5)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		strSink = err.Error()
	}
}

func BenchmarkFormatVerboseDepth5(b *testing.B) {
	err := benchDeepError(5)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		strSink = fmt.Sprintf("%+v", err)
	}
}
