# Phase 13 Summary: App Runtime and Project Package Extraction

## Result

Complete. Runtime/vault construction and project resolution behavior now live outside `internal/cli` while CLI remains command-family oriented.

## Changes

- Added `internal/app`:
  - `LoadVault`
  - `LoadRuntime`
  - `ReadVault`
  - `UpdateVault`
- Added `internal/project`:
  - project manifest resolution and diagnostics
  - project render binding conversion
  - project ID, Git root lookup, and remote normalization
- Updated `internal/cli`:
  - root helpers delegate to `internal/app`
  - project commands call `internal/project`
  - project run uses `project.Binding`
- Kept `internal/gitutil` deferred. Project-owned git helpers remain in `internal/project` until another package needs shared git helpers.

## Verification

Passed:

```bash
go test ./internal/project ./internal/cli -run 'TestProject|TestRun'
go test ./...
```

## Requirements

- ARCH-01 complete: runtime/vault construction helpers moved outside CLI.
- ARCH-02 complete: project resolution, diagnostics, binding conversion, identity, Git root lookup, and remote normalization moved to `internal/project`.
- ARCH-03 complete for this phase: `internal/cli` remains command-family oriented; no one-file-per-subcommand split was introduced.

## Follow-up

Phase 14 should extract vault diagnostics and secret edit workflow after reviewing current command output expectations.
