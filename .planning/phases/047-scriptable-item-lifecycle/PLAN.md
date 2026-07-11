# Plan: Scriptable Item Lifecycle

## Goal

Make the local item catalog fully maintainable through explicit, scriptable CLI workflows before adding an interactive editor.

## Scope

- Add application and Cobra workflows for editing item title, labels, aliases, and supported public template fields.
- Keep secret field writes on the existing stdin-only path; public metadata may use explicit typed inputs.
- Reuse vaultmodel validation and vaultjson atomic save; no workflow may mutate JSON directly.
- Preserve value-free output for list, info, diagnostics, completions, and edit confirmation.

## Non-Goals

- No TUI, Web UI, custom-template authoring UI, hosted sync, or multi-environment project model.

## Acceptance

- A user can maintain a generic or API credential item without opening the vault JSON.
- Invalid mutations fail before persistence and leave the previous vault intact.
- Secret values never enter argv or metadata output.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
