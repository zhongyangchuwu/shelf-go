# Verification: Phase 27 CLI Test Rebalancing and Boundary Verification

## Claims Verified

1. CLI tests no longer carry broad behavior-rule matrices that now belong to project/app packages.
2. CLI tests still prove command contracts and representative end-to-end flows.
3. Domain/app packages own direct reusable behavior coverage.
4. CLI no longer defines moved helper functions or directly reads vault secret internals.
5. Full test suite passes.

## Evidence

### `go test ./internal/project`

Result: Passed.

Coverage relevance:

- Project selector entry construction, invalid requests, duplicate handling, diagnostics, prefix/tag expansion, env conflicts, child env merge, and override warnings remain covered directly in `internal/project`.

### `go test ./internal/app`

Result: Passed.

Coverage relevance:

- Export selector/filter/format behavior, setup helpers, migration implementation, and manager helper primitives remain covered directly in `internal/app`.

### `go test ./internal/cli`

Result: Passed.

Coverage relevance:

- CLI command contracts still pass after trimming redundant behavior-rule tests.
- Project, run, export, setup, migrate, manager, vault, doctor, secret, and completion command families still have CLI-level coverage.

### `go test ./...`

Result: Passed.

Coverage relevance:

- All packages compile and all tests pass together.

### Boundary greps

Commands run through the grep tool:

- Pattern: `func (migratePlaintextStore|listenLoopback|managerToken|ensureInitIdentity|ensureVaultFile|ensureConfigFile|resolveInitConfigPath|relativeIfDescendant|expandInitPath|childEnv|envOverrideWarnings)` in `internal/cli`
  - Result: no matches.
- Pattern: `st\.Data\.Secrets|\.Data\.Secrets` in `internal/cli`
  - Result: no matches.

## Acceptance Criteria Mapping

| Acceptance Criterion | Evidence | Result |
| --- | --- | --- |
| CLI tests no longer carry broad behavior-rule matrices for moved project/app logic | `internal/cli/project_test.go` trimmed; project/app direct tests pass | Passed |
| Domain/app packages own behavior-rule tests directly | `go test ./internal/project` and `go test ./internal/app` passed | Passed |
| CLI tests still prove command contracts and representative workflows | `go test ./internal/cli` passed | Passed |
| Full verification confirms no user-visible behavior changes | `go test ./...` passed | Passed |
| Planning capture records final package and test ownership model | `CAPTURE.md` written for Phase 27 | Passed |

## Gaps

- No gaps identified for this refactor. Future docs can copy the ownership model from Phase 27 capture if developer docs need a non-planning version.
