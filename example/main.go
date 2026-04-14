package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tae2089/trace"
)

func main() {
	// Create a new HTTP handler with trace middleware
	http.Handle("/users/{id}", trace.ErrorMiddleware(getUserHandler))

	// Start the server
	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Server failed", "error", err)
	}
}

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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

	return nil
}

// getUserService retrieves a user by ID
func getUserService(userID string) (*User, error) {
	// Call repository layer
	user, err := repoFindUser(userID)
	if err != nil {
		return nil, trace.Wrapf(err, "service: failed to get user %s", userID)
	}
	return user, nil
}

// repoFindUser simulates database query
func repoFindUser(userID string) (*User, error) {
	// Simulate database error
	err := sql.ErrNoRows
	if err == sql.ErrNoRows {
		return nil, trace.WrapNotFound(err, fmt.Sprintf("user %s not found in database", userID))
	}
	return &User{ID: userID}, nil
}

// User represents a user entity
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
