# Shelf Go

## What This Is

Shelf Go is a personal, local, CLI-first secret manager. It stores item-based secrets in a local JSON vault and relies on external tooling such as chezmoi for encryption/sync.

## Current State

### Validated

- [x] v0.2.0 item model, JSON repository, secret CLI, label export, project item bindings, and item-vault run/export workflows are implemented.
- [x] Built-in age/v1 compatibility, v1 manager, v1 migration/import/export, and project configure TUI were removed for the personal-use CLI direction.
- [x] Local workstation config now points at `/home/han/.local/share/shelf/vault.json`, copied from the previous item sidecar after backup.

### Active

- [ ] Harden release docs and verify the item-vault-only CLI before tagging v0.2.0.

### Out of Scope

- Built-in encryption; use chezmoi or another external tool.
- v1 compatibility and automatic migration.
- Web manager and TUI editing.
- Team sharing, hosted sync, browser autofill, or daemon unlock flows.

## Constraints

- The vault JSON contains plaintext secret values; do not commit/sync it unless external tooling encrypts it.
- `.shelf.json` and config must remain value-free.
- Secret values enter through stdin for writes.
- Diagnostics, list/info/status, and completions must not print secret values.

## Package Boundaries

- `internal/cli`: Cobra command surface.
- `internal/cliui`: Lip Gloss-backed human output helpers for titles, summaries, empty states, and tables.
- `internal/app`: application workflows.
- `internal/vaultmodel`: item/field/template/profile model and validation.
- `internal/vaultjson`: JSON item-vault codec/repository.
- `internal/project`: Git project identity, manifest schema, binding resolution, child env helpers.
- `internal/config`: config and vault path resolution.
- `internal/util`: atomic write and env/shell/JSON binding formatting.

## Key Decisions

| Decision | Rationale | Outcome |
| --- | --- | --- |
| Use external encryption | User prefers chezmoi encryption and a simpler personal tool. | Removed age/jsonvault/v1 encryption code from active product. |
| Use item vault JSON as canonical file | It is easy to inspect, back up, and encrypt externally. | Default vault path is `~/.local/share/shelf/vault.json`. |
| Drop v1 compatibility | Personal-use project; compatibility burden slows release. | Removed v1 vault, manager, import/migrate/export, project selectors. |
| Keep CLI before Web/TUI | CLI is the daily workflow and easiest to verify. | Active commands are setup, doctor, secret, project, completion. |

---
*Last updated: 2026-07-10 after item-vault-only cutover planning and implementation*
