# Roadmap: Shelf Go v0.1.1

## Overview

Shelf Go v0.1.1 improves the day-to-day editing and selection experience without changing the storage model. The release adds a local manager surface, tag-based secret selection for CLI exports, and project tag bindings while keeping secret values out of manifests.

v0.1.1 release hardening is complete and remains ready for review/tagging. The newly added boundary-refactor phases are behavior-preserving internal cleanup that should run only if selected before the next tag/release; they keep the current age-encrypted JSON vault model and do not add user-visible features.

SQLite and storage backend redesign are explicitly deferred to v0.2.0.

## Phases

- [x] Phase 17: Web Manager Design Contract
- [x] Phase 18: Web Manager Editing Console
- [x] Phase 19: Secret Tag Selection
- [x] Phase 20: Project Tag Bindings
- [x] Phase 21: Script Workflow Consolidation
- [x] Phase 22: Architecture Repartition Core
- [x] Phase 23: Documentation and Usage Alignment
- [x] Phase 24: v0.1.1 Release Hardening
- [x] Phase 25: CLI Project Boundary Refactor
- [x] Phase 26: App Service Extraction
- [ ] Phase 27: CLI Test Rebalancing and Boundary Verification

## Phase Details

### Phase 17: Web Manager Design Contract

**Goal:** Define the Web manager UX, visual direction, security boundaries, and implementation constraints before rebuilding the UI.

**Depends on:** v0.1.0 release archive complete.

**Requirements:** WEB-01..WEB-06, BOUND-01, BOUND-02

**Plan:** `.planning/phases/017-web-manager-design/PLAN.md`

### Phase 18: Web Manager Editing Console

**Goal:** Rebuild the local manager as the main secret editing surface.

**Depends on:** Phase 17 complete.

**Requirements:** WEB-01..WEB-06, BOUND-01

**Plan:** `.planning/phases/018-web-manager-editing-console/PLAN.md`

### Phase 19: Secret Tag Selection

**Goal:** Add tag-based secret selection to compact CLI workflows without adding fine-grained metadata editing command groups.

**Depends on:** Phase 18 can run in parallel after shared tag selector semantics are agreed, but should not depend on Web UI internals.

**Requirements:** TAG-01, TAG-02, TAG-05, BOUND-01

**Plan:** `.planning/phases/019-secret-tag-selection/PLAN.md`

### Phase 20: Project Tag Bindings

**Goal:** Let `.shelf.json` bind tag-selected secret sets for project export/run workflows without storing values.

**Depends on:** Phase 19 complete.

**Requirements:** TAG-03, TAG-04, TAG-05

**Plan:** `.planning/phases/020-project-tag-bindings/PLAN.md`

### Phase 21: Script Workflow Consolidation

**Goal:** Move common developer/release flows out of ad-hoc manual commands and inline `justfile` recipes into reusable scripts under `scripts/`.

**Depends on:** Phase 20 complete.

**Requirements:** OPS-01, OPS-02, OPS-03

**Plan:** `.planning/phases/021-script-workflow-consolidation/PLAN.md`

### Phase 22: Architecture Repartition Core

**Goal:** Cleanly repartition internal packages and replace the vault-scoped manager command with a single local manager entrypoint.

**Depends on:** Phase 21 complete.

**Requirements:** ARCH-01, ARCH-02, BOUND-01, BOUND-02

**Success Criteria:**
1. `shelf manager` is the only manager command entrypoint; `shelf vault open` is removed.
2. `internal/manager` remains the local manager surface package and is not limited to vault-only features.
3. Project manifest schema/IO/validation moves into `internal/project`.
4. Encrypted vault core, diagnostics, locking, age, JSON, and atomic write support live under `internal/vault`.
5. Version composition moves into `internal/app`.
6. Export env/shell/JSON formatting moves from `internal/render` to `internal/exportfmt`.
7. Behavior remains unchanged apart from the intentional manager command rename.

**Plan:** `.planning/phases/022-architecture-repartition-core/PLAN.md`

### Phase 23: Documentation and Usage Alignment

**Goal:** Update user and developer docs after the architecture and command naming cutover.

**Depends on:** Phase 22 complete.

