# Plan: Legacy Migration Adapter

## Objective

Add internal migration adapters for converting v1 vault data and project manifests to v2 model data and value-free item/field bindings, without changing current CLI behavior or deleting source files.

## Scope

- Choose package boundary for v1-to-v2 migration logic.
- Implement deterministic conversion from v1 vault `path -> Secret` records to v2 `generic_secret` items.
- Preserve old paths as aliases.
- Preserve values, env bindings, descriptions, and tags.
- Build a migration map from v1 path to v2 item ID + primary field ID.
- Define a v2 project manifest binding model if needed for conversion tests.
- Convert v1 path manifest entries to value-free item/field/env bindings.
- Decide and test handling for prefix/tag manifest entries in this phase.
- Validate converted v2 vaults through `vaultmodel.Validate`.

## Non-Goals

- No public migration CLI command unless later phase requires it.
- No v2 secret CLI cutover.
- No project export/run v2 runtime integration.
- No manager adaptation.
- No age encryption adapter.

## Tasks

1. Inspect v1 vault and project manifest types and tests.
2. Select migration package/API boundary.
3. Implement v1 vault conversion with deterministic IDs and aliases.
4. Implement v1 project manifest conversion or value-free intermediate model.
5. Add tests for value/metadata preservation, no API-credential inference, value-free manifests, and validation failure handling.
6. Run focused and project-wide verification.
7. Write summary, review, verification, capture, update root planning docs, and commit.

## Acceptance Criteria

- v1 flat secrets convert to v2 `generic_secret` items.
- Converted v2 items preserve values, env bindings, descriptions as notes, and tags as labels.
- Old v1 paths are available as aliases or migration map entries.
- API credential templates are not inferred automatically.
- Project manifest conversion stores item IDs, field IDs, env names, required state, and optional title snapshots without secret values.
- Conversion validates v2 output before reporting success.

## Verification

- Focused migration package tests.
- `go test ./internal/vaultmodel ./internal/vaultjson ./<migration-package>`
- `go test ./...`
- `go vet ./...`
- `go-arch-lint check`

## Risks

- Deterministic IDs can leak old path structure if encoded directly; prefer hashed or normalized safe IDs.
- Prefix/tag manifest conversion may surprise users if expansion depends on current vault contents; record warnings and keep conversion explicit.
- Migration tests must avoid real secret values while still proving value preservation with fake fixtures.
