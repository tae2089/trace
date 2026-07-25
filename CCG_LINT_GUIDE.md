# CCG Lint & Annotation Guide for Go Code

## Executive Summary

Based on official CCG documentation and analysis of the Trace project:

- **`ccg lint` flags 8 categories**, with **`unannotated` (59) and `incomplete` (20)** being the primary findings in Trace
- **Go-specific symbols worth annotating**: interfaces, public functions, error types, service methods, handlers
- **Symbols safe to skip**: trivial getters/setters, error method implementations, internal helpers, test utilities
- **Practical strategy**: Prioritize high-value symbols first; use regex rules in `.ccg.yaml` to suppress low-value patterns

---

## What CCG Lint Detects (8 Categories)

### File-Level Categories
1. **`orphan`** — Doc file exists but source file was deleted
2. **`missing`** — Source file exists but no doc generated
3. **`stale`** — Doc file older than source file modification time

### Symbol-Level Categories
4. **`unannotated`** — Function/class/type has NO annotation at all (no `@intent`, `@domainRule`, etc.)
5. **`incomplete`** — Annotation exists but missing `@intent` tag (most important tag)
6. **`contradiction`** — Annotation has `@param` tags but code signature changed after annotation
7. **`dead-ref`** — `@see` tag points to non-existent symbol
8. **`drifted`** — Annotation exists but code changed after annotation was written

**Key insight**: `unannotated` ≠ `incomplete`. A symbol is:
- **`unannotated`** if it has zero annotations
- **`incomplete`** if it has some annotation but is missing `@intent`

---

## Trace Project Current State

```
Unannotated symbols (59):
  - Error type methods (Error, HTTPStatusCode, IsXxx, Unwrap)
  - Internal helpers (attrsToAny, cloneTraceError, contextFieldsToMap, etc.)
  - main.main
  - Wrapper functions (wrapInternal, replaceTraceError, etc.)

Incomplete annotations (20):
  - Error constructor functions (ErrorNotFound, ErrorAccessDenied, etc.)
  - Public functions without @intent (IsNotFound, NotFound, captureFrame, etc.)
  - main.User, main.getUserHandler, main.getUserService, main.repoFindUser
```

---

## Go Symbols: What to Annotate vs. Skip

### ✅ WORTH ANNOTATING (High-Value)

#### 1. **Interfaces** (especially behavior-defining ones)
```go
// @intent let callers recognize missing-resource semantics through behavior instead of concrete error types.
type ErrorNotFound interface {
    error
    IsNotFound() bool
}
```
**Why**: Interfaces define contracts. Callers need to understand what behavior they're checking for.

#### 2. **Public Functions & Methods** (especially handlers, service methods, CLI commands)
```go
// @intent verify user identity before granting system access
// @domainRule lock account after 5 consecutive failed attempts
// @sideEffect writes login attempt to audit_log table
// @mutates user.FailedAttempts, user.LockedUntil
func AuthenticateUser(username, password string) (string, error)
```
**Why**: These are entry points. Developers need to understand purpose, side effects, and business rules.

#### 3. **Error Constructor Functions**
```go
// @intent create a typed error indicating a resource was not found
// @domainRule used when database query returns zero rows
func NotFound(format string, args ...any) error
```
**Why**: Error constructors are semantic markers. Annotating them helps search find error-creation patterns.

#### 4. **Package-Level** (`@index`)
```go
// @index Typed error categories, retryability rules, and HTTP status mapping for trace errors.
package trace
```
**Why**: Package-level annotations help RAG and search understand module purpose at a glance.

#### 5. **Complex Helpers with Side Effects**
```go
// @intent capture the current call stack for error attribution
// @sideEffect calls runtime.Callers (expensive)
// @ensures returns frames in call order from innermost to outermost
func captureFrame(skip int) Frames
```
**Why**: Side effects and performance implications matter to callers.

---

### ❌ SAFE TO SKIP (Low-Value)

