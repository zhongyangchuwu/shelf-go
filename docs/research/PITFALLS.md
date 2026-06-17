# Pitfalls Research

**Domain:** CLI-first solo developer secret manager with age-encrypted portable vault
**Researched:** 2026-06-16
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Side-Channel Plaintext Files

**What goes wrong:**
The main vault is encrypted, but backups, migration files, editor temp files, logs, or test fixtures keep real secrets in plaintext.

**Why it happens:**
Encryption gets added to the happy path only, while existing write and edit helpers continue writing plaintext.

**How to avoid:**
Make the encrypted vault boundary mandatory for saves and backups. Audit `secret edit`, migration, tests, and error paths for plaintext writes.

**Warning signs:**
New files named `*.json`, `*.bak`, `*.tmp`, or editor swap files appear near vault operations and contain secret-shaped values.

**Phase to address:**
Phase 1 storage encryption and Phase 2 migration/hardening.

---

### Pitfall 2: Unsafe Localhost Editor

**What goes wrong:**
A malicious web page can trigger writes against `http://127.0.0.1:<port>` if the local manager trusts browser requests without CSRF/Origin controls.

**Why it happens:**
Developers assume localhost is private, but browsers can still send cross-origin requests to local services.

**How to avoid:**
Bind to loopback only, generate a per-session token, require CSRF tokens for writes, verify Origin/Host, and never use GET for state changes.

**Warning signs:**
Handlers accept POST/PUT/DELETE without a token or Origin check; write routes are reachable with simple form submissions.

**Phase to address:**
Localhost manager phase before edit endpoints ship.

---

### Pitfall 3: Breaking Existing CLI Workflows

**What goes wrong:**
Encryption changes command behavior, output formats, or project/run semantics, making the tool feel less reliable.

**Why it happens:**
Storage migration and command redesign are implemented together.

**How to avoid:**
Keep CLI semantics stable while changing persistence. Add regression tests around `secret`, `export`, `project`, and `run`.

**Warning signs:**
Tests need large rewrites for unrelated command behavior; users must learn new commands just to use encrypted storage.

**Phase to address:**
Phase 1 and every storage-touching phase.

---

### Pitfall 4: Ambiguous Key and Recipient Configuration

**What goes wrong:**
Users cannot tell which key decrypts the vault, or Shelf stores private identity material in config.

**Why it happens:**
Recipient public keys, identity private keys, and chezmoi conventions get mixed together.

**How to avoid:**
Store only non-secret recipients and paths in Shelf config. Document identity lookup clearly and validate missing identities with actionable errors.

**Warning signs:**
Config examples include private age identity contents, or errors say only "decrypt failed" without path/key context.

**Phase to address:**
Phase 1 config and decrypt UX.

---

### Pitfall 5: Migration Without Rollback

**What goes wrong:**
A user migrates a plaintext store to an encrypted vault, then discovers decrypt/load fails or data is missing.

**Why it happens:**
Migration writes the new format before verifying it or removes the old file too early.

**How to avoid:**
Copy/backup first, encrypt to target, decrypt target, validate full data, then report next manual cleanup steps.

**Warning signs:**
Migration code deletes or overwrites source files before decrypt-verifying the target.

**Phase to address:**
Phase 2 migration and recovery.

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Shell out to `age` | Fast prototype | External binary dependency and harder error handling | Only as a spike, not production path. |
| Keep plaintext backups | Easy rollback | Defeats encrypted vault purpose | Never for real secrets. |
| Put web handlers in `internal/cli` | Faster first patch | Large command files become harder to test | Only for tiny command glue. |
| One generic "decrypt failed" error | Simple error path | Bad UX and hard support | Never for user-facing migration. |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| age | Confusing recipients with identities | Recipients are public encryption targets; identities are private decrypt material. |
| chezmoi | Trying to control chezmoi from Shelf | Produce a normal encrypted file; let chezmoi manage it. |
| Git | Checking only `.gitignore` | Verify tracked state and whether the tracked file is encrypted. |
| Browser UI | Trusting local requests | Require session/CSRF protections and loopback binding. |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Whole-file encrypt on every write | Slow edits on huge vaults | Accept for small stores; benchmark before redesign. | Thousands of large structured secrets. |
| Repeated decrypt for UI requests | UI feels sluggish | Load once per server session, save through locked writes. | Search/edit loops over large vaults. |
| Over-abstracting backends early | Slow delivery | Keep a narrow persistence boundary first. | Before there is more than one real backend. |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Plaintext temp files | Secrets leak through `/tmp`, editor files, or backups. | Private temp dir, `0600`, cleanup, and no unnecessary value temp files. |
| GET write endpoints | Browser/link prefetch can mutate secrets. | Writes only through non-GET methods plus CSRF. |
| Long-lived unlocked server | Any local browser/process has more time to attack. | Short-lived sessions, explicit close, timeout. |
| Logging secret values | Secrets leak into terminal, CI, or debug logs. | Tests that non-value commands never print values. |
| Non-strict decode | Corrupt data can be partially accepted. | Reject trailing JSON and unknown/unsupported versions. |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Asking users to edit raw JSON for normal changes | Error-prone and unpleasant. | Local UI and focused CLI edit/set commands. |
| Hiding where the vault lives | Users cannot manage it with chezmoi. | Clear `doctor` and config output. |
| Cryptic decrypt errors | Users cannot fix key/config problems. | Distinguish missing identity, wrong identity, unreadable vault, and corrupt vault. |
| Surprising `.env` generation | Risk of committed plaintext. | Keep `.shelf.json` as manifest and make exports explicit. |

## "Looks Done But Isn't" Checklist

- [ ] **Encryption:** Main vault is encrypted, and backups/temp/migration outputs are not plaintext.
- [ ] **Migration:** New vault decrypts and validates before the old store is touched.
- [ ] **CLI compatibility:** Existing command tests pass against encrypted storage.
- [ ] **Doctor:** Reports tracked plaintext files and encrypted vault status.
- [ ] **Localhost editor:** Write routes require CSRF/session controls and loopback binding.
- [ ] **Recovery:** Users have a documented way to restore or retry after failed writes.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Plaintext backup leaked | HIGH | Rotate affected secrets; remove file from git history if committed; fix backup path. |
| Migration produced unreadable vault | MEDIUM | Restore from pre-migration copy; rerun after fixing recipient/identity config. |
| Localhost write vulnerability found | HIGH | Disable edit routes; patch token/Origin/CSRF handling; add regression tests. |
| CLI regression after encryption | MEDIUM | Revert command behavior change; preserve storage change behind compatibility tests. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Side-channel plaintext files | Storage encryption/hardening | Search generated files and tests for known secret values. |
| Unsafe localhost editor | Vault manager phase | Security tests for no GET writes, token required, Origin rejected. |
| Breaking CLI workflows | Every phase | Existing command test suite passes. |
| Ambiguous key config | Encryption config phase | Error tests for missing/wrong identity. |
| Migration without rollback | Migration phase | Migration test decrypts target and preserves source until success. |

## Sources

- `.planning/codebase/CONCERNS.md` - existing plaintext, temp-file, and validation risks.
- https://pkg.go.dev/filippo.io/age - age recipient/identity APIs.
- https://github.com/C2SP/C2SP/blob/main/age.md - age format specification.
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html - CSRF prevention guidance.

---
*Pitfalls research for: Shelf Go encrypted vault*
*Researched: 2026-06-16*
