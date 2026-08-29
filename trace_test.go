package trace_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/tae2089/trace/v3"
)

var errNotFound = errors.New("not found")

type validationError struct {
	Field string
}

func (e *validationError) Error() string {
	return "validation failed: " + e.Field
}

func TestNew(t *testing.T) {
	err := trace.New("not found")

	assertPlainFormat(t, err, "not found")
	assertDebugContainsInOrder(
		t,
		err,
		"trace_test.go",
		"TestNew",
		"not found",
	)

	if got := errors.Unwrap(err); got != nil {
		t.Fatalf("errors.Unwrap(New(...)) = %v, want nil", got)
	}
}

func TestErrorf(t *testing.T) {
	t.Run("preserves single wrapped cause", func(t *testing.T) {
		err := trace.Errorf("load user: %w", errNotFound)

		assertPlainFormat(t, err, "load user: not found")
		if !errors.Is(err, errNotFound) {
			t.Fatal("errors.Is must find the wrapped sentinel")
		}
		assertDebugContainsInOrder(
			t,
			err,
			"trace_test.go",
			"TestErrorf",
			"load user: not found",
		)
	})

	t.Run("preserves multiple wrapped causes", func(t *testing.T) {
		errUnauthorized := errors.New("unauthorized")

		err := trace.Errorf("load user: %w: %w", errNotFound, errUnauthorized)

		assertPlainFormat(t, err, "load user: not found: unauthorized")
		if !errors.Is(err, errNotFound) {
			t.Fatal("errors.Is must find the first wrapped sentinel")
		}
		if !errors.Is(err, errUnauthorized) {
			t.Fatal("errors.Is must find the second wrapped sentinel")
		}
		assertDebugContainsInOrder(
			t,
			err,
			"branch 1",
			"not found",
			"branch 2",
			"unauthorized",
		)
	})
}

func TestWrap(t *testing.T) {
	cause := &validationError{Field: "email"}

	err := trace.Wrap(cause, "validate user")
	err = trace.Wrap(err, "create user")

	assertPlainFormat(t, err, "create user: validate user: validation failed: email")
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is must find the original cause")
	}

	var got *validationError
	if !errors.As(err, &got) {
		t.Fatal("errors.As must find the typed cause")
	}
	if got.Field != "email" {
		t.Fatalf("typed cause field = %q, want email", got.Field)
	}
	if got := errors.Unwrap(err); got == nil {
		t.Fatal("errors.Unwrap(Wrap(...)) must expose the wrapped cause")
	}
	assertDebugContainsInOrder(
		t,
		err,
		"create user",
		"trace_test.go",
		"TestWrap",
		"validate user",
		"validation failed: email",
	)
}

func TestWrapf(t *testing.T) {
	err := trace.Wrapf(errNotFound, "load user %d", 42)

	assertPlainFormat(t, err, "load user 42: not found")
	if !errors.Is(err, errNotFound) {
		t.Fatal("errors.Is must find the wrapped cause")
	}
	assertDebugContainsInOrder(t, err, "load user 42", "not found")
	assertDebugContainsInOrder(t, err, "trace_test.go", "TestWrapf")
}

func TestWrapNilAndEmptyContext(t *testing.T) {
	if got := trace.Wrap(nil, "ignored"); got != nil {
		t.Fatalf("Wrap(nil, ...) = %v, want nil", got)
	}
	if got := trace.Wrapf(nil, "ignored %s", "value"); got != nil {
		t.Fatalf("Wrapf(nil, ...) = %v, want nil", got)
	}
	if got := trace.Wrap(errNotFound, ""); got != errNotFound {
		t.Fatal("Wrap with empty context must return the original error")
	}
	if got := trace.Wrapf(errNotFound, ""); got != errNotFound {
		t.Fatal("Wrapf with empty format must return the original error")
	}
}

func TestJoinedErrorDebugOutput(t *testing.T) {
	validation := trace.Wrap(&validationError{Field: "name"}, "validate user")
	persistence := trace.Wrap(errNotFound, "save user")

	err := trace.Wrap(errors.Join(validation, persistence), "create user")

	var got *validationError
	if !errors.As(err, &got) {
		t.Fatal("errors.As must find a typed error inside a joined branch")
	}
	if !errors.Is(err, errNotFound) {
		t.Fatal("errors.Is must find a sentinel inside a joined branch")
	}
	assertDebugContainsInOrder(
		t,
		err,
		"create user",
		"branch 1",
		"validate user",
		"validation failed: name",
		"branch 2",
		"save user",
		"not found",
	)
}

