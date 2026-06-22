# Requirements: Shelf Go

**Defined:** 2026-06-16
**Core Value:** A single developer can safely carry and use project secrets across machines through a portable encrypted vault, while keeping local env and `shelf run` workflows fast and simple.

## v1 Requirements

Requirements for the encrypted-vault milestone. Each requirement maps to one roadmap phase.

### Encrypted Vault

- [x] **VAULT-01**: User can configure Shelf to use an age-encrypted vault file as the durable secret store.
- [x] **VAULT-02**: User can configure one or more age recipients without storing private identity material in Shelf config.
- [x] **VAULT-03**: User can configure identity file paths or identity discovery needed to decrypt the vault.
- [x] **VAULT-04**: Shelf decrypts the vault on load, validates the plaintext store model, and exposes the existing in-memory secret model to commands.
- [x] **VAULT-05**: Shelf encrypts the full validated store to configured age recipients on save.
- [x] **VAULT-06**: Shelf rejects unreadable, undecryptable, corrupt, or unsupported vault formats with actionable errors.

### CLI Compatibility

- [x] **CLI-01**: Existing `shelf secret` read and write commands work against the encrypted vault.
- [x] **CLI-02**: Existing `shelf export` path and prefix flows work against the encrypted vault.
- [x] **CLI-03**: Existing `shelf project` manifest commands work against the encrypted vault without writing secret values to `.shelf.json`.
- [x] **CLI-04**: Existing `shelf run -- ...` and `shelf run --dry-run -- ...` work against the encrypted vault and preserve current value-printing rules.
- [x] **CLI-05**: Existing command tests or equivalent regression coverage verify that encryption did not change command semantics.

### Migration and Recovery

- [x] **MIGR-01**: User can migrate an existing plaintext Shelf JSON store into an age-encrypted vault.
- [x] **MIGR-02**: Migration preserves the original plaintext source until the encrypted target decrypts and validates successfully.
- [x] **MIGR-03**: Migration reports clear next steps for moving, deleting, or archiving the old plaintext store.
- [x] **MIGR-04**: Shelf creates backups or recovery artifacts in encrypted form when secret values are involved.
- [x] **MIGR-05**: Shelf avoids writing plaintext secret values to durable temp files during normal store operations.

### Git and Chezmoi Safety

- [x] **SAFE-01**: User can keep the encrypted vault as a normal portable file suitable for git-backed dotfile workflows such as chezmoi.
- [x] **SAFE-02**: Shelf config remains non-secret and can be reviewed or committed without exposing secret values or private identities.
- [x] **SAFE-03**: `shelf doctor` reports whether the active store is plaintext or encrypted.
- [x] **SAFE-04**: `shelf doctor` warns when plaintext secret storage appears to be tracked by git.
- [x] **SAFE-05**: `shelf doctor` confirms when the tracked/synced vault path is encrypted and value-free project manifests remain safe.

### Localhost Vault Manager

- [x] **WEB-01**: User can start a localhost-only vault manager from the CLI.
- [x] **WEB-02**: User can search and browse secret paths and non-secret metadata in the vault manager.
- [x] **WEB-03**: User can view and intentionally reveal/copy a secret value from the vault manager.
- [x] **WEB-04**: User can create, update, and delete secrets from the vault manager.
- [x] **WEB-05**: Vault-manager writes use the same validation, locking, encrypted-save, and backup rules as CLI writes.
- [x] **WEB-06**: Vault manager binds to loopback by default and uses session/write-safety controls for state-changing requests.
- [x] **WEB-07**: Vault manager does not require a permanent daemon or hosted service.

### Documentation and Verification

- [x] **DOCS-01**: Documentation explains the encrypted vault model, age recipient and identity configuration, and chezmoi-friendly workflow.
- [x] **DOCS-02**: Documentation clearly separates Shelf config, `.shelf.json` project manifests, encrypted vault data, and generated/exported env files.
- [x] **DOCS-03**: Documentation warns about plaintext export, terminal output, browser reveal/copy actions, and old plaintext store cleanup.
- [x] **TEST-01**: Automated tests cover encrypted load/save, wrong identity or missing identity errors, migration success/failure paths, and CLI compatibility.
- [x] **TEST-02**: Automated or manual verification covers localhost manager write protections and no-plaintext side files for representative edit flows.

## v2 Requirements

Deferred to future releases. Tracked but not in the current roadmap.

### Vault Expansion

- **V2-VAULT-01**: User can use password-only encryption if they do not use age keys.
- **V2-VAULT-02**: User can manage multiple vaults or profiles.
- **V2-VAULT-03**: User can inspect non-secret age recipient metadata from a vault.
- **V2-VAULT-04**: User can restore encrypted backups through a dedicated command.
- **V2-VAULT-05**: User gets better merge/conflict handling for git-managed encrypted vault updates.

### Integrations

- **V2-INT-01**: Shelf can offer optional direct chezmoi helper commands.
- **V2-INT-02**: Shelf can support hardware-key or age plugin workflows if users need them.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Team sharing | The current product is for solo developers; sharing requires identity, permissions, revocation, audit, and conflict semantics. |
| Hosted sync service | Shelf should stay local-first and portable instead of requiring a backend account. |
| Permanent daemon | Core CLI workflows should not depend on a long-running process; the vault manager should be short-lived/on-demand. |
| Browser extension or autofill | Shelf is focused on developer secrets and env workflows, not general password-manager replacement. |
| Plain `.env` as source of truth | `.env` cannot reliably encode Shelf secret identity and invites plaintext project files. |
| Direct chezmoi control in v1 | Chezmoi can manage the encrypted vault file; Shelf should not couple to chezmoi commands before the vault format is solid. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| VAULT-01 | Phase 1 | Complete |
| VAULT-02 | Phase 1 | Complete |
| VAULT-03 | Phase 1 | Complete |
| VAULT-04 | Phase 1 | Complete |
| VAULT-05 | Phase 1 | Complete |
| VAULT-06 | Phase 1 | Complete |
| CLI-01 | Phase 1 | Complete |
| MIGR-01 | Phase 2 | Complete |
| MIGR-02 | Phase 2 | Complete |
| MIGR-03 | Phase 2 | Complete |
| MIGR-04 | Phase 2 | Complete |
| MIGR-05 | Phase 2 | Complete |
| SAFE-01 | Phase 2 | Complete |
| SAFE-02 | Phase 2 | Complete |
| SAFE-03 | Phase 2 | Complete |
| SAFE-04 | Phase 2 | Complete |
| SAFE-05 | Phase 2 | Complete |
| CLI-02 | Phase 3 | Complete |
| CLI-03 | Phase 3 | Complete |
| CLI-04 | Phase 3 | Complete |
| CLI-05 | Phase 3 | Complete |
| TEST-01 | Phase 3 | Complete |
| WEB-01 | Phase 4 | Complete |
| WEB-02 | Phase 4 | Complete |
| WEB-03 | Phase 4 | Complete |
| WEB-04 | Phase 4 | Complete |
| WEB-05 | Phase 4 | Complete |
| WEB-06 | Phase 4 | Complete |
| WEB-07 | Phase 4 | Complete |
| TEST-02 | Phase 4 | Complete |
| DOCS-01 | Phase 5 | Complete |
| DOCS-02 | Phase 5 | Complete |
| DOCS-03 | Phase 5 | Complete |

**Coverage:**
- v1 requirements: 33 total
- Mapped to phases: 33
- Unmapped: 0

---
*Requirements defined: 2026-06-16*
*Last updated: 2026-06-22 after Phase 5 verification*
