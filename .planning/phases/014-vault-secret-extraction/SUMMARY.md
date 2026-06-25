# Phase 14 Summary: Vault Diagnostics and Secret Workflow Extraction

## Result

Complete. Vault diagnostic rules and `secret edit` workflow now live outside `internal/cli`.

## Changes

- Added `internal/vault`:
  - typed diagnostic levels and checks
  - `Status` report for `shelf vault status` / `check`
  - `Doctor` report for `shelf doctor`
  - vault format/load/tracking detail helpers
- Added `internal/secret`:
  - editable secret JSON model
  - `Edit` workflow for temp file creation, restrictive permissions, editor invocation, JSON parse, and store update
- Updated `internal/cli`:
  - `vault status` renders `vault.Status`
  - `doctor` renders `vault.Doctor` and keeps completion checks in CLI
  - `secret edit` calls `secret.Edit`

## Verification

Passed:

```bash
go test ./internal/vault ./internal/secret ./internal/cli -run 'Test(Vault|Doctor|SecretEdit)'
go test ./...
```

## Requirements

- ARCH-04 complete: vault diagnostic rules are reusable outside CLI through `internal/vault`.
- ARCH-05 complete: `secret edit` workflow is reusable outside CLI through `internal/secret` while preserving temp-file cleanup behavior.

## Follow-up

Phase 15 should centralize atomic file writes, canonicalize env/path validation, and split `internal/store` files by responsibility without introducing a speculative backend interface.
