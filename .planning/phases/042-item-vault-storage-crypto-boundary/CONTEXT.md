# Context: Item Vault Storage and Crypto Boundary

## Goal

Prepare v0.2.0 for a formal release by replacing the temporary plaintext item-vault sidecar with a canonical encrypted Shelf vault file and by separating encryption from the item vault model and JSON codec.

## User Direction

- `setup` should offer an interactive encryption experience similar to gopass: choose an existing key or generate a new one.
- Age remains the preferred encryption mechanism.
- Encryption should be decoupled from the vault implementation.
- User-facing names should stop using temporary `v2` language.
- Before v0.2.0 release, prioritize CLI, secret model, and file/storage correctness.
- Web, TUI, and broader project workflow improvements come after this CLI/storage cutover.

## Current State

- `internal/vaultmodel` owns the item/field/template model.
- `internal/vaultjson` owns strict plaintext JSON codec/repository behavior.
- App-layer v2 secret/project/manager-gate workflows currently construct concrete `vaultjson.Repository` instances and use `<resolved-vault-path>.v2.json` by default.
- `internal/age` already owns age encryption/decryption helpers for the v1 vault path.
- `internal/vaultmigrate` was removed; automatic v1-to-item-vault migration is deferred.
- Security docs now warn that `*.v2.json` files are plaintext cutover artifacts and must not be synced.

## Constraints

- Do not overwrite or silently migrate existing v1 encrypted vaults.
- Do not commit or sync plaintext item-vault JSON.
- Keep `.shelf.json` value-free.
- Keep secret values entering through stdin only.
- Keep manager gated until an item-vault manager exists.
- Do not add Web/TUI/project feature expansion in this phase.

## Release Decision Frame

v0.2.0 should not be treated as the official portable item-vault release until the item vault is encrypted at rest through a canonical file path with non-temporary naming.
