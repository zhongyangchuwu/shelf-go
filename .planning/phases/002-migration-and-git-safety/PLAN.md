# Plan: Migration and Git Safety

## Objective

Implement Phase 2 requirements MIGR-01 through MIGR-05 and SAFE-01 through SAFE-05 with a safe plaintext-to-encrypted migration path and doctor checks that identify unsafe tracked plaintext state.

## Scope

In scope:
- New CLI migration command for plaintext JSON store to age-encrypted vault.
- Store helpers only where needed to identify encrypted-vault vs plaintext-store files.
- Doctor checks for vault format, plaintext-at-active-vault errors, tracked plaintext warnings, and encrypted tracked-vault confirmation.
- Behavior-focused command tests.

Out of scope:
- Deleting or archiving plaintext source automatically.
- Chezmoi command integration.
- Password-only encryption.
- UI manager work.

## Tasks

1. Add store format classification helpers.
2. Add `shelf migrate --from <plaintext.json> [--to <vault.age>] [--force]`.
3. Verify migration by decrypting and validating the target before reporting success.
4. Preserve source plaintext and avoid durable plaintext temp files in migration writes.
5. Extend `shelf doctor` git safety checks using format classification and git tracked state.
6. Add command tests for migration and doctor safety behavior.
7. Run focused tests and record summary evidence.

## Acceptance Criteria

- MIGR-01: A command migrates a valid plaintext Shelf JSON store into an age-encrypted vault.
- MIGR-02: The plaintext source remains byte-for-byte unchanged after success and failure.
- MIGR-03: Migration output names the encrypted target and tells the user the plaintext source remains and should be moved/deleted/archived manually.
- MIGR-04: Replacing an existing vault creates only encrypted backup bytes for secret values.
- MIGR-05: Migration does not create durable plaintext temp or backup files.
- SAFE-01: Doctor accepts a normal encrypted vault file as portable git-trackable state.
- SAFE-02: Config remains metadata-only; no migration path writes secret values to config or `.shelf.json`.
- SAFE-03: Doctor reports active store format as encrypted, plaintext, missing, or unsupported.
- SAFE-04: Doctor fails or warns when a tracked active secret file is plaintext JSON.
- SAFE-05: Doctor confirms when a tracked active vault is encrypted and project manifests remain value-free by construction.

## Verification

- `go test ./internal/store ./internal/cli`
- `go test ./...` if focused tests pass and changed behavior crosses package boundaries.

## Risks

- Git detection can be brittle outside a repository; checks must degrade to explicit ok/warn output instead of failing healthy non-git users.
- `--force` replacement must not leak the old encrypted vault or plaintext source through plaintext backup behavior.
- Tests should inspect bytes for absence of known secret values rather than asserting age ciphertext internals.
