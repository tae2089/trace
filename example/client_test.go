package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/tae2089/trace"
)

func TestFetchUserRestoresSafeErrorResponse(t *testing.T) {
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.DiscardHandler))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	mux := http.NewServeMux()
	mux.Handle("/users/{id}", handle(getUserHandler))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	_, err := fetchUser(
		context.Background(),
		server.Client(),
		server.URL+"/users/user-123",
		"01KREQUEST",
	)
	if !trace.IsNotFound(err) {
		t.Fatalf("fetchUser() error = %T %v, want not found", err, err)
	}
	wantHTTPError := trace.HTTPError{
		Status:  http.StatusNotFound,
		Code:    trace.CodeNotFound,
		Message: "user user-123 not found in database",
	}
	if got := trace.ToHTTPError(err); got != wantHTTPError {
		t.Fatalf("ToHTTPError() = %#v, want %#v", got, wantHTTPError)
	}
	wantFields := map[string]any{
		"status_code": http.StatusNotFound,
		"request_id":  "01KREQUEST",
	}
	if got := trace.GetFields(err); !reflect.DeepEqual(got, wantFields) {
		t.Fatalf("GetFields() = %#v, want %#v", got, wantFields)
	}
	if strings.Contains(trace.DebugReport(err), "sql: no rows in result set") {
		t.Fatalf("server cause crossed the HTTP boundary: %s", trace.DebugReport(err))
	}
}

func TestFetchUserDecodesSuccessfulResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"user-123","name":"Ada"}`))
	}))
	t.Cleanup(server.Close)

	user, err := fetchUser(
		context.Background(),
		server.Client(),
		server.URL+"/users/user-123",
		"",
	)
	if err != nil {
		t.Fatalf("fetchUser() error = %v", err)
	}
	want := &User{ID: "user-123", Name: "Ada"}
	if !reflect.DeepEqual(user, want) {
		t.Fatalf("fetchUser() = %#v, want %#v", user, want)
	}
}

func TestFetchUserRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", int(maxUserResponseBytes)+1)))
	}))
	t.Cleanup(server.Close)

	_, err := fetchUser(
		context.Background(),
		server.Client(),
		server.URL+"/users/user-123",
		"",
	)
	if err == nil {
		t.Fatal("fetchUser() error = nil")
	}
	if got := trace.UserMessage(err); got != "user response exceeds 1048576 bytes" {
		t.Fatalf("UserMessage() = %q", got)
	}
}
