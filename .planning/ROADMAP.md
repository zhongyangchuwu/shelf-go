# Roadmap: Shelf Go v0.1.1

## Overview

Shelf Go v0.1.1 improves the day-to-day editing and selection experience without changing the storage model. The release rebuilds `shelf vault open` into a practical local secret console, adds tag-based secret selection for CLI exports, and lets projects bind tag-selected secret sets while keeping secret values out of manifests.

Before release, v0.1.1 still needs workflow/script cleanup, documentation updates, and architecture naming cleanup. Release hardening is intentionally the final phase, not the next phase.

SQLite and storage backend redesign are explicitly deferred to v0.2.0.

## Phases

- [x] Phase 17: Web Manager Design Contract
- [x] Phase 18: Web Manager Editing Console
- [x] Phase 19: Secret Tag Selection
- [x] Phase 20: Project Tag Bindings
- [x] Phase 21: Script Workflow Consolidation
- [ ] Phase 22: Documentation and Architecture Cleanup
- [ ] Phase 23: v0.1.1 Release Hardening

## Phase Details

### Phase 17: Web Manager Design Contract

**Goal:** Define the Web manager UX, visual direction, security boundaries, and implementation constraints before rebuilding the UI.

**Depends on:** v0.1.0 release archive complete.

**Requirements:** WEB-01..WEB-06, BOUND-01, BOUND-02

**Success Criteria:**
1. WebUI design contract describes list/search, add, edit, delete, reveal, copy, hide, tag filtering, and tag editing flows.
2. Visual direction is selected from reusable console/admin template references, with local embedded assets and no CDN requirement.
3. Technical direction keeps `net/http` and single-binary Go distribution; SPA and broad web frameworks are rejected unless explicitly re-approved.
4. Safety contract covers token URL cleanup, loopback/token/Host/Origin checks, no-store secret responses, and no persistent browser storage for secret values.

**Plan:** `.planning/phases/017-web-manager-design/PLAN.md`

### Phase 18: Web Manager Editing Console

**Goal:** Rebuild `shelf vault open` as the main secret editing surface.

**Depends on:** Phase 17 complete.

**Requirements:** WEB-01..WEB-06, BOUND-01

**Success Criteria:**
1. Manager UI lists and searches secrets by path, env, description, and tags without returning values in list responses.
2. Manager UI supports add, edit/rename, delete, tag editing, explicit reveal, hide, and copy workflows.
3. Existing manager safety tests remain covered and new tests cover token redirect, no-store secret responses, POST reveal/copy, and no value leakage in list/search responses.
4. Assets are embedded locally and release builds remain single-binary friendly.

**Plan:** `.planning/phases/018-web-manager-editing-console/PLAN.md`

### Phase 19: Secret Tag Selection

**Goal:** Add tag-based secret selection to compact CLI workflows without adding fine-grained metadata editing command groups.

**Depends on:** Phase 18 can run in parallel after shared tag selector semantics are agreed, but should not depend on Web UI internals.

**Requirements:** TAG-01, TAG-02, TAG-05, BOUND-01

**Success Criteria:**
1. `shelf secret list --tag` filters by one or more tags and remains value-free.
2. `shelf secret export --tag` exports tag-selected secrets in existing env, shell, and JSON formats.
3. Multiple `--tag` filters use AND semantics and deterministic sorted output.
4. Existing path/prefix export behavior, `--all`, and no-new-dotenv boundary remain unchanged.

**Plan:** `.planning/phases/019-secret-tag-selection/PLAN.md`

### Phase 20: Project Tag Bindings

**Goal:** Let `.shelf.json` bind tag-selected secret sets for project export/run workflows without storing values.

**Depends on:** Phase 19 complete.

**Requirements:** TAG-03, TAG-04, TAG-05

**Success Criteria:**
1. Manifest schema supports tag-selected entries with path/prefix/tag forms mutually exclusive.
2. `shelf project add --tag` records value-free tag bindings.
3. `project list`, `explain`, `export`, and `run` expand tag bindings with clear missing and conflict diagnostics.
4. Dynamic tag binding behavior is documented and covered by command tests.

**Plan:** `.planning/phases/020-project-tag-bindings/PLAN.md`

### Phase 21: Script Workflow Consolidation

**Goal:** Move common developer/release flows out of ad-hoc manual commands and inline `justfile` recipes into reusable scripts under `scripts/`.

**Depends on:** Phase 20 complete.

**Requirements:** OPS-01, OPS-02, OPS-03

**Success Criteria:**
1. Install flow currently embedded in `justfile` is captured as a script under `scripts/` and `just install` delegates to it.
2. Tag/release flow is captured as scripts under `scripts/`, including the existing tag command and GoReleaser checks/snapshot workflow.
3. Scripts are portable Bash with `set -euo pipefail`, predictable arguments, and concise usage errors.
4. `justfile` remains a thin task runner rather than the source of workflow logic.
5. Script behavior is documented enough for maintainers to release without relying on remembered manual commands.

