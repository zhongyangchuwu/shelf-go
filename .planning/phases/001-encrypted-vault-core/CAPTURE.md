# Capture

## Durable Docs Updated

- `README.md` status now records the encrypted vault core implementation.
- `docs/data-spec.md` now describes the age-encrypted vault as the durable source of truth and records lock/encrypt/write flow.
- `docs/usage-spec.md` now describes `shelf doctor` vault-health output instead of plaintext data-file output.
- `docs/codebase/ARCHITECTURE.md` now describes `store.Vault`, vault runtime loading, encrypted write flow, and vault-first anti-patterns.
- `docs/codebase/STACK.md` now lists `filippo.io/age`, default vault path, and `SHELF_VAULT`.
- `docs/codebase/INTEGRATIONS.md` now records encrypted vault storage and removes stale `SHELF_DATA`/`--data` docs.
- `docs/codebase/TESTING.md` now uses `--vault` examples and encrypted vault integration-test language.
- `docs/codebase/CONVENTIONS.md` now records validation through `store.Vault.Save`.
- `docs/codebase/CONCERNS.md` now marks the plaintext-at-rest core risk mitigated for vault mode and narrows remaining risks to migration, git safety, and editor temp files.

## Planning Records Updated

- `.planning/phases/001-encrypted-vault-core/VERIFICATION.md` created with Phase 1 claims, observed evidence, coverage, skipped checks, untested claims, and passed result.
- `.planning/phases/001-encrypted-vault-core/SUMMARY.md` updated to reference direct verification evidence and no remaining Phase 1 work.
- `.planning/REQUIREMENTS.md` marked VAULT-01 through VAULT-06 and CLI-01 complete.
- `.planning/ROADMAP.md` marked Phase 1 complete and updated coverage counts.
- `.planning/PROJECT.md` moved age-encrypted vault persistence into validated scope and updated age/portable-vault decisions.

## Learnings

- `store.Vault` is now the canonical CLI persistence boundary; command code should not call plaintext `store.Load`/`store.Save` for active secret workflows.
- The vault file format is `shelf-vault/v1\n` plus age ciphertext.
- Vault mode intentionally rejects plaintext JSON with a migration-oriented error instead of silently treating it as compatible.
- Phase 1 intentionally breaks plaintext-era `data`, `--data`, and `SHELF_DATA`; the active contract is `vault_path`, `--vault`, and `SHELF_VAULT`.
- Encrypted backups are part of the vault threat model; tests assert `.bak` bytes are encrypted and loadable.
- `secret edit` still uses a plaintext editor temp file; this is not a vault persistence leak, but remains a future hardening item.
- `doctor` validates encrypted vault loadability now, but does not yet prove git/chezmoi safety.

## Ship Inputs

- Summary:
  - Added age-backed encrypted vault persistence for core secret workflows.
  - Added vault-first config and CLI contract: `vault_path`, `recipients`, `identity_paths`, `--vault`, `SHELF_VAULT`.
  - Added actionable errors for plaintext JSON, missing/wrong identities, unsupported header, corrupt ciphertext, and invalid decrypted stores.
  - Updated docs and planning records to reflect the encrypted vault boundary.
- Verification:
  - `go test ./internal/config ./internal/store ./internal/cli`
  - `go test ./...`
  - `.planning/phases/001-encrypted-vault-core/VERIFICATION.md`
- Risks / waivers:
  - `secret edit` editor temp file remains plaintext by Phase 1 waiver.
  - Plaintext migration and git/chezmoi safety classification remain Phase 2.
  - Full export/project/run semantic hardening remains Phase 3, though the current suite passes.
