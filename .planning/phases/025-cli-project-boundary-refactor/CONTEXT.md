# Phase 25 Context: CLI Project Boundary Refactor

## Goal

Move project/session business rules out of `internal/cli` so project commands become thin Cobra adapters over `internal/project` domain services while preserving every user-visible CLI behavior.

## Constraints

- Preserve current command names, flags, output formats, stdout/stderr routing, exit behavior, vault format, and `.shelf.json` schema.
- Keep `internal/cli` responsible for Cobra wiring, flags, args, shell completions, terminal/process I/O, and child process execution.
- Move reusable project/session behavior into `internal/project`; do not create a broad app service until Phase 26.
- Keep `internal/app` changes out of this phase except callsite compatibility if needed.
- Do not add new user-visible commands or aliases.
- Keep `secret add` prompt and other terminal interaction in CLI.
- Test behavior-rule changes directly in `internal/project`; keep CLI tests for command contracts and smoke workflows.

## Decisions

- Phase 25 only targets project/session boundaries: `internal/cli/project.go`, `internal/cli/run.go`, `internal/project`, and affected tests.
- `internal/project` should own entry construction for path, prefix, and tag selectors because these are manifest/domain rules, not CLI presentation.
- `internal/project` should own environment merge utilities used by `project run` because they are project-session semantics reusable beyond Cobra.
- CLI should still execute child processes because `os/exec`, stdio wiring, and exit-code translation are command adapter responsibilities.

## Rejected Options

- Do not split every CLI command into separate files; the project explicitly prefers command-family oriented CLI files.
- Do not move Cobra completion logic into domain packages; completions are shell/Cobra adapter behavior.
- Do not combine Phase 25 with app export/setup/migrate extraction; that belongs to Phase 26.

## Open Questions

- Exact function names may change during implementation, but the service boundary should expose request/result types rather than Cobra-specific arguments.

## Canonical References

- `internal/cli/project.go`
- `internal/cli/run.go`
- `internal/project/manifest.go`
- `internal/project/resolve.go`
- `internal/cli/project_test.go`
- `internal/cli/run_test.go`

## Verification Expectations

- `internal/project` tests cover entry building, selector validation, diagnostics, prefix/tag expansion, env conflicts, child env merge, and override warnings.
- Focused CLI tests confirm project add/list/export/explain/run behavior is unchanged.
- No secret values leak in project explain/list diagnostics.
- `go test ./internal/project ./internal/cli -run 'Test(Project|Run)'` passes.
- Final gate for phase completion remains `go test ./...`.
