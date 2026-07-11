# Plan: Vault Resolver Consolidation

## Goal

Make `internal/vaultmodel` the single owner of item selection, primary-secret-field lookup, field lookup, and env-name-field lookup.

## Scope

- Add a validated, per-vault resolver that builds the registry once.
- Move selector ambiguity handling and item/field/env domain rules out of app/project helpers.
- Replace duplicate secret/project resolver helpers without changing CLI or JSON contracts.
- Add cross-workflow consistency coverage.

## Acceptance

- Secret and project workflows use the shared resolver for common item/field/env decisions.
- Project-specific V2 resolver helpers and duplicate secret helpers are deleted.
- Selector and ambiguity diagnostics remain value-free.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
