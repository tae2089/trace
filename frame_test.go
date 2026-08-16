package trace_test

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tae2089/trace/v2"
)

func TestCaptureFrameResolvesCallSite(t *testing.T) {
	_, wantFile, wantLine, _ := runtime.Caller(0)
	frame := trace.CaptureFrame(1) // must point at this line

	if got, want := frame.File(), filepath.Base(wantFile); got != want {
		t.Fatalf("File() = %q, want %q", got, want)
	}
	if got := frame.Line(); got != wantLine+1 {
		t.Fatalf("Line() = %d, want %d", got, wantLine+1)
	}
	if got := frame.Function(); !strings.Contains(got, "TestCaptureFrameResolvesCallSite") {
		t.Fatalf("Function() = %q, want it to contain the test name", got)
	}
}

func TestCaptureFrameSkipPointsAtCallerOfCaller(t *testing.T) {
	var frame trace.Frame
	capture := func() {
		frame = trace.CaptureFrame(2) // 2 = the caller of this closure
	}
	capture()

	if got := frame.Function(); !strings.Contains(got, "TestCaptureFrameSkipPointsAtCallerOfCaller") {
		t.Fatalf("Function() = %q, want the enclosing test, not the closure", got)
	}
}

func TestFrameStringFormat(t *testing.T) {
	frame := trace.CaptureFrame(1)
	s := frame.String()
	if !strings.Contains(s, "frame_test.go:") {
		t.Fatalf("String() = %q, want file:line format", s)
	}
	if !strings.Contains(s, "TestFrameStringFormat") {
		t.Fatalf("String() = %q, want the function name", s)
	}
}

func TestFrameMarshalJSONKeepsWireShape(t *testing.T) {
	frame := trace.CaptureFrame(1)
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Function string `json:"function"`
		File     string `json:"file"`
		Line     int    `json:"line"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.File != "frame_test.go" || decoded.Line == 0 || !strings.Contains(decoded.Function, "TestFrameMarshalJSONKeepsWireShape") {
		t.Fatalf("unexpected JSON payload: %s", data)
	}
}

func TestZeroFrameRendersWithoutPanic(t *testing.T) {
	var frame trace.Frame
	if got := frame.Line(); got != 0 {
		t.Fatalf("zero frame Line() = %d, want 0", got)
	}
	_ = frame.String()
	_ = frame.File()
	_ = frame.Function()
}
