# Phase 15 Summary: Shared Persistence Primitives and Store File Layout

## Result

Complete. Repeated safe write logic is centralized, validation helpers are canonicalized, and `internal/store` is split by responsibility without introducing a backend interface.

## Changes

- Added `internal/atomicfile`:
  - atomic temp write
  - explicit file mode and directory mode
  - optional sync
  - optional last-write `.bak` backup
- Updated persistence writes:
  - store/vault writes use `atomicfile` with `0600`, sync, and backup
  - manifest saves use `atomicfile` with `0644` and sync
  - setup config writes use `atomicfile` with `0600`
- Canonicalized validators:
  - `store.IsEnvName`
  - `store.ValidateEnvName`
  - `store.IsPathToken`
  - manifest and render now reuse store validation helpers
- Split `internal/store` files:
  - `io.go`: plaintext load/save and store-file write orchestration
  - `json.go`: store encode/decode and data validation
  - `store.go`: `Store` methods
  - `age.go`: age open/seal/identity loading
  - `vault.go`: file format detection and vault orchestration

## Verification

Passed:

```bash
go test ./internal/store ./internal/manifest ./internal/render ./internal/cli -run 'Test(Vault|Setup|Manifest|Export)'
go test ./...
```

## Requirements

- ARCH-06 complete: atomic file writes are centralized with explicit permissions, sync, and backup options.
- ARCH-07 complete: env name and path token validation use canonical store helpers.
- ARCH-08 complete: `internal/store` remains one package but has clearer source-file responsibilities.

## Follow-up

No backend interface was added. SQLite remains a future spike candidate only when storage/query pressure justifies it.
