# Requirements: Shelf Go

**Defined:** 2026-06-16
**Core Value:** A single developer can safely carry and use project secrets across machines through a portable encrypted vault, while keeping local env and `shelf run` workflows fast and simple.

## v1 Requirements

Requirements for the encrypted-vault milestone. Each requirement maps to one roadmap phase.

### Encrypted Vault

- [ ] **VAULT-01**: User can configure Shelf to use an age-encrypted vault file as the durable secret store.
- [ ] **VAULT-02**: User can configure one or more age recipients without storing private identity material in Shelf config.
- [ ] **VAULT-03**: User can configure identity file paths or identity discovery needed to decrypt the vault.
- [ ] **VAULT-04**: Shelf decrypts the vault on load, validates the plaintext store model, and exposes the existing in-memory secret model to commands.
- [ ] **VAULT-05**: Shelf encrypts the full validated store to configured age recipients on save.
- [ ] **VAULT-06**: Shelf rejects unreadable, undecryptable, corrupt, or unsupported vault formats with actionable errors.

### CLI Compatibility

- [ ] **CLI-01**: Existing `shelf secret` read and write commands work against the encrypted vault.
- [ ] **CLI-02**: Existing `shelf export` path and prefix flows work against the encrypted vault.
- [ ] **CLI-03**: Existing `shelf project` manifest commands work against the encrypted vault without writing secret values to `.shelf.json`.
- [ ] **CLI-04**: Existing `shelf run -- ...` and `shelf run --dry-run -- ...` work against the encrypted vault and preserve current value-printing rules.
- [ ] **CLI-05**: Existing command tests or equivalent regression coverage verify that encryption did not change command semantics.

### Migration and Recovery

- [ ] **MIGR-01**: User can migrate an existing plaintext Shelf JSON store into an age-encrypted vault.
- [ ] **MIGR-02**: Migration preserves the original plaintext source until the encrypted target decrypts and validates successfully.
- [ ] **MIGR-03**: Migration reports clear next steps for moving, deleting, or archiving the old plaintext store.
- [ ] **MIGR-04**: Shelf creates backups or recovery artifacts in encrypted form when secret values are involved.
- [ ] **MIGR-05**: Shelf avoids writing plaintext secret values to durable temp files during normal store operations.

### Git and Chezmoi Safety

- [ ] **SAFE-01**: User can keep the encrypted vault as a normal portable file suitable for git-backed dotfile workflows such as chezmoi.
- [ ] **SAFE-02**: Shelf config remains non-secret and can be reviewed or committed without exposing secret values or private identities.
- [ ] **SAFE-03**: `shelf doctor` reports whether the active store is plaintext or encrypted.
- [ ] **SAFE-04**: `shelf doctor` warns when plaintext secret storage appears to be tracked by git.
- [ ] **SAFE-05**: `shelf doctor` confirms when the tracked/synced vault path is encrypted and value-free project manifests remain safe.

### Localhost Vault Manager

- [ ] **WEB-01**: User can start a localhost-only vault manager from the CLI.
- [ ] **WEB-02**: User can search and browse secret paths and non-secret metadata in the vault manager.
- [ ] **WEB-03**: User can view and intentionally reveal/copy a secret value from the vault manager.
- [ ] **WEB-04**: User can create, update, and delete secrets from the vault manager.
- [ ] **WEB-05**: Vault-manager writes use the same validation, locking, encrypted-save, and backup rules as CLI writes.
- [ ] **WEB-06**: Vault manager binds to loopback by default and uses session/write-safety controls for state-changing requests.
- [ ] **WEB-07**: Vault manager does not require a permanent daemon or hosted service.

### Documentation and Verification

- [ ] **DOCS-01**: Documentation explains the encrypted vault model, age recipient and identity configuration, and chezmoi-friendly workflow.
- [ ] **DOCS-02**: Documentation clearly separates Shelf config, `.shelf.json` project manifests, encrypted vault data, and generated/exported env files.
- [ ] **DOCS-03**: Documentation warns about plaintext export, terminal output, browser reveal/copy actions, and old plaintext store cleanup.
- [ ] **TEST-01**: Automated tests cover encrypted load/save, wrong identity or missing identity errors, migration success/failure paths, and CLI compatibility.
- [ ] **TEST-02**: Automated or manual verification covers localhost manager write protections and no-plaintext side files for representative edit flows.

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
| VAULT-01 | TBD | Pending |
| VAULT-02 | TBD | Pending |
| VAULT-03 | TBD | Pending |
| VAULT-04 | TBD | Pending |
| VAULT-05 | TBD | Pending |
| VAULT-06 | TBD | Pending |
| CLI-01 | TBD | Pending |
| CLI-02 | TBD | Pending |
| CLI-03 | TBD | Pending |
| CLI-04 | TBD | Pending |
| CLI-05 | TBD | Pending |
| MIGR-01 | TBD | Pending |
| MIGR-02 | TBD | Pending |
| MIGR-03 | TBD | Pending |
| MIGR-04 | TBD | Pending |
| MIGR-05 | TBD | Pending |
| SAFE-01 | TBD | Pending |
| SAFE-02 | TBD | Pending |
| SAFE-03 | TBD | Pending |
| SAFE-04 | TBD | Pending |
| SAFE-05 | TBD | Pending |
| WEB-01 | TBD | Pending |
| WEB-02 | TBD | Pending |
| WEB-03 | TBD | Pending |
| WEB-04 | TBD | Pending |
| WEB-05 | TBD | Pending |
| WEB-06 | TBD | Pending |
| WEB-07 | TBD | Pending |
| DOCS-01 | TBD | Pending |
| DOCS-02 | TBD | Pending |
| DOCS-03 | TBD | Pending |
| TEST-01 | TBD | Pending |
| TEST-02 | TBD | Pending |

**Coverage:**
- v1 requirements: 32 total
- Mapped to phases: 0
- Unmapped: 32

---
*Requirements defined: 2026-06-16*
*Last updated: 2026-06-16 after initial definition*
