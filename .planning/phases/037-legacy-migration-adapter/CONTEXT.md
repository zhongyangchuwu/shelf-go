# Context: Legacy Migration Adapter

## Goal

Implement migration adapters that convert v1 flat vault records and v1 project manifests into v2 template-based items and field bindings without deleting or rewriting source data before validation succeeds.

## Constraints

- Do not change current CLI behavior unless this phase explicitly adds an internal migration service only.
- Do not delete or mutate v1 source vaults/manifests during conversion.
- Use fake secret values only in tests.
- Convert all v1 secrets to `generic_secret` by default; do not infer `api_credential` from old paths.
- Preserve secret values, env bindings, descriptions, and tags.
- Project manifest conversion must remain value-free.
- Use `internal/vaultmodel` and `internal/vaultjson` from previous phases.

## Decisions

- Implement conversion as internal package code before adding CLI commands.
- Preserve old v1 paths as v2 aliases to support migration-period lookup.
- Map v1 `Secret.Env` to an env binding on the generated `value` field.
- Map v1 description to item notes and tags to labels.
- Manifest path entries convert to item/field/env bindings using the migration map.

## Open Questions

- Exact package name and API shape for the migration adapter.
- Deterministic item ID format for converted v1 paths.
- Whether prefix/tag manifest entries should expand immediately or remain selector-like in v2 project manifests.
- How to represent migration warnings for empty selectors and env conflicts.

## Verification Expectations

- Unit tests prove v1 secret conversion preserves values and metadata.
- Unit tests prove v1 project manifest path entries convert to value-free v2 bindings.
- Conversion validates v2 output before returning success.
- Existing v1 behavior remains unchanged.
