# Plan: Project Binding TUI

## Goal

Make value-free project binding management discoverable and fast without turning Shelf into an env-file editor.

## Scope

- Add `shelf project tui` for the current Git worktree.
- Show project bindings, required/optional state, resolution status, duplicate-env conflicts, and missing-value diagnostics without values.
- Add an item/field picker for bindings, default env-name selection, explicit env override, optional marking, remove confirmation, and manifest initialization with confirmation.
- Reuse `internal/app` and `internal/project`; the TUI does not parse or write `.shelf.json` itself.
- Keep `project run` as an explicit CLI command rather than embedding a long-running child process in the TUI.

## Non-Goals

- No named dev/staging/prod environments, project discovery database, cross-repository dependency index, or automatic env loading.

## Acceptance

- Project manifests remain value-free after all TUI operations.
- Project status and error screens never display resolved values.
- The same binding decisions are available through scriptable CLI commands.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
