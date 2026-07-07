# Summary: JSON Codec and Repository Boundary

## Completed Changes

- Added `internal/vaultjson` as the v2 plaintext JSON codec/repository package.
- Added `JSONCodec` with strict decode and validated encode for `vaultmodel.Vault`.
- Added decode rejection for unknown fields, trailing JSON, unsupported versions, and invalid model data.
- Added `Repository` with `Load`, `Save`, `Path`, and injectable codec construction.
- Implemented repository save through existing atomic write/backup conventions.
- Added tests for round-trip ordering/metadata, strict decode failures, missing file behavior, save/load, invalid-save preservation, and `.bak` creation.
- Updated architecture lint config to recognize `internal/vaultjson` and allow dependencies only on `vaultmodel` and `util`.

## Files Changed

- `.go-arch-lint.yml`
- `internal/vaultjson/codec.go`
- `internal/vaultjson/repository.go`
- `internal/vaultjson/codec_test.go`
- `.planning/phases/036-json-codec-repository-boundary/CONTEXT.md`
- `.planning/phases/036-json-codec-repository-boundary/PLAN.md`

## Deviations

- `Repository.Load` returns `os.ErrNotExist` for missing files rather than returning an empty vault. This avoids hiding misconfigured paths and is explicitly covered by tests.
- The first v2 repository is plaintext JSON only. age encryption remains deferred to a later repository adapter.
- The package name is `internal/vaultjson` to distinguish v2 JSON storage from existing v1 `internal/jsonvault`.

## Evidence

- `go test ./internal/vaultjson` passed.
- `go test ./internal/vaultmodel ./internal/vaultjson` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go-arch-lint check` passed.

## Unresolved Risks

- The plaintext repository is not wired into CLI/runtime paths yet, so no user-facing storage behavior changes in this phase.
- age encryption and recipient/identity integration remain for a future adapter phase.
- Migration code still needs to decide target path/extension and backup UX when writing v2 vaults.
