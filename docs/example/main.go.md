<!-- generated-by: code-context-graph docs -->
# example/main.go

> Example HTTP application demonstrating application-owned logging, safe HTTP errors, and layered wrapping.

## Functions

### main
- **Lines:** 15–23
- **Intent:** demonstrate how an application owns request logging while trace renders safe error responses.
- **Calls:** handle

### handle
- **Lines:** 26–43
- **Intent:** keep request logging policy in the application and delegate only safe response rendering to trace.
- **Calls:** ToHTTPError, WriteError, SlogError

### getUserHandler
- **Lines:** 47–63
- **Intent:** show how handlers return errors to the application-owned HTTP adapter.
getUserHandler handles GET /users/{id} requests
- **Calls:** getUserService

### getUserService
- **Lines:** 67–74
- **Intent:** demonstrate service-layer wrapping that adds user-specific context before errors cross boundaries.
getUserService retrieves a user by ID
- **Calls:** repoFindUser, Wrapf

### repoFindUser
- **Lines:** 78–85
- **Intent:** demonstrate repository-layer translation from storage failures into typed trace errors.
repoFindUser simulates database query
- **Calls:** WrapNotFound

## Classes

### User
- **Lines:** 89–92
- **Intent:** provide a minimal response model for the trace package usage example.
User represents a user entity
