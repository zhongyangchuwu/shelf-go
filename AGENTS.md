
## Project

**Shelf Go**

Shelf Go is a fast local secret manager for solo developers and hackers who want developer-secret workflows without the friction of `.env` sprawl, general-purpose password managers, or hosted tools like Doppler. The current product is a Go CLI that manages secrets, project bindings, direct exports, and `shelf run`; the next direction is making the store a portable git-safe encrypted vault that works well with chezmoi.

Shelf is CLI-first, but it should also provide a local vault manager over localhost for search, viewing, and editing because editing complete secret objects in a terminal editor is awkward. Team sharing is intentionally out of scope for now.

**Core Value:** A single developer can safely carry and use project secrets across machines through a portable encrypted vault, while keeping local env and `shelf run` workflows fast and simple.

### Constraints

- **Security**: Vault data must be encrypted at rest before it is safe to commit or sync - file permissions alone are not sufficient for a secret manager.
- **Encryption**: age is the preferred encryption mechanism because it matches the user's existing chezmoi setup.
- **Portability**: The encrypted vault should be a normal file that can be moved, backed up, or managed by chezmoi.
- **Local-first**: Shelf should not require a hosted backend, account, or daemon for core CLI workflows.
- **CLI-first UX**: Commands must stay scriptable and predictable; the localhost vault manager is additive, not a replacement for CLI use.
- **Non-secret config**: Shelf config and `.shelf.json` project manifests must not contain secret values.
- **Brownfield architecture**: New functionality should keep the current package boundaries: CLI orchestration in `internal/cli`, persistence in `internal/store`, project manifests in `internal/manifest`, rendering in `internal/render`, and config resolution in `internal/config`.



## Technology Stack

## Languages

- Go 1.26.4 - CLI application and all production code in `cmd/shelf/main.go` and `internal/**`.
- JSON - Local secret store format in `internal/store/model.go`, project manifest format in `internal/manifest/manifest.go`, and examples in `docs/data-spec.md`.
- YAML - Optional user config file parsed by `internal/config/config.go`.
- Just - Developer task runner commands in `justfile`.

## Runtime

- Go runtime 1.26.4, declared in `go.mod`.
- Local command-line execution; the binary entry point is `cmd/shelf/main.go`.
- Unix-style filesystem behavior is used for file locks and permissions in `internal/store/lock.go` and `internal/store/io.go`.
- Go modules
- Lockfile: `go.sum` present

## Frameworks

- `github.com/spf13/cobra` v1.9.1 - Command tree, flags, command validation, completions, and execution in `internal/cli/root.go`, `internal/cli/secret.go`, `internal/cli/project.go`, `internal/cli/run.go`, and related command files.
- Go standard `testing` package - Unit and command tests in `internal/cli/*_test.go`, `internal/store/io_test.go`, and `internal/manifest/manifest_test.go`.
- Go toolchain - Build, install, and test commands in `justfile`.
- `just` - Convenience task runner for `build`, `install`, `test`, and `tag` recipes in `justfile`.

## Key Dependencies

- `github.com/spf13/cobra` v1.9.1 - Defines the entire CLI command surface; use it for new commands under `internal/cli/`.
- `gopkg.in/yaml.v3` v3.0.1 - Parses optional runtime config from `~/.config/shelf/config.yaml` in `internal/config/config.go`.
- `golang.org/x/term` v0.44.0 - Reads hidden terminal input for `shelf secret add` in `internal/cli/secret.go`.
- `github.com/spf13/pflag` v1.0.6 - Indirect Cobra flag dependency.
- `github.com/inconshreveable/mousetrap` v1.1.0 - Indirect Cobra dependency.
- `golang.org/x/sys` v0.46.0 - Indirect terminal and syscall support.

## Configuration

- Runtime config resolution is implemented in `internal/config/config.go`.
- Default config path: `~/.config/shelf/config.yaml` from `internal/config/config.go`.
- Default data path: `~/.local/share/shelf/secrets.json` from `internal/config/config.go`.
- `SHELF_CONFIG` overrides the config file path in `internal/config/config.go`.
- `SHELF_DATA` overrides the data file path in `internal/config/config.go`.
- `EDITOR` supplies the default editor when config has no editor value in `internal/config/config.go`.
- `FPATH` / `fpath` are inspected by `shelf doctor` for zsh completion installation in `internal/cli/doctor.go`.
- `go.mod` declares module `github.com/zhongyangchuwu/shelf-go`, Go version `1.26.4`, and all module dependencies.
- `go.sum` pins module checksums.
- `justfile` provides:

