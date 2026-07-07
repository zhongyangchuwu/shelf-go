# Summary: Vault Model and Validation Core

## Completed Changes

- Added `internal/vaultmodel` as the v2 in-memory vault model package.
- Added v2 model types for `Vault`, `Item`, `Field`, `Binding`, `Template`, and `FieldTemplate`.
- Added constants for v2 version, built-in template IDs, field types, sensitivity, purposes, and env binding type.
- Added built-in template registry for `generic_secret` and `api_credential`.
- Added validation for vault version, duplicate item IDs, empty IDs/titles, unknown templates, missing required fields, duplicate field IDs, built-in field metadata mismatches, invalid field values, invalid env bindings, duplicate env binding names, and public-field env bindings.
- Added runtime indexes for item ID, exact title, alias, and env binding lookup.
- Added package tests for valid built-in template examples, invalid validation cases, and index lookups.
- Updated architecture lint config to recognize `internal/vaultmodel` as a core component.
- Removed a direct CLI dependency on `internal/project` in the project configure TUI by projecting env-file detail types into `internal/app`.

## Files Changed

- `.go-arch-lint.yml`
- `internal/vaultmodel/model.go`
- `internal/vaultmodel/templates.go`
- `internal/vaultmodel/validate.go`
- `internal/vaultmodel/index.go`
- `internal/vaultmodel/vaultmodel_test.go`
- `internal/app/project.go`
- `internal/cli/project_configure.go`
- `internal/cli/project_test.go`
- `.planning/phases/035-vault-model-validation-core/CONTEXT.md`
- `.planning/phases/035-vault-model-validation-core/PLAN.md`

## Deviations

- The implementation keeps custom template loading out of `Validate`; `Vault.Templates` remains reserved for a later phase. `ValidateWithRegistry` exists for explicit registry-based validation.
- The phase includes a small architecture cleanup in project configure state projection so `go-arch-lint check` passes with the new component graph.
- No item ID generation was added; this phase validates supplied IDs only.

## Evidence

- `go test ./internal/vaultmodel` passed.
- `go test ./internal/vault ./internal/project ./internal/app ./internal/cli ./internal/vaultmodel` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go-arch-lint check` passed.

## Unresolved Risks

- Exact item ID generation and deterministic migration ID format remain for later phases.
- JSON codec/repository behavior is not implemented until Phase 36.
- Legacy aliases exist in the model and index but lifetime policy remains for migration/cutover phases.
