# State

## Current Position
- Phase: Phase 23 - Documentation and Usage Alignment
- Status: not-started
- Active Artifact: .planning/ROADMAP.md
- Next Action: Start Phase 23 planning for user/developer documentation covering `shelf manager`, tag workflows, scripts, and the final package layout.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the manager editing console and manager API hardening.
- Phase 19 implemented direct secret tag selection for `secret list` and `secret export`.
- Phase 20 implemented value-free project tag bindings for `.shelf.json`, `project add/list/rm`, and project explain/export/run resolution.
- Phase 21 completed script workflow consolidation: install now runs through `scripts/install.sh`, release check/snapshot/tag now run through one `scripts/release.sh` command surface, and `justfile` remains a thin task runner.
- Phase 22 completed architecture repartition:
  - `shelf manager` replaced `shelf vault open` with no alias.
  - Vault core moved to `internal/vault`.
  - Project manifest schema/IO/validation moved into `internal/project`.
  - Version composition moved into `internal/app`.
  - Export formatting moved into `internal/exportfmt`.
  - Final internal package set is `app`, `cli`, `config`, `exportfmt`, `manager`, `project`, `secret`, and `vault`.
- v0.1.1 still defers SQLite/storage redesign to v0.2.0 and keeps the current age-encrypted JSON vault format.

## Updated
- 2026-06-27
