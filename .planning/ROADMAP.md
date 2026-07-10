# Roadmap: Shelf Go v0.2.0

## Overview

v0.2.0 is the local item-vault CLI milestone. The product intentionally drops built-in age/v1 compatibility and uses a plaintext JSON item vault that can be encrypted externally with chezmoi or similar tooling.

This milestone is for local development and personal use. Public repository/release work is deferred until a future v1.0.0 milestone.

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
- [x] CLI output polish and `internal/cliui` extraction

## Current Wrap-up Scope

**Goal:** Review code structure, docs, planning artifacts, and local file hygiene so v0.2.0 is coherent for continued local use.

**Status:** In progress

**No release work:** no public repository preparation, GoReleaser packaging, GitHub Release, or tag workflow for v0.2.0.

## Completed Cutover

- Removed age/jsonvault/v1 vault packages and dependencies.
- Removed manager/import/migrate/v1 export/project configure surfaces.
- Converted config/setup/doctor to item vault JSON.
- Converted project manifests to concrete item-field bindings only.
- Cut local workstation config to `/home/han/.local/share/shelf/vault.json` after backup.
- Added Lip Gloss-backed `internal/cliui` for human command output.
- Verified tests, vet, and architecture lint.

## Remaining Local Wrap-up

- [ ] Finish code/framework review and apply must-fix cleanup.
- [ ] Finish `.planning/` and `docs/` cleanup.
- [ ] Run installed-binary smoke tests.

## Deferred After v0.2.0

1. More complete `shelf secret` UX polish.
2. Web manager for item vaults.
3. TUI/full-object item editing.
4. Project configure TUI.
5. Optional built-in encryption if external encryption becomes insufficient.
6. Optional migration automation with explicit review/rollback UX.
7. Public repository/release packaging for v1.0.0.

---
*Last updated: 2026-07-10 during local v0.2.0 wrap-up*