**Requirements:** DOC-01, DOC-02, ARCH-01, ARCH-02, BOUND-01, BOUND-02

**Success Criteria:**
1. User-facing docs describe manager editing, tag-based direct secret workflows, and project tag bindings.
2. Developer docs describe scripts and release workflow after Phase 21.
3. Architecture docs describe the final Phase 22 package layout and manager command naming.
4. Docs no longer treat `shelf vault open` as the primary manager entrypoint.

**Plan:** `.planning/phases/023-documentation-and-usage-alignment/PLAN.md`

### Phase 24: v0.1.1 Release Hardening

**Goal:** Prepare v0.1.1 for release only after architecture and docs cleanup are complete.

**Depends on:** Phases 18, 19, 20, 21, 22, and 23 complete.

**Requirements:** WEB-01..WEB-06, TAG-01..TAG-05, OPS-01..OPS-03, DOC-01..DOC-02, ARCH-01..ARCH-02, BOUND-01..BOUND-02, REL-011-01

**Success Criteria:**
1. README and docs reflect the implemented manager, tag workflows, scripts, and architecture names.
2. CHANGELOG has a `0.1.1` section.
3. Go tests, vet, release check, and snapshot release pass through the consolidated scripts.
4. Verification records confirm no storage format change and SQLite deferral to v0.2.0.
5. Release readiness does not rely on manual commands that should live in scripts.

**Plan:** `.planning/phases/024-v0.1.1-release-hardening/PLAN.md`

### Phase 25: CLI Project Boundary Refactor

**Goal:** Move project/session business rules out of `internal/cli` so CLI project commands become thin Cobra adapters over `internal/project` domain services.

**Depends on:** Phase 24 complete, unless the release-ready commit is tagged first and this work moves to the next milestone.

**Requirements:** ARCH-03, ARCH-05, BOUND-01, BOUND-02

**Success Criteria:**
1. `internal/project` owns project entry construction for path, prefix, and tag selectors, including optional/required state and validation against the vault store.
2. `internal/project` owns project environment utilities currently in CLI, including child env merging and parent env override warnings.
3. `internal/cli/project.go` and `internal/cli/run.go` keep Cobra wiring, flags, completions, output routing, and child process execution only.
4. Project-domain tests cover selector, diagnostic, env conflict, prefix/tag expansion, and environment-merge rules without invoking Cobra.
5. Existing project/run CLI behavior remains unchanged.

**Plan:** `.planning/phases/025-cli-project-boundary-refactor/PLAN.md`

### Phase 26: App Service Extraction

**Goal:** Move command orchestration that composes config, vault, project, export, setup, migrate, and manager helpers into `internal/app` services with result structs or strings returned to CLI.

**Depends on:** Phase 25 complete.

**Requirements:** ARCH-04, ARCH-05, BOUND-01, BOUND-02

**Success Criteria:**
1. `secret export` selection/filter/format orchestration lives behind an `internal/app` service and no longer reads vault internals directly from CLI.
2. Setup/vault init reusable file-creation and identity/vault/config orchestration live outside CLI while interactive prompts remain in CLI.
3. Migration and manager helper logic that does not depend on Cobra moves out of CLI.
4. CLI handlers call app services and print returned outputs/errors without owning reusable business rules.
5. Existing setup, vault, export, migrate, manager, and doctor behavior remains unchanged.

**Plan:** `.planning/phases/026-app-service-extraction/PLAN.md`

### Phase 27: CLI Test Rebalancing and Boundary Verification

**Goal:** Rebalance tests so `internal/cli` protects command contracts while domain/app packages own behavior-rule coverage.

**Depends on:** Phase 26 complete.

**Requirements:** ARCH-05, ARCH-06, BOUND-01, BOUND-02

**Success Criteria:**
1. CLI tests cover command wiring, flags, completions, stdout/stderr contracts, error wording, and a small number of end-to-end smoke workflows.
2. Project/app/secret/vault tests cover reusable behavior directly without temp git repositories or Cobra unless those boundaries are required.
3. No behavior-only test remains in `internal/cli` solely because the implementation currently lives there.
4. Full verification confirms no user-visible CLI behavior, vault format, manifest format, or manager route changes.
5. Planning capture records the final package boundary and test ownership model.

