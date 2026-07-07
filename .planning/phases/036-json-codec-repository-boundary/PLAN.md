# Plan: JSON Codec and Repository Boundary

## Objective

Add the v2 JSON codec and plaintext file repository boundary for `internal/vaultmodel` while keeping current v1 storage and CLI behavior unchanged.

## Scope

- Define v2 codec interface for decoding/encoding `vaultmodel.Vault`.
- Implement strict JSON decode with unsupported-version and unknown-field errors.
- Validate decoded vaults through `vaultmodel.Validate`.
- Preserve item and field ordering on round-trip.
- Implement a plaintext JSON repository with load/save behavior and atomic writes.
- Add focused codec/repository tests with fake secret values only.
- Update architecture lint config for the new storage package.

## Non-Goals

- No age encryption adapter.
- No migration command.
- No CLI cutover.
- No project manifest v2 implementation.
- No manager changes.

## Tasks

1. Select package name and component boundary.
2. Implement codec decode/encode.
3. Implement repository load/save with validation.
4. Add focused tests for round-trip, strict decode failures, save validation, and atomic persistence behavior.
5. Run focused and project-wide verification.
6. Write summary, review, verification, capture, update root planning docs, and commit.

## Acceptance Criteria

- v2 JSON round-trips a valid `vaultmodel.Vault` without reordering items or fields.
- Decode rejects unknown top-level fields, trailing JSON, unsupported versions, and invalid vault contents.
- Save rejects invalid vaults before writing.
- Repository load returns an empty v2 vault for missing files only if that behavior is explicitly chosen and tested; otherwise it returns a clear not-found error.
- Existing v1 packages and CLI behavior remain unchanged.

## Verification

- Focused tests for the new v2 JSON/repository package.
- `go test ./internal/vaultmodel ./<new-package>`
- `go test ./...`
- `go vet ./...`
- `go-arch-lint check`

## Risks

- Naming can confuse v1 `jsonvault` and v2 JSON storage; choose a package name that makes version boundary explicit.
- Returning an empty vault on missing files may hide misconfiguration; decide explicitly before implementation.
- Atomic write and backup behavior should reuse existing conventions without importing v1 storage internals into the v2 package.
