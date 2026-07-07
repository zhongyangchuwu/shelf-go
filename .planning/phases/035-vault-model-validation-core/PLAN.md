# Plan: Vault Model and Validation Core

## Objective

Add the v2 vault model and validation core as an internal package that can represent generic secrets and API credentials with field-level sensitivity and env bindings, without changing existing v1 runtime behavior.

## Scope

- Add `internal/vaultmodel` model types for `Vault`, `Item`, `Field`, `Binding`, `Template`, and `FieldTemplate`.
- Add constants for version, built-in template IDs, field types, sensitivities, purposes, and binding types.
- Add built-in template registry for `generic_secret` and `api_credential`.
- Add validation for required template fields, duplicate item IDs, duplicate field IDs, unknown templates, field metadata mismatches, invalid field values, invalid env binding names, env bindings on public fields, and duplicate env binding names.
- Add runtime index construction for item ID, title, alias, and env binding lookup.
- Add focused unit tests for valid and invalid model cases.

## Non-Goals

- No v2 JSON codec/repository.
- No migration implementation.
- No CLI command changes.
- No project manifest v2 implementation.
- No manager adaptation.
- No encryption changes.
- No item ID generator beyond validation of provided IDs.

## Tasks

1. Create `internal/vaultmodel` with model constants and structs.
2. Implement built-in templates and template lookup.
3. Implement validation helpers for vault, item, field, bindings, and typed values.
4. Implement runtime index construction.
5. Add behavior-focused package tests.
6. Run focused tests for `internal/vaultmodel` and relevant existing packages.
7. Write `SUMMARY.md`, `REVIEW.md`, `VERIFICATION.md`, and `CAPTURE.md`.
8. Update root planning artifacts and commit the completed phase.

## Acceptance Criteria

- `internal/vaultmodel` compiles and has no dependencies on CLI/app/project/manager/v1 vault packages.
- Valid `generic_secret` and `api_credential` items pass validation.
- Validation rejects duplicate item IDs, duplicate field IDs, unknown templates, missing required fields, mismatched built-in field metadata, invalid env binding names, public-field env bindings, and duplicate env binding names.
- Runtime indexes expose deterministic lookup for item IDs, exact titles, aliases, and env binding names.
- Existing v1 tests for `internal/vault`, `internal/project`, and package-wide tests still pass.

## Verification

- `go test ./internal/vaultmodel`
- `go test ./internal/vault ./internal/project ./internal/vaultmodel`
- `go test ./...`
- `go vet ./...` if tests pass.

## Risks

- Too much migration logic can leak into the core model. Keep aliases as value-free selectors only.
- Over-validating optional fields may make future custom templates harder. Restrict strict metadata matching to built-in template fields.
- Adding a new core package may require `.go-arch-lint.yml` updates before architecture checks understand the boundary.