#### 1. **Error Method Implementations** (Error, HTTPStatusCode, IsXxx, Unwrap)
```go
// ❌ Skip these:
func (e *NotFoundError) Error() string { ... }
func (e *NotFoundError) HTTPStatusCode() int { ... }
func (e *NotFoundError) IsNotFound() bool { ... }
func (e *NotFoundError) Unwrap() error { ... }
```
**Why**:
- These are mechanical implementations of interfaces
- Their behavior is obvious from the method name
- Annotating them adds noise without search value
- CCG lint will flag them as `unannotated`, but that's acceptable

#### 2. **Trivial Getters/Setters**
```go
// ❌ Skip:
func (f Frame) String() string { return fmt.Sprintf(...) }
```
**Why**: One-liners with obvious behavior don't need intent documentation.

#### 3. **Internal Helpers** (unexported functions)
```go
// ❌ Skip:
func attrsToAny(attrs []slog.Attr) map[string]any { ... }
func contextFieldsToMap(ctx context.Context) map[string]any { ... }
func copyFields(src map[string]any) map[string]any { ... }
```
**Why**:
- Internal functions are not part of the public API
- Callers can't use them anyway
- Annotating them doesn't improve search quality
- Exception: if an internal function is called from many places or has complex logic, consider annotating

#### 4. **Wrapper Functions** (thin pass-throughs)
```go
// ❌ Skip:
func wrapInternal(err error, msg string) error { ... }
func replaceTraceError(err error, ...) error { ... }
```
**Why**: These are implementation details. Annotate the public API that calls them instead.

#### 5. **Test Utilities & Mocks**
```go
// ❌ Skip:
type mockSearchBackend struct { ... }
func (m *mockSearchBackend) Query(...) { ... }
```
**Why**: Tests are excluded from `unannotated` checks anyway (CCG lint skips tests).

#### 6. **main.main**
```go
// ❌ Skip:
func main() { ... }
```
**Why**: `main` is a special case. Its purpose is obvious (entry point). Annotating it adds no value.

---

## Practical Annotation Strategy for Trace

### Phase 1: High-Value Symbols (Do First)
1. **Error constructors** (ErrorNotFound, ErrorAccessDenied, etc.)
   - Add `@intent` explaining when to use each error type
   - Add `@domainRule` for any business rules (e.g., "lock account after 5 failures")

2. **Public functions** (IsNotFound, NotFound, captureFrame, formatMessage, etc.)
   - Add `@intent` explaining purpose
   - Add `@sideEffect` if applicable (e.g., captureFrame calls runtime.Callers)

3. **Package-level** (`@index`)
   - Already done in errors.go and trace.go ✓

### Phase 2: Medium-Value Symbols (Do Second)
1. **Interfaces** (ErrorNotFound, ErrorAlreadyExists, etc.)
   - Already annotated ✓

2. **Complex internal helpers** (if called from many places)
   - Example: `framesToSerializable`, `contextFieldsToMap`
   - Add `@intent` if they're called from 3+ places

### Phase 3: Skip (Suppress in .ccg.yaml)
1. **Error method implementations** (Error, HTTPStatusCode, IsXxx, Unwrap)
   - Add regex rule to suppress `unannotated` for these patterns

2. **Trivial helpers** (String, copyFields, etc.)
   - Add regex rule to suppress

---

## .ccg.yaml Configuration for Trace

```yaml
rules:
  # Suppress unannotated for error method implementations
  - pattern: "trace\\..*Error\\.(Error|HTTPStatusCode|Is[A-Z]|Unwrap|LogValue)"
    category: unannotated
    action: ignore

  # Suppress unannotated for internal helpers
  - pattern: "trace\\.(attrsToAny|cloneTraceError|contextFieldsToMap|copyFields|debugReportWalk|errorCause|errorsWalk|framesToSerializable|httpStatusError|replaceTraceError|wrapInternal)"
    category: unannotated
    action: ignore

  # Suppress unannotated for main.main
  - pattern: "main\\.main"
    category: unannotated
    action: ignore

  # Suppress incomplete for error constructors (they're self-documenting)
  - pattern: "trace\\.(ErrorNotFound|ErrorAlreadyExists|ErrorBadParameter|ErrorNotImplemented|ErrorAccessDenied|ErrorConflict|ErrorConnectionProblem|ErrorLimitExceeded|ErrorTimeout|ErrorRetryable)"
    category: incomplete
    action: warn  # warn instead of ignore to keep visibility
```

