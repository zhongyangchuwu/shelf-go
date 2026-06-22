# Verification: Phase 6 Command Hierarchy Cutover

## Claims Checked

- CMD-01: `shelf setup` performs global onboarding.
- CMD-02: `shelf vault init` performs vault/config initialization.
- CMD-03: `shelf vault migrate` performs plaintext-to-encrypted migration.
- CMD-04: `shelf vault open` is wired under the vault command group.
- CMD-05: `shelf secret export` performs direct path/prefix export.
- CMD-06: `shelf project run -- ...` performs project runtime injection.
- CMD-07: old top-level `init`, `migrate`, `export`, `run`, and `manager` commands are absent.
- CMD-08: README and usage spec document the new hierarchy.
- VUX-01/VUX-02 partial: `shelf vault status`/`check` reports vault path, format, loadability, and migration guidance without revealing values.

## Evidence Observed

- Focused CLI tests passed: `go test ./internal/cli -run 'Test(Setup|Migrate|Export|Secret|Run|Root|Manager|Completion)'`.
- Focused vault tests passed: `go test ./internal/cli -run 'Test(Vault|Root|Manager|Migrate|Setup)'`.
- Full suite passed: `go test ./...`.
- `TestRootExcludesPreReleaseTopLevelCommands` asserts old root commands are absent.
- `TestVaultStatusReportsEncryptedLoadableVault` asserts encrypted vault status reports config, vault path, encrypted format, and loadability.
- `TestVaultStatusGuidesPlaintextMigration` asserts plaintext vault status recommends `shelf vault migrate`.
- Documentation search found new canonical command forms in README and usage spec.

## Gaps

- Project session activation/deactivation remains design-only and unimplemented by request.
- Vault UX can still be expanded in a later pass for richer recovery messages and backup restore flows.

## Result

Phase 6 verification passed. The command hierarchy cutover is implemented and tested.
