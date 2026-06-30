# Plan: Vault Import Pivot

## Objective

Correct the architecture after the gopass source spike: keep Shelf's local vault as the single runtime secret store, move gopass to an import-only role, and rename packages so domain, persistence, import, and crypto boundaries are explicit.

## Scope

- Remove runtime `source.type: gopass` behavior.
- Move package names and imports from `internal/adapters/*` to semantic packages.
- Move `internal/crypto/age` into `internal/vaultcrypto` with age-specific exported names.
- Add `shelf vault import gopass` that imports gopass CLI entries into the local vault data model.
- Update architecture lint, docs, tests, and planning state.

## Package Migration

| Current | Target | Role |
| --- | --- | --- |
| `internal/adapters/shelfvault/model.go` | `internal/vault/model.go` | Shelf vault domain model |
| `path.go`, `validate.go`, `store.go`, `reader.go` | `internal/vault/` | domain rules, in-memory store, source reader |
| `age.go`, `json.go`, `vault.go`, `io.go`, `lock.go`, `status.go` | `internal/vaultfile/` | encrypted JSON file implementation |
| `internal/crypto/age/age.go` | `internal/vaultcrypto/age.go` | age encryption boundary |
| `internal/adapters/gopass/` | `internal/importer/gopass/` | gopass import client |

## Implementation Steps

1. Move vault domain files to `internal/vault` and update packages/imports.
2. Move encrypted JSON file files to `internal/vaultfile`; update references to domain types through `internal/vault`.
3. Move age crypto to `internal/vaultcrypto`; rename exports to `AgeIdentity`, `ReadOrCreateAgeIdentity`, `EncryptAge`, and `DecryptAge`.
4. Move gopass code to `internal/importer/gopass` and change it from `source.Reader` to CLI client methods: `ListFlat`, `ShowPassword`.
5. Remove config source selector and runtime gopass branch; project workflows load only local vault data.
6. Add app import service with all-or-nothing read before write.
7. Add `shelf vault import gopass` with `--prefix`, `--command`, `--force`, and `--dry-run`.
8. Update `.go-arch-lint.yml`, `docs/architecture.md`, `.planning/STATE.md`.

## Acceptance

- Project/run/manager/secret workflows all read local Shelf vault only.
- gopass is reachable only through import command code paths.
- Current encrypted JSON age vault behavior remains unchanged.
- gopass import writes Shelf vault secrets with JSON string values.
- Existing secrets are skipped unless `--force` is passed.
- `./scripts/test.sh` passes, including architecture lint.
