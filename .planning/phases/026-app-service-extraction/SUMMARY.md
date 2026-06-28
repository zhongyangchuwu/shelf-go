# Summary: Phase 26 App Service Extraction

## Outcome

Completed app-service extraction for export, setup helper, migration, and manager helper behavior. `internal/cli` now delegates reusable orchestration to `internal/app` while preserving command adapters, prompts, output routing, and process/server lifecycle.

## Implemented

- Added `internal/app/export.go`:
  - `ExportRequest`
  - `ExportSecrets`
  - direct secret export selector/filter/format orchestration
- Added `internal/app/migrate.go`:
  - `MigratePlaintextStore`
  - plaintext source preservation, target format detection, encrypted write, and verification
- Added `internal/app/setup.go`:
  - `InitConfig`
  - config path resolution
  - age identity creation/loading
  - vault file creation
  - config file writing
  - init path expansion and relative descendant path helpers
- Added `internal/app/manager.go`:
  - loopback-only listener creation
  - manager token generation
- Refactored CLI adapters:
  - `internal/cli/export.go` parses flags/args, loads runtime, calls `app.ExportSecrets`, and prints output.
  - `internal/cli/migrate.go` resolves vaults, calls `app.MigratePlaintextStore`, and prints user guidance.
  - `internal/cli/init.go` keeps prompt collection but delegates setup file/identity/vault helpers to `internal/app`.
  - `internal/cli/manager.go` delegates loopback listener and token helpers to `internal/app`, while keeping server lifecycle and signal handling.
- Moved helper-level tests out of CLI and added app package coverage for export, setup, migration, and manager helpers.

## Deviations

- CLI still calls `config.Resolve` after setup config writing to preserve current runtime resolution behavior and output paths.
- Export formatting still receives `st.Data.Secrets` because existing `internal/exportfmt` APIs accept the map shape; app owns this bridge now instead of CLI.
- Broad CLI test reduction remains deferred to Phase 27.

## Verification Evidence

- `gofmt` ran on changed Go files.
- `go test ./internal/app` passed.
- `go test ./internal/cli -run 'Test(Export|Secret.*Export|Setup|Migrate|Manager|Vault|Doctor)'` passed.
- `go test ./...` passed.

## Files Changed

- `internal/app/export.go`
- `internal/app/migrate.go`
- `internal/app/setup.go`
- `internal/app/manager.go`
- `internal/app/export_test.go`
- `internal/app/migrate_test.go`
- `internal/app/setup_test.go`
- `internal/app/manager_test.go`
- `internal/cli/export.go`
- `internal/cli/migrate.go`
- `internal/cli/init.go`
- `internal/cli/manager.go`
- `internal/cli/manager_test.go`
- Planning artifacts under `.planning/`
