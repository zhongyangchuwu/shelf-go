# Plan: Phase 10 Vault Restore and Recovery

## Objective

Add an explicit vault recovery path for encrypted backups.

## Scope

In scope:

- Add `shelf vault restore --from <backup.age> [--to <vault.age>] [--force]`.
- Validate restore source format and decryptability before target replacement.
- Re-encrypt restored store to the configured target recipients.
- Update recovery/security docs.
- Add focused CLI tests.

Out of scope:

- Restoring plaintext JSON stores; `vault migrate` owns that path.
- Backup listing or automatic backup discovery.
- Recipient inspection or rekey commands.
- Git merge/conflict handling.

## Tasks

1. Implement restore command.
   - Register under `newVaultCmd`.
   - Parse `--from`, optional `--to`, and `--force`.
   - Build a source vault using the source path and current config identities/recipients.
   - Build a target vault from active config or `--to`.

2. Implement restore helper.
   - Require source format `encrypted-vault`.
   - Load/decrypt/validate source store.
   - Refuse existing non-empty target unless `--force`.
   - Reject plaintext JSON target even with force.
   - Save target with configured recipients and verify load + secret count.

3. Update docs.
   - Add restore command to README/reference.
   - Add troubleshooting recovery section.
   - Add security notes for identity loss and encrypted backups.

4. Verify.
   - Add restore success and failure tests.
   - Run focused vault restore tests.
   - Run `go test ./...`.

## Acceptance Criteria

- `shelf vault restore --from backup.age` restores to the active vault path when the target is missing or `--force` is supplied.
- Restore refuses existing targets without `--force`.
- Restore rejects plaintext sources and invalid/undecryptable encrypted sources.
- Restore rejects plaintext targets to avoid plaintext `.bak` preservation.
- Docs explain identity requirements and post-restore validation.

## Verification

- `go test ./internal/cli -run TestVaultRestore`
- `go test ./...`

## Risks

- `Vault.Save` creates `.bak` on overwrite; rejecting plaintext targets prevents preserving a plaintext `.bak` during restore.
- Source backup may require an identity no longer configured; restore must report the decrypt failure instead of replacing anything.
