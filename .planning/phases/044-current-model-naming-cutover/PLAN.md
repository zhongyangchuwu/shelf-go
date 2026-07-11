# Plan: Current-Model Naming Cutover

## Goal

Remove obsolete `V2` names from active code after the v1/age stack removal. Preserve only persisted schema version fields (`Vault.Version`, `Manifest.Version`, `CurrentVersion`).

## Scope

- Rename current item-vault app APIs, summaries, info types, helpers, tests, and files from `*V2` to current names.
- Rename project bindings/resolution APIs, helpers, tests, and files from `*V2` to current names.
- Use LSP renames for exported symbols and rename files without compatibility shims.
- Keep CLI syntax and persisted JSON unchanged.

## Acceptance

- No transitional `V2` identifiers or filenames remain in active `internal/` or `cmd/` code.
- No old aliases or compatibility wrappers remain.
- Persisted schema version validation remains intact.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
