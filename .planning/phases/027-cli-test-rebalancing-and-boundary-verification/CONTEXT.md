# Phase 27 Context: CLI Test Rebalancing and Boundary Verification

## Goal

Rebalance tests after Phase 25 and Phase 26 so `internal/cli` protects command contracts while project/app/domain packages own reusable behavior-rule coverage.

## Constraints

- Preserve all command behavior, vault format, manifest format, error strings, and output routing.
- Do not remove CLI smoke coverage unless equivalent domain/app coverage exists and the remaining CLI tests still prove command wiring.
- Do not move shell completion tests out of CLI; completions are Cobra behavior.
- Do not move interactive prompt tests out of CLI; prompts are adapter behavior.
- Do not add new user-facing features.

## Decisions

- Remove redundant CLI tests that now duplicate direct `internal/project` coverage from Phase 25.
- Keep CLI tests for project init/id/missing manifest/outside git, one or two add/rm/list/export smoke paths, no-leak assertions, run child execution, dry-run command contract, and completions.
- Keep direct export CLI format smoke while `internal/app` owns selector/filter behavior.
- Verify boundaries with focused searches for moved helper names and imports in addition to `go test ./...`.

## Open Questions

- None blocking. Any remaining broad CLI tests should be justified as end-to-end smoke or adapter-specific coverage.

## Verification Expectations

- `go test ./internal/project` passes.
- `go test ./internal/app` passes.
- Focused CLI tests pass.
- `go test ./...` passes.
- Grep confirms CLI no longer owns moved helper functions or direct behavior-only services.