## Platform Requirements

- Install Go 1.26.4 or compatible toolchain as declared in `go.mod`.
- Install `just` only if using `justfile`; equivalent `go build`, `go install`, and `go test` commands work without it.
- Git is required for project-aware commands such as `shelf project id`, `shelf project init`, `shelf project explain`, and `shelf run` in `internal/cli/project.go` and `internal/cli/run.go`.
- A terminal is required for hidden interactive password input in `shelf secret add` from `internal/cli/secret.go`.
- A shell and editor are required for `shelf secret edit`, which invokes `sh -c "$SHELF_EDITOR \"$SHELF_EDIT_FILE\""` in `internal/cli/secret.go`.
- Deployment target is a local CLI binary named `shelf`.
- Build output defaults to `./bin/shelf` via `justfile`.
- No server runtime, container runtime, hosted platform, or daemon process is detected.



## Conventions

## Naming Patterns

- Use lowercase package-oriented file names with short nouns or command names: `internal/cli/run.go`, `internal/cli/secret.go`, `internal/store/io.go`, `internal/manifest/validate.go`.
- Put command tests beside command implementations with `_test.go` suffix: `internal/cli/run_test.go`, `internal/cli/secret_test.go`, `internal/manifest/manifest_test.go`.
- Keep the binary entry point under `cmd/<binary>/main.go`: `cmd/shelf/main.go`.
- Export package entry points that are used across packages with PascalCase: `cli.NewRootCmd` in `internal/cli/root.go`, `store.Load` in `internal/store/io.go`, `manifest.Load` in `internal/manifest/io.go`, `config.Resolve` in `internal/config/config.go`.
- Keep Cobra command constructors package-private and named `new<Name>Cmd`: `newRunCmd` in `internal/cli/run.go`, `newSecretCmd` in `internal/cli/secret.go`, `newProjectCmd` in `internal/cli/project.go`.
- Use package-private helper functions for local behavior: `childEnv`, `envOverrideWarnings`, and `splitEnv` in `internal/cli/run.go`; `copyFile` in `internal/store/io.go`; `expandPath` in `internal/config/config.go`.
- Name test functions `Test<Behavior>` with behavior-oriented wording: `TestRunInjectsProjectSecretsIntoChild` in `internal/cli/run_test.go`, `TestLoadRejectsUnknownFields` in `internal/store/io_test.go`.
- Use short local variables for idiomatic Go scopes: `cmd`, `err`, `st`, `cfg`, `tt`, `got`, `want`.
- Use descriptive names for user-visible domain values: `configPath`, `dataPath`, `manifestPath`, `envName`, `resolvedEntries`.
- Use `want`/`got` in tests and compare output explicitly, as in `internal/cli/secret_test.go` and `internal/manifest/manifest_test.go`.
- Use package-level variables only for injectable seams needed by tests, such as `secretAddIsTerminal` and `secretAddReadPassword` in `internal/cli/secret.go`.
- Use PascalCase exported data models for cross-package contracts: `store.Store`, `store.Secret`, `manifest.Manifest`, `config.Runtime`.
- Use lowerCamelCase package-private types for local command internals: `exitCoder`, `exitCodeError`, `editableSecret`, `secretAddPrompt`.
- Use struct tags for serialized formats at model boundaries: JSON tags in `internal/store/model.go` and `internal/manifest/manifest.go`, YAML tags in `internal/config/config.go`.

## Code Style

- Use standard Go formatting. Run `gofmt` on Go files before finishing changes.
- There is no `.prettierrc`, `biome.json`, `.golangci.yml`, or other formatter config detected.
- Keep imports grouped by `gofmt`: standard library first, blank line, third-party packages, then project packages. Examples: `internal/cli/run.go`, `internal/cli/secret.go`.
- Use octal permission literals with `0o` notation for filesystem writes: `0o600` and `0o700` in `internal/store/io.go`, `internal/manifest/io.go`, and `internal/cli/secret_test.go`.
- No dedicated lint command or config is detected.
- Use `go test ./...` as the baseline verification command from `justfile`.
- Keep code compatible with the Go version declared in `go.mod`: `go 1.26.4`.

