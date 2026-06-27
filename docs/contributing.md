# Contributing

Shelf Go is a Go CLI. Public product documentation lives under `docs/`; planning, roadmap, requirements, phase records, and durable workflow state live under `.planning/`.

## Requirements

- Go 1.26.4 or a compatible Go toolchain for this module.
- Git for project-aware commands and tests.
- `just` is optional; every recipe has an equivalent Go command.

## Build

```bash
go build -o ./bin/shelf ./cmd/shelf
```

Or:

```bash
just build
```

## Test

```bash
go test ./...
```

Or:

```bash
just test
```

## Install locally

```bash
go install ./cmd/shelf
```

The `just install` recipe also installs zsh completion to `~/.zfunc/_shelf`:

```bash
just install
```

## Project layout

```text
cmd/shelf/           process entry point

internal/cli/        Cobra commands and CLI orchestration
internal/manager/    local manager surface, currently loopback HTTP/Web

internal/app/        runtime, vault construction, and version composition
internal/project/    project identity, .shelf.json schema/IO/validation, and binding resolution
internal/secret/     reusable secret workflows such as editor-based updates

internal/config/     runtime config resolution
internal/vault/      encrypted vault core, persistence, locking, diagnostics
internal/exportfmt/  env/shell/JSON export formatting
```

## Documentation policy

- Keep user-facing docs current with implemented behavior.
- Keep planning artifacts in `.planning/`, not `docs/`.
- Do not publish stale generated codebase maps or agent plans as public docs.
- Security-sensitive docs must state plaintext boundaries explicitly in `SECURITY.md`.
- Portable vault and chezmoi workflow guidance belongs in `docs/portable-vault.md`.
- Architecture package boundaries belong in `docs/architecture.md`; implementation phase history stays in `.planning/`.

## Code style

- Run `gofmt` on changed Go files.
- Keep Cobra command constructors package-private and named `new<Name>Cmd`.
- Return errors from package code instead of logging or exiting.
- Use `cmd.OutOrStdout()` and `cmd.OutOrStderr()` for command output.
- Avoid printing secret values except in explicit value-producing commands.

## Testing style

- Prefer behavior tests through real command execution helpers.
- Use real files and `t.TempDir()` for store, vault, manifest, and config behavior.
- Do not mock Git when behavior depends on a Git worktree; tests use real `git` commands.
- Do not mock child process execution for `project run`; tests execute real local shell commands.

## Release planning

Use `.planning/PROJECT.md`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`, and phase directories for planning state. Public `docs/` should describe current behavior, not internal phase history.

Release automation uses GoReleaser for GitHub Release binaries and checksums. For 0.1.x, package-manager distribution is intentionally deferred; validate release config locally before tagging:

```bash
goreleaser check
goreleaser release --clean --snapshot
```
