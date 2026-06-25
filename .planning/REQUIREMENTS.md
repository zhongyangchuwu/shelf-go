# Requirements: Shelf Go

**Defined:** 2026-06-16
**Revised:** 2026-06-25
**Core Value:** A developer can safely manage project secrets in an encrypted local vault and use them through explicit CLI, file, and child-process workflows without treating plaintext `.env` files as the source of truth.

## Current Requirements

Requirements for first public release readiness. Prior encrypted-vault, command-hierarchy, vault-UX, project-session design, safety, architecture, and release-readiness work is treated as baseline behavior that must remain passing.

### Release Readiness

- [x] **REL-01**: Release automation builds repeatable multi-platform CLI artifacts, archives, and checksums from version tags.
- [x] **REL-02**: GitHub Actions runs CI checks and tag-triggered release publishing with least necessary permissions.
- [x] **REL-03**: README explains the project motivation and primary initialization, secret, and project workflows without becoming a full command inventory.
- [x] **REL-04**: First-release changelog, verification evidence, and deferred post-0.1 work are recorded before tagging.

### Architecture Boundaries

- [x] **ARCH-01**: Runtime/vault construction helpers are owned outside `internal/cli` so future backend construction changes do not require editing every command file.
- [x] **ARCH-02**: Project manifest resolution, project diagnostics, env conflict detection, render binding conversion, project identity, git root lookup, and remote normalization are owned by `internal/project`, not `internal/cli`.
- [x] **ARCH-03**: `internal/cli` remains compact and command-family oriented, targeting roughly 3-6 files instead of one file per subcommand.
- [x] **ARCH-04**: Vault status/check/doctor diagnostic rules are reusable outside CLI through an `internal/vault` feature package.
- [x] **ARCH-05**: `secret edit` editable JSON and temp-file/editor workflow are reusable outside CLI through an `internal/secret` feature package while preserving plaintext cleanup behavior.
- [x] **ARCH-06**: Atomic file write behavior is centralized with explicit permission, sync, and backup options for vault, manifest, and config writes.
- [x] **ARCH-07**: Env name and path token validation have a single canonical implementation used by store, manifest, and render.
- [x] **ARCH-08**: `internal/store` keeps one package but separates model, store methods, JSON encode/decode, age seal/open, and vault orchestration into clearer files.

## Baseline Implemented Requirements

Already implemented and must remain working unless explicitly redesigned by the current milestone.

### Project Export UX

- [x] **PUX-01**: `shelf project export` defaults to existing `shell` output so redirected files are directly sourceable.
- [x] **PUX-02**: Explicit `--format env|shell|json` behavior remains supported, and no new `dotenv` format is introduced.
- [x] **PUX-03**: User docs recommend explicit source workflows such as `shelf project export > .env.local` and warn that generated env files are plaintext and must not be committed.

### Vault Recovery

- [x] **VREC-01**: User can recover from a single last-write encrypted `.bak` using ordinary file copy and `shelf vault status` verification.
- [x] **VREC-02**: Recovery docs explain that `.bak` is overwritten on each later vault replacement and is not a history system.
- [x] **VREC-03**: Recovery docs explain identity loss, backup recovery, and post-recovery `shelf vault status` verification.

### Safety Hardening

- [x] **SAFE-EDIT-01**: `shelf secret edit` limits plaintext temporary-file exposure with restrictive permissions and cleanup behavior.
- [x] **SAFE-MGR-01**: Local manager plaintext and token boundaries are either cheaply hardened or documented explicitly without adding a permanent daemon.
- [x] **SAFE-DOC-01**: User-facing docs name remaining plaintext boundaries and recommended cleanup behavior.

### Command Hierarchy

- [x] **CMD-01**: User can run global onboarding through `shelf setup` instead of ambiguous top-level `shelf init`.
- [x] **CMD-02**: User can initialize/configure vault storage through `shelf vault init`.
- [x] **CMD-03**: User can migrate plaintext stores through `shelf vault migrate`.
- [x] **CMD-04**: User can open the local vault manager through `shelf vault open`.
- [x] **CMD-05**: User can directly export path/prefix secrets through `shelf secret export`.
- [x] **CMD-06**: User can run project-bound commands through `shelf project run -- ...`.
- [x] **CMD-07**: The pre-release CLI removes old top-level `init`, `migrate`, `export`, `run`, and `manager` commands.
- [x] **CMD-08**: User-facing docs and command tests describe the scoped hierarchy as canonical.

