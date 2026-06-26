# State

## Current Position
- Phase: Phase 22 - Documentation and Architecture Cleanup
- Status: not-started
- Active Artifact: .planning/ROADMAP.md
- Next Action: Start Phase 22 planning for user/developer documentation and architecture naming cleanup around `shelf vault open` and `internal/manager`.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the Web manager editing console and manager API hardening.
- Phase 19 implemented direct secret tag selection for `secret list` and `secret export`.
- Phase 20 implemented value-free project tag bindings for `.shelf.json`, `project add/list/rm`, and project explain/export/run resolution.
- v0.1.1 release hardening was moved later because script consolidation, docs, and architecture naming cleanup remain before release.
- Phase 21 planning artifacts were created at `.planning/phases/021-script-workflow-consolidation/CONTEXT.md` and `PLAN.md`.
- Phase 21 completed script workflow consolidation: install now runs through `scripts/install.sh`, release check/snapshot/tag now run through one `scripts/release.sh` command surface, and `justfile` remains a thin task runner.
- Remaining planned work:
  - Phase 22: update docs and resolve architecture naming/package mismatch around `shelf vault open` and `internal/manager`.
  - Phase 23: perform final v0.1.1 release hardening after Phase 22.
- v0.1.1 still defers SQLite/storage redesign to v0.2.0 and keeps the current age-encrypted JSON vault format.

## Updated
- 2026-06-26
