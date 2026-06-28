# Verification: Phase 25 CLI Project Boundary Refactor

## Claims Verified

1. `internal/project` owns project entry construction for path, prefix, and tag selectors.
2. `internal/project` owns reusable project environment behavior for child env merge and override warnings.
3. CLI project/run handlers delegate reusable project behavior while preserving command adapter responsibilities.
4. Domain tests cover behavior-rule branches without Cobra.
5. Existing project/run CLI behavior remains unchanged.

## Evidence

### `go test ./internal/project`

Result: Passed.

Coverage from this check:

- `BuildEntry` creates path, prefix, and tag entries.
- `BuildEntry` rejects invalid selector/tag/env combinations.
- `AddEntry` preserves duplicate-entry rejection.
- `ChildEnv` replaces parent env values, appends missing values, and handles malformed parent env entries.
- `EnvOverrideWarnings` reports parent env conflicts without values.
- `ResolveEntries` covers missing required/optional secrets, env conflicts, prefix expansion order, empty prefix diagnostics, and tag expansion.

### `go test ./internal/cli -run 'Test(Project|Run)'`

Result: Passed.

Coverage from this check:

- Project CLI command contracts still work through Cobra.
- Project add/list/rm/export/explain workflows still operate with real temp git projects and vault files.
- Project run still injects secrets into child processes.
- Project run dry-run still reports env override warnings and does not execute the child command.
- Existing no-leak assertions for project explain/run paths still pass.

### `go test ./...`

Result: Passed.

Coverage from this check:

- All packages compile after the package-boundary refactor.
- Existing full test suite remains green.

## Acceptance Criteria Mapping

| Acceptance Criterion | Evidence | Result |
| --- | --- | --- |
| `internal/cli/project.go` no longer directly encodes path/prefix/tag entry construction rules | `project add` delegates to `project.AddEntry`; `go test ./internal/cli -run 'Test(Project|Run)'` passed | Passed |
| `internal/cli/run.go` no longer owns reusable child env merge or parent override warnings | `project.ChildEnv` and `project.EnvOverrideWarnings` are used by `run.go`; project environment tests passed | Passed |
| `internal/project` exposes cohesive APIs without Cobra/CLI imports | New project APIs compile in `go test ./internal/project`; no Cobra imports in new project files | Passed |
| Project behavior-rule tests run in `internal/project` without `runShelf` | New entry/environment/resolve tests construct `vault.Store` and `Manifest` directly | Passed |
| CLI tests still cover user-visible command contracts and no secret leaks | Focused CLI tests passed | Passed |
| Existing CLI behavior remains unchanged | Focused CLI tests and `go test ./...` passed | Passed |

## Gaps

- Phase 27 still owns broader test rebalancing; this phase intentionally retained some overlapping CLI smoke tests.
- Phase 26 still owns app service extraction for export/setup/migrate/manager orchestration.
