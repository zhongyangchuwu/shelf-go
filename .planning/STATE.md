# State

## Current Position

- Phase: Secret Copy and Export Simplification
- Status: complete
- Active Artifact: `.planning/phases/043-secret-copy-and-export-simplification/CAPTURE.md`
- Next Action: Select the next local CLI UX or staged refactor phase; public release work remains deferred.

## Blockers

- None.

## Recent Evidence

- Item-vault-only cutover is implemented and committed.
- Active packages are now CLI/app/cliui/clipboard/config/project/util/vaultjson/vaultmodel.
- `secret copy` now writes a resolved secret to the system clipboard without printing it; WSL uses the verified `clip.exe` bridge fallback.
- `secret export` and `project export` now support only `shell` and `json` output.
- Default vault path is `~/.local/share/shelf/vault.json`.
- Local config points at `/home/han/.local/share/shelf/vault.json`.
- Local backup before cutover: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Built-in release packaging/tagging is deferred until a future public v1.0.0 milestone.
- Latest verification passed:
  - `go test ./...`
  - `go vet ./...`
  - `go-arch-lint check`
  - `go build ./cmd/shelf`
  - installed-binary `secret copy` smoke test with a temporary vault and fake value

## Updated

- 2026-07-10
