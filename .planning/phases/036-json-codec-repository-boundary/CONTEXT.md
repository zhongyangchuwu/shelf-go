# Context: JSON Codec and Repository Boundary

## Goal

Implement v2 JSON encode/decode and a plaintext JSON repository behind a repository/codec boundary, using `internal/vaultmodel` from Phase 35.

## Constraints

- Do not change current v1 vault persistence or CLI behavior in this phase.
- Do not add age encryption yet; encrypted repository behavior is a later adapter step.
- Keep JSON strict enough to catch unsupported versions, malformed input, unknown top-level fields, and invalid model data before save.
- Preserve item ordering and field ordering on round-trip.
- Follow existing atomic write and backup conventions where applicable.

## Decisions

- Put v2 storage boundary in a new internal package that depends on `internal/vaultmodel`, not on CLI/app/project/manager.
- Keep codec pure over bytes; repository owns file I/O and atomic persistence.
- Use tests with fake values only.

## Open Questions

- Package name: `internal/vaultjson`, `internal/vaultrepo`, or similar should be chosen to avoid confusion with existing v1 `internal/jsonvault`.
- Whether the first repository writes plaintext `.json` only or uses a distinct extension to avoid confusing it with encrypted `vault.age`.
- Exact backup naming convention for v2 repository writes.

## Verification Expectations

- JSON codec round-trips v2 vaults without losing ordering or field metadata.
- Strict decode rejects unknown fields, trailing data, unsupported versions, and invalid model data.
- Repository save/load uses atomic write behavior and validates before save.
- Existing v1 tests continue to pass.
