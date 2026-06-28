# Summary: Phase 27 CLI Test Rebalancing and Boundary Verification

## Outcome

Completed the test rebalancing pass after moving reusable behavior into `internal/project` and `internal/app`. CLI tests now focus more tightly on command contracts, representative end-to-end smoke paths, no-leak assertions, error wording, and completions.

## Implemented

- Trimmed redundant project CLI tests that duplicated direct `internal/project` coverage:
  - optional/required missing matrix variants
  - prefix add/build validation variants
  - duplicate/missing/empty prefix entry construction variants
  - prefix remove variant
  - env/shell/json project export format duplicates
  - JSON value string conversion duplicate
  - optional/empty prefix diagnostics variants
  - prefix expansion sorting and prefix explain variants
- Preserved CLI coverage for:
  - project id/init/force/outside-git/missing-manifest wording
  - project explain representative diagnostics and no secret leakage
  - project add manifest persistence and value-free manifest behavior
  - encrypted-vault project export smoke
  - project add missing-manifest wording
  - project rm/list/export default smoke
  - project required-missing export failure smoke
  - parent env override no-leak behavior
  - project completions
- Verified moved helper boundaries with grep.

## Verification Evidence

- `gofmt` ran on changed Go files.
- `go test ./internal/project` passed.
- `go test ./internal/app` passed.
- `go test ./internal/cli` passed.
- `go test ./...` passed.
- Boundary grep found no moved helper function definitions in `internal/cli`.
- Boundary grep found no direct `Data.Secrets` access in `internal/cli`.

## Files Changed

- `internal/cli/project_test.go`
- `.planning/phases/027-cli-test-rebalancing-and-boundary-verification/CONTEXT.md`
- `.planning/phases/027-cli-test-rebalancing-and-boundary-verification/PLAN.md`
- Planning root artifacts under `.planning/`
