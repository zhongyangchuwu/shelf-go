# Plan: Vault Naming Cleanup

## Objective

Make the package graph names match actual responsibilities after the vault import pivot.

## Decisions

- Keep `internal/vault` as the Shelf vault domain model: data types, path/env/tag rules, in-memory store.
- Rename `internal/vaultfile` to `internal/jsonvault` because the package is the current encrypted JSON vault implementation, not every possible file backend.
- Rename `internal/vaultcrypto` to `internal/age` because the code is an age algorithm helper and does not need vault-specific state.
- Hide direct age usage from `internal/app`; setup should call through `jsonvault` so app depends on the current vault implementation, not the encryption helper.

## Target Package Roles

| Package | Role |
| --- | --- |
| `internal/vault` | Shelf vault domain model and store rules |
| `internal/jsonvault` | Current shelf-vault/v1 encrypted JSON repository |
| `internal/age` | age encryption/decryption and identity helpers |
| `internal/importer/gopass` | gopass CLI import client |

## Implementation Steps

1. Move `internal/vaultfile` to `internal/jsonvault` and rename package declarations/imports.
2. Move `internal/vaultcrypto` to `internal/age`; simplify exported names from `EncryptAge`/`DecryptAge`/`AgeIdentity` to `Encrypt`/`Decrypt`/`Identity`.
3. Add `jsonvault.ReadOrCreateAgeIdentity` wrapper so `app` has no direct dependency on `internal/age`.
4. Update `.go-arch-lint.yml` and `docs/architecture.md`.
5. Run focused tests and `./scripts/test.sh`.

## Acceptance

- Architecture graph names show `vault`, `jsonvault`, and `age` with clear roles.
- `app` no longer depends directly on age.
- Existing encrypted JSON vault behavior is unchanged.
- `./scripts/test.sh` passes.
