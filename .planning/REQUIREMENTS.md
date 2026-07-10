# Requirements: Shelf Go

## Current v0.2.0 Requirements

- [x] **ITEM-01**: Shelf stores secrets as template-based items with fields, public env-name fields, labels, aliases, profiles, validation, and value-free metadata output.
- [x] **ITEM-02**: Shelf persists the item vault as strict JSON with validation, atomic writes, and backups.
- [x] **ITEM-03**: Shelf uses the configured vault path directly and defaults to `~/.local/share/shelf/vault.json`.
- [x] **ITEM-04**: Built-in age encryption, v1 encrypted vault code, v1 manager, import/migrate, and v1 export paths are removed from active product scope.
- [x] **CLI-01**: `shelf setup` creates config and an empty item vault.
- [x] **CLI-02**: `shelf doctor` validates config, permissions, vault loadability, and shell completion without printing values.
- [x] **CLI-03**: `shelf secret create`, `set`, `get`, `list`, `info`, `export`, and `rm` operate over item vault JSON.
- [x] **CLI-04**: Secret writes read values from stdin, not argv.
- [x] **CLI-05**: Label-scoped export materializes only matching items with explicit env-name fields and keeps failure diagnostics value-free.
- [x] **CLI-06**: Human-facing secret list/create output is centralized through `internal/cliui`.
- [x] **PROJECT-01**: `.shelf.json` stores concrete item/field/env bindings only.
- [x] **PROJECT-02**: `project init/add/rm/list/status/export/run` work over item-vault bindings.
- [x] **PROJECT-03**: Project manifests stay value-free and project diagnostics do not print secret values.
- [x] **LOCAL-01**: Local workstation vault was backed up and cut over to `/home/han/.local/share/shelf/vault.json`.

## Deferred Requirements

- Built-in encryption if external encryption becomes insufficient.
- Optional reviewed import/migration tooling only if a real local need appears.
- Web manager.
- TUI full-object editing.
- Project configure TUI.
- Team/hosted sync workflows.
- Public repository/release packaging for a future v1.0.0 milestone.

## Traceability

| Requirement | Phase | Status |
| --- | --- | --- |
| ITEM-01..ITEM-03 | Phases 35, 36, 38, 42 | Complete |
| ITEM-04 | Phase 42 | Complete |
| CLI-01..CLI-06 | Phases 42 and local CLI polish | Complete |
| PROJECT-01..PROJECT-03 | Phases 39, 42 | Complete |
| LOCAL-01 | Phase 42 | Complete |

---
*Last updated: 2026-07-10 during local v0.2.0 wrap-up*
