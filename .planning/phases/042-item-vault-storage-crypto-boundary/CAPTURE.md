# Capture: Item Vault Plain JSON Cutover

## Durable Knowledge

- Shelf no longer owns encryption. Chezmoi or another external tool encrypts/syncs the vault file.
- The active vault file is plaintext JSON and contains secret values.
- The canonical default vault path is `~/.local/share/shelf/vault.json`.
- The active product is CLI-only: setup, doctor, secret, project, completion.
- Web manager, TUI editing, project configure TUI, v1 compatibility, and automatic migration are deferred.

## Code Shape

- Removed active age/v1 package stack.
- Removed manager/import/migrate/v1 export surfaces.
- Kept `vaultmodel` and `vaultjson` as model and repository boundaries.
- Project manifests are item-field bindings only.

## Local Cutover

- Backup: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- New local vault: `/home/han/.local/share/shelf/vault.json`.
- Config: `/home/han/.config/shelf/config.yaml` now points at the JSON vault.

## Next Work

1. Local v0.2.0 wrap-up: code/framework review, docs/planning cleanup, and file hygiene.
2. Installed-binary smoke tests for local workflows.
3. Future v1.0.0: public repository/release packaging if the project is made public.
4. Later product work: Web manager, TUI editing, project configure TUI, optional built-in encryption if needed.
