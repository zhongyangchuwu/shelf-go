# Phase 28 Context: Architecture Boundary Lint and CLI Adapter Slimming

## Goal
Clarify Shelf Go internal package boundaries, encode them with go-arch-lint, and slim `internal/cli` so it primarily owns Cobra command surface, flags, completions, output channels, terminal wiring, and process lifecycle.

## Problem
After Phases 25-27, `internal/cli/project.go` and `internal/cli/secret.go` still contain application orchestration and workflow logic. `internal/app`, `internal/project`, and `internal/secret` overlap because package boundaries are not enforced by tooling.

Current observed internal imports from `go list -json ./internal/...`:

```text
app      -> config, exportfmt, vault
cli      -> app, config, exportfmt, manager, project, secret, vault
config   -> external yaml only
exportfmt-> vault
manager  -> exportfmt, vault
project  -> exportfmt, vault
secret   -> vault
vault    -> config
```

## Chosen Architecture
Use a light Go-friendly Onion / ports-and-adapters shape:

```text
cmd/shelf -> cli
cli, manager -> app
app -> config, vault, project, secret, exportfmt, manager helpers as needed
project, secret, exportfmt -> vault
vault, config -> no project-internal imports
```

Package responsibilities:

- `internal/cli`: Cobra commands, flags, args validators, shell completions, terminal stdio wiring, final stdout/stderr writes, child process lifecycle.
- `internal/manager`: localhost HTTP/UI adapter and server/browser lifecycle; prefer app use cases for secret/vault/project behavior.
- `internal/app`: application use cases that compose config, vault, project, secret, export formatting, setup, migration, manager helpers, and return typed results or strings.
- `internal/project`: `.shelf.json` schema, validation, load/save, Git project identity, selector entry construction, resolution, diagnostics, env/session rules.
- `internal/secret`: secret editing/add workflow and secret-level transformations over a provided vault store.
- `internal/vault`: encrypted file storage, locking, age, JSON model, secret path/value validation; no config resolution.
- `internal/config`: runtime config resolution only.
- `internal/exportfmt`: pure env/shell/JSON rendering helpers.

## External Tool
Use `go-arch-lint` schema version 3. The config checks import direction from `.go-arch-lint.yml`; it does not impose an architecture by itself.

## Preserve
- CLI commands, flags, completions, output formats, and error wording unless unavoidable.
- Vault format and `.shelf.json` schema.
- Manager routes/responses.

## Verification Target
- `go test ./...`
- `go-arch-lint check --arch-file .go-arch-lint.yml --project-path ./`
