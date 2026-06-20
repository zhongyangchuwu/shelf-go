# Capture: Migration and Git Safety

## Durable Docs Updated

- `docs/usage-spec.md` now lists and explains `shelf migrate --from <plaintext.json> [--to <vault.age>] [--force]`, including source preservation, target verification, overwrite behavior, and cleanup guidance.
- `docs/usage-spec.md` now documents doctor vault-format and ordinary Git tracking checks for encrypted vaults versus plaintext JSON stores.
- `docs/data-spec.md` now records migration as the supported plaintext-to-vault conversion path and clarifies that replacement backups are encrypted while plaintext migration sources are not rewritten or backed up by Shelf.
- `docs/codebase/ARCHITECTURE.md` now describes the migrate command, vault format classification, and git-safety doctor flow.

## Planning Records Updated

- `.planning/phases/002-migration-and-git-safety/CONTEXT.md` created with Phase 2 goals, constraints, decisions, and verification expectations.
- `.planning/phases/002-migration-and-git-safety/PLAN.md` created with tasks, acceptance criteria, verification, and risks.
- `.planning/phases/002-migration-and-git-safety/SUMMARY.md` records implemented changes, files changed, deviations, evidence, and risks.
- `.planning/phases/002-migration-and-git-safety/VERIFICATION.md` records requirement-by-requirement evidence for MIGR-01 through MIGR-05 and SAFE-01 through SAFE-05.
- `.planning/REQUIREMENTS.md` marks MIGR-01 through MIGR-05 and SAFE-01 through SAFE-05 complete.
- `.planning/ROADMAP.md` marks Phase 2 complete and updates completed requirement count to 17.
- `.planning/PROJECT.md` moves migration and git-safety behavior into validated scope and records the source-preservation migration decision.
- `.planning/STATE.md` advances to Phase 3 `not-started`.

## Learnings

- `store.DetectFileFormat` is now the canonical way to classify active vault paths before attempting decryption.
- Plaintext `store.Load` remains valid only at explicit legacy/migration boundaries; CLI runtime secret workflows still use `store.Vault`.
- Migration must preserve the plaintext source even after success; Shelf reports manual cleanup guidance rather than deleting user data.
- Forced migration safely reuses `Vault.Save`, so replacement backups remain encrypted without a separate migration backup path.
- Doctor git-safety confidence is based on the active path's actual tracked state and detected format, not on filename conventions.
- Chezmoi integration remains operationally simple: chezmoi can track the encrypted file; Shelf only needs to verify that tracked bytes are encrypted.

## Ship Inputs

- Summary:
  - Added `shelf migrate --from <plaintext.json> [--to <vault.age>] [--force]`.
  - Added vault file format detection for encrypted vaults, plaintext stores, unsupported vault headers, unsupported content, empty files, and missing paths.
  - Extended `shelf doctor` to fail plaintext active stores, fail tracked plaintext stores, and confirm tracked encrypted vaults.
  - Updated docs and planning records for migration and git/chezmoi safety.
- Verification:
  - `go test ./internal/store ./internal/cli`
  - `go test ./...`
  - `.planning/phases/002-migration-and-git-safety/VERIFICATION.md`
- Risks / waivers:
  - Doctor checks ordinary Git tracked state and does not invoke chezmoi.
  - Doctor does not scan every possible `.shelf.json`; manifest schema remains the value-free guarantee.
  - `secret edit` editor temp-file plaintext risk remains deferred.
