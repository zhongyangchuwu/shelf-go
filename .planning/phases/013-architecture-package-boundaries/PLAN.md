# Phase 13 Plan: Architecture Package Boundaries

## Goal

Move reusable application and project-resolution behavior out of `internal/cli` while keeping the CLI package compact and command-family oriented.

## Scope

- Add `internal/app` for runtime/vault loading helpers currently owned by `internal/cli/root.go`.
- Add `internal/project` for project manifest resolution, diagnostics, binding conversion, and project identity normalization currently owned by `internal/cli/project.go`.
- Keep `internal/cli` at a small file count; do not split every command into a separate file.
- Do not add `internal/gitutil` in this phase. Project-owned git identity helpers move to `internal/project`; shared git helpers can be extracted later if vault diagnostics also need them.

## Non-goals

- No command behavior changes.
- No new storage backend or backend interface.
- No `internal/vault`, `internal/secret`, or `internal/atomicfile` extraction in this phase.
- No manager UI changes.

## Design

### Layers

```text
Top/display:      cmd/shelf, internal/cli, internal/manager
Feature support:  internal/app, internal/project
Base support:     internal/config, internal/store, internal/manifest, internal/render, internal/version
```

### Dependency rules

- `cmd/shelf -> internal/cli` only.
- `internal/cli -> internal/app`, `internal/project`, and existing base packages.
- `internal/app -> internal/config`, `internal/store`.
- `internal/project -> internal/manifest`, `internal/render`, `internal/store`, standard library git subprocesses.
- `internal/project` must not import `internal/cli`.
- `internal/store`, `internal/manifest`, `internal/render`, `internal/config` must not import `internal/cli` or feature packages.

### Package APIs

`internal/app`:

```go
func LoadVault(configPathFlag, vaultPathFlag string) (config.Runtime, *store.Vault, error)
func LoadRuntime(configPathFlag, vaultPathFlag string) (config.Runtime, *store.Store, error)
func ReadVault(configPathFlag, vaultPathFlag string, fn func(*store.Store) error) error
func UpdateVault(configPathFlag, vaultPathFlag string, fn func(*store.Store) error) error
```

`internal/project`:

```go
type Binding struct { Path, EnvName, Value string }
type Diagnostic struct { Status, Path, Message string }

func ResolveEntries(m manifest.Manifest, st *store.Store) ([]Binding, []Diagnostic)
func HasFailures([]Diagnostic) bool
func RenderDiagnostics(io.Writer, []Diagnostic)
func BindingsForRender([]Binding) []render.Binding

func ID() (string, error)
func IDBestEffort(root string) string
func IDFromRoot(root string) (string, error)
func Root() (string, error)
func NormalizeRemote(remote string) (string, error)
```

## Implementation steps

1. Create `internal/app/runtime.go` and move vault/runtime helpers from CLI.
2. Update CLI commands to pass global flag values into `app` helpers.
3. Create `internal/project/resolve.go` and move resolver/diagnostic/binding logic.
4. Create `internal/project/identity.go` and move project identity/git root/remote normalization logic.
5. Update `internal/cli/project.go` and `internal/cli/run.go` to consume `internal/project`.
6. Run focused project/CLI tests and full `go test ./...`.

## Verification

- `go test ./internal/project ./internal/cli -run 'TestProject|TestRun'`
- `go test ./...`

## Completion criteria

- CLI behavior unchanged.
- `internal/cli/root.go` no longer constructs vault/runtime directly.
- Project binding resolution is reusable outside CLI.
- No dependency cycles.
