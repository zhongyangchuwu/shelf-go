# Coding Conventions

**Analysis Date:** 2026-06-16

## Naming Patterns

**Files:**
- Use lowercase package-oriented file names with short nouns or command names: `internal/cli/run.go`, `internal/cli/secret.go`, `internal/store/io.go`, `internal/manifest/validate.go`.
- Put command tests beside command implementations with `_test.go` suffix: `internal/cli/run_test.go`, `internal/cli/secret_test.go`, `internal/manifest/manifest_test.go`.
- Keep the binary entry point under `cmd/<binary>/main.go`: `cmd/shelf/main.go`.

**Functions:**
- Export package entry points that are used across packages with PascalCase: `cli.NewRootCmd` in `internal/cli/root.go`, `store.Load` in `internal/store/io.go`, `manifest.Load` in `internal/manifest/io.go`, `config.Resolve` in `internal/config/config.go`.
- Keep Cobra command constructors package-private and named `new<Name>Cmd`: `newRunCmd` in `internal/cli/run.go`, `newSecretCmd` in `internal/cli/secret.go`, `newProjectCmd` in `internal/cli/project.go`.
- Use package-private helper functions for local behavior: `childEnv`, `envOverrideWarnings`, and `splitEnv` in `internal/cli/run.go`; `copyFile` in `internal/store/io.go`; `expandPath` in `internal/config/config.go`.
- Name test functions `Test<Behavior>` with behavior-oriented wording: `TestRunInjectsProjectSecretsIntoChild` in `internal/cli/run_test.go`, `TestLoadRejectsUnknownFields` in `internal/store/io_test.go`.

**Variables:**
- Use short local variables for idiomatic Go scopes: `cmd`, `err`, `st`, `cfg`, `tt`, `got`, `want`.
- Use descriptive names for user-visible domain values: `configPath`, `dataPath`, `manifestPath`, `envName`, `resolvedEntries`.
- Use `want`/`got` in tests and compare output explicitly, as in `internal/cli/secret_test.go` and `internal/manifest/manifest_test.go`.
- Use package-level variables only for injectable seams needed by tests, such as `secretAddIsTerminal` and `secretAddReadPassword` in `internal/cli/secret.go`.

**Types:**
- Use PascalCase exported data models for cross-package contracts: `store.Store`, `store.Secret`, `manifest.Manifest`, `config.Runtime`.
- Use lowerCamelCase package-private types for local command internals: `exitCoder`, `exitCodeError`, `editableSecret`, `secretAddPrompt`.
- Use struct tags for serialized formats at model boundaries: JSON tags in `internal/store/model.go` and `internal/manifest/manifest.go`, YAML tags in `internal/config/config.go`.

## Code Style

**Formatting:**
- Use standard Go formatting. Run `gofmt` on Go files before finishing changes.
- There is no `.prettierrc`, `biome.json`, `.golangci.yml`, or other formatter config detected.
- Keep imports grouped by `gofmt`: standard library first, blank line, third-party packages, then project packages. Examples: `internal/cli/run.go`, `internal/cli/secret.go`.
- Use octal permission literals with `0o` notation for filesystem writes: `0o600` and `0o700` in `internal/store/io.go`, `internal/manifest/io.go`, and `internal/cli/secret_test.go`.

**Linting:**
- No dedicated lint command or config is detected.
- Use `go test ./...` as the baseline verification command from `justfile`.
- Keep code compatible with the Go version declared in `go.mod`: `go 1.26.4`.

## Import Organization

**Order:**
1. Standard library imports, such as `errors`, `fmt`, `os`, `path/filepath`, `strings`.
2. Third-party imports, such as `github.com/spf13/cobra`, `gopkg.in/yaml.v3`, and `golang.org/x/term`.
3. Project imports under `github.com/zhongyangchuwu/shelf-go/internal/...`.

**Path Aliases:**
- Not applicable for Go. Use full module import paths from `go.mod`, such as `github.com/zhongyangchuwu/shelf-go/internal/store`.

## Error Handling

**Patterns:**
- Return errors instead of logging or exiting inside package code. Cobra commands use `RunE` and return errors in `internal/cli/run.go`, `internal/cli/secret.go`, and `internal/cli/project.go`.
- Wrap contextual parse and validation failures with `fmt.Errorf(...: %w)` when callers need the original error, as in `internal/store/io.go` and `internal/manifest/io.go`.
- Use `errors.Is` for sentinel filesystem cases, especially `os.ErrNotExist`, in `internal/store/io.go`, `internal/config/config.go`, and `internal/cli/run.go`.
- Use `errors.As` for typed exit behavior in `internal/cli/run.go`; `ExitCode(err)` maps command failures to CLI exit codes.
- Validate data before saving or mutating state. `store.Store.Save` and `store.Store.Set` call validation in `internal/store/io.go`; `manifest.Validate` enforces project manifest rules in `internal/manifest/validate.go`.
- Keep user-facing error strings concise and actionable, usually lower-case and without punctuation: examples include `secret not found: %s` in `internal/cli/secret.go` and `.shelf.json not found; run shelf project init` style messages in `internal/cli/project.go`.

## Logging

**Framework:** `console` / Cobra writers

**Patterns:**
- Use `cmd.OutOrStdout()` for normal command output and `cmd.OutOrStderr()` for diagnostics in Cobra commands, as in `internal/cli/run.go` and `internal/cli/secret.go`.
- Do not use a structured logging framework. No `log`, `slog`, or third-party logger pattern is used.
- Avoid printing secrets in diagnostic paths. Tests assert non-leakage in `internal/cli/secret_test.go` and `internal/cli/run_test.go`.

## Comments

**When to Comment:**
- Keep comments sparse. Add comments only where they clarify test setup or command behavior, such as `// Init the project manifest.` in `internal/cli/test_helpers_test.go`.
- Prefer readable function and test names over comments for expected behavior.

**JSDoc/TSDoc:**
- Not applicable. There is no Go doc comment convention enforced for every exported symbol in the current codebase.

## Function Design

**Size:** Use small helpers for reusable behavior and accept larger Cobra `RunE` blocks when the command flow is linear. Extract pure helpers for logic that needs direct tests, such as `childEnv`, `splitEnv`, and completion helpers in `internal/cli/run.go` and `internal/cli/secret.go`.

**Parameters:** Pass explicit primitive and domain values. Avoid global configuration except for test-injected terminal/password hooks in `internal/cli/secret.go`.

**Return Values:** Prefer `(value, error)` or `(value, bool)` depending on whether absence is exceptional. Examples: `store.Load(path) (*Store, error)`, `store.Store.Get(path) (Secret, bool)`, `config.Resolve(...) (Runtime, error)`.

## Module Design

**Exports:** Export only package APIs needed by other packages. Keep command constructors and internal helpers unexported inside `internal/cli`.

**Barrel Files:** Not used. Packages expose symbols directly from their implementation files, such as `internal/store/io.go`, `internal/manifest/io.go`, and `internal/config/config.go`.

---

*Convention analysis: 2026-06-16*
