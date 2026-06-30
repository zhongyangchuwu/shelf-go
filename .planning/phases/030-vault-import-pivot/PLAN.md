# Plan: Vault Import Pivot

## Objective

Correct the architecture after the gopass source spike: keep Shelf's local vault as the single runtime secret store, move gopass to an import-only role, and rename packages so domain, persistence, import, and crypto boundaries are explicit.

## Scope

- Remove runtime `source.type: gopass` behavior.
- Move package names and imports from `internal/adapters/*` to semantic packages.
- Add `shelf vault import gopass` that imports gopass CLI entries into the local vault data model.
- Update architecture lint, docs, tests, and planning state.

## Final Package Layout

| Package | Role |
| --- | --- |
| `internal/vault` | Shelf vault domain model, path/env/tag rules, in-memory store, source reader |
| `internal/jsonvault` | Current `shelf-vault/v1` encrypted JSON repository implementation |
| `internal/age` | age encryption/decryption and identity helpers |
| `internal/importer/gopass` | gopass import client |

## Implementation Steps

1. Move vault domain files to `internal/vault` and update packages/imports.
2. Move encrypted JSON repository files to `internal/jsonvault`; update references to domain types through `internal/vault`.
3. Move age helpers to `internal/age` with algorithm-local names: `Identity`, `ReadOrCreateIdentity`, `Encrypt`, and `Decrypt`.
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
