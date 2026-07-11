# Changelog

All notable user-facing changes will be recorded here.

## Unreleased

### Breaking changes

- Removed built-in age encryption and the v1 encrypted vault stack. Shelf now uses a local plaintext JSON item vault; use chezmoi or another external tool for encrypted sync.
- Removed v1 vault lifecycle commands and surfaces: `shelf vault`, top-level v1 `export`, v1 import/migrate, and the Web manager.
- Removed v1 path/prefix/tag project manifest support. `.shelf.json` now stores concrete item/field/env bindings.
- Removed project configure TUI; project bindings are managed through CLI commands.
- Replaced the old `group:key` secret CLI with item/field workflows.
- Removed v0.2.0 release packaging/tag automation from active local-development scope; public release work is deferred to a future v1.0.0 milestone.
- Removed the redundant `env` export format; `secret export` and `project export` now support `shell` and `json` only.

### Added

- Added `shelf setup` for config and item-vault JSON initialization.
- Added item-vault `shelf doctor` checks for config, permissions, loadability, and completion state.
- Added template-based secret item commands: `secret create`, `set --stdin`, `get`, `list --long`, `info`, `rm --yes`, and label-scoped `secret export --label`.
- Added concrete project item/field bindings for `project status`, `project add`, `project list`, `project rm`, `project export`, and `project run`.
- Added shared Lip Gloss-backed human output rendering for secret create/list output.
- Added `shelf secret copy` for clipboard-only interactive secret retrieval with value-free confirmation.

### Fixed

- Secret creation now rejects metadata flags unsupported by the selected template instead of silently dropping them.

## 0.1.1 - 2026-06-27

- Replaced the local manager entrypoint with `shelf manager` and removed the pre-release `shelf vault open` command.
- Rebuilt the local manager as a searchable editing console with add, edit, rename, delete, reveal, copy, and tag workflows.
- Hardened manager access with token URL cleanup, no-store responses, metadata-only list/detail responses, and explicit POST reveal actions.
- Added tag-based secret selection to `shelf secret list` and `shelf secret export` with repeatable AND semantics.
- Added value-free project tag bindings for `shelf project add`, `list`, `explain`, `export`, and `run`.
- Consolidated install, release check, snapshot, and tag workflows under reusable `scripts/` commands while keeping `justfile` as a thin wrapper.
- Repartitioned internal packages around `app`, `project`, `secret`, `vault`, `manager`, `config`, and `exportfmt` boundaries.
- Updated README and docs for manager editing, tag workflows, scripts, architecture, and troubleshooting.

## 0.1.0 - 2026-06-25

- Added encrypted local vault core with age recipients and identities.
- Added local secret CRUD commands, project manifests, direct export, and `shelf run`.
- Added doctor/status checks, shell completions, and release automation.
