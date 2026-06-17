# Feature Research

**Domain:** CLI-first solo developer secret manager with age-encrypted portable vault
**Researched:** 2026-06-16
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Encrypted vault at rest | A secret manager cannot keep durable plaintext when the file is git/sync managed. | HIGH | Must include backups and temp files in the threat model. |
| Age recipient configuration | Users need to choose which age recipient can decrypt the vault. | MEDIUM | Recipient config is non-secret and can live in Shelf config. |
| Identity/decrypt workflow | Commands need a reliable way to locate identities without leaking them into config. | HIGH | Must work with existing age/chezmoi habits. |
| Migration from plaintext store | Existing Shelf users need a path from `secrets.json` to encrypted vault. | MEDIUM | Should backup before migration and verify decrypt/load after migration. |
| CLI compatibility | Existing `secret`, `export`, `project`, and `run` flows must keep working. | MEDIUM | Encryption should be a storage-layer boundary, not a command rewrite. |
| Git-safe health checks | Users need confirmation that the committed/synced file is encrypted and config is non-secret. | LOW | Extend `shelf doctor`. |
| Localhost vault manager | CLI JSON editing is painful; users expect search and edit UI for vault management. | HIGH | Must be local-only and protected against browser-origin attacks. |
| Recovery behavior | Users need confidence they can recover from failed writes or bad edits. | MEDIUM | Encrypted backups and validation are part of this. |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Chezmoi-friendly vault | Fits existing dotfile workflows without a hosted account. | MEDIUM | Shelf should not need direct chezmoi integration in v1. |
| Project-aware env injection | More useful to developers than generic password managers. | Already implemented | Preserve and harden around encrypted storage. |
| Short-lived local editor | Better than `$EDITOR` JSON edits without running a permanent service. | HIGH | Start on demand, bind loopback, exit cleanly. |
| Vault format inspection | Helps users verify git-safe state without revealing secrets. | MEDIUM | Could show age header/recipients metadata where safe. |
| Portable single binary | Easy install for hackers and solo developers. | MEDIUM | Avoid external service and runtime dependencies. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Team sharing | Common secret-manager feature. | Forces identity, permissions, revocation, audit, and sync semantics too early. | Keep solo vault first. |
| Hosted sync | Convenient cross-device story. | Adds account/backend trust and reliability burden. | Portable encrypted vault managed by chezmoi/git. |
| Plain `.env` import/source of truth | Familiar developer habit. | Loses Shelf path identity and invites plaintext secrets into repos. | `.shelf.json` manifests plus export/run. |
| Permanent local daemon | Makes UI feel instant. | Adds lifecycle, attack surface, and stale-unlocked-vault risk. | On-demand localhost manager. |
| Browser extension/autofill | Password-manager parity. | Not aligned with developer-secret/env workflow. | CLI and local vault manager. |

## Feature Dependencies

```text
Encrypted vault boundary
    ├──requires──> age recipient and identity config
    ├──requires──> encrypted backup/write path
    └──enables──> git-safe chezmoi workflow

Migration from plaintext store
    └──requires──> encrypted vault boundary

Localhost vault manager
    ├──requires──> encrypted vault load/save APIs
    ├──requires──> edit-safe validation
    └──requires──> local web security controls

Doctor git-safe checks
    └──requires──> vault format detection
```

### Dependency Notes

- **Encrypted vault boundary before UI:** Editing through the localhost manager should not cement plaintext persistence.
- **Migration before default switch:** Existing users need a safe conversion path before encrypted storage becomes the normal path.
- **Security controls before edit UI:** Read-only UI is lower risk; edit UI needs CSRF and write validation from the start.

## MVP Definition

### Launch With (v1)

- [ ] age-encrypted vault file replaces plaintext durable storage.
- [ ] Shelf config can point to vault path, age recipients, and identity locations without storing secret values.
- [ ] Existing CLI commands work against the encrypted vault.
- [ ] Migration command or flow converts existing plaintext `secrets.json` safely.
- [ ] `shelf doctor` verifies encrypted vault state and warns about unsafe plaintext/git combinations.
- [ ] Localhost vault manager can search, view metadata, and edit secrets safely.

### Add After Validation (v1.x)

- [ ] Better recipient management UX after the first age path works.
- [ ] Vault inspection command that reports non-secret encryption metadata.
- [ ] Backup restore command for encrypted backups.
- [ ] Improved conflict handling for git-managed vault updates.

### Future Consideration (v2+)

- [ ] Password-based encryption for users without age keys.
- [ ] Multiple vaults or profiles.
- [ ] Hardware-key/plugin support if age plugin usage becomes important.
- [ ] Team sharing, if the product intentionally expands beyond solo use.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| age-encrypted vault | HIGH | HIGH | P1 |
| recipient/identity config | HIGH | MEDIUM | P1 |
| CLI compatibility over encrypted store | HIGH | MEDIUM | P1 |
| plaintext migration | HIGH | MEDIUM | P1 |
| doctor git-safe checks | HIGH | LOW | P1 |
| localhost edit UI | HIGH | HIGH | P2 |
| encrypted backup restore | MEDIUM | MEDIUM | P2 |
| direct chezmoi commands | LOW | MEDIUM | P3 |
| team sharing | LOW for target user | HIGH | P3 |

## Competitor Feature Analysis

| Feature | 1Password/Bitwarden | Doppler | Our Approach |
|---------|---------------------|---------|--------------|
| Developer env injection | Possible but indirect or plan-dependent. | Core hosted workflow. | Local CLI-native `project` and `run`. |
| Offline solo vault | Strong for passwords. | Not the product center. | Portable encrypted vault file. |
| Git/dotfile portability | Not the default mental model. | Not git-file native. | First-class encrypted file managed by user's tooling. |
| Local editing UX | Strong GUI, less developer-path specific. | Web service oriented. | Localhost manager focused on developer secret paths/env metadata. |

## Sources

- `.planning/PROJECT.md` - Shelf product direction and constraints.
- `.planning/codebase/ARCHITECTURE.md` and `CONCERNS.md` - current implementation and security gaps.
- `README.md`, `docs/usage-spec.md`, `docs/data-spec.md`, `docs/roadmap.md` - existing command surface and product rules.
- https://pkg.go.dev/filippo.io/age - official age Go package.
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html - local web write-safety threat model.

---
*Feature research for: Shelf Go encrypted vault*
*Researched: 2026-06-16*
