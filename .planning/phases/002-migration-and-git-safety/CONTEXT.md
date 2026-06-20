# Context: Migration and Git Safety

## Goal

Users can convert an existing plaintext Shelf JSON store into an age-encrypted vault and use `shelf doctor` to distinguish unsafe plaintext state from git/chezmoi-safe encrypted state.

## Constraints

- Migration must never delete, rename, truncate, or overwrite the plaintext source.
- Migration must write only encrypted bytes to the target vault and encrypted backup path.
- The encrypted target must be decrypted and validated before migration reports success.
- Existing package boundaries stay intact: CLI orchestration in `internal/cli`, persistence in `internal/store`, config resolution in `internal/config`.
- `shelf doctor` must not print secret values.
- Config and `.shelf.json` remain non-secret; age recipients and identity paths are acceptable config metadata, not private key material.

## Decisions

- Add a top-level `shelf migrate` command for plaintext-to-vault conversion.
- Use explicit `--from` and `--to` flags; default `--to` is the active configured vault path.
- Require configured recipients to encrypt the target vault and configured identities to verify it after write.
- Refuse to overwrite an existing target unless `--force` is set.
- Keep backup behavior inside the existing encrypted write path; replacing an existing vault creates an encrypted `.bak` through `store.Vault.Save`.
- Extend `doctor` with format-aware checks: encrypted vault files are ok; plaintext JSON at the active vault path is a failure requiring migration.

## Open Questions

- None for Phase 2 execution.

## Verification Expectations

- Unit/command tests prove migration success, no plaintext target bytes, source preservation, overwrite refusal, verification failure handling, and encrypted backup behavior.
- Doctor tests prove encrypted state is recognized and tracked plaintext JSON is reported as unsafe.
- Final focused check: `go test ./internal/store ./internal/cli`.
