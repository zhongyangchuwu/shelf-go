# Technical Debt Review

**Scope:** Current `cmd/shelf`, all active `internal/**` packages, root/user documentation, security documentation, tests, and architecture lint.

## Baseline

- `bash scripts/test.sh` passed before remediation.
- Go diagnostics reported no issues in reviewed source files.
- No active TODO/FIXME/HACK/deprecated markers were found in production code or user docs.

## Findings

## Finding: Existing vault backup permissions can remain too broad

- Severity: High
- Location: `internal/util/atomicfile.go`, `copyFile`
- Problem: `os.OpenFile` applies `mode` only when it creates the backup. An existing `<vault>.bak` keeps its prior mode, so a previously broad backup can continue to contain plaintext secret values after every later vault write.
- Cost: Vault backup confidentiality can diverge from the `0o600` primary vault contract.
- Fix: Explicitly chmod the opened backup file to the requested mode before copying.
- Verification: Create a `0o644` pre-existing backup, write with `Backup: true` and `FileMode: 0o600`, then assert prior content and repaired `0o600` permissions.

## Finding: Create summary hard-codes built-in env field IDs

- Severity: Medium
- Location: `internal/cli/secret.go`, `primaryEnvFromInfo`
- Problem: The CLI supports custom templates and resolves env-name roles generically, but the create summary recognizes only `value_env` and `credential_env`.
- Cost: Creating a supported custom template can succeed while presenting an empty env line in the safe success summary.
- Fix: Populate non-serialized primary-env metadata in `app.SecretInfo` through the canonical resolver and render that field directly.
- Verification: Create a custom template with a differently named env field and assert both app metadata and CLI summary contain its env name.

## Finding: Architecture and security documentation describe removed behavior

- Severity: Medium
- Location: `docs/architecture.md`, `SECURITY.md`
- Problem: Architecture docs say TUI flows are planned/removed; security docs name the removed `secret set` command and say no TUI editor exists.
- Cost: Users and maintainers receive an incorrect product boundary and unsafe/stale command guidance.
- Fix: Document the shipped secret/project workspaces, their app/clipboard boundaries, and current `secret update --stdin` write path.
- Verification: Documentation review against active command constructors and architecture lint boundaries.

## Finding: Local impossible panic and dead setup helper

- Severity: Low
- Location: `internal/app/secret.go`, `internal/cli/init.go`, `internal/app/setup.go`
- Problem: String JSON encoding cannot fail, yet `rawString` panics; `initConfigExists` is unreferenced; two status labels allocate temporary maps for boolean lookup.
- Cost: Avoidable failure vocabulary, dead code, and small needless allocations obscure simple workflows.
- Fix: Use allocation-free string quoting, delete the unused helper/import, and use direct conditionals.
- Verification: Existing secret/setup behavior tests and full repository verification.

## Review Result

All confirmed findings above are bounded remediations. No blocker was found in current CLI, TUI, vault-model, project-resolution, or documentation behavior beyond the listed items.
