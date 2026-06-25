# Capture: Phase 12 Restore Simplification

## Durable Docs Updated

- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/security.md`
- `docs/troubleshooting.md`

## Planning Records Updated

- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/012-restore-simplification/CONTEXT.md`
- `.planning/phases/012-restore-simplification/PLAN.md`
- `.planning/phases/012-restore-simplification/SUMMARY.md`
- `.planning/phases/012-restore-simplification/VERIFICATION.md`

## Learnings

- A dedicated restore command is not justified while backups are single-slot encrypted files.
- Manual copy plus `shelf vault status` is simpler and keeps the portable-file model obvious.
- Rich history/restore would need a separate phase with timestamped encrypted snapshots or another explicit history design.

## Ship Inputs

- `go test ./...` passed.
- Public docs no longer advertise `shelf vault restore`.
