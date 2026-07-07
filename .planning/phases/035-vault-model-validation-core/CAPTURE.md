# Capture: Vault Model and Validation Core

## Durable Docs Updated

- None outside `.planning/`. User-facing docs should update when v2 storage or CLI behavior becomes user-visible.

## Planning Records Updated

- `.planning/ROADMAP.md` marks Phase 35 complete and leaves Phase 36 as the next implementation phase.
- `.planning/REQUIREMENTS.md` marks `V2-IMPL-01..V2-IMPL-02` complete.
- `.planning/PROJECT.md` records the v2 model/validation core as validated project state.
- `.planning/STATE.md` advances to Phase 36.
- Phase-local `SUMMARY.md`, `REVIEW.md`, and `VERIFICATION.md` capture implementation evidence.

## Learnings

- The v2 core model can be added incrementally without changing v1 vault persistence or CLI behavior.
- Architecture lint is useful during phased work: it caught an existing CLI-to-project dependency while adding the new core package.
- `internal/app` is the right projection boundary for TUI state that would otherwise force CLI to import domain packages directly.
- Built-in template validation should stay strict for known fields, while custom template behavior remains a later decision.

## Ship Inputs

- Phase 36 should build the v2 JSON codec and repository against `internal/vaultmodel` without reworking CLI behavior.
- Phase 36 should decide strict JSON decoding behavior for unknown fields and unsupported versions.
- Migration and CLI phases should depend on `Validate` and `NewIndex` rather than duplicating model checks.
