# Context: Item Vault Plain JSON Cutover

## Goal

Cut Shelf Go to the personal-use item-vault CLI: remove built-in age encryption, v1 encrypted vault compatibility, v1 manager/import/migration paths, and temporary `v2` storage naming. Keep local `shelf secret`, `shelf project`, `shelf setup`, and `shelf doctor` working over a configured JSON item vault.

## User Direction

- Do not preserve v1 compatibility; this is a personal project.
- Remove age-encryption-related code for now.
- Chezmoi owns encryption/sync outside Shelf.
- Ensure the local Shelf experience stays normal on this workstation.
- Prioritize CLI, secret model, and file/storage correctness before Web/TUI/project expansion.

## Current Target

- Canonical vault path: `~/.local/share/shelf/vault.json` unless config/flag/env override.
- Config contains `vault_path` only; no recipients or identity paths.
- Item vault JSON is plaintext and must be protected externally.
- Active CLI commands: setup, doctor, secret, project, completion.
- Removed active surfaces: age package, jsonvault/v1 vault packages, manager, import, migrate, v1 export, project configure TUI, v1 project selectors.

## Constraints

- Do not print secret values in diagnostics, list/info/status, completions, or planning artifacts.
- Keep `.shelf.json` value-free.
- Secret values enter through stdin only.
- Local migration/cutover must back up existing local config and vault JSON first.
