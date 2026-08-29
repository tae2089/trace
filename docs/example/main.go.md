<!-- generated-by: code-context-graph docs -->
# example/main.go

## Functions

### main
- **Lines:** 14–28
- **Intent:** demonstrate plain error text, standard inspection, and opt-in trace formatting.
- **Calls:** loadUser

### loadUser
- **Lines:** 31–39
- **Intent:** add service-layer context without changing the application error meaning.
- **Calls:** validateUserID, queryUser, Wrapf, Wrapf

### validateUserID
- **Lines:** 42–47
- **Intent:** show an application-owned typed error preserved through Trace wrappers.
- **Calls:** Errorf

### queryUser
- **Lines:** 50–52
- **Intent:** create a traced origin that keeps a sentinel error visible to errors.Is.
- **Calls:** Errorf

### Error
- **Lines:** 60–62
- **Intent:** expose the application error message preserved by Trace.

## Classes

### ValidationError
- **Lines:** 55–57
- **Intent:** represent application-owned validation semantics outside the Trace package.
