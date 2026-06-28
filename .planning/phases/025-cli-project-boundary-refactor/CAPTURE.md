# Capture: Phase 25 CLI Project Boundary Refactor

## Durable Knowledge

- `internal/project` now owns project entry construction from selectors and project environment behavior.
- `internal/cli/project.go` should remain a Cobra adapter: root discovery, manifest file loading/saving, runtime loading, flags, completions, and output routing.
- `internal/cli/run.go` should keep process lifecycle and exit-code translation, but should not own reusable env merge rules.
- Behavior-rule tests should prefer direct `internal/project` tests with constructed `vault.Store` and `Manifest` values.

## Follow-On Work

- Phase 26 should extract app-level orchestration for `secret export`, setup/vault init, migrate, manager helpers, and related command workflows.
- Phase 27 should reduce redundant CLI behavior-rule tests after app/domain tests cover the same branches.

## Documentation Impact

- No user-facing documentation update is required; command behavior did not change.
- Developer architecture docs may need a final update after Phase 27 completes the full boundary/test ownership model.
