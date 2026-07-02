# Changelog

All notable user-facing changes will be recorded here.

## Unreleased

### Breaking changes

- Replaced `shelf project explain` with `shelf project status`; status now reports project binding resolution and value-free local env-file summaries.

### Added

- Added `shelf project configure` for interactive, value-free project binding setup from env-file variables to existing vault secrets.

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
