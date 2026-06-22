# Capture: Phase 3 Project Workflow Compatibility

## Durable Docs Updated

- None. User-facing documentation remains Phase 5 scope.

## Planning Records Updated

- Phase context, plan, summary, review, and verification were created under `.planning/phases/003-project-workflow-compatibility/`.
- Root planning docs need Phase 3 completion status and Phase 4 readiness updates.

## Learnings

- The CLI compatibility surface already uses the vault boundary; Phase 3 required regression coverage, not production rewiring.
- `runShelf` test helper turns `--vault` tests into real age-encrypted vault tests by creating an identity/config pair and injecting `--config`.
- `run --dry-run` only prints override warnings and `inject <ENV>` lines; it intentionally does not echo the child command.
- No-value diagnostics should continue testing with distinct secret and parent env values to catch both leak classes.

## Ship Inputs

- Phase 3 has automated evidence through `go test ./...`.
- Phase 4 can start from a stable encrypted CLI workflow and focus on localhost manager safety controls.
