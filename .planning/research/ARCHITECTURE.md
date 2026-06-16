# Architecture Research

**Domain:** CLI-first solo developer secret manager with age-encrypted portable vault
**Researched:** 2026-06-16
**Confidence:** HIGH

## Standard Architecture

### System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                         User Surfaces                        │
├───────────────────────────────┬─────────────────────────────┤
│ CLI commands                   │ Localhost vault manager      │
│ internal/cli                   │ internal/cli + internal/web  │
└───────────────┬───────────────┴──────────────┬──────────────┘
                │                              │
                ▼                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Secret Domain Operations                   │
│ store model, validation, paths, project resolution, render   │
└───────────────────────────────┬─────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Vault Persistence Boundary                  │
│ decrypt/load/validate  ←→  mutate in memory  ←→ encrypt/save │
└───────────────────────────────┬─────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Portable Files on Disk                    │
│ encrypted vault.age, non-secret config.yaml, .shelf.json     │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| CLI surface | Existing command UX and orchestration | Continue using Cobra in `internal/cli`. |
| Store model | In-memory plaintext representation and validation | Keep `internal/store.Data`, `Secret`, path grammar, and validation. |
| Vault codec | Serialize, encrypt, decrypt, and validate vault bytes | Add a store/vault boundary under `internal/store` or a focused `internal/vault` package. |
| Age key config | Resolve recipients and identities | Extend `internal/config` with non-secret vault settings. |
| Local web server | Search/edit UI and JSON API | Short-lived loopback HTTP server with embedded assets. |
| Project resolver | Convert manifests to env bindings | Preserve existing `internal/cli/project.go` behavior, later extract if the web UI needs it. |

## Recommended Project Structure

```text
internal/
├── cli/                 # Cobra commands, including vault migration and serve/open command
├── config/              # Runtime config, vault path, recipients, identity paths
├── store/               # Data model, validation, locking, load/save orchestration
├── vault/               # Optional package for age codec and vault format helpers
├── web/                 # Optional package for localhost manager handlers and assets
├── manifest/            # Existing project manifest model
├── render/              # Existing export rendering
└── version/             # Existing version behavior
```

### Structure Rationale

- **Keep `store` as the model owner:** Commands already depend on `store.Data` semantics; encryption should not force command rewrites.
- **Use `vault` only if boundaries get clearer:** If age codec logic grows, separate it from generic store validation while keeping the public load/save API simple.
- **Use `web` for HTTP-specific code:** Do not put handlers and HTML/API details into large CLI files.
- **Keep config non-secret:** Recipient public keys and vault paths are configuration; identities remain private files supplied by the user.

## Architectural Patterns

### Pattern 1: Encryption Boundary Around Existing Store

**What:** Commands keep operating on plaintext `store.Data` in memory, while disk persistence encrypts/decrypts at the load/save boundary.

**When to use:** This is the right fit because Shelf already has working CLI and validation semantics.

**Trade-offs:** Minimizes command churn, but all temp/backups/migration paths must go through the same boundary.

### Pattern 2: Format Version Wrapper

**What:** Store an encrypted payload that decrypts to a versioned JSON object. Keep version checks strict.

**When to use:** Required before changing durable format from plaintext JSON to age-encrypted bytes.

**Trade-offs:** Needs migration tooling, but prevents ambiguous file interpretation later.

### Pattern 3: Short-Lived Localhost Manager

**What:** `shelf vault open` or similar starts a local server, opens a browser to a tokenized URL, and exits on user action or timeout.

**When to use:** Use for search/edit workflows that are poor in terminal JSON.

**Trade-offs:** Better UX, but must defend against cross-origin browser requests and accidental long-lived unlocked state.

## Data Flow

### Encrypted CLI Read Flow

```text
CLI command
    ↓
config.Resolve vault path + identities
    ↓
store.Load
    ↓
read encrypted vault bytes
    ↓
age decrypt
    ↓
JSON decode + strict validation
    ↓
command reads secret data
```

### Encrypted CLI Write Flow

```text
Mutating CLI command
    ↓
lock vault path
    ↓
load/decrypt latest vault
    ↓
validate mutation
    ↓
serialize plaintext in memory
    ↓
age encrypt to configured recipients
    ↓
atomic write + encrypted backup
    ↓
unlock
```

### Localhost Edit Flow

```text
shelf vault open
    ↓
bind 127.0.0.1 random port
    ↓
generate one-session token / CSRF material
    ↓
browser opens tokenized URL
    ↓
GET search/list metadata
    ↓
POST/PUT edit secret with CSRF + Origin checks
    ↓
same locked encrypted save path as CLI
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1 developer, hundreds of secrets | Single encrypted file and whole-file load/save are fine. |
| Thousands of secrets | Add tests/benchmarks before changing; consider indexes after decrypt. |
| Multi-device git workflows | Handle conflict detection and recovery before introducing sync logic. |
| Teams | Requires a different roadmap: recipient rotation, audit, sharing, revocation. |

### Scaling Priorities

1. **First bottleneck:** Security correctness, not performance. Fix plaintext-at-rest and unsafe temp/backups first.
2. **Second bottleneck:** UX around editing and recovery. Provide local UI and encrypted backups.
3. **Third bottleneck:** Merge/conflict behavior for git-managed encrypted blobs.

## Anti-Patterns

### Anti-Pattern 1: Encrypting Only the Main File

**What people do:** Encrypt `secrets.json` but leave `.bak`, edit temp files, logs, or migration copies in plaintext.
**Why it's wrong:** Attackers and accidental git commits will find the side files.
**Do this instead:** Treat every durable or temporary persistence path as part of the vault threat model.

### Anti-Pattern 2: Making the Browser UI a Trusted App Without Controls

**What people do:** Bind a localhost UI and assume local means safe.
**Why it's wrong:** Browsers can send cross-origin requests to localhost services.
**Do this instead:** Bind loopback, use unguessable session tokens, validate Origin, require CSRF tokens, and avoid GET writes.

### Anti-Pattern 3: Replacing Existing CLI Semantics During Encryption

**What people do:** Redesign commands while changing storage.
**Why it's wrong:** It mixes product risk with security migration risk.
**Do this instead:** Preserve command behavior first, then improve UX in separate phases.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| age | Library integration through `filippo.io/age` | Prefer library for single-binary behavior. |
| chezmoi | File-level interoperability | Let chezmoi manage the encrypted vault; Shelf should not shell out to chezmoi in v1. |
| Git | Existing project-aware behavior and doctor checks | Add encrypted-vault checks; keep `.shelf.json` value-free. |
| Browser | Loopback-only manager | Treat write endpoints as security-sensitive. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| CLI to store/vault | Function calls | Keep load/save APIs narrow. |
| Web UI to store/vault | Handler calls into same domain layer | Do not duplicate secret mutation logic in handlers. |
| Config to vault | Runtime settings | Public recipients are config; private identities are file references, not stored secret material. |
| Manifest to render | Existing resolver | Preserve env-name conflict checks. |

## Sources

- `.planning/codebase/ARCHITECTURE.md` - current layered CLI architecture.
- `.planning/codebase/CONCERNS.md` - plaintext storage and temp-file risks.
- https://pkg.go.dev/filippo.io/age - Go age library.
- https://pkg.go.dev/net/http - Go HTTP server primitives.
- https://go.dev/blog/routing-enhancements - modern ServeMux routing capabilities.
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html - CSRF controls for state-changing browser requests.

---
*Architecture research for: Shelf Go encrypted vault*
*Researched: 2026-06-16*