---

## Lint Interpretation for Trace

### Current Report
```
Summary: 0 orphan, 0 missing, 0 stale, 59 unannotated, 0 contradiction, 0 dead-ref, 20 incomplete, 0 drifted
```

**Interpretation**:
- ✓ No file-level issues (orphan/missing/stale)
- ✓ No contradiction or dead-ref issues
- ⚠️ 59 unannotated: mostly error methods and internal helpers (acceptable to suppress)
- ⚠️ 20 incomplete: error constructors and public functions (should add `@intent`)

### After Applying Suppressions
Expected result:
```
Summary: 0 orphan, 0 missing, 0 stale, ~10 unannotated, 0 contradiction, 0 dead-ref, 20 incomplete, 0 drifted
```

The remaining 10 `unannotated` would be:
- `main.main`
- `main.User`, `main.getUserHandler`, `main.getUserService`, `main.repoFindUser`
- A few other edge cases

These are acceptable because they're either entry points or test/example code.

---

## Go-Specific Annotation Patterns

### Error Types
```go
// ✅ Good: Explains the semantic meaning
// @intent let callers recognize missing-resource semantics through behavior instead of concrete error types.
type ErrorNotFound interface {
    error
    IsNotFound() bool
}

// ✅ Good: Explains when to use
// @intent create a typed error indicating a resource was not found
// @domainRule used when database query returns zero rows
func NotFound(format string, args ...any) error
```

### Interfaces
```go
// ✅ Good: Explains the contract
// @intent define the behavior that all trace-aware errors must implement
type TraceError interface {
    Error() string
    Unwrap() error
    GetFields() map[string]any
}
```

### Public Functions
```go
// ✅ Good: Explains purpose and side effects
// @intent wrap an error with additional context and capture the current call stack
// @sideEffect calls runtime.Callers (expensive, ~1-2µs per call)
// @ensures returns a TraceError with frames in call order
func Wrap(err error, msg string, args ...any) error
```

### Handlers/Service Methods
```go
// ✅ Good: Explains business logic
// @intent validate user credentials and create an authenticated session
// @domainRule lock account after 5 consecutive failed attempts
// @sideEffect writes login attempt to audit_log table
// @mutates user.FailedAttempts, user.LockedUntil
// @requires user.IsActive == true
// @ensures err == nil implies valid JWT with 24h expiry
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error)
```

---

## Lint Workflow for Trace

### Step 1: Run Lint
```bash
ccg lint --namespace trace
```

### Step 2: Review Findings
- Identify which `unannotated` symbols are truly low-value (error methods, internal helpers)
- Identify which `incomplete` symbols need `@intent` added

### Step 3: Add Suppressions to .ccg.yaml
- Use regex patterns to suppress low-value symbols
- Keep `warn` action for symbols you want visibility on but don't want to fail CI

### Step 4: Add Annotations to High-Value Symbols
- Error constructors: add `@intent` + `@domainRule` if applicable
- Public functions: add `@intent` + `@sideEffect` if applicable
- Handlers/service methods: add full annotation block

### Step 5: Rerun Lint
```bash
ccg lint --namespace trace --strict
```

Expected result: 0 errors (or only warnings for intentionally suppressed symbols)

---

## Key Takeaways

| Symbol Type | Annotate? | Why |
|---|---|---|
| Interface | ✅ Yes | Defines contract; callers need to understand behavior |
| Public function | ✅ Yes | Entry point; developers need to understand purpose |
| Error constructor | ✅ Yes | Semantic marker; helps search find error patterns |
| Package-level | ✅ Yes | Helps RAG understand module purpose |
| Error method (Error, IsXxx, Unwrap) | ❌ No | Mechanical implementation; obvious from name |
| Trivial getter/setter | ❌ No | One-liner; obvious behavior |
| Internal helper | ❌ No | Not part of public API; suppress in .ccg.yaml |
| Wrapper function | ❌ No | Implementation detail; annotate the public API instead |
| Test utility | ❌ No | Tests are excluded from lint anyway |

**Bottom line**: Annotate the **public API surface** and **business logic**. Skip the **mechanical implementations** and **internal plumbing**.
