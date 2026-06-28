# Capture: Phase 26 App Service Extraction

## Durable Knowledge

- `internal/app` now owns reusable command orchestration for direct secret export, plaintext migration, setup file/identity/vault helpers, and manager helper primitives.
- `internal/cli` should keep command adapter behavior only: Cobra flags/args, prompts, completions, output routing, signal handling, server/process lifecycle, and user guidance text.
- `internal/app` services should accept request structs or explicit values and return strings/results/errors; they should not import Cobra or CLI.
- Setup prompting remains CLI-owned because it depends on command stdin/stdout.
- Manager server lifecycle remains CLI-owned because it depends on command process lifetime and OS signals.

## Follow-On Work

- Phase 27 should rebalance CLI tests now that Phase 25 and Phase 26 provide direct project/app behavior coverage.
- Phase 27 should record the final package/test ownership model for maintainers.

## Documentation Impact

- No user-facing documentation update is required; command behavior did not change.
- Developer architecture docs should be updated after Phase 27 if the package-boundary model is documented outside planning artifacts.
