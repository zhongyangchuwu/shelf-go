# Verification: Phase 26 App Service Extraction

## Claims Verified

1. `internal/app` owns reusable export orchestration.
2. `internal/app` owns plaintext migration implementation.
3. `internal/app` owns reusable setup file/identity/vault helpers while prompts remain in CLI.
4. `internal/app` owns manager loopback/token helpers while server lifecycle remains in CLI.
5. CLI command behavior remains unchanged.

## Evidence

### `go test ./internal/app`

Result: Passed.

Coverage from this check:

- Export prefix selection filters secrets without env by default.
- Export `--all`-equivalent behavior derives env names for secrets without env.
- Tag export selection preserves AND semantics.
- Export validates missing selector, no-match, and unsupported format errors.
- Manager listener rejects non-loopback addresses and accepts loopback addresses.
- Manager token generation returns non-empty random tokens.
- Config file writing uses relative descendant paths.
- Vault file creation is idempotent.
- Relative path helper rejects outside paths.
- Plaintext migration preserves source bytes and writes encrypted target content.
- Plaintext migration refuses existing encrypted targets without force.

### `go test ./internal/cli -run 'Test(Export|Secret.*Export|Setup|Migrate|Manager|Vault|Doctor)'`

Result: Passed.

Coverage from this check:

- CLI export commands still read encrypted vaults and emit shell/env/json formats.
- Setup and vault init still create files and preserve existing vault contents.
- Migration CLI still reports cleanup guidance, refuses existing targets, and supports force migration with encrypted backup.
- Manager command remains registered.
- Vault status/check and doctor-related command paths still pass focused checks.

### `go test ./...`

Result: Passed.

Coverage from this check:

- All packages compile with the new app services.
- Full test suite remains green after CLI-to-app extraction.

## Acceptance Criteria Mapping

| Acceptance Criterion | Evidence | Result |
| --- | --- | --- |
| `internal/cli/export.go` no longer owns reusable export orchestration | CLI delegates to `app.ExportSecrets`; app and CLI export tests passed | Passed |
| `internal/cli/migrate.go` no longer owns plaintext migration implementation | CLI delegates to `app.MigratePlaintextStore`; app and CLI migrate tests passed | Passed |
| Setup helpers live outside CLI while prompts stay in CLI | `app.Ensure*` helpers and `InitConfig` added; CLI still owns prompt collection; setup tests passed | Passed |
| Manager loopback/token helpers live outside CLI | `app.ListenLoopback` and `app.ManagerToken` added; app manager tests passed | Passed |
| `internal/app` services do not import Cobra or CLI | New app service files compile without Cobra imports | Passed |
| Existing command behavior remains unchanged | Focused CLI tests and `go test ./...` passed | Passed |

## Gaps

- Phase 27 still owns broad CLI test rebalancing and final package-boundary capture.
- Export formatter APIs still accept the vault secret map; app now owns that bridge, but a future formatter API could further reduce map exposure.
