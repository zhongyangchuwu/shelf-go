# State

## Current Position

- Phase: Safe Catalog TUI
- Status: design drafted — awaiting user review
- Active Artifact: `.planning/phases/048-safe-catalog-tui/UI-SPEC.md`
- Next Action: User review of the UI specification; do not implement before approval.

## Blockers

- None.

## Recent Evidence

- Item-vault-only cutover is implemented and committed.
- Active packages are now CLI/app/cliui/clipboard/config/project/util/vaultjson/vaultmodel.
- `secret copy` now writes a resolved secret to the system clipboard without printing it; WSL tries `clip.exe` then verified `wl-copy` fallback.
- `secret export` and `project export` now support only `shell` and `json` output.
- Transitional implementation `V2` names were removed; persisted schema version fields remain.
- `vaultmodel.Resolver` now owns shared item selector, field, and env-name resolution.
- `vaultmodel.Registry` now owns primary-secret and env-name template-role lookup for item creation and resolution.
- `secret update` atomically changes title, labels, aliases, or one typed public field; secret fields remain stdin-only.
- Phase 48 UI/technical design is drafted; no TUI code or runtime dependencies have been added.
- The proposed personal workflow sequence is scriptable item lifecycle → safe catalog TUI → TUI editing → project binding TUI → credential health audit.
- TUI remains optional and keyboard-first; it must never display secret values and must reuse application workflows.
- Named dev/staging/prod environments are not planned for the personal-use path.
- Default vault path is `~/.local/share/shelf/vault.json`.
- Local config points at `/home/han/.local/share/shelf/vault.json`.
- Local backup before cutover: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Built-in release packaging/tagging is deferred until a future public v1.0.0 milestone.
- Latest verification passed:
  - `go test ./...`
  - `go vet ./...`
  - `go-arch-lint check`
  - `go build ./cmd/shelf`
  - installed-binary `secret update` smoke test with a temporary vault and fake value

## Updated

- 2026-07-12
