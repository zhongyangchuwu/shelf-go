# Testing Patterns

**Analysis Date:** 2026-06-16

## Test Framework

**Runner:**
- Go standard `testing` package from the toolchain declared in `go.mod`.
- Config: Not detected. There is no `go test` config file; package discovery uses standard Go conventions.

**Assertion Library:**
- Standard library only. Tests use `t.Fatalf`, `strings.Contains`, `reflect.DeepEqual`, and direct value comparisons.

**Run Commands:**
```bash
go test ./...              # Run all tests
go test ./internal/cli     # Run CLI package tests
go test -cover ./...       # Run all tests with coverage summary
```

## Test File Organization

**Location:**
- Tests are co-located with packages under test: `internal/cli/*_test.go`, `internal/store/io_test.go`, `internal/manifest/manifest_test.go`.
- Packages without tests currently include `cmd/shelf`, `internal/config`, `internal/render`, and `internal/version`.

**Naming:**
- Use `*_test.go` file names and `Test<Behavior>` functions.
- Shared test helpers live in `internal/cli/test_helpers_test.go`.

**Structure:**
```text
internal/
├── cli/
│   ├── *_test.go              # CLI command, completion, git, editor, concurrency tests
│   └── test_helpers_test.go   # runShelf, runShelfWithInput, runGit, prompt helpers
├── manifest/
│   └── manifest_test.go       # manifest round-trip and validation tests
└── store/
    └── io_test.go             # store JSON load and strict decoding tests
```

## Test Structure

**Suite Organization:**
```go
func TestValidateRejectsInvalidManifestRules(t *testing.T) {
	tests := []struct {
		name    string
		in      Manifest
		wantErr string
	}{ /* cases */ }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
```

**Patterns:**
- Use table-driven tests for validation matrices, as in `internal/manifest/manifest_test.go` and `internal/store/io_test.go`.
- Use `t.TempDir()` for isolated filesystem state and `t.Chdir()` for command behavior that depends on the current repository root, as in `internal/cli/run_test.go` and `internal/cli/project_test.go`.
- Use `t.Setenv()` for environment-dependent behavior, including `EDITOR`, `FPATH`, and parent environment override tests in `internal/cli/secret_test.go`, `internal/cli/doctor_test.go`, and `internal/cli/run_test.go`.
- Assert exact command output when output is stable, such as `export APP_TOKEN=edited\n` in `internal/cli/secret_test.go`.
- Assert substring containment for diagnostics and JSON snippets where ordering or surrounding output is less central, as in `internal/cli/run_test.go` and `internal/manifest/manifest_test.go`.

## Mocking

**Framework:** Standard library seams and helper functions. No gomock, testify, or other mocking package is used.

**Patterns:**
```go
func withPromptPassword(t *testing.T, password string) {
	t.Helper()
	origIsTerminal := secretAddIsTerminal
	origReadPassword := secretAddReadPassword
	secretAddIsTerminal = func(int) bool { return true }
	secretAddReadPassword = func(int) ([]byte, error) { return []byte(password), nil }
	t.Cleanup(func() {
		secretAddIsTerminal = origIsTerminal
		secretAddReadPassword = origReadPassword
	})
}
```

**What to Mock:**
- Mock terminal and password behavior through package-level function variables in `internal/cli/secret.go` and restore with `t.Cleanup()` from `internal/cli/test_helpers_test.go`.
- Mock editor behavior by writing executable shell scripts into `t.TempDir()` and setting `EDITOR`, as in `internal/cli/secret_test.go`.
- Use real Cobra commands via `NewRootCmd()` rather than mocking command handlers; helper functions in `internal/cli/test_helpers_test.go` capture stdout/stderr with `bytes.Buffer`.

**What NOT to Mock:**
- Do not mock filesystem reads/writes for store, manifest, or CLI command tests. Use `t.TempDir()` and real files.
- Do not mock git when behavior depends on git root detection. Tests call real `git init` through `runGit` in `internal/cli/test_helpers_test.go`.
- Do not mock child process execution for `shelf run`; tests execute `sh -c` commands in `internal/cli/run_test.go`.

## Fixtures and Factories

**Test Data:**
```go
manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
	t.Fatalf("write manifest: %v", err)
}
```

**Location:**
- Fixtures are inline inside tests. There is no dedicated `testdata/` directory.
- Reusable CLI setup lives in `internal/cli/test_helpers_test.go`, especially `setupProjectTest`, `runShelf`, `runShelfWithInput`, and `runGit`.
- Small test-only helpers can live at the bottom of a test file, such as `boolPtr` in `internal/manifest/manifest_test.go`.

## Coverage

**Requirements:** None enforced in repository configuration.

**View Coverage:**
```bash
go test -cover ./...
```

## Test Types

**Unit Tests:**
- Cover validation and pure helper behavior with table tests and direct function calls. Examples: `manifest.Validate` in `internal/manifest/manifest_test.go`, `store.Load` strict JSON behavior in `internal/store/io_test.go`, `childEnv` in `internal/cli/run_test.go`, and completion helpers in `internal/cli/secret_test.go`.

**Integration Tests:**
- CLI tests exercise Cobra command execution, persisted JSON files, git repositories, editor scripts, shell child processes, file locking, and environment injection. Examples: `internal/cli/run_test.go`, `internal/cli/project_test.go`, and `internal/cli/secret_test.go`.

**E2E Tests:**
- No separate E2E framework is used. The closest E2E coverage is package-level CLI execution through `runShelf` in `internal/cli/test_helpers_test.go`.

## Common Patterns

**Async Testing:**
```go
const count = 20
var wg sync.WaitGroup
errs := make(chan error, count)
for i := 0; i < count; i++ {
	i := i
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := runShelf(t, "--data", data, "secret", "set", fmt.Sprintf("app:key_%02d", i), fmt.Sprintf("value-%02d", i))
		errs <- err
	}()
}
```

**Error Testing:**
```go
_, err := runShelf(t, "--data", data, "secret", "rm", "app:token")
if err == nil {
	t.Fatalf("expected rm on missing to fail")
}
```

**Secret-Safety Testing:**
```go
if strings.Contains(out, "secret") || strings.Contains(out, "parent") {
	t.Fatalf("dry-run leaked env value:\n%s", out)
}
```

---

*Testing analysis: 2026-06-16*
