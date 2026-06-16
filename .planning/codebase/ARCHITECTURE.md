<!-- refreshed: 2026-06-16 -->
# Architecture

**Analysis Date:** 2026-06-16

## System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                         CLI Surface                          │
├──────────────────┬──────────────────┬───────────────────────┤
│ Secret Commands  │ Project Commands │ Runtime/Health Cmds   │
│ `internal/cli/`  │ `internal/cli/`  │ `internal/cli/`       │
└────────┬─────────┴────────┬─────────┴──────────┬────────────┘
         │                  │                     │
         ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain and Runtime Layers                 │
│ `internal/config`, `internal/store`, `internal/manifest`,    │
│ `internal/render`, `internal/version`                        │
└─────────────────────────────────────────────────────────────┘
         │                  │                     │
         ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│ Local Files and Child Processes                              │
│ config YAML, secrets JSON, `.shelf.json`, git, shell command │
└─────────────────────────────────────────────────────────────┘
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

**Overall:** Layered local CLI with thin Cobra command handlers over file-backed domain packages.

**Key Characteristics:**
- Keep command construction and UX in `internal/cli/*.go`; keep reusable domain behavior in `internal/store`, `internal/manifest`, `internal/config`, and `internal/render`.
- Persist state in local files, not a database or server: the secret store is JSON, config is YAML, and project bindings live in `.shelf.json`.
- Use explicit load-mutate-save flows. Mutating secret-store commands acquire a file lock through `store.LockFile` before loading and saving.
- Treat project resolution as a shared CLI helper: `resolveProjectEntries` in `internal/cli/project.go` is reused by `project explain`, `project export`, and `run`.

## Layers

**Executable Layer:**
- Purpose: Start the process and convert returned errors into stderr output and exit status.
- Location: `cmd/shelf/main.go`
- Contains: `main()`
- Depends on: `internal/cli`
- Used by: Go build output in `bin/shelf`

**CLI Command Layer:**
- Purpose: Define user commands, parse flags/args through Cobra, route to domain packages, and write stdout/stderr.
- Location: `internal/cli`
- Contains: command factories such as `newSecretCmd`, `newProjectCmd`, `newRunCmd`, `newDoctorCmd`, and helpers such as `loadRuntime`.
- Depends on: `github.com/spf13/cobra`, `internal/config`, `internal/store`, `internal/manifest`, `internal/render`, `internal/version`, Go `os/exec`.
- Used by: `cmd/shelf/main.go` and CLI tests in `internal/cli/*_test.go`.

**Configuration Layer:**
- Purpose: Resolve runtime paths and editor selection.
- Location: `internal/config/config.go`
- Contains: `Config`, `Runtime`, `Resolve`, default paths, YAML parsing, path expansion.
- Depends on: `gopkg.in/yaml.v3`, Go `os` and `filepath`.
- Used by: `internal/cli/root.go`, `internal/cli/init.go`, `internal/cli/doctor.go`.

**Secret Store Layer:**
- Purpose: Own the secret data model, path grammar, JSON validation, atomic persistence, backups, prefix listing, and write locking.
- Location: `internal/store`
- Contains: `Data`, `Secret`, `Info`, `Store`, `SecretID`, `Load`, `Save`, `Set`, `Update`, `Delete`, `LockFile`.
- Depends on: Go standard library only.
- Used by: `internal/cli`, `internal/manifest/validate.go`, `internal/render/export.go`.

**Project Manifest Layer:**
- Purpose: Own `.shelf.json` schema, validation, duplicate detection, and atomic project manifest persistence.
- Location: `internal/manifest`
- Contains: `Manifest`, `Entry`, `Load`, `Save`, `Validate`, `AddEntry`, `RemoveEntry`, `FindEntry`.
- Depends on: `internal/store` for path validation.
- Used by: `internal/cli/project.go`, `internal/cli/run.go`.

**Render Layer:**
- Purpose: Convert stored JSON values and secret paths into environment variable bindings.
- Location: `internal/render/export.go`
- Contains: `Binding`, `EnvName`, `ValueString`, `Env`, `Shell`, `JSON`, `ShellQuote`.
- Depends on: `internal/store`.
- Used by: `internal/cli/export.go`, `internal/cli/project.go`.

## Data Flow

### Direct Secret Write Path

