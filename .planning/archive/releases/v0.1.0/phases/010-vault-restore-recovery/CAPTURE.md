# Capture: Phase 10 Vault Restore and Recovery

## Durable Docs Updated

- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/security.md`
- `docs/troubleshooting.md`

## Planning Records Updated

- `.planning/phases/010-vault-restore-recovery/CONTEXT.md`
- `.planning/phases/010-vault-restore-recovery/PLAN.md`
- `.planning/phases/010-vault-restore-recovery/SUMMARY.md`
- `.planning/phases/010-vault-restore-recovery/VERIFICATION.md`

## Learnings

- Restore should stay encrypted-only; plaintext JSON remains migration's responsibility.
- Rejecting plaintext targets avoids preserving plaintext `.bak` files during restore overwrite.
- Recovery depends on age identity availability; this belongs in user-facing troubleshooting and security docs.

## Ship Inputs

- `go test ./internal/cli -run TestVaultRestore` passed.
- `go test ./...` passed.
- Next planned phase is secret edit and manager safety hardening.
