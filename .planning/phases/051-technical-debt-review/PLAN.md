# Plan: Repository Technical Debt Review

**Status:** Implemented and verified.

## Goal

Review the active Go packages and user-facing documentation, then remove confirmed correctness, security, maintainability, and documentation debt without expanding the personal-use product scope.

## Scope

- Active production packages under `cmd/shelf` and `internal/**`.
- User-facing root/docs/security/contributor documentation.
- Current tests, architecture lint, and verification scripts.

## Confirmed Remediation

1. Preserve restrictive backup permissions whenever a vault backup path already exists.
2. Derive create-result primary env metadata from template roles instead of built-in field IDs.
3. Remove an impossible `panic`, an unused setup helper, and needless map allocations in status labels.
4. Align stale architecture and security documentation with shipped secret/project TUI workspaces and the unified `secret update` command.

## Non-goals

- New product capabilities, provider integrations, release packaging, encryption, migration, Web UI, or team sync.
- Redesigning the current plaintext/external-encryption model.

## Acceptance

- Existing backup files are rewritten with the requested restrictive mode.
- A custom template’s primary env name appears in the safe create summary.
- No stale documentation claims that shipped TUI workflows are absent or that `secret set` is active.
- Focused regression tests and `bash scripts/test.sh` pass.