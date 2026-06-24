# Verification: Phase 10 Vault Restore and Recovery

## Claims Checked

- User can restore an encrypted vault backup through `shelf vault restore`.
- Restore refuses existing targets unless `--force` is supplied.
- Restore validates restore sources before replacing targets.
- Plaintext source and target paths are rejected.
- Docs explain restore usage, identity requirements, and `shelf vault status` verification.

## Evidence Observed

- `go test ./internal/cli -run TestVaultRestore` passed.
- `go test ./...` passed.
- `TestVaultRestoreEncryptedBackup` restores a `.bak` backup and verifies `secret get` returns the restored value.
- `TestVaultRestoreRefusesExistingTargetWithoutForce` proves the target remains unchanged without `--force`.
- `TestVaultRestoreRejectsPlaintextSource` rejects plaintext JSON and points to `shelf vault migrate`.
- `TestVaultRestoreRejectsInvalidEncryptedSource` rejects invalid encrypted-vault content before target replacement.
- `TestVaultRestoreRejectsPlaintextTarget` rejects plaintext target paths even with `--force`.
- Docs search found restore usage in README, getting started, reference, security, and troubleshooting docs.

## Coverage

- Successful restore.
- Existing target protection.
- Force restore behavior through `.bak` restore test.
- Invalid encrypted source rejection.
- Plaintext source and target rejection.
- Public recovery docs.

## Gaps

- No test for restoring with a different recipient set; current behavior re-encrypts to configured target recipients and is covered indirectly by the normal vault save/load path.

## Result

Passed.
