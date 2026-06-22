# Context: Phase 3 Project Workflow Compatibility

## Goal

Prove and preserve that the developer workflows built on `shelf export`, `shelf project`, and `shelf run` continue to work unchanged when the durable store is an age-encrypted vault.

## Constraints

- Keep Shelf CLI-first: export, project binding, and runtime injection must stay scriptable and predictable.
- Do not change `.shelf.json` schema or store secret values in project manifests.
- Reuse the existing vault loading boundary through `loadVault`, `loadRuntime`, `readVault`, and `updateVault`; do not add plaintext-specific command paths.
- Preserve current value-printing rules: export/project export intentionally print values, while `project explain`, dry-run, diagnostics, and errors must not print secret values.
- Tests should exercise real encrypted vault load/decrypt behavior through the existing age-backed test config helper, not mocks.
- Keep package boundaries: CLI orchestration in `internal/cli`, rendering in `internal/render`, manifests in `internal/manifest`, persistence in `internal/store`.

## Decisions

- Treat existing `--vault` CLI tests as encrypted-vault tests because `runShelf` injects an age recipient/identity config when `--vault` is present.
- Add explicit compatibility coverage where current tests are missing or ambiguous: direct `shelf export` exact/prefix formats, manifest value-free persistence, encrypted vault ciphertext checks, and no-leak dry-run/explain assertions.
- Prefer small test-helper improvements over production changes unless implementation gaps are found.
- Phase 3 can remain a single implementation plan because production code already routes export/project/run through shared vault loaders.

## Open Questions

- Does any existing compatibility test still bypass the encrypted vault helper by omitting `--vault` on commands that load the store?
- Are there direct export tests outside `secret_test.go`, or does Phase 3 need dedicated `export_test.go` coverage?
- Do project manifest mutation tests assert absence of secret values in `.shelf.json` after `project add`?
- Do dry-run and explain tests cover encrypted vault ciphertext remaining value-free after project/run workflows?

## Verification Expectations

- Run targeted CLI compatibility tests for export, project, and run commands.
- Run `go test ./internal/cli ./internal/store ./internal/render ./internal/manifest` after implementation.
- Run `go test ./...` before closing the phase.
- Verification evidence must distinguish value-printing commands from no-secret-value diagnostics.
