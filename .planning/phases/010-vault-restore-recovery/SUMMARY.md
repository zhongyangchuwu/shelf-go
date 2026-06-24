# Summary: Phase 10 Vault Restore and Recovery

## Completed Changes

- Added `shelf vault restore --from <backup.age> [--to <vault.age>] [--force]`.
- Restore accepts encrypted Shelf vault sources only, decrypts and validates the source, writes the target through encrypted vault save, and verifies the restored target loads.
- Restore refuses existing targets without `--force`.
- Restore rejects plaintext JSON sources and plaintext JSON targets.
- Updated README, getting started, reference, troubleshooting, and security docs with restore usage, identity requirements, and post-restore `shelf vault status` verification.

## Files Changed

- `internal/cli/manager.go`
- `internal/cli/restore.go`
- `internal/cli/restore_test.go`
- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/security.md`
- `docs/troubleshooting.md`
- `.planning/phases/010-vault-restore-recovery/CONTEXT.md`
- `.planning/phases/010-vault-restore-recovery/PLAN.md`

## Deviations

- None. Restore does not support plaintext JSON sources; migration remains the plaintext conversion path.

## Evidence

- `go test ./internal/cli -run TestVaultRestore` passed.
- `go test ./...` passed.
- Docs search found restore references in README, getting started, reference, security, troubleshooting, and Phase 10 planning artifacts.

## Unresolved Risks

- If all matching private age identities are lost, Shelf cannot recover encrypted vaults or backups. Docs now state this explicitly.
