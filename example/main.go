// @index Example HTTP application demonstrating application-owned logging, safe HTTP errors, and layered wrapping.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tae2089/trace/v2"
	"github.com/tae2089/trace/v2/tracehttp"
)

// @intent demonstrate how an application owns request logging while trace renders safe error responses.
func main() {
	http.Handle("/users/{id}", handle(getUserHandler))

	// Start the server
	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Server failed", "error", err)
	}
}

// @intent keep request logging policy in the application and delegate only safe response rendering to trace.
func handle(handler tracehttp.ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			httpError := tracehttp.ToHTTPError(err)
			slog.Error("Request failed",
				trace.SlogError(err),
				slog.Int("status_code", httpError.Status),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			requestID := r.Header.Get("X-Request-ID")
			if writeErr := tracehttp.WriteError(w, err, requestID); writeErr != nil {
				slog.Error("Failed to write error response", "error", writeErr)
			}
		}
	}
}

// @intent show how handlers return errors to the application-owned HTTP adapter.
// getUserHandler handles GET /users/{id} requests
func getUserHandler(w http.ResponseWriter, r *http.Request) error {
	// Extract user ID from URL
	userID := r.PathValue("id")

	// Call service layer
	user, err := getUserService(userID)
	if err != nil {
		return err
	}

	// Return user data
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		return trace.Wrap(err, "encode user response")
	}
	return nil
}

// @intent demonstrate service-layer wrapping that adds user-specific context before errors cross boundaries.
// getUserService retrieves a user by ID
func getUserService(userID string) (*User, error) {
	// Call repository layer
	user, err := repoFindUser(userID)
	if err != nil {
		return nil, trace.Wrapf(err, "service: failed to get user %s", userID)
	}
	return user, nil
}

// @intent demonstrate repository-layer translation from storage failures into typed trace errors.
// repoFindUser simulates database query
func repoFindUser(userID string) (*User, error) {
	// Simulate database error
	err := sql.ErrNoRows
	if err == sql.ErrNoRows {
		return nil, trace.WrapNotFound(err, fmt.Sprintf("user %s not found in database", userID))
	}
	return &User{ID: userID}, nil
}

// @intent provide a minimal response model for the trace package usage example.
// User represents a user entity
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
