# Shelf Go

## What This Is

Shelf Go is a local-first encrypted secret environment manager for solo developers. It keeps developer secrets in an age-encrypted portable vault, lets projects declare value-free environment bindings in `.shelf.json`, and provides predictable ways to inspect, export, inject, and edit those secrets without treating plaintext `.env` files as the source of truth.

Shelf optimizes for correctness first and usability second: secret values must not leak into config, manifests, backups, or unexpected files; common workflows should still be comfortable through clear command namespaces, project-aware exports/runs, and local TUI/Web-style editing rather than raw JSON editing when possible.

## Core Value

A developer can safely manage project secrets in an encrypted local vault and use them through explicit CLI, file, and child-process workflows without treating plaintext `.env` files as the source of truth.

## Requirements

### Validated

- [x] Secret CRUD exists through the `shelf secret` command group, including add, set, get, list, info, edit, and remove.
- [x] Secret identity uses stable `group_path:key` paths with metadata stored beside each secret.
- [x] Age-encrypted vault persistence exists for core secret workflows, including vault config, recipients, identity paths, encrypted writes, encrypted backups, and actionable vault load errors.
- [x] Plaintext-to-vault migration exists, preserving the source until the encrypted target decrypts and validates successfully.
- [x] Project manifests exist through `.shelf.json` so projects can declare required secret paths without storing values.
- [x] Project binding management exists through project commands for init, explain, add, rm, list, and export.
- [x] Runtime injection exists through a project-aware run workflow and value-free dry-run behavior.
- [x] Direct secret export exists for exact paths and prefixes in shell, env, and JSON formats.
- [x] Local health checks exist through `shelf doctor`.
- [x] Mutating secret-store commands use a write-side lock and atomic save behavior.
- [x] Git safety checks exist through `shelf doctor`, distinguishing plaintext JSON from encrypted vault files and flagging tracked plaintext stores as unsafe.
- [x] A localhost-only on-demand vault manager exists for browsing/searching metadata, intentional reveal, and create/update/delete over encrypted storage.
- [x] User-facing documentation explains encrypted vault setup, age recipients and identity paths, chezmoi-safe storage, value-free manifests, plaintext exports, migration cleanup, and manager reveal risks.
- [x] Pre-release command hierarchy is scoped and canonical: global onboarding is `shelf setup`, vault lifecycle is `shelf vault`, direct export is `shelf secret export`, and project runtime injection is `shelf project run`.
- [x] Vault status/check diagnostics report config, vault path, recipient configuration, format, loadability, and recovery guidance without revealing values.
- [x] Project session activation/deactivation/shell semantics were designed under `shelf project` and intentionally left unimplemented for now because hook-based shell mutation is not the minimal default workflow.
- [x] Project export defaults to sourceable shell output, while explicit env and JSON formats remain available and no dotenv format is added.
- [x] Vault restore exists for encrypted backups, validates restored contents before replacement, and documents identity-loss recovery limits.

### Active

- [ ] Continue safety and minimal project env UX milestone: harden plaintext boundaries for secret edit and local manager workflows.

### Out of Scope

- Team or organization sharing - Shelf is currently for one developer, so user management, invitations, permissions, and shared vault coordination are deferred.
- Hosted secret service - the product should remain local-first and not depend on a SaaS backend for core workflows.
- Replacing chezmoi - Shelf should produce and manage a portable encrypted vault file; chezmoi can continue managing dotfiles.
- General password-manager replacement - Shelf is focused on developer secrets, env bindings, project runtime workflows, and local editing, not browser autofill, credit cards, identities, or family vaults.
- Plain `.env` as the source of truth - `.env` files may be generated/exported, but Shelf's source of truth is the encrypted vault plus project manifests.
- Hook-based project activation in current scope - explicit export/source and child-process workflows are preferred until hook complexity is clearly justified.

## Context

The existing repository is a Go CLI using Cobra. The command layer lives in `internal/cli`, reusable secret-store behavior lives in `internal/store`, project manifests live in `internal/manifest`, output rendering lives in `internal/render`, local manager HTTP behavior lives in `internal/manager`, and runtime config resolution lives in `internal/config`.

