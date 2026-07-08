# Capture: Manager Adaptation and Cutover

## Durable Knowledge

- The current local manager is a v1 path/env/tag editor.
- V2 vaults use template items and fields; a correct v2 manager needs a different editing model.
- Phase 40 explicitly gates the v1 manager when a valid v2 vault is present instead of adapting the UI prematurely.
- The gate belongs in `internal/app` before `LoadVault`, not in `internal/manager`, because manager should remain a narrow surface independent of app/vaultmodel/vaultjson.

## Docs Updated

- `README.md`: manager section and safety notes now describe the v1 manager and v2 gate.
- `docs/getting-started.md`: manager section now identifies v1 behavior and CLI fallback for v2.
- `docs/reference.md`: `shelf manager` now documents the v2 cutover gate.
- `docs/troubleshooting.md`: added a v2 manager gate troubleshooting section.
- `docs/architecture.md`: runtime construction and local manager sections now describe app-level v2 gating.
- `docs/portable-vault.md`: v2 JSON vault row notes the manager gate.
- `CHANGELOG.md`: unreleased breaking changes include the v2 manager gate.

## Planning Updated

- `.planning/PROJECT.md`: records Phase 40 as validated and marks v0.2.0 implementation track ready for milestone audit.
- `.planning/REQUIREMENTS.md`: marks `V2-IMPL-07` complete.
- `.planning/ROADMAP.md`: marks Phase 40 complete.
- `.planning/STATE.md`: advances to v0.2.0 implementation-track audit readiness.

## Follow-Up Inputs

- Audit the v0.2.0 track before release/cutover decisions.
- Decide whether the next milestone should prioritize v2 age wrapping, public migration/cutover commands, or a real v2 manager/editor.
- Keep the v2 manager future design item/field-first; do not revive the raw v1 path editor for v2 items.
