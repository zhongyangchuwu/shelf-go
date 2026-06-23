# Plan: Phase 7 Vault UX Hardening

## Scope

Keep the encrypted vault contract unchanged. Harden CLI diagnostics and docs around existing vault lifecycle commands.

## Tasks

1. Add shared vault diagnostic guidance in `internal/cli` so `vault status` and `doctor` produce consistent next steps.
2. Extend status/check behavior for missing recipients, missing identities, plaintext stores, unsupported vault formats, and undecryptable vaults.
3. Add focused CLI tests for the new recovery guidance and `vault check` alias.
4. Update README and usage spec vault docs with first-run, migration cleanup, status/check, and open flows.
5. Run focused verification for CLI vault/doctor/manager/migration behavior plus store vault tests.

## Acceptance Criteria

- `shelf vault status` and `shelf vault check` do not reveal values.
- Plaintext stores recommend `shelf vault migrate --from <plaintext.json> --to <vault.age>` and manual plaintext cleanup.
- Missing recipients recommend `shelf vault init --recipient ...` or `shelf setup --recipient ...` before writes.
- Missing identities and wrong identities recommend fixing `identity_paths` or rerunning `shelf vault init`.
- Unsupported/corrupt/undecryptable vaults have concise recovery guidance without suggesting destructive overwrite first.
- Docs describe canonical vault commands only.

## Verification

- `go test ./internal/cli -run 'Test(Vault|Doctor|Manager|Migrate|Setup)'`
- `go test ./internal/store -run 'TestVault'`
- `go test ./...`
