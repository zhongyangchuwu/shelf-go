# Summary: Phase 19 Secret Tag Selection

## Completed Changes

- Added store-level tag matching with AND semantics.
- Added deterministic `Store.ListByTags(prefix, tags)` selection.
- Added `shelf secret list --tag` with repeatable tag filters.
- Changed `shelf secret export` to accept optional path/prefix when `--tag` is provided.
- Added `shelf secret export --tag` support for env, shell, and JSON output through existing render paths.
- Preserved default env-only export behavior and `--all` derived-env behavior for tag-selected exports.
- Added focused store and CLI tests for tag selection.

## Files Changed

- `internal/store/store.go`
- `internal/store/store_test.go`
- `internal/cli/secret.go`
- `internal/cli/export.go`
- `internal/cli/tag_selection_test.go`
- `.planning/phases/019-secret-tag-selection/CONTEXT.md`
- `.planning/phases/019-secret-tag-selection/PLAN.md`

## Deviations

- None.

## Evidence

- `go test ./internal/store` passed.
- `go test ./internal/cli -run 'TestSecret.*Tag|TestSecretExport|TestExportPrefix|TestSecretSetGetListInfoExport'` passed.
- `go test ./...` passed.
- LSP workspace diagnostics reported no Go issues.

## Unresolved Risks

- Phase 20 must reuse `store.HasTags`/`Store.ListByTags` for project tag bindings to avoid divergent tag semantics.
