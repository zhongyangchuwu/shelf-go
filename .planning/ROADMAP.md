# Roadmap: Shelf Go v0.1.1

## Overview

Shelf Go v0.1.0 has shipped. The active planning workspace is now open for v0.1.1 selection; no implementation phase is selected yet.

## Phases

No active v0.1.1 phases yet.

## Phase Details

None. Add a phase after selecting the next v0.1.1 milestone and mapping current requirements.

## Future Candidates

- Manager UI redesign: improve `shelf vault open` visual design, information architecture, and edit/reveal UX; keep the current loopback manager functional while improving local editing speed and safety clarity.
- Native Windows smoke tests: verify `shelf setup`, secret set/get, and project run on a real Windows runner.
- SQLite storage spike: investigate SQLite as an encrypted vault payload or metadata/search layer only if JSON schema/search/history pressure becomes real. Any design must preserve encrypted-at-rest safety and avoid plaintext SQLite WAL, journal, or temp files.
- Password-only encryption: consider only if users need a no-age-key workflow and the threat model remains clear.
- Multiple vaults or profiles: consider after single-vault workflows show concrete pressure.
- Chezmoi helper commands: consider optional integration while preserving Shelf's portable encrypted-file model.
- Package-manager distribution: consider Homebrew/Scoop or similar after initial usage validates demand.
- `internal/gitutil`: create only if both `internal/project` and future `internal/vault` need shared git subprocess helpers.

## Explicit Non-Goals Until Selected

- No active `project activate` / `project deactivate` shell hook implementation.
- No `project shell` wrapper unless a later phase shows clear value over `project run -- $SHELL`.
- No new `dotenv` format; use existing `shell` output for sourceable files.
- No team sharing, hosted sync, or permanent daemon.
- No SQLite backend implementation without a dedicated spike.
- No speculative repository/service abstraction beyond packages required by current code.

## Progress

| Phase | Status | Requirements | Plan | Completion Date |
|-------|--------|--------------|------|-----------------|
| None selected for v0.1.1 | Ready | - | - | - |

## Archived Releases

- v0.1.0: `.planning/archive/releases/v0.1.0/SUMMARY.md`

---
*Last updated: 2026-06-26 after archiving v0.1.0 planning history*
