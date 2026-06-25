# Changelog

All notable user-facing changes will be recorded here.

## Unreleased

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
