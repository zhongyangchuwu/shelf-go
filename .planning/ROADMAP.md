# Roadmap: Safety and Minimal Project Env UX

## Overview

Shelf already has the encrypted vault baseline, scoped command hierarchy, project manifests, project runtime injection, direct export, doctor/status diagnostics, and localhost vault manager. The next pre-release milestone keeps the product small: avoid shell hooks and session wrappers, prefer explicit file/source workflows, and prioritize safety hardening plus recoverability before release infrastructure.

## Phases

- [x] Phase 9: Project Export Shell Default
- [x] Phase 10: Vault Restore and Recovery Docs
- [x] Phase 11: Secret Edit and Manager Safety Hardening

## Phase Details

### Phase 9: Project Export Shell Default

**Goal:** Make the default project export output directly sourceable without adding a new dotenv format or shell hook workflow.

**Depends on:** Completed command hierarchy and project workflow compatibility.

**Requirements:** PUX-01, PUX-02, PUX-03

**Success Criteria:**
1. `shelf project export` defaults to the existing `shell` format, matching `shelf secret export`.
2. `--format env|shell|json` remains available; no `dotenv` format is added.
3. User docs recommend explicit workflows such as `shelf project export > .env.local` and `source .env.local`, with plaintext and git-ignore warnings.
4. Tests cover the default format and preserve explicit format behavior.

**Plans:** `.planning/phases/009-project-export-shell-default/PLAN.md`

### Phase 10: Vault Restore and Recovery Docs

**Goal:** Make encrypted backup recovery explicit and testable.

**Depends on:** Phase 9 complete.

**Requirements:** VREC-01, VREC-02, VREC-03

**Success Criteria:**
1. User can restore a validated encrypted vault backup through a vault-scoped command or documented manual flow.
2. Restore refuses unsafe overwrite by default and validates decrypted store contents before replacing a vault.
3. Troubleshooting and security docs explain identity loss, backup restore, and post-restore `shelf vault status` verification.
4. Tests cover restore success, overwrite refusal, invalid backup rejection, and value-free diagnostics.

**Plans:** `.planning/phases/010-vault-restore-recovery/PLAN.md`

### Phase 11: Secret Edit and Manager Safety Hardening

**Goal:** Reduce plaintext exposure in interactive editing and local manager workflows without adding a daemon or complex UI.

**Depends on:** Phase 10 complete.

**Requirements:** SAFE-EDIT-01, SAFE-MGR-01, SAFE-DOC-01

**Success Criteria:**
1. `shelf secret edit` temporary files use restrictive permissions and are cleaned on success and failure paths where possible.
2. Local manager safety gaps are either cheaply hardened or documented explicitly; no permanent daemon is introduced.
3. Docs name remaining plaintext boundaries and recommended close/cleanup behavior.
4. Focused tests cover temp-file permissions/cleanup and any manager behavior changes.

**Plans:** `.planning/phases/011-edit-manager-safety/PLAN.md`

## Explicit Non-Goals for This Milestone

- No `project activate` / `project deactivate` shell hook implementation.
- No `project shell` wrapper unless a later phase shows clear value over `project run -- $SHELL`.
- No new `dotenv` format; use existing `shell` output for sourceable files.
- No team sharing, hosted sync, permanent daemon, or release packaging work in this milestone.

## Progress

| Phase | Status | Requirements | Plans | Completion Date |
|-------|--------|--------------|-------|-----------------|
| Phase 9: Project Export Shell Default | Complete | PUX-01..PUX-03 | `.planning/phases/009-project-export-shell-default/PLAN.md` | 2026-06-24 |
| Phase 10: Vault Restore and Recovery Docs | Complete | VREC-01..VREC-03 | `.planning/phases/010-vault-restore-recovery/PLAN.md` | 2026-06-24 |
| Phase 11: Secret Edit and Manager Safety Hardening | Complete | SAFE-EDIT-01, SAFE-MGR-01, SAFE-DOC-01 | `.planning/phases/011-edit-manager-safety/PLAN.md` | 2026-06-24 |

---
*Last updated: 2026-06-24 after selecting the safety and minimal project env UX milestone*
