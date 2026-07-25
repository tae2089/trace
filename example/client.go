// @index Client-side example for safely restoring typed trace errors from public HTTP error responses.
package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/tae2089/trace"
)

const maxUserResponseBytes int64 = 1 << 20

// @intent demonstrate a bounded HTTP client flow that reconstructs only public typed error semantics.
// @domainRule error responses are decoded through ReadErrorResponse before successful JSON is unmarshaled.
// fetchUser requests one user and restores safe typed errors returned by the server.
func fetchUser(
	ctx context.Context,
	client *http.Client,
	userURL string,
	requestID string,
) (*User, error) {
	if client == nil {
		client = http.DefaultClient
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, userURL, nil)
	if err != nil {
		return nil, trace.Wrap(err, "build user request")
	}
	if requestID != "" {
		request.Header.Set("X-Request-ID", requestID)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, trace.Wrap(err, "request user")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxUserResponseBytes+1))
	if err != nil {
		return nil, trace.Wrap(err, "read user response body")
	}
	if int64(len(body)) > maxUserResponseBytes {
		return nil, trace.Errorf("user response exceeds %d bytes", maxUserResponseBytes)
	}
	if err := trace.ReadErrorResponse(response.StatusCode, body); err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, trace.Wrap(err, "decode user response")
	}
	return &user, nil
}
