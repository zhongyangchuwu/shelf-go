# Summary: Phase 12 Restore Simplification

## Completed Changes

- Removed the `shelf vault restore` command and restore tests.
- Removed restore command registration from `shelf vault`.
- Updated public docs to describe manual recovery from the single last-write encrypted `.bak` file.
- Updated planning records to reflect minimal recovery instead of a dedicated restore command.

## Files Changed

- `internal/cli/manager.go`
- `internal/cli/restore.go` removed
- `internal/cli/restore_test.go` removed
- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/security.md`
- `docs/troubleshooting.md`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/012-restore-simplification/CONTEXT.md`
- `.planning/phases/012-restore-simplification/PLAN.md`

## Deviations

- None. Recovery is now explicitly manual and minimal.

## Evidence

- Pending final verification in `VERIFICATION.md`.

## Unresolved Risks

- Manual recovery has fewer pre-copy guardrails than the removed command. `shelf vault status` remains the post-copy verification step.
