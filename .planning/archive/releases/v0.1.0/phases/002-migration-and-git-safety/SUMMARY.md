# Summary: Migration and Git Safety

## Completed Changes

- Added `store.DetectFileFormat` and file-format constants for missing, empty, encrypted vault, plaintext store, unsupported vault, and unsupported content.
- Added top-level `shelf migrate --from <plaintext.json> [--to <vault.age>] [--force]`.
- Migration loads the source through the legacy plaintext store reader, writes through `store.Vault.Save`, decrypts and validates the target, and checks the source bytes are unchanged before reporting success.
- Migration refuses an existing target by default, refuses plaintext JSON as a target, and uses existing encrypted backup behavior when `--force` replaces an encrypted vault.
- `shelf doctor` now reports vault format separately from vault loadability.
- `shelf doctor` fails plaintext JSON active vault paths with migration guidance.
- `shelf doctor` fails tracked plaintext secret stores and confirms tracked encrypted vaults.
- Added regression coverage for migration success, overwrite refusal, encrypted backup behavior, tracked plaintext detection, and tracked encrypted vault confirmation.

## Files Changed

- `internal/store/vault.go`
- `internal/cli/root.go`
- `internal/cli/migrate.go`
- `internal/cli/doctor.go`
- `internal/cli/doctor_test.go`
- `internal/cli/migrate_test.go`
- `.planning/STATE.md`
- `.planning/phases/002-migration-and-git-safety/CONTEXT.md`
- `.planning/phases/002-migration-and-git-safety/PLAN.md`

## Deviations

- Phase 2 was executed directly after creating context and plan because the user asked to continue development from current planning artifacts.
- Doctor git checks still use ordinary git tracking only; direct chezmoi command integration remains out of scope per roadmap.

## Evidence

- `go test ./internal/store ./internal/cli` passed.
- `go test ./...` passed.

Focused tests prove:

- Plaintext migration writes encrypted vault bytes and preserves source bytes.
- Migrated vault can be read by existing `shelf secret get` command.
- Existing targets are not changed without `--force`.
- Forced migration creates an encrypted `.bak` without known plaintext secret strings.
- Doctor fails active plaintext vault format.
- Doctor fails tracked plaintext secret stores.
- Doctor confirms tracked encrypted vault files.

## Unresolved Risks

- Full phase verification and capture are not yet recorded in `VERIFICATION.md` and `CAPTURE.md`.
- Doctor does not inspect arbitrary `.shelf.json` manifests beyond relying on the schema contract that manifests store paths and env names, not values.
