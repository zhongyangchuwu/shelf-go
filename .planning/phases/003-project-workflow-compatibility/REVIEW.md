# Review: Phase 3 Project Workflow Compatibility

## Scope Reviewed

- Direct export compatibility tests in `internal/cli/export_test.go`.
- Project manifest/export compatibility tests in `internal/cli/project_test.go`.
- Run dry-run compatibility tests in `internal/cli/run_test.go`.
- Existing command loading path in `internal/cli/root.go`, `internal/cli/export.go`, `internal/cli/project.go`, and `internal/cli/run.go` as context.

## Findings

- No vault-boundary bypass found. Export, project export/explain, and run load the store through `loadRuntime`, which constructs a configured `store.Vault` and decrypts through `vault.Load`.
- No production value leak found in no-value diagnostics under tested paths. `project explain` and `run --dry-run` print paths/env names/warnings, not values.
- Existing run dry-run semantics intentionally do not print the child command. The new test now asserts current behavior instead of expanding scope.

## Fixes Applied

- Added regression tests instead of changing production code.
- Corrected the dry-run test expectation to avoid changing established CLI output.

## Waivers

- No separate manual CLI transcript was captured because automated tests execute Cobra command paths with real age-encrypted vault files.

## Remaining Risks

- None identified for Phase 3 acceptance criteria.
