# Context: Phase 10 Vault Restore and Recovery

## Goal

Make encrypted vault backup recovery explicit, safe, and documented through a vault-scoped restore workflow.

## Constraints

- Restore must operate on encrypted Shelf vault files, not plaintext JSON stores.
- Restore must validate decrypted backup contents before replacing a target vault.
- Restore must refuse overwriting an existing target unless `--force` is supplied.
- Restore must not create plaintext backups as a side effect.
- Keep command placement under `shelf vault`.
- Keep implementation inside current package boundaries: CLI orchestration in `internal/cli`, vault persistence in `internal/store`, docs in `docs/`.

## Decisions

- Add `shelf vault restore --from <backup.age> [--to <vault.age>] [--force]`.
- Default restore target is the active configured vault path.
- Restore decrypts the source with configured identities, validates the store model, then saves it to the target using configured recipients.
- Plaintext JSON restore sources are rejected; users should use `shelf vault migrate` for plaintext stores.
- Plaintext target paths are rejected instead of overwritten to avoid `writeStoreFile` preserving a plaintext `.bak`.

## Open Questions

None for this phase.

## Verification Expectations

- Tests cover restore success, overwrite refusal, force restore, plaintext source rejection, invalid encrypted backup rejection, and target plaintext rejection.
- Docs explain backup restore, identity requirements, and `shelf vault status` verification after restore.
