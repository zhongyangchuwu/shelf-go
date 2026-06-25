# Roadmap: Pre-release Architecture Refactor

## Overview

Shelf has completed the encrypted vault baseline, command hierarchy cutover, project workflow compatibility, vault UX hardening, safety hardening, release-readiness docs, and minimal project env UX. The next pre-release milestone reduces architecture friction before the first public release by moving reusable behavior out of `internal/cli` while keeping the CLI package compact and command-family oriented.

## Architecture Direction

### Layers

```text
Top/display:      cmd/shelf, internal/cli, internal/manager
Feature support:  internal/app, internal/project, later internal/vault and internal/secret
Base support:     internal/config, internal/store, internal/manifest, internal/render, internal/version, later internal/atomicfile
```

### Decisions

- Keep `internal/cli` at roughly 3-6 files, grouped by command family, not one file per command.
- Move reusable application/runtime construction to `internal/app`.
- Move project identity and manifest resolution to `internal/project`.
- Do not create `internal/gitutil` in the first extraction. Project-owned git helpers live in `internal/project`; shared git utilities can be extracted later only if both project and vault diagnostics need them.
- Defer storage backend interfaces until a real second backend spike begins.

## Phases

- [x] Phase 13: App Runtime and Project Package Extraction
- [x] Phase 14: Vault Diagnostics and Secret Workflow Extraction
- [x] Phase 15: Shared Persistence Primitives and Store File Layout

## Phase Details

### Phase 13: App Runtime and Project Package Extraction

**Goal:** Move reusable runtime/vault loading and project resolution behavior out of `internal/cli` without changing command behavior.

**Depends on:** Completed safety and minimal project env UX milestone.

**Requirements:** ARCH-01, ARCH-02, ARCH-03

**Success Criteria:**
1. `internal/app` owns runtime/vault load, read, and update helpers currently in `internal/cli/root.go`.
2. `internal/project` owns manifest resolution, project diagnostics, render binding conversion, project ID, git root lookup, and remote normalization currently in `internal/cli/project.go`.
3. `internal/cli` remains command-family oriented and does not split into one file per subcommand.
4. Existing project, run, and full test suites pass.

**Plan:** `.planning/phases/013-architecture-package-boundaries/PLAN.md`

### Phase 14: Vault Diagnostics and Secret Workflow Extraction

**Goal:** Move reusable vault status/doctor diagnostics and plaintext edit workflow out of `internal/cli`.

**Depends on:** Phase 13 complete.

**Requirements:** ARCH-04, ARCH-05

**Success Criteria:**
1. `internal/vault` owns vault status/check/doctor diagnostic rules and returns typed diagnostic records for CLI rendering.
2. `internal/secret` owns `secret edit` editable JSON and temp-file/editor lifecycle.
3. `internal/cli/vault.go`, `internal/cli/doctor.go`, and `internal/cli/secret.go` remain thin command-family files.
4. Vault, doctor, manager, and secret edit tests pass.

**Plan:** `.planning/phases/014-vault-secret-extraction/PLAN.md`

### Phase 15: Shared Persistence Primitives and Store File Layout

**Goal:** Remove duplicated persistence primitives and make `internal/store` easier to evolve without introducing speculative backend interfaces.

**Depends on:** Phase 14 complete.

**Requirements:** ARCH-06, ARCH-07, ARCH-08

**Success Criteria:**
1. Atomic write behavior is centralized in a small shared helper with explicit mode, sync, and backup options.
2. Env name and path token validation have one canonical implementation.

3. `internal/store` separates store model/methods, JSON encode/decode, age seal/open, and vault orchestration into clearer files within the same package.
4. Store, manifest, render, setup, and full test suites pass.

**Plan:** `.planning/phases/015-persistence-store-layout/PLAN.md`

## Future Candidates

- SQLite storage spike: investigate SQLite as an encrypted vault payload or metadata/search layer only if JSON schema/search/history pressure becomes real. Any design must preserve encrypted-at-rest safety and avoid plaintext SQLite WAL, journal, or temp files.
- Dolt is not a current vault-storage candidate: it is powerful for versioned SQL data, but too heavy for Shelf's portable encrypted-file model and weakens useful diff/history unless secrets or metadata are exposed.
- `internal/gitutil`: create only if both `internal/project` and future `internal/vault` need shared git subprocess helpers.

## Explicit Non-Goals for This Milestone

- No command behavior changes.
- No `project activate` / `project deactivate` shell hook implementation.
- No `project shell` wrapper unless a later phase shows clear value over `project run -- $SHELL`.
- No new `dotenv` format; use existing `shell` output for sourceable files.
- No team sharing, hosted sync, permanent daemon, or release packaging work in this milestone.
- No SQLite backend implementation.
- No speculative repository/service abstraction beyond packages required by the current code.

## Progress

| Phase | Status | Requirements | Plan | Completion Date |
|-------|--------|--------------|------|-----------------|
| Phase 13: App Runtime and Project Package Extraction | Complete | ARCH-01..ARCH-03 | `.planning/phases/013-architecture-package-boundaries/PLAN.md` | 2026-06-25 |
| Phase 14: Vault Diagnostics and Secret Workflow Extraction | Complete | ARCH-04..ARCH-05 | `.planning/phases/014-vault-secret-extraction/PLAN.md` | 2026-06-25 |
| Phase 15: Shared Persistence Primitives and Store File Layout | Complete | ARCH-06..ARCH-08 | `.planning/phases/015-persistence-store-layout/PLAN.md` | 2026-06-25 |

---
*Last updated: 2026-06-25 after completing Phase 15 shared persistence primitives and store layout*
