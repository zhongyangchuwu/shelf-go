# Summary: Legacy Migration Adapter

## Completed Changes

- Added `internal/vaultmigrate` for internal v1-to-v2 migration logic.
- Added deterministic conversion from v1 `vault.Data` flat secret records to v2 `vaultmodel.Vault` items.
- Converted every v1 secret to the `generic_secret` template by default; no API credential inference is performed.
- Preserved v1 secret values, explicit or derived env bindings, descriptions as notes, tags as labels, and old paths as aliases.
- Added deterministic non-path-leaking item IDs derived from hashed v1 paths.
- Added path-to-item migration mappings for project manifest conversion.
- Added value-free v2 project manifest binding model and validation.
- Converted v1 project manifest path, prefix, and tag entries into item/field/env bindings with warnings for empty prefix/tag expansion.
- Registered `vaultmigrate` in architecture lint.
- Added focused migration tests.

## Files Changed

- `.go-arch-lint.yml`
- `internal/vaultmigrate/model.go`
- `internal/vaultmigrate/vault.go`
- `internal/vaultmigrate/project.go`
- `internal/vaultmigrate/vaultmigrate_test.go`
- `.planning/phases/037-legacy-migration-adapter/CONTEXT.md`
- `.planning/phases/037-legacy-migration-adapter/PLAN.md`

## Deviations

- Prefix and tag manifest entries are expanded immediately against the migration mapping instead of being preserved as selector-like v2 manifest entries. This keeps the v2 manifest value-free and executable by item/field bindings.
- Duplicate converted bindings are deduplicated by item/field/env key. This prevents prefix plus tag overlap from writing duplicate v2 bindings.
- The adapter is internal only; no public migration CLI command is added in this phase.

## Evidence

- `go test ./internal/vaultmigrate` passed.
- `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go-arch-lint check` passed.

## Unresolved Risks

- The public migration command and file write orchestration remain for a later phase.
- The v2 project manifest model is internal to migration and not yet used by project export/run.
- Alias lifetime after migration remains a cutover policy decision.
