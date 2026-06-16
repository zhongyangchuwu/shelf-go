# Shelf Go

## What This Is

Shelf Go is a fast local secret manager for solo developers and hackers who want developer-secret workflows without the friction of `.env` sprawl, general-purpose password managers, or hosted tools like Doppler. The current product is a Go CLI that manages secrets, project bindings, direct exports, and `shelf run`; the next direction is making the store a portable git-safe encrypted vault that works well with chezmoi.

Shelf is CLI-first, but it should also provide a local vault manager over localhost for search, viewing, and editing because editing complete secret objects in a terminal editor is awkward. Team sharing is intentionally out of scope for now.

## Core Value

A single developer can safely carry and use project secrets across machines through a portable encrypted vault, while keeping local env and `shelf run` workflows fast and simple.

## Requirements

### Validated

- [x] Secret CRUD exists through the `shelf secret` command group, including add, set, get, list, info, edit, and remove.
- [x] Secret identity uses stable `group_path:key` paths with metadata stored beside each secret.
- [x] Direct export exists for exact paths and prefixes in shell, env, and JSON formats.
- [x] Project manifests exist through `.shelf.json` so projects can declare required secret paths without storing values.
- [x] Project binding management exists through `shelf project init`, `explain`, `add`, `rm`, `list`, and `export`.
- [x] Runtime injection exists through `shelf run -- ...` and `shelf run --dry-run -- ...`.
- [x] Local health checks exist through `shelf doctor`.
- [x] Mutating secret-store commands use a write-side lock and atomic save behavior.

### Active

- [ ] Replace the plaintext JSON store boundary with an age-encrypted vault file while preserving existing command semantics after decrypt/load and before encrypt/save.
- [ ] Make the encrypted vault safe to commit or sync through git-backed dotfile workflows, especially chezmoi-managed repositories.
- [ ] Keep Shelf's configuration non-secret and portable, with clear separation between config files, project manifests, and encrypted vault data.
- [ ] Preserve fast CLI workflows for setting, listing, exporting, project binding, and `shelf run`.
- [ ] Add a localhost-only vault manager that can search, inspect metadata, and edit secrets.
- [ ] Improve secret editing so common edit workflows do not require manually editing full JSON in `$EDITOR`.
- [ ] Harden the storage layer around encryption, backups, migration, and validation so backups and temporary files do not reintroduce plaintext-at-rest risk.

### Out of Scope

- Team or organization sharing - Shelf is currently for one developer, so user management, invitations, permissions, and shared vault coordination are deferred.
- Hosted secret service - the product should remain local-first and not depend on a SaaS backend for v1.
- Replacing chezmoi - Shelf should produce and manage a portable encrypted vault file; chezmoi can continue managing dotfiles.
- General password-manager replacement - Shelf is focused on developer secrets, env bindings, and project runtime workflows, not browser autofill, credit cards, identities, or family vaults.
- Plain `.env` as the source of truth - `.env` files may be generated/exported, but Shelf's source of truth is the encrypted vault plus project manifests.

## Context

The existing repository is a Go CLI using Cobra. The command layer lives in `internal/cli`, reusable secret-store behavior lives in `internal/store`, project manifests live in `internal/manifest`, output rendering lives in `internal/render`, and runtime config resolution lives in `internal/config`.

The current store is a small local JSON file with `version: 1` and a flat map of canonical secret paths to secret objects. That model is intentionally simple and should remain the plaintext in-memory model after decryption. The key architectural change is the persistence boundary: load should decrypt and validate; save should validate, serialize, encrypt, and atomically persist.

The current docs already treat plaintext storage as an MVP compromise and call out encryption as the main missing feature. The existing codebase map also flags backup handling, editor temp files, store migration, prefix matching, and trailing JSON validation as important hardening concerns.

The target user uses chezmoi and age encryption today. Shelf should fit that mental model by making the vault portable and git-safe, rather than inventing a new sync system.

## Constraints

- **Security**: Vault data must be encrypted at rest before it is safe to commit or sync - file permissions alone are not sufficient for a secret manager.
- **Encryption**: age is the preferred encryption mechanism because it matches the user's existing chezmoi setup.
- **Portability**: The encrypted vault should be a normal file that can be moved, backed up, or managed by chezmoi.
- **Local-first**: Shelf should not require a hosted backend, account, or daemon for core CLI workflows.
- **CLI-first UX**: Commands must stay scriptable and predictable; the localhost vault manager is additive, not a replacement for CLI use.
- **Non-secret config**: Shelf config and `.shelf.json` project manifests must not contain secret values.
- **Brownfield architecture**: New functionality should keep the current package boundaries: CLI orchestration in `internal/cli`, persistence in `internal/store`, project manifests in `internal/manifest`, rendering in `internal/render`, and config resolution in `internal/config`.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use age encryption for the vault | The target workflow already uses age with chezmoi, and age fits portable file encryption better than a hosted secret manager. | Pending |
| Keep the vault as a portable file | A normal encrypted file can be managed by git and chezmoi without building sync infrastructure. | Pending |
| Keep Shelf CLI-first | The core audience is solo developers and hackers who need fast terminal workflows for env export and runtime injection. | Pending |
| Add a localhost vault manager with editing | CLI JSON editing is painful; a local UI can improve search and edits while staying local-first. | Pending |
| Exclude team sharing from v1 | Team sharing would force identity, permissions, sharing protocols, and conflict handling before the solo workflow is solid. | Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason.
2. Requirements validated? -> Move to Validated with phase reference.
3. New requirements emerged? -> Add to Active.
4. Decisions to log? -> Add to Key Decisions.
5. "What This Is" still accurate? -> Update if drifted.

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections.
2. Core Value check - still the right priority?
3. Audit Out of Scope - reasons still valid?
4. Update Context with current state.

---
*Last updated: 2026-06-16 after initialization*
