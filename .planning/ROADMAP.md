# Roadmap: Shelf Go Encrypted Vault

**Created:** 2026-06-16
**Mode:** Vertical MVP
**Granularity:** Standard

## Overview

This roadmap turns the existing Shelf Go CLI into a git-safe, age-encrypted, portable secret manager for solo developers. Each phase should leave the product more usable while preserving the existing CLI workflows.

| Phase | Name | Goal | Requirements | UI Hint |
|-------|------|------|--------------|---------|
| 1 | Encrypted Vault Core | Existing secret commands can use an age-encrypted vault as durable storage. | VAULT-01, VAULT-02, VAULT-03, VAULT-04, VAULT-05, VAULT-06, CLI-01 | no |
| 2 | Migration and Git Safety | Users can safely migrate plaintext stores and verify git/chezmoi-safe state. | MIGR-01, MIGR-02, MIGR-03, MIGR-04, MIGR-05, SAFE-01, SAFE-02, SAFE-03, SAFE-04, SAFE-05 | no |
| 3 | Project Workflow Compatibility | Existing export, project, and run workflows work unchanged over encrypted storage. | CLI-02, CLI-03, CLI-04, CLI-05, TEST-01 | no |
| 4 | Localhost Vault Manager | Users can search, reveal, and edit secrets through a local manager safely. | WEB-01, WEB-02, WEB-03, WEB-04, WEB-05, WEB-06, WEB-07, TEST-02 | yes |
| 5 | Documentation and Release Hardening | The encrypted-vault workflow is documented, verified, and ready for real use. | DOCS-01, DOCS-02, DOCS-03 | no |

## Phases

### Phase 1: Encrypted Vault Core

**Goal:** Existing `shelf secret` workflows can read and write an age-encrypted vault file without changing command semantics.
**Mode:** mvp

**Requirements:** VAULT-01, VAULT-02, VAULT-03, VAULT-04, VAULT-05, VAULT-06, CLI-01

**Success Criteria:**
1. A user can configure Shelf to store secrets in an age-encrypted vault file.
2. Shelf config supports vault path, recipients, and identity locations without embedding private identity material.
3. `shelf secret set/get/list/info/edit/rm` can operate on the encrypted vault.
4. Wrong identity, missing identity, corrupt vault, and unsupported format errors are actionable.
5. The plaintext store model remains internal to load/decrypt and save/encrypt boundaries.

**Key Risks:**
- Plaintext backups or temp files can accidentally bypass the encrypted boundary.
- Recipient and identity configuration can become confusing if errors are too generic.

### Phase 2: Migration and Git Safety

**Goal:** Users can convert existing plaintext Shelf data into an encrypted portable vault and verify that git/chezmoi workflows are safe.
**Mode:** mvp

**Requirements:** MIGR-01, MIGR-02, MIGR-03, MIGR-04, MIGR-05, SAFE-01, SAFE-02, SAFE-03, SAFE-04, SAFE-05

**Success Criteria:**
1. A migration flow encrypts an existing plaintext JSON store and verifies the new vault by decrypting and validating it.
2. The source plaintext store remains untouched until the encrypted target is proven readable.
3. Backup and recovery artifacts containing secret values are encrypted.
4. `shelf doctor` distinguishes plaintext stores from encrypted vaults.
5. `shelf doctor` warns on tracked plaintext secret files and confirms encrypted/value-free states where possible.

**Key Risks:**
- Migration can create data-loss risk if it overwrites source files too early.
- Git safety checks can produce false confidence if they only inspect path names, not tracked state and format.

### Phase 3: Project Workflow Compatibility

**Goal:** The developer workflows that make Shelf useful for projects continue to work over encrypted storage.
**Mode:** mvp

**Requirements:** CLI-02, CLI-03, CLI-04, CLI-05, TEST-01

**Success Criteria:**
1. `shelf export` exact-path and prefix flows render env, shell, and JSON output from encrypted storage.
2. `shelf project` manifest commands continue to resolve paths and prefixes without storing values in `.shelf.json`.
3. `shelf run -- ...` injects resolved secrets into child processes from encrypted storage.
4. `shelf run --dry-run` preserves current no-secret-value output behavior.
5. Regression coverage proves current command semantics survive the storage change.

