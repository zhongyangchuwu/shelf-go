# Roadmap: Shelf Go v0.2.0

## Overview

v0.2.0 is the local item-vault CLI milestone. The product intentionally drops built-in age/v1 compatibility and uses a plaintext JSON item vault that can be encrypted externally with chezmoi or similar tooling.

This milestone is for local development and personal use. Public repository/release work is deferred until a future v1.0.0 milestone.

**Current version:** `v0.2.0` is the local milestone label and the default CLI version. It does not create a public tag, release, or packaging workflow.

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

**Status:** Personal Workflow UX phases are complete. Shelf now covers scriptable item lifecycle, wide-screen secret and project binding workspaces, and value-free credential health signals without a multi-environment manifest model.

- [x] Phase 47: Scriptable Item Lifecycle — added safe, scriptable item mutations before interactive editing.
- [x] Phase 48: Wide-Screen Secret Workspace TUI — keyboard-first catalog plus item create, update, copy, and remove actions.
- [x] Phase 49: Project Binding TUI — value-free current-project manifest management and binding picker.
- [x] Phase 50: Credential Health Audit — value-free expired and upcoming-expiry checks in CLI and TUI.

**Execution order:** Phase 47 is the CLI/API foundation. Phase 48 delivers the complete personal secret workspace on wide terminals. Phase 49 adds project bindings to the established workspace. Phase 50 surfaces operational health in both CLI and TUI.

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

## Completed Phase 48

- Added `shelf tui` with interactive-terminal preflight, a 100×24 minimum, alternate-screen cleanup, and a resize-required view.
- Added safe catalog/search/filter/detail views, including AND label filtering and scrollable detail content.
- Added masked create/edit forms, value-free review and confirmation views, app-backed create/update/remove actions, and clipboard-only copy.
- Preserved value-free TUI state and architecture boundaries; full checks and pseudo-terminal lifecycle smoke passed.

## Completed Phase 49

- Added `shelf project tui` with the existing terminal preflight, alternate-screen lifecycle, and resize-required view.
- Added safe project-workspace/candidate app DTOs and value-free binding-status resolution without materializing values.
- Added manifest initialization confirmation, candidate picker/filter, env override and optional form, remove confirmation, and resolution diagnostics.
- Kept manifest/vault I/O in app and project export/run as explicit CLI actions.

## Completed Phase 50

- Added `shelf secret audit` with value-free text/JSON output, configurable UTC warning horizon, and deterministic health exit codes.
- Added app health classifications for expired, due, missing, invalid, and healthy expiry metadata without reading secret field values.
- Added safe health indicators to secret/project TUI views and project status; project export/run semantics remain unchanged.

## Completed Technical Debt Review

- Repaired existing vault-backup permissions on every backup write and added regression coverage.
- Replaced built-in env-field assumptions in create summaries with canonical template-role resolution.
- Removed a package panic/dead setup helper and corrected stale architecture, security, contributor, quickstart, and troubleshooting documentation.


## Completed Catppuccin Mocha Panel TUI Refinement

- Reframed secret and project workspaces as persistent two-panel layouts with rounded borders, active-pane markers, and a shared fixed Catppuccin Mocha palette.
- Preserved value-free rendering, current keymaps, resize behavior, and app/clipboard boundaries.

## Completed Phase 53

- Replaced V2 field instances with V3 atomic `id`/`value` storage and template-owned field definitions; V2 vaults are rejected without automatic migration.
- Replaced role metadata with validated `env_<field>` and `validity_<secret-field>` conventions, including field-level lifecycle audit.
- Replaced `shelf secret export` with `shelf env export`, projecting explicitly mapped secret and public fields.
- Allowed project binding/export/run for eligible public and secret fields, and grouped structural metadata in value-free TUI detail.
- Added V3 model, codec, app, CLI, project, TUI, full-suite, temporary-vault, and pseudo-terminal verification; published a manual migration guide.

## Deferred Beyond the Personal Workflow UX Plan

1. Web manager for item vaults.
2. Optional built-in encryption if external encryption becomes insufficient.
3. Optional migration automation with explicit review/rollback UX.
4. Team/hosted sync workflows.
5. Public repository/release packaging for v1.0.0.

---
*Last updated: 2026-07-13 after Item-First Environment Projection V3*
