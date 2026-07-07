# Context: Vault Model and Validation Core

## Goal

Implement the v2 in-memory vault model and validation core described by Phase 34 without changing current v1 CLI behavior, current v1 vault persistence, project manifests, migration commands, or manager behavior.

## Constraints

- Keep the new model behind a new internal package boundary until storage and CLI cutover phases are ready.
- Do not change `internal/vault`, `internal/jsonvault`, `internal/project`, `internal/app`, `internal/cli`, or manager behavior in this phase except for dependency-boundary configuration if needed.
- Do not implement JSON repository, encryption, migration, CLI commands, project manifest v2, or manager UI in this phase.
- Secret values in tests may use fake literals only.
- Use Go standard library only unless a concrete implementation need appears.
- Preserve existing architecture direction: adapters depend inward; core model packages do not depend on CLI/app/manager.

## Decisions

- Create a new package `internal/vaultmodel` for v2 model types, built-in templates, validation, and runtime indexes.
- Keep v1 package `internal/vault` untouched so current command behavior remains stable while v2 work lands incrementally.
- Use string-backed constants for field types, sensitivity, purpose, and binding types so JSON representation stays explicit.
- Implement validation as a pure function over `Vault` plus built-in template registry.
- Implement indexes after validation for item ID, title, alias, and env binding lookup.

## Open Questions

- Exact generated item ID format is deferred; validation only requires non-empty unique IDs in this phase.
- Alias lifetime after migration remains deferred to migration/cutover phases.
- Custom template loading remains deferred; this phase only supports built-in templates.

## Verification Expectations

- Focused package tests prove valid built-in-template items and validation failure cases.
- Existing v1 vault/project behavior remains unaffected by passing existing relevant tests.
- Architecture lint either ignores the new core package safely or is updated to represent the new boundary.
