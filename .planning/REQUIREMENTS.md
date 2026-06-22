# Requirements: Shelf Go

**Defined:** 2026-06-16
**Revised:** 2026-06-22
**Core Value:** A developer can safely manage project secrets in an encrypted local vault and inject them into commands, projects, and shell workflows with predictable, reversible behavior.

## Current Requirements

Requirements for the pre-release command hierarchy and vault UX milestone. Prior encrypted-vault implementation is treated as existing baseline behavior that must remain passing while command names are simplified.

### Command Hierarchy

- [ ] **CMD-01**: User can run global onboarding through `shelf setup` instead of ambiguous top-level `shelf init`.
- [ ] **CMD-02**: User can initialize/configure vault storage through `shelf vault init` using the existing config, recipient, identity, and vault-path behavior.
- [ ] **CMD-03**: User can migrate plaintext stores through `shelf vault migrate` so migration is clearly a vault lifecycle operation.
- [ ] **CMD-04**: User can open the local vault manager through `shelf vault open` so the UI entrypoint is clearly vault-scoped.
- [ ] **CMD-05**: User can directly export path/prefix secrets through `shelf secret export` so direct export is distinct from project manifest export.
- [ ] **CMD-06**: User can run project-bound commands through `shelf project run -- ...` so runtime injection is clearly tied to `.shelf.json`.
- [ ] **CMD-07**: The pre-release CLI removes old top-level commands whose names now obscure scope: `init`, `migrate`, `export`, `run`, and `manager`.
- [ ] **CMD-08**: User-facing docs and command tests describe the new hierarchy as the canonical command surface.

### Vault UX

- [ ] **VUX-01**: User can inspect vault configuration and loadability through a vault-scoped status/check command without revealing secret values.
- [ ] **VUX-02**: Vault commands provide concise next steps for missing recipients, missing identities, plaintext legacy stores, unsupported formats, and undecryptable vaults.
- [ ] **VUX-03**: Vault docs explain the recommended first-run flow, age identity/recipient setup, migration cleanup, and local manager opening flow.
- [ ] **VUX-04**: Vault UX verification covers encrypted save/load, migration, doctor/status behavior, and manager write safety after the command hierarchy change.

### Future Project Sessions

- [ ] **SES-01**: Project activation/deactivation is planned under `shelf project activate` and `shelf project deactivate`, not top-level commands.
- [ ] **SES-02**: Project shell entry is planned under `shelf project shell`, not as a top-level command.
- [ ] **SES-03**: Activation/deactivation design records how previous env values are restored rather than simply unset.
- [ ] **SES-04**: Activation design records the need for a shell hook/function because a child CLI process cannot mutate the parent shell environment directly.

## Baseline Implemented Requirements

Already implemented and must remain working unless explicitly redesigned by the current milestone.

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

## Deferred Requirements

Tracked for future releases. These are not current implementation commitments.

### Project Sessions

- **V2-SES-01**: User can activate the current project environment in the current shell through `shelf project activate` after installing a shell hook.
- **V2-SES-02**: User can restore the previous shell environment through `shelf project deactivate` without losing pre-existing env values.
- **V2-SES-03**: User can enter an isolated project environment through `shelf project shell` without installing shell hooks.

### Vault Expansion

- **V2-VAULT-01**: User can use password-only encryption if they do not use age keys.
- **V2-VAULT-02**: User can manage multiple vaults or profiles.
- **V2-VAULT-03**: User can inspect non-secret age recipient metadata from a vault.
- **V2-VAULT-04**: User can restore encrypted backups through a dedicated command.
- **V2-VAULT-05**: User gets better merge/conflict handling for git-managed encrypted vault updates.

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
| Plain `.env` as source of truth | `.env` cannot reliably encode Shelf secret identity and invites plaintext secrets into repos. |
| Direct chezmoi control in current scope | Chezmoi can manage the encrypted vault file; Shelf should not couple to chezmoi commands before the vault format and command hierarchy are stable. |
| Backward-compatible pre-release aliases | The project has not been published; simpler command cutover is more valuable than compatibility shims. |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CMD-01 | Phase 6 | Planned |
| CMD-02 | Phase 6 | Planned |
| CMD-03 | Phase 6 | Planned |
| CMD-04 | Phase 6 | Planned |
| CMD-05 | Phase 6 | Planned |
| CMD-06 | Phase 6 | Planned |
| CMD-07 | Phase 6 | Planned |
| CMD-08 | Phase 6 | Planned |
| VUX-01 | Phase 7 | Planned |
| VUX-02 | Phase 7 | Planned |
| VUX-03 | Phase 7 | Planned |
| VUX-04 | Phase 7 | Planned |
| SES-01 | Phase 8 | Planned |
| SES-02 | Phase 8 | Planned |
| SES-03 | Phase 8 | Planned |
| SES-04 | Phase 8 | Planned |
| BASE-VAULT-01..09 | Completed encrypted-vault milestone | Complete |
| BASE-CLI-01..05 | Completed encrypted-vault milestone | Complete |
| BASE-SAFE-01..07 | Completed encrypted-vault milestone | Complete |

**Coverage:**
- Current requirements: 16 total
- Mapped to phases: 16
- Unmapped: 0

---
*Last updated: 2026-06-22 for command hierarchy and vault UX planning*
