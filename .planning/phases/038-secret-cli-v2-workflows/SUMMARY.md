# Summary: Secret CLI v2 Workflows

## Completed Changes

- Replaced the old `group:key` secret CLI surface with v2 item/field commands: `secret create`, `secret set --stdin`, `secret get`, `secret list [--long]`, `secret info`, and `secret rm --yes`.
- Refactored `internal/vaultmodel` so env names are ordinary public fields (`value_env`, `credential_env`) instead of `Field.Bindings`.
- Added template metadata for env-name fields through `FieldTemplate.EnvFor` and `PurposeEnvName`.
- Added profile support in the registry, including the `openai` profile over `api_credential`.
- Updated `internal/vaultmodel` validation and indexes for env-name public fields and duplicate env-name lookup.
- Updated `internal/vaultjson` tests for the revised v2 shape.
- Updated `internal/vaultmigrate` so migrated v1 env names become `value_env` public fields and project mapping remains value-free.
- Added `internal/app/secret_v2.go` for v2 app-layer create, set, get, list, info, remove, selector resolution, and env lookup.
- Added direct v2 app service tests and v2 CLI tests for stdin-only writes, profile defaults, env-name field updates, value-free info/list, ambiguity errors, removed legacy paths, and removal.
- Removed obsolete v1 CLI tests for direct secret export and tag selection because those command paths are intentionally removed/deferred in the v2 secret CLI.
- Updated user docs, changelog, architecture docs, and planning docs for the v2 secret CLI cutover.
- Registered `app -> vaultmodel/vaultjson` as an allowed architecture dependency for the v2 app service boundary.

## Files Changed

- `.go-arch-lint.yml`
- `CHANGELOG.md`
- `README.md`
- `docs/architecture.md`
- `docs/getting-started.md`
- `docs/portable-vault.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `internal/app/secret_v2.go`
- `internal/app/secret_v2_test.go`
- `internal/app/secret_service_test.go`
- `internal/cli/secret.go`
- `internal/cli/secret_test.go`
- `internal/cli/test_helpers_test.go`
- `internal/cli/export_test.go` (removed)
- `internal/cli/tag_selection_test.go` (removed)
- `internal/vaultjson/codec_test.go`
- `internal/vaultmigrate/vault.go`
- `internal/vaultmigrate/vaultmigrate_test.go`
- `internal/vaultmodel/index.go`
- `internal/vaultmodel/model.go`
- `internal/vaultmodel/templates.go`
- `internal/vaultmodel/validate.go`
- `internal/vaultmodel/vaultmodel_test.go`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`

## Deviations

- `secret copy` was deferred. Clipboard behavior needs a small platform adapter and was explicitly excluded from the first v2 CLI cutover.
- Direct `secret export` was removed/deferred rather than ported. Project/runtime v2 bindings are Phase 39, so direct export needs a later design over v2 item/field selectors.
- The Phase 38 v2 CLI uses the plaintext `vaultjson` repository boundary. To avoid overwriting the v1 encrypted vault by default, v2 commands write to `<resolved-vault-path>.v2.json` unless `--vault` is explicitly supplied.
- Existing project/export/run/manager code remains v1-backed until Phase 39/40.

## Evidence

- `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate ./internal/app ./internal/cli` passed.
- `go test ./internal/app -run 'Test.*V2'` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go-arch-lint check` passed.

## Unresolved Risks

- Phase 38 v2 JSON files contain plaintext secret values until age wrapping is wired into the v2 repository.
- Project manifest v2 export/run integration is not implemented yet, so daily project workflows still need Phase 39.
- The local manager still targets v1 vault records and must be adapted or gated in Phase 40.