**Plan:** TBD

## Future Candidates

- SQLite/storage redesign for v0.2.0: reconsider only after defining threat model, artifact leakage checklist, migration path, and release-build impact.
- Native Windows smoke tests: verify `shelf setup`, secret set/get, and project run on a real Windows runner.
- Password-only encryption: consider only if users need a no-age-key workflow and the threat model remains clear.
- Multiple vaults or profiles: consider after single-vault workflows show concrete pressure.
- Chezmoi helper commands: consider optional integration while preserving Shelf's portable encrypted-file model.
- Package-manager distribution: consider Homebrew/Scoop or similar after initial usage validates demand.

## Explicit Non-Goals for v0.1.1

- No SQLite implementation, SQLite spike, or storage backend abstraction.
- No new vault file format.
- No `secret meta` or `secret tag` command group; manager is the primary editing surface.
- No SPA requirement, hosted frontend, CDN dependency, or permanent daemon.
- No `project activate` / `project deactivate` shell hook implementation.
- No `project shell` wrapper unless a later phase shows clear value over `project run -- $SHELL`.
- No new `dotenv` format; use existing `shell` output for sourceable files.
- No team sharing or hosted sync.
- No release hardening before architecture and documentation cleanup phases complete.

## Progress

| Phase | Status | Requirements | Plan | Completion Date |
|-------|--------|--------------|------|-----------------|
| Phase 17: Web Manager Design Contract | Complete | WEB-01..WEB-06, BOUND-01..BOUND-02 | `.planning/phases/017-web-manager-design/PLAN.md` | 2026-06-26 |
| Phase 18: Web Manager Editing Console | Complete | WEB-01..WEB-06, BOUND-01 | `.planning/phases/018-web-manager-editing-console/PLAN.md` | 2026-06-26 |
| Phase 19: Secret Tag Selection | Complete | TAG-01..TAG-02, TAG-05, BOUND-01 | `.planning/phases/019-secret-tag-selection/PLAN.md` | 2026-06-26 |
| Phase 20: Project Tag Bindings | Complete | TAG-03..TAG-05 | `.planning/phases/020-project-tag-bindings/PLAN.md` | 2026-06-26 |
| Phase 21: Script Workflow Consolidation | Complete | OPS-01..OPS-03 | `.planning/phases/021-script-workflow-consolidation/PLAN.md` | 2026-06-26 |
| Phase 22: Architecture Repartition Core | Complete | ARCH-01..ARCH-02, BOUND-01..BOUND-02 | `.planning/phases/022-architecture-repartition-core/PLAN.md` | 2026-06-27 |
| Phase 23: Documentation and Usage Alignment | Complete | DOC-01..DOC-02, ARCH-01..ARCH-02, BOUND-01..BOUND-02 | `.planning/phases/023-documentation-and-usage-alignment/PLAN.md` | 2026-06-27 |
| Phase 24: v0.1.1 Release Hardening | Complete | WEB-01..WEB-06, TAG-01..TAG-05, OPS-01..OPS-03, DOC-01..DOC-02, ARCH-01..ARCH-02, BOUND-01..BOUND-02, REL-011-01 | `.planning/phases/024-v0.1.1-release-hardening/PLAN.md` | 2026-06-27 |
| Phase 25: CLI Project Boundary Refactor | Complete | ARCH-03, ARCH-05, BOUND-01..BOUND-02 | `.planning/phases/025-cli-project-boundary-refactor/PLAN.md` | 2026-06-28 |
| Phase 26: App Service Extraction | Complete | ARCH-04, ARCH-05, BOUND-01..BOUND-02 | `.planning/phases/026-app-service-extraction/PLAN.md` | 2026-06-28 |
| Phase 27: CLI Test Rebalancing and Boundary Verification | Planning | ARCH-05..ARCH-06, BOUND-01..BOUND-02 | TBD | - |

## Archived Releases

- v0.1.0: `.planning/archive/releases/v0.1.0/SUMMARY.md`

---
*Last updated: 2026-06-28 after completing Phase 26 app service extraction*
