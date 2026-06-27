# Plan: Phase 22 Architecture Repartition Core

## Objective

Refactor Shelf Go's internal architecture so package names match product boundaries and the manager command has one clear pre-release entrypoint.

## Scope

- Replace `shelf vault open` with `shelf manager`.
- Keep `internal/manager` as the local manager surface package.
- Move `internal/manifest` into `internal/project`.
- Rename `internal/store` to `internal/vault` and merge old `internal/vault` diagnostics plus `internal/atomicfile` into it.
- Move `internal/version` into `internal/app`.
- Rename `internal/render` to `internal/exportfmt`.
- Update imports, package names, tests, and minimal architecture references needed to keep code/docs consistent.
- Do not perform the full user/developer documentation rewrite; that is the next phase.

## Tasks

1. Update root planning artifacts to split architecture, docs, and release phases.
2. Move vault core files/tests into `internal/vault` and resolve package naming conflicts.
3. Move project manifest files/tests into `internal/project`.
4. Move version composition into `internal/app`.
5. Rename `internal/render` to `internal/exportfmt` and update callers.
6. Replace `shelf vault open` registration/tests with `shelf manager`.
7. Run formatter and focused tests while resolving import/package issues.
8. Update phase summary, verification, capture, and state.

## Acceptance Criteria

- `internal/store`, `internal/manifest`, `internal/atomicfile`, and `internal/version` are removed.
- Vault core and diagnostics live under `internal/vault`.
- Project manifest schema/IO/validation live under `internal/project`.
- Export formatting lives under `internal/exportfmt`.
- `shelf manager` is the only manager command entrypoint.
- `shelf vault open` is not registered.
- Behavior remains unchanged apart from the intentional manager command rename.
- Planning records show docs deferred to the next phase and release hardening deferred after docs.

## Verification

- `gofmt` on changed Go files.
- `go test ./internal/vault`
- `go test ./internal/project`
- `go test ./internal/cli -run 'Test.*Manager|TestVault'`
- `go test ./...`
- `go vet ./...`

## Risks

- Moving `store` to `vault` touches many imports; missed references will fail compile.
- Merging old vault diagnostics with store vault core can create naming conflicts; resolve by keeping existing exported diagnostic names stable where possible.
- Removing `vault open` is an intentional user-visible breaking change; tests and docs must not imply alias compatibility.
- Broad documentation remains incomplete until the next phase; minimal architecture references must still be accurate enough after code moves.
