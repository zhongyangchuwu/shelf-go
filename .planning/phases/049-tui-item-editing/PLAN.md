# Plan: TUI Item Editing

## Goal

Let a personal developer create and maintain item-based secrets interactively while preserving the CLI's safe write and output contracts.

## Scope

- Add TUI flows for create, edit public metadata, add/remove labels and aliases, set a field, copy a field, and confirm deletion.
- Compose Huh forms inside the Bubble Tea application; validate against the same application and vaultmodel rules as CLI commands.
- Use masked input for secret values, save through `internal/app`, and return to a refreshed value-free detail view.
- Delegate copy to `internal/clipboard`; display only a value-free success/failure message.
- Keep CLI commands as equally capable, scriptable alternatives.

## Non-Goals

- No secret reveal pane, clipboard history management, custom-template authoring, or external-editor temporary files.

## Acceptance

- Every mutation has an explicit cancel path and destructive operations require confirmation.
- Secret values are neither rendered nor included in TUI errors, logs, snapshots, or confirmation messages.
- TUI and CLI produce equivalent persisted items for supported operations.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
