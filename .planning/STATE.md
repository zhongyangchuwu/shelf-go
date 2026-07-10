# State

## Current Position

- Phase: v0.2.0 Local Wrap-up
- Status: reviewing code/framework structure, planning/docs freshness, and file hygiene
- Active Artifact: `.planning/ROADMAP.md`
- Next Action: Apply review findings, then run local verification and installed-binary smoke tests.

## Blockers

- None.

## Recent Evidence

- Item-vault-only cutover is implemented and committed.
- Active packages are now CLI/app/cliui/config/project/util/vaultjson/vaultmodel.
- Default vault path is `~/.local/share/shelf/vault.json`.
- Local config points at `/home/han/.local/share/shelf/vault.json`.
- Local backup before cutover: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Built-in release packaging/tagging is deferred until a future public v1.0.0 milestone.
- Latest verification before wrap-up passed:
  - `go test ./...`
  - `go vet ./...`
  - `go-arch-lint check`

## Updated

- 2026-07-10