### Vault UX

- [x] **VUX-01**: User can inspect vault configuration and loadability through a vault-scoped status/check command without revealing secret values.
- [x] **VUX-02**: Vault commands provide concise next steps for missing recipients, missing identities, plaintext legacy stores, unsupported formats, and undecryptable vaults.
- [x] **VUX-03**: Vault docs explain first-run flow, age identity/recipient setup, migration cleanup, and local manager opening flow.
- [x] **VUX-04**: Vault UX verification covers encrypted save/load, migration, doctor/status behavior, and manager write safety after the command hierarchy change.

### Encrypted Vault Baseline

- [x] **BASE-VAULT-01**: User can configure Shelf to use an age-encrypted vault file as the durable secret store.
- [x] **BASE-VAULT-02**: User can configure one or more age recipients without storing private identity material in Shelf config.
- [x] **BASE-VAULT-03**: User can configure identity file paths or identity discovery needed to decrypt the vault.
- [x] **BASE-VAULT-04**: Shelf decrypts the vault on load, validates the plaintext store model, and exposes the existing in-memory secret model to commands.
- [x] **BASE-VAULT-05**: Shelf encrypts the full validated store to configured age recipients on save.
- [x] **BASE-VAULT-06**: Shelf rejects unreadable, undecryptable, corrupt, or unsupported vault formats with actionable errors.
- [x] **BASE-VAULT-07**: User can migrate an existing plaintext Shelf JSON store into an age-encrypted vault while preserving the original source until the encrypted target validates successfully.
- [x] **BASE-VAULT-08**: Shelf creates backups or recovery artifacts in encrypted form when secret values are involved.
- [x] **BASE-VAULT-09**: Shelf avoids writing plaintext secret values to durable temp files during normal store operations.

### Project and Secret Baseline

- [x] **BASE-CLI-01**: Existing secret read and write commands work against the encrypted vault.
- [x] **BASE-CLI-02**: Existing direct export path and prefix flows work against the encrypted vault.
- [x] **BASE-CLI-03**: Existing project manifest commands work against the encrypted vault without writing secret values to `.shelf.json`.
- [x] **BASE-CLI-04**: Existing runtime injection and dry-run behavior work against the encrypted vault and preserve value-printing rules.
- [x] **BASE-CLI-05**: Regression coverage verifies that encryption did not change command semantics.

### Safety and Local Manager Baseline

- [x] **BASE-SAFE-01**: User can keep the encrypted vault as a normal portable file suitable for git-backed dotfile workflows such as chezmoi.
- [x] **BASE-SAFE-02**: Shelf config remains non-secret and can be reviewed or committed without exposing secret values or private identities.
- [x] **BASE-SAFE-03**: `shelf doctor` reports whether the active store is plaintext or encrypted.
- [x] **BASE-SAFE-04**: `shelf doctor` warns when plaintext secret storage appears to be tracked by git.
- [x] **BASE-SAFE-05**: User can start a localhost-only vault manager for metadata search, intentional reveal, and create/update/delete over encrypted storage.
- [x] **BASE-SAFE-06**: Vault-manager writes use the same validation, locking, encrypted-save, and backup rules as CLI writes.
- [x] **BASE-SAFE-07**: Vault manager binds to loopback by default and uses session/write-safety controls for state-changing requests.

### Project Session Design Baseline

- [x] **SES-01**: Project activation/deactivation was analyzed under `shelf project` and intentionally left unimplemented for now.
- [x] **SES-02**: Project shell entry was analyzed under `shelf project` and intentionally left unimplemented for now.
- [x] **SES-03**: Activation/deactivation design records restore semantics if hooks are ever implemented later.
- [x] **SES-04**: Activation design records that current-shell mutation requires a shell hook/function.

