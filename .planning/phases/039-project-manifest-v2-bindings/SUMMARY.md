# Summary: Project Manifest v2 Bindings

## Completed Changes

- Added `.shelf.json` version 2 support in `internal/project.Manifest` with concrete `bindings` records: `item`, `field`, `env`, optional `required`, and optional `title`.
- Kept version 1 manifest support for existing path/prefix/tag workflows.
- Added strict v2 manifest validation for required binding fields, env-name syntax, duplicate item/field bindings, and version-specific schema separation.
- Added v2 manifest JSON round-trip behavior that writes `bindings: []` for empty v2 manifests while preserving v1 `secrets: []` behavior.
- Added v2 project binding resolution over `vaultmodel.Vault` with value-free diagnostics for missing items, missing fields, missing values, non-secret fields, and duplicate env names.
- Wired `project status`, `project export`, and `project run` to dispatch v1 vs v2 manifest resolution.
- Wired `project add`, `project list`, `project rm`, completion candidates, and configure-state entries for v2 manifests.
- Added `project add --field` for v2 manifests; it defaults to the primary secret field and derives env names from public env-name fields unless `--env` is supplied.
- Updated docs and planning artifacts to describe v2 project bindings.

## Files Changed

- `.go-arch-lint.yml`
- `CHANGELOG.md`
- `README.md`
- `docs/architecture.md`
- `docs/getting-started.md`
- `docs/portable-vault.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `internal/app/project.go`
- `internal/app/run.go`
- `internal/app/project_v2_test.go`
- `internal/cli/project.go`
- `internal/cli/project_test.go`
- `internal/project/manifest.go`
- `internal/project/manifest_validate.go`
- `internal/project/manifest_test.go`
- `internal/project/resolve_v2.go`
- `internal/project/resolve_v2_test.go`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/039-project-manifest-v2-bindings/CONTEXT.md`
- `.planning/phases/039-project-manifest-v2-bindings/PLAN.md`

## Deviations

- `project init` still creates version 1 manifests to preserve existing v1 encrypted-vault project workflows during cutover. Version 2 manifests can be created by writing `{"version":2,"bindings":[]}` or by later migration/cutover tooling.
- Prefix/tag-style selectors are not added to v2 manifests. Phase 39 keeps v2 concrete as item/field/env bindings; v1 prefix/tag behavior and migration expansion remain supported separately.
- No public migration command was added.
- V2 repository encryption remains deferred; v2 project resolution uses the same Phase 38 v2 JSON path behavior.

## Evidence

- `go test ./internal/project ./internal/app ./internal/cli` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go-arch-lint check` passed.

## Unresolved Risks

- Version 2 project workflows depend on plaintext v2 JSON storage until age wrapping lands.
- Version 1 and version 2 project models coexist during cutover; later phases should reduce this branch once migration/cutover is explicit.
- The local manager is still not v2-aware; Phase 40 owns adaptation or gating.
