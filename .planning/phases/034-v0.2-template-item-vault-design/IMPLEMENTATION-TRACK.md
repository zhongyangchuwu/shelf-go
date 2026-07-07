# Implementation Track: v0.2.0 Template Item Vault

## Phase 35: Vault Model and Validation Core

Goal: Implement the v2 in-memory model, built-in templates, validation, and runtime indexes without changing CLI behavior or persisted v1 vaults.

Requirements: V2-IMPL-01, V2-IMPL-02

Acceptance:

- New internal package owns v2 model types and validation.
- Built-in templates `generic_secret` and `api_credential` are available through a registry.
- Validation rejects duplicate item IDs, duplicate field IDs, unknown templates, missing required fields, invalid env bindings, and duplicate env names.
- Existing v1 tests still pass.

## Phase 36: JSON Codec and Repository Boundary

Goal: Implement v2 JSON encode/decode and a plaintext JSON repository behind the repository/codec boundary.

Requirements: V2-IMPL-03

Acceptance:

- v2 JSON round-trips without losing ordering or public/secret field metadata.
- Unknown fields and unsupported versions fail clearly.
- Repository load/save has atomic write and backup behavior consistent with current storage conventions.

## Phase 37: Legacy Migration Adapter

Goal: Convert v1 flat vault records and project manifests into v2 items and field bindings without deleting source data.

Requirements: V2-IMPL-04

Acceptance:

- v1 secrets convert to `generic_secret` items.
- v1 env/description/tags map to v2 field bindings/notes/labels.
- Project manifests convert to value-free item/field bindings.
- Migration validates output before replacing any target file.

## Phase 38: Secret CLI v2 Workflows

Goal: Add common non-interactive secret workflows over the v2 model.

Requirements: V2-IMPL-05

Acceptance:

- `create`, `set --stdin`, `get`, `copy`, `list --long`, and `info` contracts work with v2 items/fields.
- Value-free commands do not print secret field values.
- Existing script workflows have clear replacement paths.

## Phase 39: Project Manifest v2 Bindings

Goal: Implement project manifest v2 item/field env bindings and project export/run resolution.

Requirements: V2-IMPL-06

Acceptance:

- Project manifests remain value-free.
- Export/run materialize only explicitly bound secret fields.
- Missing item/field and duplicate env diagnostics are value-free.
- v1 manifests have a tested migration path.

## Phase 40: Manager Adaptation and Cutover

Goal: Adapt or temporarily gate the local manager for v2 items after CLI/core behavior is stable.

Requirements: V2-IMPL-07

Acceptance:

- Manager list/search remains value-free.
- Reveal/copy flows materialize secret fields only through explicit actions.
- If manager support is delayed, CLI output and docs clearly state the temporary limitation.
