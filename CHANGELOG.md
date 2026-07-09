# Changelog

All notable user-facing changes will be recorded here.

## Unreleased

### Breaking changes

- Replaced `shelf project explain` with `shelf project status`; status now reports project binding resolution and value-free local env-file summaries.
- Replaced the old `group:key` secret CLI with v2 template-based `secret create`, `set --stdin`, `get`, `list --long`, `info`, and `rm --yes` item/field workflows.
- Removed `secret add`, `secret edit`, and old path/prefix direct secret export from the v2 secret CLI; full interactive editing is deferred to later v2 surfaces.
- Gated `shelf manager` when the resolved v2 vault exists; the v1 Web manager no longer opens over v2 item/field vault workflows.
- Removed the internal v1-to-v2 migration adapter from the v0.2.0 release scope; automatic migration is deferred until backup, dry-run, user review, and rollback requirements are designed.

### Added

- Added `shelf project configure` as a TUI for value-free project binding setup, including env-file binding, existing-secret add, remove, review, and apply flows.
- Added v2 vault profiles with an `openai` profile over the `api_credential` template.
- Added env-name public fields such as `value_env` and `credential_env` so env names are ordinary item fields instead of special field bindings.
- Added v2 `.shelf.json` project bindings with concrete item/field/env records for `project status`, `project export`, `project run`, `project add`, `project list`, and `project rm`.
- Added label-scoped v2 `shelf secret export --label LABEL` in shell, env, and JSON formats for intentional shell initialization without storing values in shell config.
- Changed v2 label export to skip items without an explicit env-name field, and v1-to-v2 migration now leaves derived env names out of item env fields.

- Added `scripts/test.sh` / `just test` as the single local verification entrypoint, running Go tests, `go vet`, and go-arch-lint architecture checks.
- Simplified `justfile` to keep one test entrypoint and thin install/tag wrappers.

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

- Added age-encrypted vault storage with `shelf-vault/v1` envelope.
- Added vault setup, status/check, migration, and localhost manager commands under `shelf vault`.
- Added project-aware secret binding commands under `shelf project`.
- Added sourceable `shelf project export` shell output as the default project export format.
- Added direct secret export for exact paths and prefixes under `shelf secret export`.
- Added minimal last-write encrypted `.bak` recovery documentation.
- Hardened `shelf secret edit` temporary file permissions and local manager safety tests.
- Added shell completion fixes for project manifest and secret prefix arguments.
- Added public README, security policy, portable vault guide, reference, troubleshooting, contributing, and architecture docs.
- Added repeatable GitHub release automation with GoReleaser archives and checksums.
