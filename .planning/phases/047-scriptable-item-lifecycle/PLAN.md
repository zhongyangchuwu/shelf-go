# Plan: Scriptable Item Lifecycle

## Goal

Make the local item catalog fully maintainable through explicit, scriptable CLI workflows before adding an interactive editor.

## Scope

- `shelf secret update <selector>` is the sole item-mutation command: it changes title, labels, aliases, or one field in one atomic operation.
- A public field accepts explicit repeatable `--value`; any field without `--value` reads stdin. When stdin is an interactive terminal, Shelf prompts for the selected field and masks secret input.
- Secret field values never accept `--value`; they must enter through stdin. Field writes default to the template primary secret field only when no other update flags are present.
- Reuse vaultmodel validation and vaultjson atomic save; no workflow may mutate JSON directly. Preserve value-free output for list, info, diagnostics, completions, and edit confirmation.

## Non-Goals

- No TUI, Web UI, custom-template authoring UI, hosted sync, or multi-environment project model.

## Acceptance

- A user can maintain a generic or API credential item without opening the vault JSON.
- Invalid mutations fail before persistence and leave the previous vault intact.
- Secret values never enter argv or metadata output.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