1. User invokes a mutating command such as `shelf secret set ...`; Cobra routes through `NewRootCmd` (`internal/cli/root.go:10`) to secret command handlers (`internal/cli/secret.go`).
2. The handler calls `loadRuntimeForWrite`, which resolves paths with `config.Resolve`, locks `<data-path>.lock` with `store.LockFile`, then loads latest store data with `store.Load` (`internal/cli/root.go:46`).
3. The command validates and mutates in-memory store state through methods such as `Store.Set`, `Store.Update`, or `Store.Delete` (`internal/store/io.go:111`).
4. The handler calls `Store.Save`, which validates the full store, creates a `.bak` backup when the file exists, writes a temp file with mode `0600`, syncs it, and renames it into place (`internal/store/io.go:57`).
5. The deferred unlock releases the flock from `internal/store/lock.go`.

### Direct Export Path

1. User invokes `shelf export <path-or-prefix>`; Cobra routes to `newExportCmd` (`internal/cli/export.go:10`).
2. The command loads runtime and store with `loadRuntime` (`internal/cli/root.go:32`).
3. It resolves either an exact path via `Store.Get` or a prefix via `Store.List`, then optionally filters to secrets with explicit `Env` metadata (`internal/cli/export.go:20`).
4. It renders output through `render.JSON`, `render.Env`, or `render.Shell` (`internal/render/export.go`).

### Project Export Path

1. User invokes `shelf project export`; the command discovers the Git root using `git rev-parse --show-toplevel` (`internal/cli/project.go:474`).
2. It loads `.shelf.json` with `manifest.Load` from the Git root (`internal/manifest/io.go:12`).
3. It loads the configured secret store with `loadRuntime` (`internal/cli/root.go:32`).
4. `resolveProjectEntries` expands path and prefix manifest entries, checks required/optional missing values, derives env names, converts values to strings, and detects duplicate env names (`internal/cli/project.go:346`).
5. The selected project renderer writes env, shell, or JSON output via `render.*Bindings` (`internal/cli/project.go:411`).

### Runtime Injection Path

1. User invokes `shelf run -- command args...`; Cobra routes to `newRunCmd` (`internal/cli/run.go:41`).
2. The command loads `.shelf.json`, loads the secret store, and calls `resolveProjectEntries` (`internal/cli/run.go:49`).
3. Diagnostics are written to stderr; any `fail` diagnostic stops execution before child process launch (`internal/cli/run.go:62`).
4. In dry-run mode, the command prints override warnings and env names only (`internal/cli/run.go:72`).
5. In execution mode, `childEnv` overlays resolved bindings onto `os.Environ`, then `exec.Command` runs the requested program with inherited stdin and command stdout/stderr (`internal/cli/run.go:82`).
6. Child non-zero exits are wrapped in `exitCodeError`; `cmd/shelf/main.go` returns that exit code through `cli.ExitCode` (`internal/cli/run.go:30`).

**State Management:**
- Application state is file-backed. Global in-memory state is avoided except for test seams around interactive password input in `internal/cli/secret.go`.
- `Store` is an in-memory snapshot of one JSON file and is not shared across commands.
- Mutating store writes must use `loadRuntimeForWrite` to serialize concurrent writers with `store.LockFile`.
- Project manifest writes are atomic but do not use the secret-store lock because `.shelf.json` is a separate project file.

## Key Abstractions

**Cobra Command Factories:**
- Purpose: Each command file builds commands with `new*Cmd` functions and `RunE` closures.
- Examples: `internal/cli/root.go`, `internal/cli/secret.go`, `internal/cli/project.go`, `internal/cli/run.go`.
- Pattern: Keep command-local flags as closure variables, call domain packages from `RunE`, and write via `cmd.OutOrStdout()` / `cmd.OutOrStderr()`.

**Runtime:**
- Purpose: Resolved paths and editor executable for one command invocation.
- Examples: `internal/config/config.go`, `internal/cli/root.go`, `internal/cli/init.go`.
- Pattern: Use `config.Resolve(configFlag, dataFlag)`; do not read config paths directly in command handlers.

**Store:**
- Purpose: File-backed secret collection with validated path and value semantics.
- Examples: `internal/store/model.go`, `internal/store/io.go`.
- Pattern: Use `store.Load(path)` for reads; use `loadRuntimeForWrite` plus `Store.Save` for writes.

**SecretID and Path Grammar:**
- Purpose: Enforce the `group_path:key` identity format.
- Examples: `internal/store/path.go`, `docs/data-spec.md`.
- Pattern: Validate every externally supplied secret path with `store.ValidatePath` or `store.ParseSecretID`.

**Manifest Entry:**
- Purpose: Represent either an exact secret path or a prefix binding in `.shelf.json`.
- Examples: `internal/manifest/manifest.go`, `internal/manifest/validate.go`.
- Pattern: Exactly one of `Entry.Path` or `Entry.Prefix` must be set; prefix entries cannot carry `Env`.

