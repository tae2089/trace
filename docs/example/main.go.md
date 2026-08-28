<!-- generated-by: code-context-graph docs -->
# example/main.go

> Example HTTP application demonstrating application-owned logging, safe HTTP errors, and layered wrapping.

## Functions

### main
- **Lines:** 17–25
- **Intent:** demonstrate how an application owns request logging while trace renders safe error responses.
- **Calls:** handle

### handle
- **Lines:** 28–45
- **Intent:** keep request logging policy in the application and delegate only safe response rendering to trace.
- **Calls:** SlogError, ToHTTPError, WriteError

### getUserHandler
- **Lines:** 49–65
- **Intent:** show how handlers return errors to the application-owned HTTP adapter.
getUserHandler handles GET /users/{id} requests
- **Calls:** getUserService, Wrap

### getUserService
- **Lines:** 69–76
- **Intent:** demonstrate service-layer wrapping that adds user-specific context before errors cross boundaries.
getUserService retrieves a user by ID
- **Calls:** repoFindUser, Wrapf

### repoFindUser
- **Lines:** 80–87
- **Intent:** demonstrate repository-layer translation from storage failures into typed trace errors.
repoFindUser simulates database query
- **Calls:** WrapNotFound

## Classes

### User
- **Lines:** 91–94
- **Intent:** provide a minimal response model for the trace package usage example.
User represents a user entity
