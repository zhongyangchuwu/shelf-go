# Capture: JSON Codec and Repository Boundary

## Durable Docs Updated

- None outside `.planning/`. User-facing docs should update when v2 storage becomes user-visible.

## Planning Records Updated

- `.planning/ROADMAP.md` marks Phase 36 complete and Phase 37 executing.
- `.planning/REQUIREMENTS.md` marks `V2-IMPL-03` complete.
- `.planning/PROJECT.md` records the v2 JSON repository boundary as validated project state.
- `.planning/STATE.md` advances to Phase 37.
- Phase-local `SUMMARY.md`, `REVIEW.md`, and `VERIFICATION.md` capture implementation evidence.

## Learnings

- Keeping v2 JSON storage in `internal/vaultjson` avoids confusion with the existing v1 `internal/jsonvault` package.
- Returning `os.ErrNotExist` for missing v2 vault files is safer than silently creating an empty vault during early migration/cutover work.
- Structural JSON round-trip tests are less brittle than full golden strings because Go struct field order is an implementation detail of the model layout.
- Reusing `util.AtomicWrite` keeps backup and atomic-save behavior consistent across v1 and v2 storage.

## Ship Inputs

- Phase 37 should use `vaultmodel.Validate` and `vaultjson.Repository` to validate converted v2 output before writing.
- Migration should keep v1 source files unchanged and treat `os.ErrNotExist` from v2 targets explicitly.
- Future encrypted repository work must preserve the no-plaintext-temp-file security direction from Phase 34.
