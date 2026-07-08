# Capture: Project Manifest v2 Bindings

## Durable Docs Updated

- `README.md` now describes version 1 vs version 2 `.shelf.json` manifest contents.
- `docs/reference.md` now documents v2 project add syntax, v2 manifest JSON, and shared manifest rules.
- `docs/getting-started.md` now distinguishes v1 selector manifests from v2 concrete item/field bindings.
- `docs/architecture.md` now describes v1 store-backed resolution and v2 vaultmodel-backed resolution.
- `docs/portable-vault.md` now describes `.shelf.json` as value-free for both legacy paths and v2 item/field/env bindings.
- `docs/troubleshooting.md` now covers v2 missing binding and env conflict cases.
- `CHANGELOG.md` records v2 project binding support.

## Planning Records Updated

- `.planning/PROJECT.md` records Phase 39 as validated and moves active work to manager adaptation/gating.
- `.planning/REQUIREMENTS.md` marks `V2-IMPL-06` complete.
- `.planning/ROADMAP.md` marks Phase 39 complete.
- `.planning/STATE.md` advances to Phase 40.
- Phase-local `SUMMARY.md`, `REVIEW.md`, and `VERIFICATION.md` capture implementation and evidence.

## Learnings

- Concrete v2 project bindings keep runtime diagnostics simple and value-free: item/field/env records can fail without expanding selectors or inspecting unrelated fields.
- Keeping `project init` on v1 during cutover avoids breaking established encrypted-vault workflows while v2 storage is still plaintext JSON.
- A custom manifest marshaler prevents empty v1/v2 manifests from losing their required array field on save/load round trips.
- Project-level duplicate env conflict detection remains the right place for env collisions because the v2 vault intentionally permits duplicate env-name public field values.

## Ship Inputs

- Phase 40 must decide whether the local manager is adapted to v2 items now or gated to avoid editing the wrong vault model.
- Future migration/cutover work should provide a public path from v1 `.shelf.json` entries to v2 concrete bindings.
- V2 repository age wrapping remains required before v2 vault files are safe to sync or commit.
