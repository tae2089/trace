<!-- generated-by: code-context-graph docs -->
# example/client.go

> Client-side example for safely restoring typed trace errors from public HTTP error responses.

## Functions

### fetchUser
- **Lines:** 19–59
- **Intent:** demonstrate a bounded HTTP client flow that reconstructs only public typed error semantics.
- **Domain Rules:**
  - error responses are decoded through ReadErrorResponse before successful JSON is unmarshaled.
fetchUser requests one user and restores safe typed errors returned by the server.
- **Calls:** Wrap, Wrap, Wrap, Wrap, Errorf, ReadErrorResponse
