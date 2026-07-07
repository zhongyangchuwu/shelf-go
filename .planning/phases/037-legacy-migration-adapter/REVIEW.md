# Review: Legacy Migration Adapter

## Scope Reviewed

- `internal/vaultmigrate/model.go`
- `internal/vaultmigrate/vault.go`
- `internal/vaultmigrate/project.go`
- `internal/vaultmigrate/vaultmigrate_test.go`
- `.go-arch-lint.yml`

## Findings

- No blocking findings.
- `ConvertVault` validates v1 input and v2 output, preserving values and metadata without writing files.
- Converted items use `generic_secret`; no automatic API credential inference exists.
- Item IDs are deterministic and hashed so they do not directly encode old path text.
- Project manifest conversion is value-free; it stores item ID, field ID, env name, required state, and title snapshots only.
- Prefix/tag entries are expanded deterministically and warnings are emitted when selectors match nothing.
- Architecture boundary is narrow: `vaultmigrate` depends on `project`, `vault`, and `vaultmodel` only.

## Fixes Applied

- Derived env names are included in v2 field bindings so old secrets without explicit env still produce usable project bindings.
- Migration mappings include labels for tag expansion but exclude those labels from JSON serialization.
- Converted project manifests are validated before returning success.
- Tests assert manifests do not contain fake secret values.

## Waivers

- No public CLI migration command is added; Phase 37 intentionally provides internal conversion logic first.
- No file-writing migration orchestration is included; output validation and repository writing can be composed in a later command/service phase.

## Remaining Risks

- Prefix/tag expansion depends on the currently migrated vault contents and should be explained in any future user-facing migration command.
- Future project manifest v2 runtime may choose a different on-disk package/type; this internal migration shape may need adaptation.
