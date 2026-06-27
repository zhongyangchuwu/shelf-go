# Shelf Go

## What This Is

Shelf Go is a local-first encrypted secret environment manager for solo developers. It keeps developer secrets in an age-encrypted portable vault, lets projects declare value-free environment bindings in `.shelf.json`, and provides predictable ways to inspect, export, inject, and edit those secrets without treating plaintext `.env` files as the source of truth.

Shelf optimizes for correctness first and usability second: secret values must not leak into config, manifests, backups, or unexpected files; common workflows should stay comfortable through clear command namespaces, project-aware exports/runs, tag-based selection, and a local Web console for editing rather than raw JSON editing.

## Core Value

A developer can safely manage project secrets in an encrypted local vault and use them through explicit CLI, file, and child-process workflows without treating plaintext `.env` files as the source of truth.

## Requirements

### Validated

- [x] v0.1.0 is published with age-encrypted portable vault storage, scoped CLI command groups, value-free project manifests, project export/run workflows, localhost vault manager, release automation, and public documentation.
- [x] Completed v0.1.0 planning history is archived at `.planning/archive/releases/v0.1.0/`.
- [x] v0.1.1 Web manager design and editing console provide add/edit/delete/reveal/copy/tag workflows over the local vault without changing the storage format.
- [x] v0.1.1 direct secret CLI workflows support tag-based list/export selection with repeatable AND semantics.
- [x] v0.1.1 project manifests support value-free tag bindings for project export/run workflows.
- [x] v0.1.1 consolidates install/tag/release workflows into reusable `scripts/` Bash scripts and keeps `justfile` thin.
- [x] v0.1.1 repartitions internal packages and uses `shelf manager` as the single local manager entrypoint.
- [x] v0.1.1 updates user and developer docs for Web manager editing, tag workflows, scripted workflows, and final package layout.

### Active

- [ ] v0.1.1 keeps the current age-encrypted JSON vault format; SQLite/storage redesign is deferred to v0.2.0.
- [ ] v0.1.1 release hardening happens only after scripts, docs, and architecture cleanup complete.

### Out of Scope

- Team or organization sharing - Shelf is currently for one developer, so user management, invitations, permissions, and shared vault coordination are deferred.
- Hosted secret service - the product should remain local-first and not depend on a SaaS backend for core workflows.
- Replacing chezmoi - Shelf should produce and manage a portable encrypted vault file; chezmoi can continue managing dotfiles.
- General password-manager replacement - Shelf is focused on developer secrets, env bindings, project runtime workflows, and local editing, not browser autofill, credit cards, identities, or family vaults.
- Plain `.env` as the source of truth - `.env` files may be generated/exported, but Shelf's source of truth is the encrypted vault plus project manifests.
- Hook-based project activation in current scope - explicit export/source and child-process workflows are preferred until hook complexity is clearly justified.
- SQLite or storage backend replacement in v0.1.1 - storage model changes are deferred to v0.2.0 discussion.

## Context

The repository is a Go CLI using Cobra. The display layer lives in `cmd/shelf`, `internal/cli`, and `internal/manager`. Feature support lives in `internal/app`, `internal/project`, and `internal/secret`; base support lives in `internal/config`, `internal/vault`, and `internal/exportfmt`.

The v0.1.0 release is published and archived. v0.1.1 focuses on editing UX and tag-based workflows rather than storage migration. Manager editing, tag workflows, script consolidation, and architecture repartitioning are implemented; remaining pre-release work is docs and release hardening.

## Constraints