## Import Organization

- Not applicable for Go. Use full module import paths from `go.mod`, such as `github.com/zhongyangchuwu/shelf-go/internal/store`.

## Error Handling

- Return errors instead of logging or exiting inside package code. Cobra commands use `RunE` and return errors in `internal/cli/run.go`, `internal/cli/secret.go`, and `internal/cli/project.go`.
- Wrap contextual parse and validation failures with `fmt.Errorf(...: %w)` when callers need the original error, as in `internal/store/io.go` and `internal/manifest/io.go`.
- Use `errors.Is` for sentinel filesystem cases, especially `os.ErrNotExist`, in `internal/store/io.go`, `internal/config/config.go`, and `internal/cli/run.go`.
- Use `errors.As` for typed exit behavior in `internal/cli/run.go`; `ExitCode(err)` maps command failures to CLI exit codes.
- Validate data before saving or mutating state. `store.Store.Save` and `store.Store.Set` call validation in `internal/store/io.go`; `manifest.Validate` enforces project manifest rules in `internal/manifest/validate.go`.
- Keep user-facing error strings concise and actionable, usually lower-case and without punctuation: examples include `secret not found: %s` in `internal/cli/secret.go` and `.shelf.json not found; run shelf project init` style messages in `internal/cli/project.go`.

## Logging

- Use `cmd.OutOrStdout()` for normal command output and `cmd.OutOrStderr()` for diagnostics in Cobra commands, as in `internal/cli/run.go` and `internal/cli/secret.go`.
- Do not use a structured logging framework. No `log`, `slog`, or third-party logger pattern is used.
- Avoid printing secrets in diagnostic paths. Tests assert non-leakage in `internal/cli/secret_test.go` and `internal/cli/run_test.go`.

## Comments

- Keep comments sparse. Add comments only where they clarify test setup or command behavior, such as `// Init the project manifest.` in `internal/cli/test_helpers_test.go`.
- Prefer readable function and test names over comments for expected behavior.
- Not applicable. There is no Go doc comment convention enforced for every exported symbol in the current codebase.

## Function Design

## Module Design



## Architecture

## System Overview

```text

```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Executable entry point | Construct and execute the root Cobra command, print top-level errors, map errors to process exit codes | `cmd/shelf/main.go` |
| Root command | Define global flags, version, command registration, and shared runtime loaders | `internal/cli/root.go` |
| Secret commands | CRUD and interactive secret editing/add flows over the local store | `internal/cli/secret.go` |
| Export command | Export a path or prefix directly from the secret store in shell/env/JSON formats | `internal/cli/export.go` |
| Project commands | Manage `.shelf.json`, resolve project bindings, export project env, derive Git project identity | `internal/cli/project.go` |
| Run command | Resolve project env bindings and execute a child process with injected environment | `internal/cli/run.go` |
| Init command | Create config and data files with the configured paths | `internal/cli/init.go` |
| Doctor command | Check config resolution, store loadability, file permissions, Git tracking, and completion install state | `internal/cli/doctor.go` |
| Config layer | Resolve runtime config from flags, environment variables, config YAML, and defaults | `internal/config/config.go` |
| Store layer | Load, validate, mutate, lock, and atomically save the JSON secret store | `internal/store/io.go`, `internal/store/lock.go`, `internal/store/model.go`, `internal/store/path.go`, `internal/store/validate.go` |
| Manifest layer | Load, validate, mutate, and save project manifest entries | `internal/manifest/manifest.go`, `internal/manifest/io.go`, `internal/manifest/validate.go` |
| Render layer | Convert secrets and resolved bindings into env, shell, and JSON output | `internal/render/export.go` |
| Version layer | Compose semantic version plus VCS revision from Go build info | `internal/version/version.go` |

## Pattern Overview