**Plan:** `.planning/phases/021-script-workflow-consolidation/PLAN.md`

### Phase 22: Documentation and Architecture Cleanup

**Goal:** Improve docs and clean up naming/package architecture before release, especially around the Web manager naming mismatch.

**Depends on:** Phase 21 complete.

**Requirements:** DOC-01, DOC-02, ARCH-01, ARCH-02

**Success Criteria:**
1. User-facing docs describe Web manager editing, tag-based direct secret workflows, and project tag bindings.
2. Developer docs describe scripts and release workflow after Phase 21.
3. Architecture review resolves the mismatch between command name `shelf vault open`, package path `internal/manager`, and user-facing function naming.
4. Any package/command naming changes keep CLI behavior intentional, documented, and covered by tests.
5. The `internal/manager` placement is either justified in architecture docs or refactored to a clearer package boundary.

**Plan:** TBD.

### Phase 23: v0.1.1 Release Hardening

**Goal:** Prepare v0.1.1 for release only after scripts, docs, and architecture cleanup are complete.

**Depends on:** Phases 18, 19, 20, 21, and 22 complete.

**Requirements:** WEB-01..WEB-06, TAG-01..TAG-05, OPS-01..OPS-03, DOC-01..DOC-02, ARCH-01..ARCH-02, BOUND-01..BOUND-02, REL-011-01

**Success Criteria:**
1. README and docs reflect the implemented Web manager, tag workflows, scripts, and architecture names.
2. CHANGELOG has a `0.1.1` section.
3. Go tests, vet, release check, and snapshot release pass through the consolidated scripts.
4. Verification records confirm no storage format change and SQLite deferral to v0.2.0.
5. Release readiness does not rely on manual commands that should live in scripts.

**Plan:** TBD.

## Future Candidates

- SQLite/storage redesign for v0.2.0: reconsider only after defining threat model, artifact leakage checklist, migration path, and release-build impact.
- Native Windows smoke tests: verify `shelf setup`, secret set/get, and project run on a real Windows runner.
- Password-only encryption: consider only if users need a no-age-key workflow and the threat model remains clear.
- Multiple vaults or profiles: consider after single-vault workflows show concrete pressure.
- Chezmoi helper commands: consider optional integration while preserving Shelf's portable encrypted-file model.
- Package-manager distribution: consider Homebrew/Scoop or similar after initial usage validates demand.
- `internal/gitutil`: create only if both `internal/project` and future `internal/vault` need shared git subprocess helpers.

## Explicit Non-Goals for v0.1.1

- No SQLite implementation, SQLite spike, or storage backend abstraction.
- No new vault file format.
- No `secret meta` or `secret tag` command group; WebUI is the primary editing surface.
- No SPA requirement, hosted frontend, CDN dependency, or permanent daemon.
- No `project activate` / `project deactivate` shell hook implementation.
- No `project shell` wrapper unless a later phase shows clear value over `project run -- $SHELL`.
- No new `dotenv` format; use existing `shell` output for sourceable files.
- No team sharing or hosted sync.
- No release hardening before script, docs, and architecture cleanup phases complete.

## Progress

| Phase | Status | Requirements | Plan | Completion Date |
|-------|--------|--------------|------|-----------------|
| Phase 17: Web Manager Design Contract | Complete | WEB-01..WEB-06, BOUND-01..BOUND-02 | `.planning/phases/017-web-manager-design/PLAN.md` | 2026-06-26 |
| Phase 18: Web Manager Editing Console | Complete | WEB-01..WEB-06, BOUND-01 | `.planning/phases/018-web-manager-editing-console/PLAN.md` | 2026-06-26 |
| Phase 19: Secret Tag Selection | Complete | TAG-01..TAG-02, TAG-05, BOUND-01 | `.planning/phases/019-secret-tag-selection/PLAN.md` | 2026-06-26 |
| Phase 20: Project Tag Bindings | Complete | TAG-03..TAG-05 | `.planning/phases/020-project-tag-bindings/PLAN.md` | 2026-06-26 |
| Phase 21: Script Workflow Consolidation | Complete | OPS-01..OPS-03 | `.planning/phases/021-script-workflow-consolidation/PLAN.md` | 2026-06-26 |
| Phase 22: Documentation and Architecture Cleanup | Not Started | DOC-01..DOC-02, ARCH-01..ARCH-02 | TBD | - |
| Phase 23: v0.1.1 Release Hardening | Not Started | WEB-01..WEB-06, TAG-01..TAG-05, OPS-01..OPS-03, DOC-01..DOC-02, ARCH-01..ARCH-02, BOUND-01..BOUND-02, REL-011-01 | TBD | - |

## Archived Releases

- v0.1.0: `.planning/archive/releases/v0.1.0/SUMMARY.md`

---
*Last updated: 2026-06-26 after completing script workflow consolidation*