- **Security:** Vault data must remain encrypted at rest before it is safe to commit or sync; file permissions alone are not sufficient for a secret manager.
- **Correctness:** Commands must fail before mutating shell/project/vault state when required inputs are missing, env bindings conflict, vault decrypt fails, or validation fails.
- **Predictability:** Command namespaces must describe the object they operate on: global setup, vault lifecycle, secret records, or project manifests/sessions.
- **Encryption:** age remains the preferred v0.1.1 encryption mechanism because it matches the user's existing chezmoi setup.
- **Portability:** The encrypted vault should remain a normal file that can be moved, backed up, or managed by chezmoi.
- **Local-first:** Shelf should not require a hosted backend, account, CDN, or daemon for core CLI/Web manager workflows.
- **Usability:** CLI workflows must stay scriptable; full editing should be comfortable in the local Web manager.
- **Non-secret config:** Shelf config and `.shelf.json` project manifests must not contain secret values.
- **Brownfield architecture:** Keep command orchestration in `internal/cli`, reusable feature workflows in feature packages, encrypted vault core in `internal/vault`, project manifest behavior in `internal/project`, export formatting in `internal/exportfmt`, local manager behavior in `internal/manager`, and config resolution in `internal/config`.
- **Workflow automation:** Common install, tag, and release flows should live in scripts instead of only in `justfile` or remembered manual commands.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use age encryption for the vault | The target workflow already uses age with chezmoi, and age fits portable file encryption better than a hosted secret manager. | Implemented for core vault persistence; retained for v0.1.1. |
| Keep the vault as a portable file | A normal encrypted file can be managed by git and chezmoi without building sync infrastructure. | Implemented as `shelf-vault/v1`; storage replacement deferred to v0.2.0. |
| Preserve plaintext sources during migration | Deleting or rewriting the old store before validating the new vault creates data-loss risk. | Migration leaves the plaintext source unchanged and reports manual cleanup guidance after encrypted target verification. |
| Keep Shelf CLI-first but not CLI-only | The core audience needs fast terminal workflows, but raw JSON editing is a poor UX for full secret objects. | CLI commands remain scriptable; v0.1.1 makes the local Web manager the main editing surface. |
| Move project-dependent workflows under `project` | Commands that read `.shelf.json` have project scope and should not look like global operations. | Implemented for `project run`; v0.1.1 adds project tag binding. |
| Use `setup` for app/global onboarding | Top-level `init` conflicts with project initialization semantics. | Implemented as `shelf setup`; `shelf vault init` owns explicit vault lifecycle initialization. |
| Use `manager` for the local management surface | The manager can grow beyond vault-only panels, so it should not live under `vault open`. | `shelf manager` is the single local manager entrypoint; `vault` remains for vault lifecycle commands. |
| Use `secret export` for direct path/prefix export | Direct export operates on vault secret paths, while `project export` operates on `.shelf.json` bindings. | Implemented under `shelf secret export`; v0.1.1 extends it with tag selection. |
| Exclude team sharing from v1 | Team sharing would force identity, permissions, revocation, audit, and conflict handling before the solo workflow is solid. | Kept out of scope. |
| Prefer explicit export/source over shell hooks | Hook-based activation mutates parent-shell state implicitly and adds restore complexity; sourceable shell output keeps behavior visible and easy to audit. | `project export` defaults to shell output; activate/deactivate/shell remains deferred. |
| Defer storage-engine changes | JSON inside an age-encrypted vault keeps the security and portability model simple. SQLite is worth future discussion but not part of editing UX delivery. | Current storage remains age-encrypted JSON through v0.1.1; SQLite moves to v0.2.0 consideration. |
| Keep reusable workflows out of `internal/cli` | CLI files should stay command-family oriented and not own behavior needed by tests, manager, or future UX. | `internal/app`, `internal/project`, `internal/vault`, `internal/secret`, `internal/manager`, and `internal/exportfmt` own reusable behavior outside CLI. |
| Keep CLI editing compact | Fine-grained `meta`/`tag` edit commands increase command surface while WebUI is the intended editing surface. | v0.1.1 does not add `secret meta` or `secret tag`; CLI focuses on list/export/project tag application flows. |
| Release hardening is final, not next | Install/tag/release scripts, docs, and architecture naming cleanup are prerequisites for a maintainable release. | Phase 22 is architecture repartitioning, Phase 23 is docs alignment, Phase 24 is release hardening. |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Requirements invalidated? -> Move to Out of Scope with reason.
2. Requirements validated? -> Move to Validated with phase reference.
3. New requirements emerged? -> Add to Active.
4. Decisions to log? -> Add to Key Decisions.
5. "What This Is" still accurate? -> Update if drifted.

---
*Last updated: 2026-06-27 after completing architecture repartition core*
