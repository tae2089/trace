package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestLoadUserPreservesSentinel(t *testing.T) {
	err := loadUser("user-123")
	if err == nil {
		t.Fatal("loadUser() error = nil")
	}
	if !errors.Is(err, errUserNotFound) {
		t.Fatal("errors.Is must find errUserNotFound")
	}
	if got, want := err.Error(), "load user user-123: query user user-123: user not found"; got != want {
		t.Fatalf("loadUser() error = %q, want %q", got, want)
	}
}

func TestLoadUserPreservesTypedError(t *testing.T) {
	err := loadUser("")
	if err == nil {
		t.Fatal("loadUser() error = nil")
	}

	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("loadUser() error = %T %v, want ValidationError", err, err)
	}
	if validation.Field != "user_id" {
		t.Fatalf("ValidationError.Field = %q, want user_id", validation.Field)
	}
}

func TestLoadUserDebugOutputIncludesTrace(t *testing.T) {
	err := loadUser("user-123")
	debug := fmt.Sprintf("%+v", err)

	for _, want := range []string{
		"load user user-123",
		"query user user-123",
		"example/main.go",
		"loadUser",
		"queryUser",
	} {
		if !strings.Contains(debug, want) {
			t.Fatalf("debug output %q does not contain %q", debug, want)
		}
	}
}
