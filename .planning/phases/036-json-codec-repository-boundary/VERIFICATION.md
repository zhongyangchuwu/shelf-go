# Verification: JSON Codec and Repository Boundary

## Claims Checked

- v2 JSON codec round-trips valid vaults without losing item/field ordering or metadata.
- Decode rejects unknown fields, trailing JSON, unsupported versions, and invalid model data.
- Repository load reports missing files as `os.ErrNotExist`.
- Repository save/load round-trips through disk.
- Repository save rejects invalid vaults before replacing an existing target.
- Repository second save creates `.bak` containing the first save.
- Architecture lint accepts the new package boundary.

## Evidence Observed

- `internal/vaultjson/codec_test.go` covers codec and repository contracts with fake secret values only.
- `Repository.Save` calls codec encode before `util.AtomicWrite`, so invalid vaults fail before filesystem writes.
- `.go-arch-lint.yml` registers `vaultjson` with dependencies on `vaultmodel` and `util` only.

Commands run:

```text
go test ./internal/vaultjson
go test ./internal/vaultmodel ./internal/vaultjson
go test ./...
go vet ./...
go-arch-lint check
```

All commands passed.

## Coverage

- `V2-IMPL-03`: Covered by `internal/vaultjson` codec/repository implementation and focused tests for version checks, strict decode, round-trip preservation, atomic save, and backup behavior.

## Gaps

- v2 repository is not used by app/CLI yet.
- age encryption adapter is not implemented.
- Legacy migration is not implemented; this is Phase 37.

## Result

Pass. Phase 36 acceptance criteria are satisfied and work can proceed to Phase 37: Legacy Migration Adapter.
