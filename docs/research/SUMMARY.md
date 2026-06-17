# Research Summary

**Domain:** CLI-first solo developer secret manager with age-encrypted portable vault
**Researched:** 2026-06-16
**Confidence:** HIGH

## Key Findings

**Stack:** Keep the existing Go/Cobra CLI architecture and use `filippo.io/age` v1.3.1 as the core encryption dependency. Go `net/http` plus embedded static assets is sufficient for a first localhost vault manager; add routing/CSRF libraries only if the standard library path becomes too hand-rolled.

**Table Stakes:** The next product increment must make durable storage age-encrypted, portable, and safe for git/chezmoi workflows while preserving existing `secret`, `export`, `project`, and `run` commands. Migration, doctor checks, encrypted backups, and clear key configuration are not optional for a credible secret manager.

**Watch Out For:** The main risks are side-channel plaintext files, unsafe localhost edit endpoints, ambiguous recipient/identity configuration, and breaking existing CLI behavior while changing storage.

## Recommended Direction

1. Build the age-encrypted vault boundary first.
2. Preserve the current plaintext in-memory `store.Data` model after decrypt/load.
3. Add non-secret config for vault path and age recipients/identity paths.
4. Add a verified migration path from plaintext JSON to encrypted vault.
5. Extend `shelf doctor` to prove git-safe encrypted state.
6. Add the localhost vault manager after the storage boundary is secure.

## Suggested v1 Scope

- age-encrypted vault file as the source of truth.
- Existing CLI commands work against encrypted storage.
- Migration from current plaintext store.
- Git/chezmoi-safe `doctor` checks.
- Localhost vault manager for search, metadata view, and editing.
- Security hardening for backups, temp files, local web writes, and strict decode.

## Defer

- Team sharing.
- Hosted sync.
- Permanent daemon.
- Password-manager browser autofill.
- Direct chezmoi orchestration.
- Multiple vaults/profiles unless needed by implementation.

## Roadmap Implications

The roadmap should be vertical MVP style:

1. First phase should deliver a minimally usable encrypted vault through the existing CLI.
2. Second phase should handle migration, recovery, and doctor hardening.
3. Third phase should add the localhost manager with write-safety controls.
4. Later phases can improve recipient UX, restore tooling, and conflict handling.

## Sources

- `.planning/PROJECT.md`
- `.planning/codebase/ARCHITECTURE.md`
- `.planning/codebase/CONCERNS.md`
- `README.md`
- `docs/usage-spec.md`
- `docs/data-spec.md`
- https://pkg.go.dev/filippo.io/age
- https://github.com/C2SP/C2SP/blob/main/age.md
- https://pkg.go.dev/net/http
- https://go.dev/blog/routing-enhancements
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html

---
*Research summary for: Shelf Go encrypted vault*
*Researched: 2026-06-16*
