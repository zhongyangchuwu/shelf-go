# Review: Project Manifest v2 Bindings

## Scope Reviewed

- `internal/project/manifest.go`
- `internal/project/manifest_validate.go`
- `internal/project/resolve_v2.go`
- `internal/app/project.go`
- `internal/app/run.go`
- `internal/cli/project.go`
- project/app/CLI v2 tests
- user docs and planning records

## Findings

- No blocking findings.
- V2 project manifests are concrete and value-free: bindings contain item IDs, field IDs, env names, required state, and optional title only.
- `project export` and `project run` are the only Phase 39 project flows that materialize bound v2 field values.
- `project status`, diagnostics, `project list`, completions, and configure state remain value-free.
- Duplicate env names are allowed in the vault but fail at project resolution when multiple project bindings would inject the same env name.
- V1 project manifests still resolve through the existing encrypted-vault store path, so current v1 project tests remain valid.
- V2 app/CLI tests use explicit temp v2 JSON vault paths and fake secret values.

## Fixes Applied

- Added a custom manifest JSON marshaler so v1 empty manifests keep `secrets: []` and v2 empty manifests keep `bindings: []`.
- Added `project add --field` for v2 manifests instead of overloading selectors with field semantics only.
- Added architecture lint allowance for `internal/project -> internal/vaultmodel` because v2 resolution needs the v2 item/field domain model.
- Updated stale docs that described only v1 path/prefix/tag manifests.

## Waivers

- `project init` remains v1 by default during cutover to avoid breaking existing v1 project-add/export/run workflows.
- Public migration/cutover command remains deferred.
- V2 vault files remain plaintext Phase 38 JSON and are documented as not safe to sync/commit.
- Manager adaptation is deferred to Phase 40.

## Remaining Risks

- Coexistence of v1 and v2 manifest branches adds temporary complexity.
- V2 project workflows are only as safe as the plaintext v2 repository boundary until encryption is added.
- `project configure` can list and apply v2 bindings, but richer v2-specific UX remains basic.