- Keep command construction and UX in `internal/cli/*.go`; keep reusable domain behavior in `internal/store`, `internal/manifest`, `internal/config`, and `internal/render`.
- Persist state in local files, not a database or server: the secret store is JSON, config is YAML, and project bindings live in `.shelf.json`.
- Use explicit load-mutate-save flows. Mutating secret-store commands acquire a file lock through `store.LockFile` before loading and saving.
- Treat project resolution as a shared CLI helper: `resolveProjectEntries` in `internal/cli/project.go` is reused by `project explain`, `project export`, and `run`.

## Layers

- Purpose: Start the process and convert returned errors into stderr output and exit status.
- Location: `cmd/shelf/main.go`
- Contains: `main()`
- Depends on: `internal/cli`
- Used by: Go build output in `bin/shelf`
- Purpose: Define user commands, parse flags/args through Cobra, route to domain packages, and write stdout/stderr.
- Location: `internal/cli`
- Contains: command factories such as `newSecretCmd`, `newProjectCmd`, `newRunCmd`, `newDoctorCmd`, and helpers such as `loadRuntime`.
- Depends on: `github.com/spf13/cobra`, `internal/config`, `internal/store`, `internal/manifest`, `internal/render`, `internal/version`, Go `os/exec`.
- Used by: `cmd/shelf/main.go` and CLI tests in `internal/cli/*_test.go`.
- Purpose: Resolve runtime paths and editor selection.
- Location: `internal/config/config.go`
- Contains: `Config`, `Runtime`, `Resolve`, default paths, YAML parsing, path expansion.
- Depends on: `gopkg.in/yaml.v3`, Go `os` and `filepath`.
- Used by: `internal/cli/root.go`, `internal/cli/init.go`, `internal/cli/doctor.go`.
- Purpose: Own the secret data model, path grammar, JSON validation, atomic persistence, backups, prefix listing, and write locking.
- Location: `internal/store`
- Contains: `Data`, `Secret`, `Info`, `Store`, `SecretID`, `Load`, `Save`, `Set`, `Update`, `Delete`, `LockFile`.
- Depends on: Go standard library only.
- Used by: `internal/cli`, `internal/manifest/validate.go`, `internal/render/export.go`.
- Purpose: Own `.shelf.json` schema, validation, duplicate detection, and atomic project manifest persistence.
- Location: `internal/manifest`
- Contains: `Manifest`, `Entry`, `Load`, `Save`, `Validate`, `AddEntry`, `RemoveEntry`, `FindEntry`.
- Depends on: `internal/store` for path validation.
- Used by: `internal/cli/project.go`, `internal/cli/run.go`.
- Purpose: Convert stored JSON values and secret paths into environment variable bindings.
- Location: `internal/render/export.go`
- Contains: `Binding`, `EnvName`, `ValueString`, `Env`, `Shell`, `JSON`, `ShellQuote`.
- Depends on: `internal/store`.
- Used by: `internal/cli/export.go`, `internal/cli/project.go`.

## Data Flow

### Direct Secret Write Path

### Direct Export Path

### Project Export Path

### Runtime Injection Path

- Application state is file-backed. Global in-memory state is avoided except for test seams around interactive password input in `internal/cli/secret.go`.
- `Store` is an in-memory snapshot of one JSON file and is not shared across commands.
- Mutating store writes must use `loadRuntimeForWrite` to serialize concurrent writers with `store.LockFile`.
- Project manifest writes are atomic but do not use the secret-store lock because `.shelf.json` is a separate project file.

## Key Abstractions

- Purpose: Each command file builds commands with `new*Cmd` functions and `RunE` closures.
- Examples: `internal/cli/root.go`, `internal/cli/secret.go`, `internal/cli/project.go`, `internal/cli/run.go`.
- Pattern: Keep command-local flags as closure variables, call domain packages from `RunE`, and write via `cmd.OutOrStdout()` / `cmd.OutOrStderr()`.
- Purpose: Resolved paths and editor executable for one command invocation.
- Examples: `internal/config/config.go`, `internal/cli/root.go`, `internal/cli/init.go`.
- Pattern: Use `config.Resolve(configFlag, dataFlag)`; do not read config paths directly in command handlers.
- Purpose: File-backed secret collection with validated path and value semantics.
- Examples: `internal/store/model.go`, `internal/store/io.go`.
- Pattern: Use `store.Load(path)` for reads; use `loadRuntimeForWrite` plus `Store.Save` for writes.
- Purpose: Enforce the `group_path:key` identity format.
- Examples: `internal/store/path.go`, `docs/data-spec.md`.
- Pattern: Validate every externally supplied secret path with `store.ValidatePath` or `store.ParseSecretID`.
- Purpose: Represent either an exact secret path or a prefix binding in `.shelf.json`.
- Examples: `internal/manifest/manifest.go`, `internal/manifest/validate.go`.
- Pattern: Exactly one of `Entry.Path` or `Entry.Prefix` must be set; prefix entries cannot carry `Env`.
- Purpose: Internal CLI representation of concrete env bindings from a manifest plus store.
- Examples: `internal/cli/project.go`.
- Pattern: Use `resolveProjectEntries` instead of reimplementing path/prefix expansion.
- Purpose: Output-facing env name/value pair.
- Examples: `internal/render/export.go`.
- Pattern: Convert values with `render.ValueString` and emit with the format-specific binding functions.

