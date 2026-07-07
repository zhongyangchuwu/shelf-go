# Review: JSON Codec and Repository Boundary

## Scope Reviewed

- `internal/vaultjson/codec.go`
- `internal/vaultjson/repository.go`
- `internal/vaultjson/codec_test.go`
- `.go-arch-lint.yml`

## Findings

- No blocking findings.
- The v2 package boundary is clear: `vaultjson` depends on `vaultmodel` and `util`, not on CLI/app/project/manager or v1 vault packages.
- Decode uses `DisallowUnknownFields` and rejects trailing JSON.
- Encode and Save validate through `vaultmodel.Validate` before returning bytes or writing files.
- Repository missing-file behavior is explicit and tested as `os.ErrNotExist`.
- Atomic write and backup behavior reuse existing `util.AtomicWrite` instead of duplicating filesystem code.

## Fixes Applied

- Removed a corrupted duplicate test file created during parallel test authoring and kept one canonical `codec_test.go` suite.
- Adjusted tests to assert structured round-trip behavior instead of fragile full-file JSON field ordering.
- Ensured invalid string-list decode reaches typed value validation by including required built-in field purpose metadata.
- Added `vaultjson` to `.go-arch-lint.yml` with a narrow dependency allowance.

## Waivers

- No age encryption adapter is included; Phase 36 explicitly scopes plaintext JSON only.
- No CLI/runtime integration is included; cutover must wait for migration and command phases.

## Remaining Risks

- Future codec work may need v1/v2 version sniffing once migration writes real files.
- Future encrypted repository must avoid plaintext temp-file leakage and should not reuse plaintext repository paths blindly.
