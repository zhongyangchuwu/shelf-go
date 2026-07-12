# Plan: Safe Catalog TUI

## Goal

Add an optional, wide-screen, keyboard-first `shelf tui` that provides safe catalog browsing plus core interactive item workflows: create, update, copy, and remove.

## Scope

- Add `internal/tui` on Bubble Tea v2, Bubbles v2, Huh v2, and existing Lip Gloss v2.
- Require a 100×24 interactive terminal; show a resize-required screen rather than implementing a narrow layout.
- Implement a persistent lazygit-like workspace: left catalog, right detail/editor, keyboard help, search, template/label filters, resize handling, and graceful terminal restoration.
- Render only value-free catalog/detail data; secret fields expose only metadata and `value_set`.
- Add create, update, copy, and remove actions through existing app and clipboard boundaries. Secret form input is masked; no secret is revealed in the workspace.
- Keep the TUI free of direct vault/config/manifest I/O. Existing scriptable Cobra commands remain authoritative.

## Non-Goals

- No secret reveal, `.env` editor, named runtime environments, project-manifest mutation, custom-template authoring, daemon, or Web UI.
- `get`, export, run, setup, doctor, and completion remain CLI-first because they materialize stdout data, launch a child process, or are one-shot system actions.

## Acceptance

- `shelf tui` starts and exits cleanly on a supported wide interactive terminal and reports resize-required on a smaller terminal.
- Catalog, filter, detail, form, confirmation, error, and help views cannot contain a secret value.
- TUI create/update/delete use app workflows; copy uses the clipboard boundary without rendering the copied value.
- TUI state transitions have deterministic model tests without a real terminal.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.

## Design Inputs

See `RESEARCH.md` for library and security rationale and `UI-SPEC.md` for the reviewed interaction, safety, architecture, and verification contract.

## Implementation

- Added a value-free `app.SecretCatalog` DTO/query and atomic create support for labels and aliases.
- Added `internal/tui` with Bubble Tea lifecycle, Bubbles text input and viewport, Huh masked multi-step forms, adaptive Lip Gloss styles, and deterministic state tests.
- Added `shelf tui` with stdin/stdout TTY preflight and preserved app/clipboard-only mutation boundaries.
- Added value-free behavior tests, command preflight coverage, architecture rules, and pseudo-terminal wide/narrow smoke coverage.
