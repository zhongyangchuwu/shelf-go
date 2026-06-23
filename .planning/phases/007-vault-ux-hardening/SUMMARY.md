# Summary: Phase 7 Vault UX Hardening

## Completed Changes

- Added shared CLI guidance helpers for vault format and vault load failures.
- Extended `shelf vault status` / `shelf vault check` with recipient-count diagnostics and recovery guidance.
- Reused the same recovery guidance from `shelf doctor` for vault format/load failures.
- Added focused tests for missing recipients, missing identities, plaintext legacy stores, unsupported vault formats, undecryptable vaults, and the `vault check` alias.
- Updated README and usage spec first-run, status/check, migration cleanup, and vault manager guidance.

## Files Changed

- `internal/cli/manager.go`
- `internal/cli/doctor.go`
- `internal/cli/manager_test.go`
- `internal/cli/doctor_test.go`
- `README.md`
- `docs/usage-spec.md`
- `.planning/phases/007-vault-ux-hardening/CONTEXT.md`
- `.planning/phases/007-vault-ux-hardening/PLAN.md`

## Deviations

- No vault storage behavior changed. The phase stayed in CLI diagnostics and docs.

## Evidence

- `go test ./internal/cli -run 'Test(Vault|Doctor|Manager|Migrate|Setup)'` passed.
- `go test ./internal/store -run 'TestVault'` passed.
- `go test ./...` passed.
