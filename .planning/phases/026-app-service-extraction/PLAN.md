# Plan: Phase 26 App Service Extraction

## Objective

Move cross-package command orchestration from `internal/cli` into `internal/app` services without changing CLI behavior.

## Scope

- Extract direct secret export selection/filter/format orchestration from `internal/cli/export.go`.
- Extract plaintext migration implementation from `internal/cli/migrate.go`.
- Extract setup/vault-init reusable file and config/vault creation helpers from `internal/cli/init.go` while preserving prompts in CLI.
- Extract manager loopback listener and token generation helpers from `internal/cli/manager.go`.
- Add app package tests for reusable behavior.
- Keep CLI command contracts unchanged.

## Non-Goals

- No command rename, alias, or new command.
- No vault file format or manifest schema change.
- No project/session domain changes; Phase 25 completed those.
- No broad CLI test cleanup; Phase 27 owns test rebalancing.
- No moving shell completions or prompts into `internal/app`.

## Tasks

1. Add `internal/app` export service.
   - Define `ExportSecrets(st *vault.Store, req ExportRequest) (string, error)`.
   - Preserve path/prefix/tag selector behavior, `--all` filtering, format dispatch, and error strings.
   - Avoid direct CLI access to vault internals except where formatter APIs require store data.
2. Refactor `internal/cli/export.go`.
   - Parse CLI args/flags.
   - Load runtime/store.
   - Call `app.ExportSecrets` and print the returned output.
3. Add `internal/app` migration service.
   - Move plaintext migration implementation behind `MigratePlaintextStore(sourcePath string, targetVault *vault.Vault, force bool) error`.
   - Keep CLI responsible for resolving target vault and printing guidance.
4. Add setup helper/service APIs.
   - Move config path resolution, identity ensure, vault file ensure, config file ensure, path expansion, and relative path helper into app.
   - Keep CLI `initConfig.fill` prompt collection in CLI, or replace it with a request built by CLI and passed to app.
5. Add manager helper APIs.
   - Move loopback listener validation and token generation into app.
   - Keep server construction, serve loop, signal handling, and printed URL in CLI.
6. Add app package tests.
   - Cover export selector behavior and format errors.
   - Cover manager loopback validation/token shape.
   - Cover setup helper behavior where practical without prompting.
   - Keep migration coverage equivalent to existing CLI tests.
7. Format and verify.
   - Run `gofmt` on changed Go files.
   - Run focused app/CLI tests.
   - Run full `go test ./...`.
8. Record summary, verification, capture, update root planning, and commit Phase 26.

## Acceptance Criteria

- `internal/cli/export.go` no longer owns reusable export selector/filter/format orchestration.
- `internal/cli/migrate.go` no longer owns plaintext migration implementation.
- Reusable setup file/identity/vault creation helpers live outside CLI while prompts stay in CLI.
- Manager loopback validation and token generation live outside CLI.
- `internal/app` services do not import Cobra or CLI.
- Existing command behavior remains unchanged.

## Verification

- `gofmt` on changed Go files.
- `go test ./internal/app`
- `go test ./internal/cli -run 'Test(Export|Secret.*Export|Init|Migrate|Manager|Vault|Doctor)'`
- `go test ./...`

## Risks

- Setup paths and relative config values are easy to subtly change; preserve current helpers exactly when moving.
- Export format behavior may leak values by design; tests must distinguish intended export output from diagnostic/no-leak paths.
- Migration must continue preserving plaintext source bytes exactly and verifying encrypted target load.
- Manager listener helper must preserve loopback-only validation.
