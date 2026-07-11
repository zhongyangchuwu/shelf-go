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
- [x] Phase 43: Secret Copy and Export Simplification
- [x] Phase 44: Current-Model Naming Cutover
- [x] Phase 45: Vault Resolver Consolidation
- [x] Phase 46: Template Role Lookup Consolidation

## Next Local Work

**Status:** Ready for the next local CLI UX or staged refactor phase.

**Most recent artifacts:** `.planning/phases/046-template-role-lookup-consolidation/VERIFICATION.md`, `.planning/phases/046-template-role-lookup-consolidation/CAPTURE.md`

**No release work:** no public repository preparation, GoReleaser packaging, GitHub Release, or tag workflow for v0.2.0.

## Completed Cutover

- Removed age/jsonvault/v1 vault packages and dependencies.
- Removed manager/import/migrate/v1 export/project configure surfaces.
- Converted config/setup/doctor to item vault JSON.
- Converted project manifests to concrete item-field bindings only.
- Cut local workstation config to `/home/han/.local/share/shelf/vault.json` after backup.
- Added Lip Gloss-backed `internal/cliui` for human command output.
- Verified tests, vet, and architecture lint.

## Completed Phase 43

- Added clipboard-only `secret copy`, including WSL bridge fallback.
- Removed `env` export; `shell` and `json` remain.
- Verified full checks and an installed-binary temporary-vault clipboard smoke test.

## Completed Phases 44–45

- Removed transitional `V2` implementation names while preserving persisted schema version fields.
- Centralized item selector, field, and env-name rules in `vaultmodel.Resolver`.
- Reused resolver registry/index construction for repository-loaded vaults.

## Completed Phase 46

- Centralized primary-secret and env-name template-role lookup in `vaultmodel.Registry`.
- Reused the registry API in secret item construction and instance resolution.

## Deferred After v0.2.0

1. More complete `shelf secret` UX polish.
2. Web manager for item vaults.
3. TUI/full-object item editing.
4. Project configure TUI.
5. Optional built-in encryption if external encryption becomes insufficient.
6. Optional migration automation with explicit review/rollback UX.
7. Public repository/release packaging for v1.0.0.

---
*Last updated: 2026-07-11 after template-role lookup consolidation*
