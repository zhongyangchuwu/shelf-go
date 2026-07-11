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

## Planned Personal Workflow UX

**Status:** Phase 47 complete; Phase 48 Safe Catalog TUI design is awaiting user review. These phases reinforce Shelf as a local, item-first catalog with explicit runtime projections; they do not add a multi-environment manifest model.

- [x] Phase 47: Scriptable Item Lifecycle — added safe, scriptable item mutations before interactive editing.
- [ ] Phase 48: Safe Catalog TUI — keyboard-first, value-free catalog browsing, search, filtering, and detail views.
- [ ] Phase 49: TUI Item Editing — create, edit, set, copy, and delete flows using the same app workflows as CLI.
- [ ] Phase 50: Project Binding TUI — value-free current-project manifest management and binding picker.
- [ ] Phase 51: Credential Health Audit — value-free expired and upcoming-expiry checks in CLI and TUI.

**Execution order:** Phase 47 is the CLI/API foundation. Phase 48 proves safe TUI lifecycle and read paths before Phase 49 introduces secret writes. Phase 50 builds on both. Phase 51 is independent of TUI editing but follows it to surface operational data in the established UI.

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

## Completed Phase 47

- Added atomic `secret update` changes for item title, labels, aliases, and typed public fields.
- Kept secret field writes stdin-only and all update output and errors value-free.

## Deferred Beyond the Personal Workflow UX Plan

1. Web manager for item vaults.
2. Optional built-in encryption if external encryption becomes insufficient.
3. Optional migration automation with explicit review/rollback UX.
4. Team/hosted sync workflows.
5. Public repository/release packaging for v1.0.0.

---
*Last updated: 2026-07-12 after scriptable item lifecycle*
