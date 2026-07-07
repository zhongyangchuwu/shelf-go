# Capture: Legacy Migration Adapter

## Durable Docs Updated

- None outside `.planning/`. User-facing docs should update when a public migration command or v2 CLI cutover becomes available.

## Planning Records Updated

- `.planning/ROADMAP.md` marks Phase 37 complete and Phase 38 executing.
- `.planning/REQUIREMENTS.md` marks `V2-IMPL-04` complete.
- `.planning/PROJECT.md` records the legacy migration adapter as validated project state.
- `.planning/STATE.md` advances to Phase 38.
- Phase-local `SUMMARY.md`, `REVIEW.md`, and `VERIFICATION.md` capture implementation evidence.

## Learnings

- v1 paths can be preserved as aliases without keeping them as the v2 product model.
- Deterministic hashed item IDs make migration idempotent without leaking old path text in IDs.
- Prefix/tag v1 manifest entries are safest as expanded item/field bindings in v2 manifests because v2 runtime should resolve explicit fields.
- Value-free project manifest conversion can be tested by marshaling output and checking fake secret values are absent.

## Ship Inputs

- Phase 38 can build v2 secret CLI workflows on `vaultmodel`, `vaultjson`, and `vaultmigrate` without touching current v1 command behavior until cutover is explicit.
- Any public migration command should report selector expansion warnings and never delete v1 sources automatically.
- Project manifest v2 runtime should reuse or intentionally replace the internal binding shape defined here.
