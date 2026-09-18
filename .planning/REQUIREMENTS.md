# Requirements: Shelf Go

## Current v0.2.0 Requirements

- [x] **ITEM-01**: Shelf stores schema-V4 item vaults with atomic `id`/`value` fields, template-owned definitions, `env_<field>` mappings, `validity_<secret-field>` lifecycle metadata, labels, descriptions, profiles, validation, and value-free metadata output.
- [x] **ITEM-02**: Shelf persists the item vault as strict JSON with validation, atomic writes, and backups.
- [x] **ITEM-03**: Shelf uses the configured vault path directly and defaults to `~/.local/share/shelf/vault.json`.
- [x] **ITEM-04**: Built-in age encryption, v1 encrypted vault code, v1 manager, import/migrate, and v1 export paths are removed from active product scope.
- [x] **CLI-01**: `shelf setup` creates config and an empty item vault.
- [x] **CLI-02**: `shelf doctor` validates config, permissions, vault loadability, and shell completion without printing values.
- [x] **CLI-03**: `shelf secret create`, `update`, `get`, `list`, `info`, `rm`, and `audit` operate over item vault JSON; `shelf env export` materializes mapped environment fields.
- [x] **CLI-04**: Secret writes read values from stdin, not argv.
- [x] **CLI-05**: Label-scoped env export materializes matching explicitly mapped secret and public fields and keeps duplicate/missing failure diagnostics value-free.
- [x] **CLI-06**: Human-facing secret list/create output is centralized through `internal/cliui`.
- [x] **CLI-07**: `shelf secret copy` copies one resolved secret to the system clipboard without printing the value; `secret get` remains stdout-oriented.
- [x] **PROJECT-01**: `.shelf.json` stores concrete item/field/env bindings only.
- [x] **PROJECT-02**: `project init/add/rm/list/status/export/run` work over item-vault bindings.
- [x] **PROJECT-03**: Project manifests stay value-free and project diagnostics do not print secret values.
- [x] **LOCAL-01**: Local workstation vault was backed up and cut over to `/home/han/.local/share/shelf/vault.json`.

## Completed Personal Workflow Requirements

- [x] **UX-01**: Developers can update item identity, labels, descriptions, and supported public metadata through scriptable workflows without editing vault JSON.
- [x] **TUI-01**: `shelf tui` provides a keyboard-first, value-free item catalog with search, filters, detail, and safe terminal lifecycle handling.
- [x] **TUI-02**: TUI mutations reuse app workflows, mask secret writes, and never render a secret value.
- [x] **TUI-03**: `shelf project tui` manages current-project value-free bindings without editing `.shelf.json` directly.
- [x] **OPS-01**: Field-level lifecycle audit reports unknown, permanent, invalid, expired, and upcoming tracked secret fields without values.
- [x] **V4-01**: The active vault schema is V4; item aliases are removed, `description` is public metadata, and item selectors accept ID or title only. Historical V2/V3 cutovers remain documented without compatibility aliases or automatic migration.
- [x] **V4-02**: Project bindings/export/run support eligible secret and public `string`, `concealed`, and `url` fields while manifests remain value-free.

## Maintenance State

There is no active feature roadmap after `v0.2.0`. Shelf is retained as a research artifact/reference implementation. Built-in encryption, hosted sync, accounts, daemon scheduling, provider adapters, automatic rotation, Web management, and public release packaging are not planned. Reopening feature development requires a concrete workflow gap that cannot be handled cleanly by an established secret store or the surrounding Personal AI OS.

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
| TUI-03 | Phase 49 | Complete |
| OPS-01 | Phase 50 | Complete |
| V4-01..V4-02 | Phase 53 plus the V4 description cutover | Complete |

---
*Last updated: 2026-09-18 after the V4/current-state and maintenance review*
