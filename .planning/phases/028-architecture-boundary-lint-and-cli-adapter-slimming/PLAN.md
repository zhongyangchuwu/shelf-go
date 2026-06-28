# Phase 28 Plan: Architecture Boundary Lint and CLI Adapter Slimming

## Acceptance Criteria
1. `.go-arch-lint.yml` records the intended import boundaries for the current internal packages.
2. `go-arch-lint check --arch-file .go-arch-lint.yml --project-path ./` passes.
3. `internal/cli/project.go` keeps Cobra command surface and output routing, while project manifest/vault/format orchestration lives in `internal/app`.
4. `internal/cli/secret.go` keeps Cobra command surface and output routing, while secret set/get/list/info/rm orchestration lives in `internal/app` and interactive add workflow lives in `internal/secret`.
5. `internal/cli` no longer imports low-level packages just to perform reusable behavior; remaining imports are command-surface or unavoidable adapter wiring.
6. `internal/vault` does not import `internal/config`.
7. Existing behavior is preserved: `go test ./...` passes.

## Steps
1. Add `.go-arch-lint.yml` for the target light Onion / adapter-app-domain boundary.
2. Extract project command use cases into `internal/app/project.go`.
3. Extract secret command use cases into `internal/app/secret.go` and `internal/secret/add.go`.
4. Split CLI completion/helper code if it clarifies command files.
5. Route manager and vault boundary imports through app or tighten lint rules only after code is ready.
6. Run gofmt on changed Go files.
7. Verify with targeted tests, full tests, and go-arch-lint.
8. Record summary, verification, capture, update root planning, commit.

## Non-Goals
- No public CLI rename or flag changes.
- No new user-visible commands.
- No vault or manifest format change.
- No broad Clean Architecture directory migration.