**Key Risks:**
- Project resolution can regress if encrypted load errors are not surfaced consistently.
- Value-printing rules can drift while commands are being reworked.

### Phase 4: Localhost Vault Manager

**Goal:** Users can search, reveal, copy, create, update, and delete secrets through a localhost-only manager without introducing unsafe write paths.
**Mode:** mvp

**Requirements:** WEB-01, WEB-02, WEB-03, WEB-04, WEB-05, WEB-06, WEB-07, TEST-02

**UI hint:** yes

**Success Criteria:**
1. A CLI command starts a loopback-only local vault manager without requiring a permanent daemon.
2. The manager supports search and browsing of paths and non-secret metadata.
3. Secret values are revealed or copied only through intentional user actions.
4. Create, update, and delete actions reuse existing validation, locking, and encrypted-save behavior.
5. State-changing routes use session/write-safety controls such as tokenized access, CSRF protection, and Origin/Host validation.

**Key Risks:**
- Localhost browser requests can still be forged by malicious pages if write controls are weak.
- UI edit paths can duplicate validation logic and diverge from CLI behavior.

### Phase 5: Documentation and Release Hardening

**Goal:** The encrypted-vault workflow is clear enough to use safely and has final verification evidence for release.
**Mode:** mvp

**Requirements:** DOCS-01, DOCS-02, DOCS-03

**Success Criteria:**
1. Documentation explains the encrypted vault model, age recipients, identity configuration, and chezmoi-friendly usage.
2. Documentation clearly separates config, `.shelf.json`, encrypted vault data, and generated/exported env files.
3. Documentation warns about plaintext exports, terminal output, browser reveal/copy actions, and old plaintext store cleanup.
4. Release verification confirms every v1 requirement is implemented or explicitly deferred through a documented decision.

**Key Risks:**
- Users can still misuse the tool if migration cleanup and plaintext export risks are under-documented.
- Docs can drift from the final CLI flags and config fields if written before implementation settles.

## Coverage

| Requirement | Phase | Status |
|-------------|-------|--------|
| VAULT-01 | Phase 1 | Pending |
| VAULT-02 | Phase 1 | Pending |
| VAULT-03 | Phase 1 | Pending |
| VAULT-04 | Phase 1 | Pending |
| VAULT-05 | Phase 1 | Pending |
| VAULT-06 | Phase 1 | Pending |
| CLI-01 | Phase 1 | Pending |
| MIGR-01 | Phase 2 | Pending |
| MIGR-02 | Phase 2 | Pending |
| MIGR-03 | Phase 2 | Pending |
| MIGR-04 | Phase 2 | Pending |
| MIGR-05 | Phase 2 | Pending |
| SAFE-01 | Phase 2 | Pending |
| SAFE-02 | Phase 2 | Pending |
| SAFE-03 | Phase 2 | Pending |
| SAFE-04 | Phase 2 | Pending |
| SAFE-05 | Phase 2 | Pending |
| CLI-02 | Phase 3 | Pending |
| CLI-03 | Phase 3 | Pending |
| CLI-04 | Phase 3 | Pending |
| CLI-05 | Phase 3 | Pending |
| TEST-01 | Phase 3 | Pending |
| WEB-01 | Phase 4 | Pending |
| WEB-02 | Phase 4 | Pending |
| WEB-03 | Phase 4 | Pending |
| WEB-04 | Phase 4 | Pending |
| WEB-05 | Phase 4 | Pending |
| WEB-06 | Phase 4 | Pending |
| WEB-07 | Phase 4 | Pending |
| TEST-02 | Phase 4 | Pending |
| DOCS-01 | Phase 5 | Pending |
| DOCS-02 | Phase 5 | Pending |
| DOCS-03 | Phase 5 | Pending |

**Coverage Summary:**
- v1 requirements: 33 total
- Mapped to phases: 33
- Unmapped: 0

---
*Roadmap created: 2026-06-16*
