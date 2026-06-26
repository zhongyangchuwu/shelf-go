# Shelf Go

## What This Is

Shelf Go is a local-first encrypted secret environment manager for solo developers. It keeps developer secrets in an age-encrypted portable vault, lets projects declare value-free environment bindings in `.shelf.json`, and provides predictable ways to inspect, export, inject, and edit those secrets without treating plaintext `.env` files as the source of truth.

Shelf optimizes for correctness first and usability second: secret values must not leak into config, manifests, backups, or unexpected files; common workflows should still be comfortable through clear command namespaces, project-aware exports/runs, and local Web-style editing rather than raw JSON editing when possible.

## Core Value

A developer can safely manage project secrets in an encrypted local vault and use them through explicit CLI, file, and child-process workflows without treating plaintext `.env` files as the source of truth.

## Requirements

### Validated

- [x] v0.1.0 is published with age-encrypted portable vault storage, scoped CLI command groups, value-free project manifests, project export/run workflows, localhost vault manager, release automation, and public documentation.
- [x] Completed v0.1.0 planning history is archived at `.planning/archive/releases/v0.1.0/`.

### Active

- [ ] Select the v0.1.1 release or implementation milestone.

### Out of Scope

- Team or organization sharing - Shelf is currently for one developer, so user management, invitations, permissions, and shared vault coordination are deferred.
- Hosted secret service - the product should remain local-first and not depend on a SaaS backend for core workflows.
- Replacing chezmoi - Shelf should produce and manage a portable encrypted vault file; chezmoi can continue managing dotfiles.
- General password-manager replacement - Shelf is focused on developer secrets, env bindings, project runtime workflows, and local editing, not browser autofill, credit cards, identities, or family vaults.
- Plain `.env` as the source of truth - `.env` files may be generated/exported, but Shelf's source of truth is the encrypted vault plus project manifests.
- Hook-based project activation in current scope - explicit export/source and child-process workflows are preferred until hook complexity is clearly justified.

## Context

The repository is a Go CLI using Cobra. The display layer lives in `cmd/shelf`, `internal/cli`, and `internal/manager`; feature support lives in `internal/app`, `internal/project`, `internal/vault`, and `internal/secret`; base support lives in `internal/config`, `internal/store`, `internal/manifest`, `internal/render`, `internal/atomicfile`, and `internal/version`.

The v0.1.0 release is published and archived. New work should start by selecting a v0.1.1 milestone, adding current requirements, and creating fresh phase directories under `.planning/phases/`.

## Constraints

- **Security:** Vault data must be encrypted at rest before it is safe to commit or sync; file permissions alone are not sufficient for a secret manager.
- **Correctness:** Commands must fail before mutating shell/project/vault state when required inputs are missing, env bindings conflict, vault decrypt fails, or validation fails.
- **Predictability:** Command namespaces must describe the object they operate on: global setup, vault lifecycle, secret records, or project manifests/sessions.
- **Encryption:** age is the preferred encryption mechanism because it matches the user's existing chezmoi setup.
- **Portability:** The encrypted vault should be a normal file that can be moved, backed up, or managed by chezmoi.
- **Local-first:** Shelf should not require a hosted backend, account, or daemon for core CLI workflows.
- **Usability:** CLI workflows must stay scriptable, but editing and browsing secrets should support better local interfaces than full-object terminal editing alone.
- **Non-secret config:** Shelf config and `.shelf.json` project manifests must not contain secret values.
- **Brownfield architecture:** Keep command orchestration in `internal/cli`, reusable feature workflows in feature packages, persistence in `internal/store`, project manifests in `internal/manifest`, rendering in `internal/render`, local manager behavior in `internal/manager`, and config resolution in `internal/config`.

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
| Prefer explicit export/source over shell hooks | Hook-based activation mutates parent-shell state implicitly and adds restore complexity; sourceable shell output keeps behavior visible and easy to audit. | `project export` defaults to shell output; activate/deactivate/shell remains deferred. |
| Defer storage-engine changes | JSON inside an age-encrypted vault keeps the security and portability model simple. SQLite is worth a future spike only when schema/search/history pressure appears; Dolt is too heavy for vault storage and conflicts with encrypted-at-rest secret semantics. | Current storage remains age-encrypted JSON; SQLite is recorded as a deferred candidate, Dolt is not. |
| Keep reusable workflows out of `internal/cli` | CLI files should stay command-family oriented and not own behavior needed by tests, manager, or future UX. | `internal/app`, `internal/project`, `internal/vault`, `internal/secret`, and `internal/atomicfile` own reusable behavior. |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Requirements invalidated? -> Move to Out of Scope with reason.
2. Requirements validated? -> Move to Validated with phase reference.
3. New requirements emerged? -> Add to Active.
4. Decisions to log? -> Add to Key Decisions.
5. "What This Is" still accurate? -> Update if drifted.

---
*Last updated: 2026-06-26 after archiving v0.1.0 planning history*
