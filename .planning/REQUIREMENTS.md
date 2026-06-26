# Requirements: Shelf Go

**Defined:** 2026-06-16
**Revised:** 2026-06-26
**Core Value:** A developer can safely manage project secrets in an encrypted local vault and use them through explicit CLI, file, and child-process workflows without treating plaintext `.env` files as the source of truth.

## Current Requirements

No v0.1.1 requirements are selected yet. Completed v0.1.0 requirements and evidence are archived at `.planning/archive/releases/v0.1.0/SUMMARY.md` and `.planning/archive/releases/v0.1.0/VERIFICATION.md`.

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

### Release and Distribution

- **V2-REL-01**: Shelf can add optional package-manager distribution after initial usage validates demand.
- **V2-REL-02**: Shelf can add native Windows smoke tests for setup, secret set/get, and project run on a real Windows runner.

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
| Dedicated vault restore command | Current backups are ordinary encrypted vault files and single-slot only; a command adds surface area without enough value. Manual copy plus `shelf vault status` is simpler. |
| Immediate SQLite backend | SQLite remains a future spike candidate; current storage is age-encrypted JSON until storage/query pressure justifies another design. |
| Broad one-file-per-command CLI split | `internal/cli` should stay command-family oriented rather than becoming a large directory of tiny files. |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| None selected for v0.1.1 | - | - |

**Coverage:**
- Current requirements: 0 total
- Mapped to phases: 0
- Unmapped: 0
- Completed v0.1.0 requirements: archived at `.planning/archive/releases/v0.1.0/SUMMARY.md`

---
*Last updated: 2026-06-26 after archiving v0.1.0 planning history*
