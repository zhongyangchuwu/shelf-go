# State

## Current Position
- Phase: Phase 15 - Shared Persistence Primitives and Store File Layout
- Status: complete
- Active Artifact: .planning/phases/015-persistence-store-layout/SUMMARY.md
- Next Action: Review architecture refactor milestone and decide whether to proceed to release readiness or further refactor.

## Blockers
- None

## Recent Evidence
- Phase 6 command hierarchy cutover completed on 2026-06-22.
- Implemented canonical commands: `shelf setup`, `shelf vault init`, `shelf vault migrate`, `shelf vault status`/`check`, `shelf vault open`, `shelf secret export`, and `shelf project run`.
- Old top-level `init`, `migrate`, `export`, `run`, and `manager` commands are intentionally absent.
- Phase 7 vault UX hardening completed on 2026-06-23.
- `shelf vault status`/`check` and `shelf doctor` now give recovery guidance for missing recipients, missing identities, plaintext stores, unsupported vault formats, and undecryptable vaults.
- Phase 8 project session activation/deactivation/shell design completed on 2026-06-23; implementation remains out of scope.
- `go test ./internal/cli -run 'Test(Vault|Doctor|Manager|Migrate|Setup)'` passed.
- `go test ./internal/store -run 'TestVault'` passed.
- `go test ./...` passed.
- Replaced stale public docs with a minimal open-source documentation set: README plus getting started, security, reference, troubleshooting, and contributing docs. Planning documents remain under `.planning/`.
- Safety and minimal project env UX milestone selected on 2026-06-24.
- Phase 9 plans `shelf project export` defaulting to existing shell output while retaining explicit `env`, `shell`, and `json` formats and avoiding a new dotenv format.
- Phase 9 project export shell default completed on 2026-06-24.
- `shelf project export` now defaults to sourceable shell output; explicit `env`, `shell`, and `json` formats remain available.
- `go test ./internal/cli -run TestProjectExport` passed.
- `go test ./...` passed.
- Phase 10 minimal vault backup recovery completed on 2026-06-24.
- Shelf recovery is intentionally manual: copy the single last-write encrypted `.bak` over the active vault, then run `shelf vault status`.
- `go test ./...` passed.
- Phase 11 secret edit and manager safety hardening completed on 2026-06-24.
- `shelf secret edit` temp files are explicitly `0600` and covered by cleanup tests.
- Manager localhost/token/cookie boundaries are covered by focused tests.
- `go test ./internal/cli -run TestSecretEdit` passed.
- `go test ./internal/manager -run TestManager` passed.
- `go test ./...` passed.
- Phase 12 restore simplification removed `shelf vault restore`; the command was unnecessary for single-slot `.bak` recovery.
- SQLite recorded as a deferred storage spike candidate; Dolt is not a current vault-storage candidate.
- Release readiness docs/infra added on 2026-06-25: LICENSE, SECURITY.md, CHANGELOG.md, CI workflow, and portable vault guide.
- `go test ./...` passed after release readiness updates.
- `go build ./cmd/shelf` passed after release readiness updates.
- Pre-release architecture refactor milestone selected on 2026-06-25.
- Phase 13 plan created to move runtime/vault loading into `internal/app` and project resolution/identity into `internal/project` while keeping `internal/cli` command-family oriented.
- Phase 13 completed on 2026-06-25: added `internal/app`, added `internal/project`, updated CLI project/run commands, and kept `internal/gitutil` deferred.
- `go test ./internal/project ./internal/cli -run 'TestProject|TestRun'` passed.
- `go test ./...` passed after Phase 13 extraction.
- Phase 14 plan created on 2026-06-25 to extract vault diagnostics and `secret edit` workflow without command behavior changes.
- Phase 14 completed on 2026-06-25: added `internal/vault`, added `internal/secret`, updated `vault status`, `doctor`, and `secret edit` to call feature packages.
- `go test ./internal/vault ./internal/secret ./internal/cli -run 'Test(Vault|Doctor|SecretEdit)'` passed.
- `go test ./...` passed after Phase 14 extraction.
- Phase 15 plan created on 2026-06-25 to centralize atomic writes, canonicalize validators, and split store file responsibilities without backend interfaces.
- Phase 15 completed on 2026-06-25: added `internal/atomicfile`, canonicalized env/path validators, and split `internal/store` into clearer files.
- `go test ./internal/store ./internal/manifest ./internal/render ./internal/cli -run 'Test(Vault|Setup|Manifest|Export)'` passed.
- `go test ./...` passed after Phase 15 extraction.

## Updated
- 2026-06-25
