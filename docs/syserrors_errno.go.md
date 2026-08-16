<!-- generated-by: code-context-graph docs -->
# syserrors_errno.go

## Functions

### convertErrno
- **Lines:** 15–31
- **Intent:** map the syscall errno values that carry a clear trace category.
- **Domain Rules:**
  - refused, reset, unreachable, and broken-pipe errnos are transport failures.
  - descriptor-table exhaustion is a limit, not a transport failure.
- **Ensures:**
  - reports false when no errno rule applies, leaving the caller's fallbacks in charge.
- **Calls:** ConnectionProblem, WrapLimitExceeded, Timeout
