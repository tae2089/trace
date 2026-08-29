// Package main demonstrates Trace v3 core error tracking.
package main

import (
	"errors"
	"fmt"

	"github.com/tae2089/trace/v3"
)

var errUserNotFound = errors.New("user not found")

// @intent demonstrate plain error text, standard inspection, and opt-in trace formatting.
func main() {
	if err := loadUser("user-123"); err != nil {
		if errors.Is(err, errUserNotFound) {
			fmt.Println("not found")
		}

		var validation *ValidationError
		if errors.As(err, &validation) {
			fmt.Println(validation.Field)
		}

		fmt.Println(err)
		fmt.Printf("%+v\n", err)
	}
}

// @intent add service-layer context without changing the application error meaning.
func loadUser(userID string) error {
	if err := validateUserID(userID); err != nil {
		return trace.Wrapf(err, "load user %s", userID)
	}
	if err := queryUser(userID); err != nil {
		return trace.Wrapf(err, "load user %s", userID)
	}
	return nil
}

// @intent show an application-owned typed error preserved through Trace wrappers.
func validateUserID(userID string) error {
	if userID == "" {
		return trace.Errorf("validate user id: %w", &ValidationError{Field: "user_id"})
	}
	return nil
}

// @intent create a traced origin that keeps a sentinel error visible to errors.Is.
func queryUser(userID string) error {
	return trace.Errorf("query user %s: %w", userID, errUserNotFound)
}

// @intent represent application-owned validation semantics outside the Trace package.
type ValidationError struct {
	Field string
}

// @intent expose the application error message preserved by Trace.
func (e *ValidationError) Error() string {
	return "validation failed"
}
