# Requirements: Shelf Go

**Defined:** 2026-06-16
**Revised:** 2026-06-28
**Core Value:** A developer can safely manage project secrets in an encrypted local vault and use them through explicit CLI, file, and child-process workflows without treating plaintext `.env` files as the source of truth.

## Current Requirements

Requirements for v0.1.1 editing experience, tag-based workflows, workflow scripts, architecture cleanup, documentation, release preparation, and the optional pre-tag CLI boundary refactor. Completed v0.1.0 requirements and evidence are archived at `.planning/archive/releases/v0.1.0/SUMMARY.md` and `.planning/archive/releases/v0.1.0/VERIFICATION.md`.

### Web Manager Editing

- [x] **WEB-01**: The local manager provides a searchable, understandable secret console with path, env, description, tag, and value-set metadata visible without revealing secret values.
- [x] **WEB-02**: The local manager supports adding, editing, renaming, and deleting secret records, including value, env, description, and tags.
- [x] **WEB-03**: The local manager supports explicit reveal, hide, and copy flows without returning secret values in list/search responses or storing them in browser-local persistent storage.
- [x] **WEB-04**: The local manager preserves and strengthens local-only safety boundaries: loopback binding, token/cookie access, Host/Origin checks, token removal from the visible URL after first load, and no-store responses for secret-bearing endpoints.
- [x] **WEB-05**: The local manager uses embedded local assets only; no CDN, hosted frontend, or permanent daemon dependency is required.
- [x] **WEB-06**: The local manager adopts a polished console visual direction based on a reusable HTML/CSS design system or template reference, with the implementation kept compatible with Go's single-binary distribution.

### Tag-Based CLI and Project Workflows

- [x] **TAG-01**: `shelf secret list` can filter secrets by one or more tags while keeping output value-free and deterministic.
- [x] **TAG-02**: `shelf secret export` can select secrets by one or more tags using the existing env, shell, and JSON formats and the existing `--all` behavior.
- [x] **TAG-03**: Project manifests can declare tag-selected secret sets without storing secret values.
- [x] **TAG-04**: `shelf project add`, `list`, `explain`, `export`, and `run` support tag-selected bindings with clear expansion, missing-secret, and env-conflict diagnostics.
- [x] **TAG-05**: Multiple tag selectors use AND semantics for v0.1.1 so exported/project-bound sets stay narrow and predictable.

### Scripted Workflow Cleanup

- [x] **OPS-01**: Install flow currently embedded in `justfile` is moved to a reusable script under `scripts/`, and `just install` delegates to that script.
- [x] **OPS-02**: Tag and release preparation flows are moved from ad-hoc manual commands and inline `justfile` recipes into reusable scripts under `scripts/`.
- [x] **OPS-03**: Scripted workflows have clear usage, argument validation, and keep `justfile` as a thin task runner.

### Architecture Cleanup

- [x] **ARCH-01**: The manager entrypoint is renamed to `shelf manager`, and the vault-scoped `shelf vault open` command is removed before release.
- [x] **ARCH-02**: The internal package layout is repartitioned so vault core, project manifest handling, application composition, and export formatting have clear package names and dependency direction.

- [x] **ARCH-03**: Project/session business rules currently embedded in `internal/cli/project.go` and `internal/cli/run.go` move into `internal/project`, including selector entry construction, diagnostics-adjacent rules, environment merging, and override warnings.
- [x] **ARCH-04**: Cross-package command orchestration that composes config, vault, export, setup, migrate, and manager helper behavior moves into `internal/app` services while CLI keeps prompts, flags, output routing, completions, and process lifecycle.
- [x] **ARCH-05**: `internal/cli` remains a Cobra adapter layer and does not own reusable behavior needed by tests, the manager, or future UX surfaces.
- [x] **ARCH-06**: Tests are rebalanced so behavior-rule coverage lives beside the owning domain/app package, while CLI tests cover command contracts, completions, output channels, error wording, and a small number of smoke workflows.

### Documentation Cleanup

- [x] **DOC-01**: User-facing docs describe manager editing, direct tag list/export, and project tag bindings.
- [x] **DOC-02**: Developer docs describe install/tag/release scripts and the final internal architecture so maintainers do not rely on remembered manual commands or stale package maps.

### Scope Boundary and Release

- [x] **BOUND-01**: v0.1.1 does not introduce fine-grained CLI metadata-editing command groups such as `secret meta` or `secret tag`; full editing remains centered in the manager and existing compact secret commands.
- [x] **BOUND-02**: v0.1.1 keeps the current age-encrypted JSON vault format and does not implement or spike SQLite; storage redesign is deferred to v0.2.0 planning.
- [x] **REL-011-01**: v0.1.1 release readiness is checked only after architecture and documentation cleanup are complete.

## Deferred Requirements

Tracked for future releases. These are not current implementation commitments.

