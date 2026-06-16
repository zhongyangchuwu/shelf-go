# Stack Research

**Domain:** CLI-first solo developer secret manager with age-encrypted portable vault
**Researched:** 2026-06-16
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.4 | CLI, store, local HTTP server | Already used by the repo; fast single binary distribution fits developer tooling. |
| Cobra | v1.9.1 | CLI command tree | Already established in `internal/cli`; keep command UX consistent. |
| `filippo.io/age` | v1.3.1 | Encrypt and decrypt vault files | Official Go age implementation; supports age recipients/identities and SSH identity compatibility through subpackages. |
| Go `net/http` | stdlib | Localhost vault manager | Enough for a small local UI/API; avoids framework weight and daemon assumptions. |
| Embedded static assets | stdlib `embed` | Ship UI in the Shelf binary | Keeps installation as one binary and avoids runtime asset paths. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `filippo.io/age/agessh` | v1.3.1 | SSH key recipient/identity support | Use if Shelf should decrypt with existing SSH keys in addition to native age identities. |
| `filippo.io/age/armor` | v1.3.1 | ASCII armor for encrypted data | Use only if a text vault format is explicitly needed; binary `.age` files are fine for git/chezmoi. |
| `github.com/adrg/xdg` | v0.5.3 | XDG path handling | Optional; current config resolution already works, so introduce only if path behavior grows. |
| `github.com/gorilla/csrf` | v1.7.3 | CSRF protection | Use if the localhost manager has browser sessions and state-changing POST/PUT/DELETE endpoints. |
| `github.com/go-chi/chi/v5` | v5.3.0 | HTTP routing | Optional; Go 1.22+ `ServeMux` routing is likely enough for v1. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go test ./...` | Unit and command verification | Required after storage-boundary changes. |
| `just test` | Existing project shortcut | Equivalent to `go test ./...`. |
| `shelf doctor` | Runtime health checks | Extend for encrypted vault and git-safe checks. |
| `age-keygen` / age CLI | Manual interoperability checks | Useful for validating that Shelf vaults are ordinary age files. |

## Installation

```bash
go get filippo.io/age@v1.3.1

# Optional only if needed later:
go get github.com/gorilla/csrf@v1.7.3
go get github.com/go-chi/chi/v5@v5.3.0
go get github.com/adrg/xdg@v0.5.3
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `filippo.io/age` | Shelling out to `age` binary | Only if exact CLI compatibility is more important than a self-contained binary. |
| Native age recipients | Password-only encryption | Use password-only later for users without key management; not the best first fit for chezmoi. |
| Go `net/http` | Full web framework | Use a framework only if the local vault manager grows complex route/middleware needs. |
| Single encrypted vault file | Directory of encrypted secret files | Use per-secret files only if merge conflict behavior becomes the dominant problem. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Plaintext JSON as durable storage | Unsafe for git, backups, and synced dotfiles. | age-encrypted vault file with plaintext only in memory. |
| Hosted secret backend for v1 | Violates local-first solo developer goal. | Portable encrypted file managed by Shelf and optionally chezmoi. |
| Long-running daemon as a prerequisite | Adds lifecycle and trust complexity before the core vault is secure. | On-demand CLI plus optional short-lived localhost manager. |
| Browser cookies without CSRF protection | Localhost apps can still receive forged browser requests. | Local bind, random session token, Origin checks, and CSRF tokens for writes. |

## Stack Patterns by Variant

**If the user already has age/chezmoi:**
- Store recipients in non-secret Shelf config.
- Store encrypted vault in a path chezmoi can manage.
- Decrypt with the user's configured identity at command time.

**If the user has only SSH keys:**
- Use `agessh` identity parsing where appropriate.
- Make explicit which SSH keys are supported and document agent/key-file behavior.

**If the localhost manager edits secrets:**
- Bind to `127.0.0.1` only by default.
- Generate an unguessable per-session token and open the browser to a tokenized URL.
- Keep writes behind POST/PUT/DELETE, Origin validation, and CSRF protection.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| `filippo.io/age@v1.3.1` | Go module project | Latest version reported by `go list -m -versions` on 2026-06-16. |
| `github.com/go-chi/chi/v5@v5.3.0` | Go HTTP server | Optional; current stdlib routing may be enough. |
| `github.com/gorilla/csrf@v1.7.3` | Go HTTP handlers | Optional; useful if hand-written CSRF handling becomes risky. |

## Sources

- https://pkg.go.dev/filippo.io/age - official Go age package and API overview.
- https://github.com/C2SP/C2SP/blob/main/age.md - age format specification.
- https://pkg.go.dev/net/http - Go HTTP server package reference.
- https://go.dev/blog/routing-enhancements - Go 1.22 ServeMux method and wildcard routing.
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html - CSRF threat model and prevention guidance.
- Local verification: `go list -m -versions filippo.io/age` returned `v1.3.1` as the newest listed module version on 2026-06-16.

---
*Stack research for: Shelf Go encrypted vault*
*Researched: 2026-06-16*
