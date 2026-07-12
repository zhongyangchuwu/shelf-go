# Requirements: Shelf Go

## Current v0.2.0 Requirements

- [x] **ITEM-01**: Shelf stores secrets as template-based items with fields, public env-name fields, labels, aliases, profiles, validation, and value-free metadata output.
- [x] **ITEM-02**: Shelf persists the item vault as strict JSON with validation, atomic writes, and backups.
- [x] **ITEM-03**: Shelf uses the configured vault path directly and defaults to `~/.local/share/shelf/vault.json`.
- [x] **ITEM-04**: Built-in age encryption, v1 encrypted vault code, v1 manager, import/migrate, and v1 export paths are removed from active product scope.
- [x] **CLI-01**: `shelf setup` creates config and an empty item vault.
- [x] **CLI-02**: `shelf doctor` validates config, permissions, vault loadability, and shell completion without printing values.
- [x] **CLI-03**: `shelf secret create`, `update`, `get`, `list`, `info`, `export`, and `rm` operate over item vault JSON.
- [x] **CLI-04**: Secret writes read values from stdin, not argv.
- [x] **CLI-05**: Label-scoped export materializes only matching items with explicit env-name fields and keeps failure diagnostics value-free.
- [x] **CLI-06**: Human-facing secret list/create output is centralized through `internal/cliui`.
- [x] **CLI-07**: `shelf secret copy` copies one resolved secret to the system clipboard without printing the value; `secret get` remains stdout-oriented.
- [x] **PROJECT-01**: `.shelf.json` stores concrete item/field/env bindings only.
- [x] **PROJECT-02**: `project init/add/rm/list/status/export/run` work over item-vault bindings.
- [x] **PROJECT-03**: Project manifests stay value-free and project diagnostics do not print secret values.
- [x] **LOCAL-01**: Local workstation vault was backed up and cut over to `/home/han/.local/share/shelf/vault.json`.

## Planned Personal Workflow Requirements

- [x] **UX-01**: Developers can update item identity, labels, aliases, and supported public metadata through scriptable workflows without editing vault JSON.
- [x] **TUI-01**: `shelf tui` provides a keyboard-first, value-free item catalog with search, filters, detail, and safe terminal lifecycle handling.
- [x] **TUI-02**: TUI mutations reuse app workflows, mask secret writes, and never render a secret value.
- [ ] **TUI-03**: `shelf project tui` manages current-project value-free bindings without editing `.shelf.json` directly.
- [ ] **OPS-01**: Credential expiry audit reports expired and upcoming credentials without values.

## Deferred Requirements

- Built-in encryption if external encryption becomes insufficient.
- Optional reviewed import/migration tooling only if a real local need appears.
- Web manager.
- Team/hosted sync workflows.
- Public repository/release packaging for a future v1.0.0 milestone.

## Traceability

| Requirement | Phase | Status |
| --- | --- | --- |
| ITEM-01..ITEM-03 | Phases 35, 36, 38, 42 | Complete |
| ITEM-04 | Phase 42 | Complete |
| CLI-01..CLI-07 | Phases 42, 43, and local CLI polish | Complete |
| PROJECT-01..PROJECT-03 | Phases 39, 42 | Complete |
| LOCAL-01 | Phase 42 | Complete |
| UX-01 | Phase 47 | Complete |
| TUI-01 | Phase 48 | Complete |
| TUI-02 | Phase 48 | Complete |
| TUI-03 | Phase 49 | Planned |
| OPS-01 | Phase 50 | Planned |

---
*Last updated: 2026-07-12 after wide-screen secret workspace TUI*
