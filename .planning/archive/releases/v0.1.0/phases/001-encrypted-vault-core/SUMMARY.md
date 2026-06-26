# Phase 1 Summary: Encrypted Vault Core

## Status

Execution continued with a breaking vault-first refactor. The current code removes plaintext `data` compatibility from the CLI/config contract, introduces a dedicated `store.Vault` persistence type, and routes secret writes through unified vault update helpers.

## Implemented

- Added `filippo.io/age` dependency for age encryption.
- Replaced plaintext-era config naming with vault-first fields:
  - `vault_path`
  - `recipients`
  - `identity_paths`
- Replaced CLI store override:
  - removed `--data`
  - added `--vault`
- Replaced `SHELF_DATA` with `SHELF_VAULT` in config resolution.
- Added `store.Vault` and `store.VaultOptions`:
  - `NewVault`
  - `Vault.Load`
  - `Vault.Save`
  - `Vault.Lock`
  - `Vault.Read`
  - `Vault.Update`
- Simplified `store.Store` back toward the plaintext in-memory model.
- Kept plaintext `store.Load` / `store.Save(path, st)` for legacy/internal plaintext readers, but CLI runtime uses `store.Vault`.
- Added Shelf vault header:
  - `shelf-vault/v1\n`
- Added age recipient parsing and identity-file parsing.
- Added actionable error paths for:
  - plaintext JSON used as an encrypted vault
  - missing identity paths
  - wrong/no matching identity
  - unsupported vault header
  - invalid decrypted store JSON
- Wired `internal/cli/root.go` to build a `store.Vault` from runtime config.
- Routed mutating secret commands through `updateVault` / `Vault.Update`.
- Updated `shelf init` to create vault-first config and an encrypted empty vault.
- `shelf init --force` preserves an existing vault instead of resetting secret contents.
- Updated `shelf doctor` language to vault-oriented checks and added encrypted vault load validation.
- Added config YAML-template resolution test.
- Updated CLI tests to exercise encrypted vault-backed commands by default through `--vault`.

## Files Changed

- `go.mod`
- `go.sum`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/store/io.go`
- `internal/store/vault.go`
- `internal/store/vault_test.go`
- `internal/cli/root.go`
- `internal/cli/init.go`
- `internal/cli/init_test.go`
- `internal/cli/doctor.go`
- `internal/cli/doctor_test.go`
- `internal/cli/secret.go`
- `internal/cli/secret_test.go`
- `internal/cli/test_helpers_test.go`
- Other CLI tests updated from `--data` to `--vault`.
- `.planning/STATE.md`
- `.planning/phases/001-encrypted-vault-core/PLAN.md`

## Verification Evidence

- `go test ./internal/config ./internal/store ./internal/cli` passed after verification coverage updates.
- `go test ./...` passed after verification coverage updates.
- `VERIFICATION.md` records direct evidence for every Phase 1 success criterion.

Focused tests cover:

- vault save/load encrypts store bytes;
- encrypted replacement backup is loadable and does not contain plaintext secret values;
- vault mode rejects legacy plaintext JSON;
- missing and wrong identity errors are actionable;
- unsupported vault header, corrupt ciphertext, and invalid decrypted store errors are actionable;
- YAML template config resolution for `vault_path`, recipients, identity paths, and editor;
- CLI `secret set/get/list/info/edit/rm` works through encrypted vault config;
- `shelf init` creates reusable vault config and preserves existing vault contents on force.

## Deviations

- The earlier Phase 1 compatibility decision for `data`, `--data`, and `SHELF_DATA` was intentionally superseded by the user's breaking-change direction on 2026-06-18.
- `secret edit` editor temp-file hardening remains deferred per CONTEXT D-12. Vault persistence temp files are encrypted; editor temp files are not changed here.
- `shelf doctor` now validates encrypted vault loading, but full git/chezmoi safety classification remains Phase 2 scope.

## Remaining Phase 1 Work

- None. Phase 2 owns plaintext migration and git/chezmoi safety classification.