The current encrypted vault milestone is complete but was built before the final command hierarchy was settled. The repo has not been published, so no backward-compatible aliases are required. Simplicity is preferred over keeping old command spellings.

The current command ambiguity is concrete: top-level `shelf init` initializes global config and vault state, while `shelf project init` initializes `.shelf.json`; top-level `shelf run` implicitly depends on project state; top-level `shelf export` exports secrets directly rather than project bindings; top-level `shelf manager` opens the vault UI but does not say what it manages.

## Constraints

- **Security:** Vault data must be encrypted at rest before it is safe to commit or sync; file permissions alone are not sufficient for a secret manager.
- **Correctness:** Commands must fail before mutating shell/project/vault state when required inputs are missing, env bindings conflict, vault decrypt fails, or validation fails.
- **Predictability:** Command namespaces must describe the object they operate on: global setup, vault lifecycle, secret records, or project manifests/sessions.
- **Encryption:** age is the preferred encryption mechanism because it matches the user's existing chezmoi setup.
- **Portability:** The encrypted vault should be a normal file that can be moved, backed up, or managed by chezmoi.
- **Local-first:** Shelf should not require a hosted backend, account, or daemon for core CLI workflows.
- **Usability:** CLI workflows must stay scriptable, but editing and browsing secrets should support better local interfaces than full-object terminal editing alone.
- **Non-secret config:** Shelf config and `.shelf.json` project manifests must not contain secret values.
- **Brownfield architecture:** New functionality should keep the current package boundaries: CLI orchestration in `internal/cli`, persistence in `internal/store`, project manifests in `internal/manifest`, rendering in `internal/render`, local manager behavior in `internal/manager`, and config resolution in `internal/config`.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use age encryption for the vault | The target workflow already uses age with chezmoi, and age fits portable file encryption better than a hosted secret manager. | Implemented for core vault persistence. |
| Keep the vault as a portable file | A normal encrypted file can be managed by git and chezmoi without building sync infrastructure. | Implemented as `shelf-vault/v1` age-encrypted file; doctor confirms tracked encrypted vaults and fails tracked plaintext stores. |
| Preserve plaintext sources during migration | Deleting or rewriting the old store before validating the new vault creates data-loss risk. | Migration leaves the plaintext source unchanged and reports manual cleanup guidance after encrypted target verification. |
| Keep Shelf CLI-first but not CLI-only | The core audience needs fast terminal workflows, but raw JSON editing is a poor UX for full secret objects. | CLI commands remain scriptable; the local manager is a valid usability feature, not scope creep. |
| Move project-dependent workflows under `project` | Commands that read `.shelf.json` have project scope and should not look like global operations. | Implemented for `project run`; future `project activate/deactivate/shell` is designed here. |
| Use `setup` for app/global onboarding | Top-level `init` conflicts with project initialization semantics. | Implemented as `shelf setup`; `shelf vault init` owns explicit vault lifecycle initialization. |
| Use `vault` for vault lifecycle and local manager entrypoints | Initializing, migrating, inspecting, and opening the vault are vault operations, not secret-record operations. | Implemented as `vault init`, `vault migrate`, `vault status`/`check`, and `vault open`. |
| Use `secret export` for direct path/prefix export | Direct export operates on vault secret paths, while `project export` operates on `.shelf.json` bindings. | Implemented under `shelf secret export`. |
| Exclude team sharing from v1 | Team sharing would force identity, permissions, revocation, audit, and conflict handling before the solo workflow is solid. | Kept out of scope. |
| Prefer explicit export/source over shell hooks | Hook-based activation mutates parent-shell state implicitly and adds restore complexity; sourceable shell output keeps behavior visible and easy to audit. | Current milestone changes `project export` default to shell output and keeps activate/deactivate/shell deferred. |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Requirements invalidated? -> Move to Out of Scope with reason.
2. Requirements validated? -> Move to Validated with phase reference.
3. New requirements emerged? -> Add to Active.
4. Decisions to log? -> Add to Key Decisions.
5. "What This Is" still accurate? -> Update if drifted.

---
*Last updated: 2026-06-24 after selecting the safety and minimal project env UX milestone*