## Entry Points

- Location: `cmd/shelf/main.go`
- Triggers: Running the compiled `shelf` binary.
- Responsibilities: Execute root command, print returned errors, exit with command-specific code.
- Location: `internal/cli/root.go`
- Triggers: `cli.NewRootCmd()`.
- Responsibilities: Register persistent flags and subcommands; expose shared runtime loading helpers.
- Location: `internal/store/io.go`
- Triggers: `store.Load`, `Store.Save`, `Store.Set`, `Store.Update`, `Store.Delete`.
- Responsibilities: Load and persist the JSON store with validation and atomic writes.
- Location: `internal/manifest/io.go`
- Triggers: `manifest.Load`, `manifest.Save`.
- Responsibilities: Parse strict JSON, reject unknown/trailing content, validate schema, persist atomically.
- Location: `internal/cli/project.go`
- Triggers: `project explain`, `project export`, and `run`.
- Responsibilities: Resolve manifest entries into env bindings and diagnostics.
- Location: `internal/cli/run.go`
- Triggers: `shelf run -- ...`.
- Responsibilities: Build the child environment, execute the requested command, and propagate child exit codes.

## Architectural Constraints

- **Threading:** The CLI uses a single process execution path. Concurrent writers are handled by OS file locks in `internal/store/lock.go`, not goroutines.
- **Global state:** Normal runtime logic avoids mutable globals. `internal/cli/secret.go` exposes package-level function variables for terminal detection/password reading so tests can replace them.
- **Circular imports:** Current package direction is acyclic: `cmd/shelf` → `internal/cli`; `internal/cli` → domain packages; `internal/manifest` and `internal/render` → `internal/store`; `internal/store` depends only on the standard library.
- **Storage boundary:** Secret persistence is intentionally isolated in `internal/store`; future encryption or backend changes should wrap or replace `Load`/`Save` without changing CLI command semantics.
- **Git dependency:** Project identity and manifest discovery depend on a Git worktree through `git` subprocess calls in `internal/cli/project.go`.
- **Secret exposure:** `secret get`, `export`, `project export`, and `run` intentionally materialize secret values. `secret info`, `secret list`, `doctor`, and completion helpers should not print secret values.

## Anti-Patterns

### Bypassing Runtime Resolution

### Mutating Store Without Locking

### Reimplementing Project Binding Expansion

### Adding Domain Logic to `cmd/shelf`

## Error Handling

- Wrap parse/load errors with context using `fmt.Errorf("...: %w", err)` in packages such as `internal/store/io.go` and `internal/manifest/io.go`.
- Use Cobra `SilenceUsage` and `SilenceErrors` on the root command so returned errors are not duplicated (`internal/cli/root.go`).
- Use diagnostic lines for project resolution problems, then return one summary error when any required entry fails (`internal/cli/project.go`, `internal/cli/run.go`).
- Preserve child exit statuses with `exitCodeError` and `ExitCode` in `internal/cli/run.go`.

## Cross-Cutting Concerns



## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.

## OMP Workflow

Non-trivial work follows the omp-workflow cadence: classify, plan, execute, review, verify, capture, ship.

- Fast mode for small, clear, low-risk edits.
- Focused mode for bounded work benefiting from written intent.
- Full mode for phased, multi-session, risky, or coordinated work.

Keep planning artifacts in `.planning/`; update `STATE.md` when workflow position changes.