## Deferred Requirements

Tracked for future releases. These are not current implementation commitments.

### Project Sessions

- **V2-SES-01**: User can activate the current project environment in the current shell through `shelf project activate` after installing a shell hook, if future UX evidence justifies hook complexity.
- **V2-SES-02**: User can restore the previous shell environment through `shelf project deactivate` without losing pre-existing env values, if activation is implemented.
- **V2-SES-03**: User can enter an isolated project environment through `shelf project shell` only if it proves clearer than `shelf project run -- $SHELL`.

### Vault Expansion

- **V2-VAULT-01**: User can use password-only encryption if they do not use age keys.
- **V2-VAULT-02**: User can manage multiple vaults or profiles.
- **V2-VAULT-03**: User can inspect non-secret age recipient metadata from a vault.
- **V2-VAULT-04**: User gets better merge/conflict handling for git-managed encrypted vault updates.
- **V2-VAULT-05**: Shelf can spike SQLite as a future encrypted vault payload or metadata/search storage option when schema/query pressure justifies it; any design must preserve encrypted-at-rest vault safety and avoid plaintext WAL/journal/temp files.

### Integrations and Editing

- **V2-INT-01**: Shelf can offer optional direct chezmoi helper commands.
- **V2-INT-02**: Shelf can support hardware-key or age plugin workflows if users need them.
- **V2-UI-01**: Shelf can provide a TUI or improved field-specific editor when that improves local editing speed without weakening vault safety.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Team sharing | The current product is for solo developers; sharing requires identity, permissions, revocation, audit, and conflict semantics. |
| Hosted sync service | Shelf should stay local-first and portable instead of requiring a backend account. |
| Permanent daemon | Core CLI workflows should not depend on a long-running process; the vault manager should be short-lived/on-demand. |
| Browser extension or autofill | Shelf is focused on developer secrets and env workflows, not general password-manager replacement. |
| Plain `.env` as source of truth | `.env` files may be generated/exported, but Shelf's source of truth is the encrypted vault plus project manifests. |
| New dotenv export format | Existing `shell` output is already sourceable; adding another format increases surface area without enough value. |
| Hook-based project activation in current scope | Shell hooks mutate parent-shell state implicitly and add complexity; explicit export/source workflows are preferred for now. |
| Backward-compatible pre-release aliases | The project has not been published; simpler command cutover is more valuable than compatibility shims. |
| Dedicated vault restore command | Current backups are ordinary encrypted vault files and single-slot only; a command adds surface area without enough value. Manual copy plus `shelf vault status` is simpler. |
| Immediate SQLite backend | SQLite remains a future spike candidate; this milestone only improves architecture boundaries needed for future evolution. |
| Broad one-file-per-command CLI split | `internal/cli` should stay command-family oriented rather than becoming a large directory of tiny files. |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| REL-01..REL-04 | Phase 16 | Complete |
| ARCH-01..ARCH-03 | Phase 13 | Complete |
| ARCH-04..ARCH-05 | Phase 14 | Complete |
| ARCH-06..ARCH-08 | Phase 15 | Complete |
| PUX-01..PUX-03 | Phase 9 | Complete |
| VREC-01..VREC-03 | Phase 10 | Complete |
| SAFE-EDIT-01, SAFE-MGR-01, SAFE-DOC-01 | Phase 11 | Complete |
| CMD-01..CMD-08 | Phase 6 | Complete |
| VUX-01..VUX-04 | Phase 7 | Complete |
| SES-01..SES-04 | Phase 8 | Complete |
| BASE-VAULT-01..09 | Completed encrypted-vault milestone | Complete |
| BASE-CLI-01..05 | Completed encrypted-vault milestone | Complete |
| BASE-SAFE-01..07 | Completed encrypted-vault milestone | Complete |

**Coverage:**
- Current requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0
- Deferred storage spike: SQLite only; Dolt is not a current vault-storage candidate.
- Deferred manager UI redesign: post-0.1 visual and UX improvement for `shelf vault open`.

---
*Last updated: 2026-06-25 after completing Phase 16 first release readiness*