func TestDebugOutputIsDeterministic(t *testing.T) {
	err := trace.Wrap(trace.Errorf("query user: %w", errNotFound), "load user")

	first := fmt.Sprintf("%+v", err)
	second := fmt.Sprintf("%+v", err)

	if first != second {
		t.Fatalf("debug output must be deterministic\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestDebugOutputHandlesCycles(t *testing.T) {
	err := trace.Wrap(cyclicError{}, "cycle")

	debug := fmt.Sprintf("%+v", err)

	for _, want := range []string{"cycle", "cycle detected"} {
		if !strings.Contains(debug, want) {
			t.Fatalf("debug output %q does not contain %q", debug, want)
		}
	}
}

func TestDebugOutputHandlesNonComparableCycles(t *testing.T) {
	err := trace.Wrap(nonComparableCyclicError{}, "cycle")

	debug := fmt.Sprintf("%+v", err)

	for _, want := range []string{"cycle", "depth limit reached"} {
		if !strings.Contains(debug, want) {
			t.Fatalf("debug output %q does not contain %q", debug, want)
		}
	}
}

func TestDebugOutputDoesNotInventNonComparableCycles(t *testing.T) {
	second := nonComparableChainError{values: []int{2}}
	first := nonComparableChainError{values: []int{1}, next: second}

	err := trace.Wrap(first, "chain")

	debug := fmt.Sprintf("%+v", err)
	if strings.Contains(debug, "cycle detected") {
		t.Fatalf("debug output invented a cycle for a finite chain:\n%s", debug)
	}
	if strings.Contains(debug, "depth limit reached") {
		t.Fatalf("debug output hit the fallback limit for a finite two-node chain:\n%s", debug)
	}
	if got := strings.Count(debug, "error: repeated non-comparable"); got != 2 {
		t.Fatalf("debug output rendered %d non-comparable nodes, want 2:\n%s", got, debug)
	}
}

type cyclicError struct{}

func (cyclicError) Error() string {
	return "cyclic"
}

func (e cyclicError) Unwrap() error {
	return e
}

type nonComparableCyclicError struct {
	values []int
}

func (nonComparableCyclicError) Error() string {
	return "non-comparable cyclic"
}

func (e nonComparableCyclicError) Unwrap() error {
	return e
}

type nonComparableChainError struct {
	values []int
	next   error
}

func (nonComparableChainError) Error() string {
	return "repeated non-comparable"
}

func (e nonComparableChainError) Unwrap() error {
	return e.next
}

func TestConcurrentFormattingAndTraversal(t *testing.T) {
	err := trace.Wrap(trace.Errorf("query user: %w", errNotFound), "load user")

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				if !errors.Is(err, errNotFound) {
					t.Error("errors.Is lost the original cause")
				}
				_ = err.Error()
				_ = fmt.Sprintf("%+v", err)
			}
		}()
	}
	wg.Wait()
}

func TestPublicAPI(t *testing.T) {
	expected := []string{"Errorf", "New", "Wrap", "Wrapf"}
	got := exportedNames(t)

	if !slices.Equal(got, expected) {
		t.Fatalf("exported API = %v, want %v", got, expected)
	}
}

func exportedNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	names := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil && decl.Name.IsExported() {
					names = append(names, decl.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if spec.Name.IsExported() {
							names = append(names, spec.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if name.IsExported() {
								names = append(names, name.Name)
							}
						}
					}
				}
			}
		}
	}

	slices.Sort(names)
	return names
}

func assertPlainFormat(t *testing.T, err error, want string) {
	t.Helper()

	formats := map[string]string{
		"Error": err.Error(),
		"%s":    fmt.Sprintf("%s", err),
		"%v":    fmt.Sprintf("%v", err),
	}
	for name, got := range formats {
		if got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
		if strings.Contains(got, ".go") || strings.Contains(got, "Test") {
			t.Fatalf("%s leaked debug location in %q", name, got)
		}
	}
}

func assertDebugContainsInOrder(t *testing.T, err error, wants ...string) {
	t.Helper()

	debug := fmt.Sprintf("%+v", err)
	offset := 0
	for _, want := range wants {
		index := strings.Index(debug[offset:], want)
		if index < 0 {
			t.Fatalf("debug output missing %q after offset %d:\n%s", want, offset, debug)
		}
		offset += index + len(want)
	}
}
