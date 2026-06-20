# Verification: Migration and Git Safety

## Claims Checked

- MIGR-01: User can migrate an existing plaintext Shelf JSON store into an age-encrypted vault.
- MIGR-02: Migration preserves the original plaintext source until the encrypted target decrypts and validates successfully.
- MIGR-03: Migration reports clear next steps for moving, deleting, or archiving the old plaintext store.
- MIGR-04: Shelf creates backups or recovery artifacts in encrypted form when secret values are involved.
- MIGR-05: Shelf avoids writing plaintext secret values to durable temp files during normal store operations.
- SAFE-01: User can keep the encrypted vault as a normal portable file suitable for git-backed dotfile workflows such as chezmoi.
- SAFE-02: Shelf config remains non-secret and can be reviewed or committed without exposing secret values or private identities.
- SAFE-03: `shelf doctor` reports whether the active store is plaintext or encrypted.
- SAFE-04: `shelf doctor` warns when plaintext secret storage appears to be tracked by git.
- SAFE-05: `shelf doctor` confirms when the tracked/synced vault path is encrypted and value-free project manifests remain safe.

## Evidence Observed

- `go test ./internal/store ./internal/cli` passed.
- `go test ./...` passed.
- `TestMigratePlaintextStoreToEncryptedVault` proves `shelf migrate --from` creates encrypted target bytes, preserves source bytes, emits cleanup guidance, and the migrated value is readable through `shelf secret get`.
- `TestMigrateRefusesExistingTargetWithoutForce` proves migration does not replace an existing vault unless `--force` is supplied.
- `TestMigrateForceCreatesEncryptedBackup` proves forced replacement creates `.bak` bytes without the known old secret, new secret, or secret path strings and the target decrypts to the new value.
- `TestDoctorReportsHealthyStore` proves doctor reports `ok vault format` and `ok vault loads` for encrypted vaults.
- `TestDoctorFailsInvalidStore` proves doctor fails unsupported active vault content at the format gate.
- `TestDoctorFailsTrackedPlaintextStore` proves tracked plaintext JSON is reported as both unsafe format and unsafe git tracking.
- `TestDoctorConfirmsTrackedEncryptedVault` proves tracked encrypted vault files are confirmed as encrypted.

## Coverage

| Requirement | Evidence | Result |
|-------------|----------|--------|
| MIGR-01 | Migration command test creates encrypted target and reads it through existing CLI read path. | Pass |
| MIGR-02 | Migration test compares source bytes before and after success; refusal test compares target bytes before and after failure. | Pass |
| MIGR-03 | Migration test checks output includes plaintext-source preservation guidance. | Pass |
| MIGR-04 | Forced migration test checks replacement backup does not contain known plaintext secret data. | Pass |
| MIGR-05 | Migration writes through `store.Vault.Save`; tests inspect target and backup bytes for absence of known plaintext secret strings. | Pass |
| SAFE-01 | Tracked encrypted vault test confirms ordinary git can track the encrypted vault path and doctor reports it as encrypted. | Pass |
| SAFE-02 | Config flow remains `vault_path`, public recipients, and identity paths only; migration writes no config or manifest secret values. | Pass |
| SAFE-03 | Doctor tests cover encrypted, plaintext, and unsupported active store formats. | Pass |
| SAFE-04 | Tracked plaintext test fails git tracking with `tracked plaintext secret store is unsafe`. | Pass |
| SAFE-05 | Tracked encrypted test reports `tracked vault is encrypted`; manifests remain value-free by schema and docs. | Pass |

## Gaps

- Doctor does not run chezmoi-specific commands; ordinary git tracking is the implemented safety signal for Phase 2.
- Doctor does not scan every project manifest in a repository; it relies on `.shelf.json` validation prohibiting `value`, fallback plaintext, shell commands, and template expressions.
- Tests assert absence of representative known plaintext strings in encrypted artifacts; they do not attempt cryptanalysis of age ciphertext.

## Result

Phase 2 verification passed. Migration and git-safety requirements MIGR-01 through MIGR-05 and SAFE-01 through SAFE-05 have direct automated evidence or schema-backed evidence with documented gaps.
