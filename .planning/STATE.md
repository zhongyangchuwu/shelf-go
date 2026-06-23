# State

## Current Position
- Phase: pre-release command hierarchy and vault UX milestone
- Status: complete
- Active Artifact: .planning/ROADMAP.md
- Next Action: Select the next release or implementation milestone.

## Blockers
- None

## Recent Evidence
- Phase 6 command hierarchy cutover completed on 2026-06-22.
- Implemented canonical commands: `shelf setup`, `shelf vault init`, `shelf vault migrate`, `shelf vault status`/`check`, `shelf vault open`, `shelf secret export`, and `shelf project run`.
- Old top-level `init`, `migrate`, `export`, `run`, and `manager` commands are intentionally absent.
- Phase 7 vault UX hardening completed on 2026-06-23.
- `shelf vault status`/`check` and `shelf doctor` now give recovery guidance for missing recipients, missing identities, plaintext stores, unsupported vault formats, and undecryptable vaults.
- Phase 8 project session activation/deactivation/shell design completed on 2026-06-23; implementation remains out of scope.
- `go test ./internal/cli -run 'Test(Vault|Doctor|Manager|Migrate|Setup)'` passed.
- `go test ./internal/store -run 'TestVault'` passed.
- `go test ./...` passed.

## Updated
- 2026-06-23