**Resolved Project Entry:**
- Purpose: Internal CLI representation of concrete env bindings from a manifest plus store.
- Examples: `internal/cli/project.go`.
- Pattern: Use `resolveProjectEntries` instead of reimplementing path/prefix expansion.

**Render Binding:**
- Purpose: Output-facing env name/value pair.
- Examples: `internal/render/export.go`.
- Pattern: Convert values with `render.ValueString` and emit with the format-specific binding functions.

## Entry Points

**Process Entry:**
- Location: `cmd/shelf/main.go`
- Triggers: Running the compiled `shelf` binary.
- Responsibilities: Execute root command, print returned errors, exit with command-specific code.

**Root CLI Entry:**
- Location: `internal/cli/root.go`
- Triggers: `cli.NewRootCmd()`.
- Responsibilities: Register persistent flags and subcommands; expose shared runtime loading helpers.

**Secret Store Entry:**
- Location: `internal/store/io.go`
- Triggers: `store.Load`, `Store.Save`, `Store.Set`, `Store.Update`, `Store.Delete`.
- Responsibilities: Load and persist the JSON store with validation and atomic writes.

**Project Manifest Entry:**
- Location: `internal/manifest/io.go`
- Triggers: `manifest.Load`, `manifest.Save`.
- Responsibilities: Parse strict JSON, reject unknown/trailing content, validate schema, persist atomically.

**Project Resolution Entry:**
- Location: `internal/cli/project.go`
- Triggers: `project explain`, `project export`, and `run`.
- Responsibilities: Resolve manifest entries into env bindings and diagnostics.

**Child Process Entry:**
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

**What happens:** A command reads `SHELF_CONFIG`, `SHELF_DATA`, default paths, or config YAML directly.
**Why it's wrong:** It duplicates precedence rules and can disagree with `--config`, `--data`, and relative config data paths.
**Do this instead:** Call `config.Resolve` through `loadRuntime` or `loadRuntimeForWrite` in `internal/cli/root.go`.

### Mutating Store Without Locking

**What happens:** A command calls `store.Load`, mutates `Store.Data`, and then calls `Store.Save` without `store.LockFile`.
**Why it's wrong:** Concurrent writes can lose updates because each command works from a snapshot.
**Do this instead:** Use `loadRuntimeForWrite` in `internal/cli/root.go` before any write to the secret data file.

### Reimplementing Project Binding Expansion

**What happens:** A new command separately expands manifest paths/prefixes and derives env names.
**Why it's wrong:** Required/optional semantics, prefix diagnostics, env conflict detection, and value conversion can drift.
**Do this instead:** Reuse `resolveProjectEntries` in `internal/cli/project.go` or move it into a shared package before adding non-CLI consumers.

### Adding Domain Logic to `cmd/shelf`

**What happens:** The executable entry point starts resolving config, reading stores, or constructing commands directly.
**Why it's wrong:** Tests use `cli.NewRootCmd` directly; putting behavior in `cmd/shelf/main.go` bypasses the testable command surface.
**Do this instead:** Keep `cmd/shelf/main.go` as a thin adapter and place command behavior in `internal/cli`.

## Error Handling

**Strategy:** Return errors from command `RunE` functions and lower layers; centralize user-facing top-level error printing in `cmd/shelf/main.go`.

**Patterns:**
- Wrap parse/load errors with context using `fmt.Errorf("...: %w", err)` in packages such as `internal/store/io.go` and `internal/manifest/io.go`.
- Use Cobra `SilenceUsage` and `SilenceErrors` on the root command so returned errors are not duplicated (`internal/cli/root.go`).
- Use diagnostic lines for project resolution problems, then return one summary error when any required entry fails (`internal/cli/project.go`, `internal/cli/run.go`).
- Preserve child exit statuses with `exitCodeError` and `ExitCode` in `internal/cli/run.go`.

## Cross-Cutting Concerns

**Logging:** No logging framework is used. Commands write human/script output to Cobra stdout/stderr handles; tests capture those handles through `runShelf` in `internal/cli/test_helpers_test.go`.

**Validation:** Store path/value validation lives in `internal/store/path.go` and `internal/store/validate.go`. Manifest validation lives in `internal/manifest/validate.go` and delegates exact secret path validation to `internal/store`.

**Authentication:** Not applicable. The CLI has no network service authentication. Local access is controlled by filesystem permissions on config, data, lock, and manifest files.

**Persistence safety:** Secret store writes use a backup, temp file, fsync, and rename in `internal/store/io.go`; project manifest writes use temp file, fsync, and rename in `internal/manifest/io.go`.

**Completion:** Shell completion is generated by Cobra through `internal/cli/completion.go`; dynamic secret path completion loads the runtime store in `internal/cli/secret.go`.

---

*Architecture analysis: 2026-06-16*