### Storage v0.2.0 Candidates

- **V2-STORE-01**: Reconsider SQLite or another storage model only in v0.2.0 planning, after defining the threat model, artifact leakage checklist, migration path, and release-build impact.
- **V2-STORE-02**: If SQLite is reconsidered, compare plaintext SQLite, SQLCipher/encrypted SQLite, and age-wrapped in-memory SQLite against portability, WAL/journal/temp leakage, and chezmoi sync constraints before implementation.

### Project Sessions

- **V2-SES-01**: User can activate the current project environment in the current shell through `shelf project activate` after installing a shell hook, if future UX evidence justifies hook complexity.
- **V2-SES-02**: User can restore the previous shell environment through `shelf project deactivate` without losing pre-existing env values, if activation is implemented.
- **V2-SES-03**: User can enter an isolated project environment through `shelf project shell` only if it proves clearer than `shelf project run -- $SHELL`.

### Vault Expansion

- **V2-VAULT-01**: User can use password-only encryption if they do not use age keys.
- **V2-VAULT-02**: User can manage multiple vaults or profiles.
- **V2-VAULT-03**: User can inspect non-secret age recipient metadata from a vault.
- **V2-VAULT-04**: User gets better merge/conflict handling for git-managed encrypted vault updates.

### Integrations and Editing

- **V2-INT-01**: Shelf can offer optional direct chezmoi helper commands.
- **V2-INT-02**: Shelf can support hardware-key or age plugin workflows if users need them.
- **V2-UI-01**: Shelf can provide a TUI or improved field-specific editor when that improves local editing speed without weakening vault safety.

### Release and Distribution

- **V2-REL-01**: Shelf can add optional package-manager distribution after initial usage validates demand.
- **V2-REL-02**: Shelf can add native Windows smoke tests for setup, secret set/get, and project run on a real Windows runner.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Team sharing | The current product is for solo developers; sharing requires identity, permissions, revocation, audit, and conflict semantics. |
| Hosted sync service | Shelf should stay local-first and portable instead of requiring a backend account. |
| Permanent daemon | Core CLI workflows should not depend on a long-running process; the manager should be short-lived/on-demand. |
| Browser extension or autofill | Shelf is focused on developer secrets and env workflows, not general password-manager replacement. |
| Plain `.env` as source of truth | `.env` files may be generated/exported, but Shelf's source of truth is the encrypted vault plus project manifests. |
| New dotenv export format | Existing `shell` output is already sourceable; adding another format increases surface area without enough value. |
| Hook-based project activation in current scope | Shell hooks mutate parent-shell state implicitly and add complexity; explicit export/source workflows are preferred for now. |
| Dedicated vault restore command | Current backups are ordinary encrypted vault files and single-slot only; a command adds surface area without enough value. Manual copy plus `shelf vault status` is simpler. |
| SQLite in v0.1.1 | v0.1.1 is about editing UX and tag workflows; storage model redesign is deferred to v0.2.0. |
| Fine-grained CLI metadata edit subcommands | The manager is the primary editing surface; CLI should stay compact and focus on scriptable application workflows. |
| Broad one-file-per-command CLI split | `internal/cli` should stay command-family oriented rather than becoming a large directory of tiny files. |
| Release hardening as next phase | Architecture and docs cleanup must happen before v0.1.1 release readiness. |
| Compatibility alias for old manager command | The project is pre-release; keeping both `shelf manager` and `shelf vault open` would make one feature have two names. |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| WEB-01..WEB-06 | Phase 17, Phase 18 | Complete |
| TAG-01..TAG-02 | Phase 19 | Complete |
| TAG-03..TAG-05 | Phase 20 | Complete |
| OPS-01..OPS-03 | Phase 21 | Complete |
| ARCH-01..ARCH-02 | Phase 22 | Complete |
| DOC-01..DOC-02 | Phase 23 | Complete |
| ARCH-03 | Phase 25 | Complete |
| ARCH-04 | Phase 26 | Complete |
| ARCH-05 | Phase 25..Phase 27 | Complete |
| ARCH-06 | Phase 27 | Complete |
| BOUND-01 | Phase 17..Phase 27 | Complete |
| BOUND-02 | Phase 17..Phase 27 | Complete |
| REL-011-01 | Phase 24 | Complete |

**Coverage:**
- Current requirements: 27 total
- Mapped to phases: 27
- Unmapped: 0
- Completed in v0.1.1 so far: WEB-01..WEB-06, TAG-01..TAG-05, OPS-01..OPS-03, ARCH-01..ARCH-06, DOC-01..DOC-02, BOUND-01..BOUND-02, REL-011-01
- Completed v0.1.0 requirements: archived at `.planning/archive/releases/v0.1.0/SUMMARY.md`

---
*Last updated: 2026-06-28 after completing Phase 27 CLI test rebalancing*
