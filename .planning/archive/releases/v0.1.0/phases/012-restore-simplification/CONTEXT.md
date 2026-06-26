# Context: Phase 12 Restore Simplification

## Goal

Remove `shelf vault restore` because current backups are single-slot encrypted files and manual copy plus `shelf vault status` is the simpler recovery model.

## Constraints

- Keep vault recovery minimal.
- Do not add history, backup rotation, Dolt, database storage, or restore abstractions.
- Preserve encrypted `.bak` generation for last-write safety.
- Document `.bak` as a single-slot last-write backup, not a history system.

## Decisions

- Delete `shelf vault restore` command and tests.
- Recovery docs should show explicit file copy/rename workflow.
- `shelf vault status` and `shelf doctor` remain the verification commands after manual recovery.
- Future history/restore commands require a separate product decision.

## Open Questions

None.

## Verification Expectations

- `shelf vault restore` is absent from command registration and docs.
- Docs explain manual `.bak` recovery and single-slot behavior.
- Tests pass after removal.
