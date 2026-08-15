<!-- generated-by: code-context-graph docs -->
# example/main.go

> Example HTTP application demonstrating application-owned logging, safe HTTP errors, and layered wrapping.

## Functions

### main
- **Lines:** 16–24
- **Intent:** demonstrate how an application owns request logging while trace renders safe error responses.
- **Calls:** handle

### handle
- **Lines:** 27–44
- **Intent:** keep request logging policy in the application and delegate only safe response rendering to trace.
- **Calls:** SlogError, ToHTTPError, WriteError

### getUserHandler
- **Lines:** 48–64
- **Intent:** show how handlers return errors to the application-owned HTTP adapter.
getUserHandler handles GET /users/{id} requests
- **Calls:** getUserService, Wrap

### getUserService
- **Lines:** 68–75
- **Intent:** demonstrate service-layer wrapping that adds user-specific context before errors cross boundaries.
getUserService retrieves a user by ID
- **Calls:** repoFindUser, Wrapf

### repoFindUser
- **Lines:** 79–86
- **Intent:** demonstrate repository-layer translation from storage failures into typed trace errors.
repoFindUser simulates database query
- **Calls:** WrapNotFound

## Classes

### User
- **Lines:** 90–93
- **Intent:** provide a minimal response model for the trace package usage example.
User represents a user entity
