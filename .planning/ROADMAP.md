# Roadmap: Shelf Go Command Hierarchy and Vault UX

## Overview

Shelf already has the encrypted-vault baseline: age-encrypted storage, migration, project manifests, runtime injection, direct export, doctor checks, and a localhost vault manager. The next pre-release milestone simplifies the CLI before any public compatibility burden exists: command names must expose scope, project-dependent workflows must live under `shelf project`, vault lifecycle must live under `shelf vault`, and future activate/deactivate work must be planned before implementation.

## Phases

- [x] Phase 6: Command Hierarchy Cutover
- [x] Phase 7: Vault UX Hardening
- [x] Phase 8: Project Session Design

## Phase Details

### Phase 6: Command Hierarchy Cutover

**Goal:** Replace ambiguous top-level commands with scoped command namespaces before release.

**Depends on:** Completed encrypted-vault milestone.

**Requirements:** CMD-01, CMD-02, CMD-03, CMD-04, CMD-05, CMD-06, CMD-07, CMD-08

**Success Criteria:**
1. `shelf setup` performs the current global config/vault onboarding behavior.
2. `shelf vault init`, `shelf vault migrate`, and `shelf vault open` perform the current vault init, migration, and local manager behavior.
3. `shelf secret export` performs current direct path/prefix export behavior.
4. `shelf project run -- ...` performs current `.shelf.json` runtime injection and dry-run behavior.
5. Old top-level `init`, `migrate`, `export`, `run`, and `manager` commands are absent from the root command list.
6. README and usage docs present only the new canonical command hierarchy.

**Plans:** `.planning/phases/006-command-hierarchy-cutover/PLAN.md`

### Phase 7: Vault UX Hardening

**Goal:** Improve vault-specific usability and diagnostics without changing the encrypted storage contract.

**Depends on:** Phase 6 complete.

**Requirements:** VUX-01, VUX-02, VUX-03, VUX-04

**Success Criteria:**
1. A vault-scoped status/check command reports config path, vault path, file format, loadability, and safe next steps without revealing values.
2. Missing recipients, missing identities, plaintext legacy stores, unsupported vault formats, and undecryptable vaults produce concise recovery guidance.
3. Docs explain first-run setup, vault init, vault migrate, vault status/check, vault open, and plaintext cleanup.
4. Verification covers encrypted load/save, migration, status/check behavior, doctor behavior, and manager write safety under the new command hierarchy.

**Plans:** `.planning/phases/007-vault-ux-hardening/PLAN.md`

### Phase 8: Project Session Design

**Goal:** Plan venv-like project session workflows without implementing activate/deactivate/shell yet.

**Depends on:** Phase 6 complete.

**Requirements:** SES-01, SES-02, SES-03, SES-04

**Success Criteria:**
1. A design artifact defines `shelf project activate`, `shelf project deactivate`, and `shelf project shell` semantics.
2. The design records why activation requires a shell hook/function to mutate the current shell environment.
3. The design specifies reversible env restore behavior for variables that existed before activation.
4. The design defines no-value dry-run/preview output and conflict handling for repeated activation or project switching.
5. Implementation remains explicitly out of scope for this phase.

**Plans:** `.planning/phases/008-project-session-design/PLAN.md`

## Progress

| Phase | Status | Requirements | Plans | Completion Date |
|-------|--------|--------------|-------|-----------------|
| Phase 6: Command Hierarchy Cutover | Complete | CMD-01..CMD-08 | `.planning/phases/006-command-hierarchy-cutover/PLAN.md` | 2026-06-22 |
| Phase 7: Vault UX Hardening | Complete | VUX-01..VUX-04 | `.planning/phases/007-vault-ux-hardening/PLAN.md` | 2026-06-23 |
| Phase 8: Project Session Design | Complete | SES-01..SES-04 | `.planning/phases/008-project-session-design/PLAN.md` | 2026-06-23 |

---
*Last updated: 2026-06-23 after completing command hierarchy, vault UX, and project session design phases*
