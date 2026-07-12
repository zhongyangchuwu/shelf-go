# Contributing

Shelf Go is currently a local-development Go CLI. Product documentation lives under `docs/`; planning, roadmap, requirements, phase records, and durable workflow state live under `.planning/`.

## Requirements

- Go 1.26.4 or a compatible Go toolchain for this module.
- Git for project-aware commands and tests.
- `just` is optional; every recipe has an equivalent command.

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
./scripts/test.sh
```

Or:

```bash
just test
```

## Install locally

Use the reusable install script:

```bash
./scripts/install.sh
```

It runs `go install ./cmd/shelf` and writes zsh completion. Optional environment overrides:

- `GOBIN`
- `SHELF_COMPLETION_DIR`
- `SHELF_COMPLETION_FILE`

The `just install` recipe delegates to the same script.

## Project layout

```text
cmd/shelf/           process entry point

internal/cli/        Cobra command surface, flags, completions, and output routing
internal/cliui/      Lip Gloss-backed human output helpers for titles, summaries, empty states, and tables
internal/tui/          Bubble Tea secret and project-binding workspaces backed by app and clipboard boundaries
internal/clipboard/  System clipboard adapter for explicit `secret copy` side effects
internal/app/        application workflows for setup, doctor, secret, project, and run

internal/vaultmodel/ item/field/template model, validation, profiles, and indexes
internal/vaultjson/  strict JSON codec/repository for the local item vault
internal/project/    Git project identity, .shelf.json schema/IO/validation, and item binding resolution
internal/config/     config and vault path resolution
internal/util/       atomic write and env/shell/JSON binding formatting
```

`docs/architecture.md` is the source of truth for dependency direction. Keep `internal/cli` command-family oriented; put human output rendering in `internal/cliui`, explicit clipboard effects in `internal/clipboard`, and reusable behavior in `internal/app`, `internal/project`, `internal/vaultmodel`, or `internal/vaultjson`.

## Documentation policy

- Keep user-facing docs current with implemented behavior.
- Keep planning artifacts in `.planning/`, not `docs/`.
- Do not publish stale generated codebase maps or agent plans as product docs.
- Security-sensitive docs must state plaintext boundaries explicitly in `SECURITY.md`.
- Portable vault and chezmoi workflow guidance belongs in `docs/portable-vault.md`.
- Architecture package boundaries belong in `docs/architecture.md`; implementation phase history stays in `.planning/`.

## Code style

- Run `gofmt` on changed Go files.
- Keep Cobra command constructors package-private and named `new<Name>Cmd`.
- Return errors from package code instead of logging or exiting.
- Use `cmd.OutOrStdout()` and `cmd.OutOrStderr()` in Cobra commands.
- Use `internal/cliui` for human-facing command output.
- Avoid printing secret values except in explicit value-producing commands: `secret get`, `secret export`, `project export`, and `project run`. `secret copy` writes only to the clipboard and confirms without printing the value.

## Testing style

- Prefer behavior tests through real command execution helpers.
- Use real files and `t.TempDir()` for store, vault, manifest, and config behavior.
- Do not mock Git when behavior depends on a Git worktree; tests use real `git` commands.
- Do not mock child process execution for `project run`; tests execute real local shell commands.

## Local wrap-up

This repository is not being prepared for public release yet. For v0.2.0 local development wrap-up, keep verification boring:

```bash
./scripts/test.sh
go build ./cmd/shelf
go install ./cmd/shelf
shelf doctor
```

Public release packaging, tags, and GitHub Release automation are deferred until a future public v1.0.0 milestone.
