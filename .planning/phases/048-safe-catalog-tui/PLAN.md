# Plan: Safe Catalog TUI

## Goal

Add an optional, keyboard-first `shelf tui` for fast local browsing and selection of structured secret items without displaying secret values.

## Scope

- Add `internal/tui` on Bubble Tea v2, Bubbles v2, and the existing Lip Gloss v2 styling stack. Huh v2 is deferred until the editing forms in Phase 49.
- Implement full-window catalog list, text search, label/template filters, item detail, keyboard help, resize handling, and graceful terminal restoration.
- Render only value-free data: title, ID, template, labels, aliases, public metadata, and secret `value_set` state.
- Keep the TUI read-only in this phase; it delegates loading to `internal/app` and never reads vault JSON directly.
- Reject non-interactive stdin/stdout with a clear error; all existing Cobra commands remain unchanged and scriptable.

## Non-Goals

- No in-TUI secret reveal, editing, project mutation, custom-template authoring, or daemon.

## Acceptance

- `shelf tui` starts and exits cleanly on supported interactive terminals.
- List, filter, detail, error, and help views cannot contain a secret value.
- TUI state transitions have deterministic model tests without a real terminal.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.

## Research

See `RESEARCH.md` for library and security rationale.
