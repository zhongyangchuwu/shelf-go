# Roadmap: Shelf Go v0.2.0

## Overview

v0.2.0 is now the item-vault CLI release. The product intentionally drops built-in age/v1 compatibility and uses a plaintext JSON item vault that users can encrypt externally with chezmoi.

## Completed v0.2.0 Work

- [x] Phase 34: Template Item Vault Design
- [x] Phase 35: Vault Model and Validation Core
- [x] Phase 36: JSON Codec and Repository Boundary
- [x] Phase 37: Legacy Migration Adapter (retired by audit)
- [x] Phase 38: Secret CLI Item Workflows
- [x] Phase 39: Project Manifest Item Bindings
- [x] Phase 40: Manager Gate (superseded by manager removal)
- [x] Phase 41: Milestone Audit and Migration Scope Reset
- [x] Phase 42: Item Vault Plain JSON Cutover

## Phase 42: Item Vault Plain JSON Cutover

**Goal:** Remove built-in age/v1 code and make the local item-vault CLI work over canonical JSON storage.

**Status:** Complete

**Artifacts:** `.planning/phases/042-item-vault-storage-crypto-boundary/CONTEXT.md`, `.planning/phases/042-item-vault-storage-crypto-boundary/PLAN.md`

**Completed:**

- Removed age/jsonvault/v1 vault packages and dependencies.
- Removed manager/import/migrate/v1 export/project configure surfaces.
- Converted config/setup/doctor to item vault JSON.
- Converted project manifests to concrete item-field bindings only.
- Cut local workstation config to `/home/han/.local/share/shelf/vault.json` after backup.
- Verified tests, vet, and architecture lint.

## Remaining Before v0.2.0 Release

- [ ] Final release docs pass.
- [ ] Manual smoke test of the installed binary after commit.
- [ ] Tag/release preparation.

## Deferred After v0.2.0

1. Web manager for item vaults.
2. TUI/full-object item editing.
3. Project configure TUI.
4. Optional built-in encryption if external encryption becomes insufficient.
5. Optional migration automation with explicit review/rollback UX.

---
*Last updated: 2026-07-10 after item-vault-only cutover*
