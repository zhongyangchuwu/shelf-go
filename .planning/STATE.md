# State

## Current Position
- Phase: Phase 21 - v0.1.1 Release Hardening
- Status: not-started
- Active Artifact: .planning/ROADMAP.md
- Next Action: Start Phase 21 release hardening: docs, changelog, vet, release check, and snapshot release.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the Web manager editing console and manager API hardening.
- Phase 19 implemented direct secret tag selection for `secret list` and `secret export`.
- Phase 20 implemented value-free project tag bindings for `.shelf.json`, `project add/list/rm`, and project explain/export/run resolution.
- Verification observed for Phase 20: `go test ./internal/manifest`, `go test ./internal/project`, focused `go test ./internal/cli`, `go test ./...`, and LSP workspace diagnostics passed.
- v0.1.1 still defers SQLite/storage redesign to v0.2.0 and keeps the current age-encrypted JSON vault format.

## Updated
- 2026-06-26
