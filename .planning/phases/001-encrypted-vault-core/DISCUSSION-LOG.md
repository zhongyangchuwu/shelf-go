# Phase 1: Encrypted Vault Core - Discussion Log

> **Supplementary audit trail — not a required OMP artifact.**
> Decisions are captured in CONTEXT.md; this log preserves the alternatives considered.
**Date:** 2026-06-16
**Phase:** 1-Encrypted Vault Core
**Areas discussed:** Vault config contract, Age recipients and identities, Vault file format, Plaintext side-file policy

---

## Vault Config Contract

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit vault fields with compatibility aliases | Add vault-oriented config for encrypted storage while keeping `data`, `--data`, and `SHELF_DATA` usable as active store-path aliases in Phase 1. | Yes |
| Reuse existing data field only | Minimize config changes but keep encrypted vault concepts hidden behind a plaintext-era name. | |
| Hard break to vault-only config | Cleanest long-term naming but risks breaking existing scriptable workflows too early. | |

**User's choice:** Approved the recommended direction.
**Notes:** This keeps encrypted storage conceptually clear while preserving command compatibility.

---

## Age Recipients and Identities

| Option | Description | Selected |
|--------|-------------|----------|
| Recipients plus identity paths in config | Store public age recipients and filesystem paths to identities, never private identity contents. | Yes |
| Identity contents in config | Simplifies lookup but violates the non-secret config constraint. | |
| Implicit age defaults only | Reduces config surface but makes errors and portability less explicit. | |

**User's choice:** Approved the recommended direction.
**Notes:** Errors should identify missing/unreadable identities and wrong/no matching identity cases.

---

## Vault File Format

| Option | Description | Selected |
|--------|-------------|----------|
| Shelf envelope/header | Add enough Shelf-specific format/version metadata to distinguish unsupported format, plaintext legacy JSON, corrupt vaults, and decrypt failures. | Yes |
| Raw age-encrypted store JSON | Maximizes simplicity but weakens diagnostics and future format evolution. | |

**User's choice:** Approved the recommended direction.
**Notes:** Exact envelope shape is left to research and planning.

---

## Plaintext Side-File Policy

| Option | Description | Selected |
|--------|-------------|----------|
| Encrypt durable store and backups now; defer editor temp hardening | Meets Phase 1 encrypted-at-rest promise without pulling later edit UX hardening into this phase. | Yes |
| Harden all plaintext side files now | Stronger security posture but risks expanding Phase 1 beyond the encrypted store boundary. | |
| Only encrypt primary vault now | Too weak because existing `.bak` behavior would preserve secret values durably in plaintext. | |

**User's choice:** Approved the recommended direction.
**Notes:** Durable temp/write files and backups from store persistence must contain encrypted bytes. `secret edit` temp-file cleanup remains a later hardening concern unless implementation naturally touches it.

---

## the agent's Discretion

- Exact config key names.
- Exact envelope/header layout.
- Exact age library integration.
- Whether store encryption is represented by options, a backend interface, or a small persistence wrapper, as long as package boundaries and command semantics remain intact.

## Deferred Ideas

- Migration from existing plaintext stores belongs to Phase 2.
- Doctor git/chezmoi safety checks belong to Phase 2.
- Export, project, and run compatibility belongs to Phase 3.
- Localhost vault manager belongs to Phase 4.
- Documentation and release hardening belong to Phase 5.
