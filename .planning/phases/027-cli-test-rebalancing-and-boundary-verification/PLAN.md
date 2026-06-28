# Plan: Phase 27 CLI Test Rebalancing and Boundary Verification

## Objective

Reduce behavior-rule duplication in `internal/cli` tests and verify that CLI now acts as an adapter layer over project/app/domain packages.

## Scope

- Trim redundant project CLI tests now covered by `internal/project` tests.
- Trim redundant export CLI tests now covered by `internal/app` tests while preserving CLI format smoke.
- Keep CLI tests for Cobra command wiring, flags, completions, output/error contracts, no-leak assertions, and end-to-end smoke workflows.
- Add or adjust verification records and root planning state.

## Non-Goals

- No production behavior changes beyond test-only adjustments unless a boundary violation is found.
- No new app/project service extraction; Phase 25 and Phase 26 completed service movement.
- No removal of interactive prompt or completion tests from CLI.

## Tasks

1. Identify CLI tests that duplicate project/app behavior-rule tests.
2. Remove or narrow redundant CLI tests only where direct project/app coverage exists.
3. Keep CLI smoke tests for:
   - project id/init/error wording
   - project add/list/rm/export/explain representative workflows
   - project no-leak assertions
   - run child execution/dry-run contracts
   - export format dispatch
   - setup/migrate/manager/vault command contracts
   - completions
4. Run format/tests.
5. Run boundary greps for moved helper names and direct old responsibilities.
6. Record summary, verification, capture, update root planning state, and commit Phase 27.

## Acceptance Criteria

- `internal/cli` tests no longer carry broad behavior-rule matrices for project resolution, project entry building, direct export filtering, app migration helpers, or manager helper primitives.
- Domain/app packages own those behavior-rule tests directly.
- CLI tests still prove command contracts and representative end-to-end flows.
- Full verification confirms no user-visible behavior changes.
- Planning capture records final package and test ownership model.

## Verification

- `gofmt` on changed Go files.
- `go test ./internal/project`
- `go test ./internal/app`
- `go test ./internal/cli`
- `go test ./...`
- `grep` checks for moved helper names in `internal/cli`.

## Risks

- Over-trimming CLI tests can remove command wiring coverage. Keep smoke paths for each command family.
- Some CLI tests intentionally cover no-leak behavior through real command output; keep those even if domain tests cover underlying resolution.
