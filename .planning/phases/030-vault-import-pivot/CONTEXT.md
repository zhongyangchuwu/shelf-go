# Context: Vault Import Pivot

## Problem

The previous gopass read-source MVP made gopass a runtime project source. That creates two facts of secret truth: project/run can read gopass while secret CRUD and manager still read the local Shelf vault.

## Decision

Shelf's product model is a local encrypted vault. External systems such as gopass are import sources, not runtime backends.

## Target Model

- `internal/vault`: Shelf vault domain model, path/env/tag validation, in-memory store, and `source.Reader` adapter over a local store.
- `internal/vaultfile`: current encrypted JSON file implementation, including file format detection, JSON encoding, file locking, and save/load/update orchestration.
- `internal/vaultcrypto`: vault encryption boundary. Current functions are age-specific and named with `Age` to leave room for GPG.
- `internal/importer/gopass`: gopass CLI import client. It is not a `source.Reader` and is not selected at runtime by project workflows.

## Import Semantics

- Command target: `shelf vault import gopass`.
- gopass path `a/b/c` maps to Shelf path `a/b:c` by turning the last slash into a colon.
- gopass entries without a slash are skipped as unmappable.
- Imported values are always JSON strings, even if the text looks like JSON.
- Existing Shelf secrets are skipped unless `--force` is passed.
- Read failures abort before writing, so import does not partially update the vault.
- Metadata (`env`, `description`, `tags`) is not imported in the MVP.

## Out of Scope

- gopass as a runtime backend/source.
- writing back to gopass.
- multi-source project resolution.
- SQL/NoSQL vault implementations; the refactor only makes current JSON file persistence distinct from the domain model.
