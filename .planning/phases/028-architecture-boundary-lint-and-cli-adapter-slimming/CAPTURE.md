# Phase 28 Capture

## Decisions
- Use a light Onion / ports-and-adapters boundary instead of broad Clean Architecture directories.
- Keep package names capability-oriented: `cli`, `manager`, `app`, `project`, `secret`, `vault`, `config`, `exportfmt`.
- Enforce import direction with `go-arch-lint` rather than relying on conventions.
- Keep test files excluded from arch lint so package tests can construct lower-level fixtures directly.

## Lessons
- Helper functions alone would not make CLI lighter; the main fix was moving usecase orchestration and workflow state out of CLI.
- `app` should own config-aware orchestration; `vault` should not import `config`.
- Manager is an adapter, but allowing CLI to import `manager` is pragmatic for server lifecycle construction while manager production code depends only on app.

## Follow-Up Candidates
- Add a script or CI task for `go-arch-lint check` after deciding where release checks should include it.
- Revisit whether `project -> exportfmt` should be removed by moving render binding conversion into app if project domain grows.
