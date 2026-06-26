# State

## Current Position
- Phase: Phase 19 - Secret Tag Selection
- Status: not-started
- Active Artifact: .planning/ROADMAP.md
- Next Action: Start Phase 19 planning for tag-based `secret list` and `secret export` selectors.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- v0.1.1 milestone selected: Web manager editing console plus tag-based secret/project workflows.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the Web manager editing console and manager API hardening.
- Verification observed: `go test ./internal/manager`, `go test ./...`, and LSP workspace diagnostics passed.
- v0.1.1 still defers SQLite/storage redesign to v0.2.0 and keeps the current age-encrypted JSON vault format.
- v0.1.1 keeps CLI editing compact: no `secret meta` or `secret tag` command group; CLI changes focus on tag list/export and project tag bindings.

## Updated
- 2026-06-26
