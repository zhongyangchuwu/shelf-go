# Research: Safe Catalog TUI

## Decision Supported

Add an optional, keyboard-first TUI as the interactive front-end for Shelf's local item catalog. Keep all existing Cobra commands scriptable and authoritative.

## Recommended Stack

- `charm.land/bubbletea/v2`: stateful Model-Update-View application, full-window or inline rendering, terminal lifecycle, key handling, and efficient cell-based rendering.
- `github.com/charmbracelet/bubbles/v2`: list, table, text input, viewport, and paginator components where they reduce custom state.
- `github.com/charmbracelet/huh/v2`: create/edit forms embedded as a Bubble Tea model; supports field validation and masked inputs.
- Existing `charm.land/lipgloss/v2`: shared styling dependency already present in this module.

Bubble Tea's documented Model/Init/Update/View architecture keeps screen state testable without an interactive terminal. Huh forms are Tea models and compose directly into a parent screen.

## Security Constraints

- TUI list, search, filters, detail panes, errors, and confirmation dialogs must never render secret field values.
- Secret fields show only metadata and `value_set`; no reveal screen is planned.
- Copy delegates to the existing clipboard boundary. Secret writes use a masked input, save immediately, and do not log submitted content.
- A TUI cannot make environment-variable exposure safer; `project run` remains explicit and child-process scoped.

## Architecture

Add `internal/tui` as an optional UI adapter depending on `app`, `clipboard`, Bubble Tea, Bubbles, Huh, and Lip Gloss. It must not read or write vault JSON, project manifests, or config directly. Existing application workflows remain the only mutation and resolution paths.

`cli` owns command parsing and TTY preflight, then invokes `tui`. Non-TTY invocation returns a clear error; scriptable Cobra commands retain their current behavior.

## Sources

- Bubble Tea v2 README: https://github.com/charmbracelet/bubbletea
- Bubble Tea package documentation: https://pkg.go.dev/charm.land/bubbletea/v2
- Bubbles: https://github.com/charmbracelet/bubbles
- Huh: https://github.com/charmbracelet/huh
