# Plan: Scriptable Item Lifecycle

## Goal

Make the local item catalog fully maintainable through explicit, scriptable CLI workflows before adding an interactive editor.

## Scope

- Add `shelf secret update <selector>` backed by an app workflow; update title, add/remove labels, add/remove aliases, or one public field in one atomic operation.
- Use `--field FIELD --value VALUE` for typed public-field changes, with repeatable values for string lists, and `--clear-field` for explicit clearing.
- Keep secret field writes on the existing `secret set --stdin` path; update rejects attempts to write secret fields or to pass values without an explicit public field.
- Reuse vaultmodel validation and vaultjson atomic save; no workflow may mutate JSON directly. Preserve value-free output for list, info, diagnostics, completions, and edit confirmation.

## Non-Goals

- No TUI, Web UI, custom-template authoring UI, hosted sync, or multi-environment project model.

## Acceptance

- A user can maintain a generic or API credential item without opening the vault JSON.
- Invalid mutations fail before persistence and leave the previous vault intact.
- Secret values never enter argv or metadata output.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
