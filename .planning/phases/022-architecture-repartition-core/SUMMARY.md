# Summary: Phase 22 Architecture Repartition Core

## Completed Changes

- Replaced the vault-scoped manager entrypoint with a single top-level `shelf manager` command.
- Removed `shelf vault open` registration and tests now assert it is absent.
- Renamed `internal/store` to `internal/vault` and merged old vault diagnostics plus atomic write support into the vault core package.
- Moved `.shelf.json` schema, IO, validation, and tests from `internal/manifest` into `internal/project`.
- Moved version composition from `internal/version` into `internal/app`.
- Renamed `internal/render` to `internal/exportfmt`.
- Updated imports, package names, tests, and minimal architecture references.
- Split roadmap sequencing into architecture repartition, docs alignment, and release hardening phases.

## Files Changed

- `internal/cli/*`
- `internal/app/runtime.go`
- `internal/app/version.go`
- `internal/exportfmt/export.go`
- `internal/manager/*`
- `internal/project/*`
- `internal/secret/*`
- `internal/vault/*`
- `cmd/shelf/main.go`
- `README.md`
- `docs/architecture.md`
- `docs/contributing.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/022-architecture-repartition-core/*`

## Deviations

- Minimal docs were updated in this phase to avoid stale command/package references after the code cutover. The full docs rewrite remains deferred to Phase 23.
- `internal/project` and `internal/secret` were kept as independent feature packages per architecture decision instead of being folded into `internal/app`.
- `internal/manager` package name was kept because it represents the broader local manager surface, not only vault or Web UI.

## Evidence

- `go test ./internal/vault` passed.
- `go test ./internal/project` passed.
- `go test ./internal/cli -run 'Test.*Manager|TestVault'` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- LSP workspace diagnostics reported no Go issues.
- `grep` found no active code/docs references to removed internal package paths.
- `glob internal/*` shows the final package set: `app`, `cli`, `config`, `exportfmt`, `manager`, `project`, `secret`, `vault`.

## Unresolved Risks

- Phase 23 still needs the full user/developer documentation pass for manager, tag workflows, scripts, and final architecture.
- Phase 24 still needs release hardening and final snapshot release evidence.
