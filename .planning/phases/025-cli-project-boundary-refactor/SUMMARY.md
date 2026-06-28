# Summary: Phase 25 CLI Project Boundary Refactor

## Outcome

Completed the project/session boundary refactor for the first slice: reusable project entry construction and project environment behavior now live in `internal/project`, while CLI handlers remain responsible for Cobra wiring, file loading, output routing, and child process execution.

## Implemented

- Added `internal/project/entry.go`:
  - `AddEntryRequest`
  - `BuildEntry`
  - `AddEntry`
- Added `internal/project/environment.go`:
  - `ChildEnv`
  - `EnvOverrideWarnings`
- Refactored `internal/cli/project.go`:
  - `project add` now delegates path/prefix/tag selector entry construction and manifest mutation to `internal/project`.
  - `project explain` uses `project.EnvOverrideWarnings`.
- Refactored `internal/cli/run.go`:
  - `project run` uses `project.ChildEnv` for child environment construction.
  - `project run --dry-run` uses `project.EnvOverrideWarnings`.
  - CLI still owns `exec.Command`, stdio wiring, and exit-code translation.
- Added direct domain tests in `internal/project`:
  - entry construction and invalid request coverage
  - duplicate entry handling
  - child env merge behavior
  - parent env override warnings
  - resolve diagnostics, prefix expansion, tag expansion, and env conflict coverage
- Updated `internal/cli/run_test.go` to keep CLI dry-run warning/no-leak coverage while moving pure env merge coverage to `internal/project`.

## Deviations

- Phase 25 kept some CLI tests as smoke coverage instead of aggressively deleting all overlapping cases. This preserves end-to-end confidence before Phase 27 does the broader test rebalancing.
- `internal/project.AddEntry` currently returns an updated `Manifest` value plus the added `Entry`; this keeps mutation behavior explicit without introducing an app service.

## Verification Evidence

- `gofmt` ran on changed Go files.
- `go test ./internal/project` passed.
- `go test ./internal/cli -run 'Test(Project|Run)'` passed.
- `go test ./...` passed.

## Files Changed

- `internal/project/entry.go`
- `internal/project/environment.go`
- `internal/project/entry_test.go`
- `internal/project/environment_test.go`
- `internal/project/resolve_test.go`
- `internal/cli/project.go`
- `internal/cli/run.go`
- `internal/cli/run_test.go`
- Planning artifacts under `.planning/`
