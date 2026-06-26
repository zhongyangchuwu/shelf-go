# Phase 15 Plan: Shared Persistence Primitives and Store File Layout

## Goal

Centralize repeated safe file-write behavior, canonicalize validation helpers, and split `internal/store` source files by responsibility without changing command behavior or introducing a speculative backend interface.

## Scope

- Add `internal/atomicfile` for atomic writes with explicit file mode, directory mode, sync, and optional backup behavior.
- Use `atomicfile` from store vault/plain JSON writes, manifest saves, and setup config writes.
- Expose canonical env-name and path-token validation helpers from `internal/store`.
- Update manifest and render to use store validators.
- Split `internal/store/io.go` and `internal/store/vault.go` into clearer files within the same package.

## Non-goals

- No SQLite backend.
- No storage backend interface.
- No command output changes.
- No public format changes.

## Verification

- `go test ./internal/store ./internal/manifest ./internal/render ./internal/cli -run 'Test(Vault|Setup|Manifest|Export)'`
- `go test ./...`
