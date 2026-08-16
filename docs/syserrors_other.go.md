<!-- generated-by: code-context-graph docs -->
# syserrors_other.go

## Functions

### convertErrno
- **Lines:** 8–10
- **Intent:** keep ConvertSystemError compiling on platforms that do not define the errno constants.
- **Ensures:**
  - always reports false so the caller falls back to its portable rules.
