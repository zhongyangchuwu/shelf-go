# State

## Current Position
- Phase: Phase 20 - Project Tag Bindings
- Status: not-started
- Active Artifact: .planning/ROADMAP.md
- Next Action: Start Phase 20 planning for value-free project manifest tag bindings.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- v0.1.1 milestone selected: Web manager editing console plus tag-based secret/project workflows.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the Web manager editing console and manager API hardening.
- Phase 19 implemented direct secret tag selection for `secret list` and `secret export`.
- Verification observed for Phase 19: `go test ./internal/store`, focused `go test ./internal/cli`, `go test ./...`, and LSP workspace diagnostics passed.
- v0.1.1 still defers SQLite/storage redesign to v0.2.0 and keeps the current age-encrypted JSON vault format.
- v0.1.1 keeps CLI editing compact: no `secret meta` or `secret tag` command group; CLI changes focus on tag list/export and project tag bindings.

## Updated
- 2026-06-26
