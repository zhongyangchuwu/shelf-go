# State

## Current Position

- Phase: Item Vault Plain JSON Cutover
- Status: implementation verified locally; commit pending
- Active Artifact: `.planning/phases/042-item-vault-storage-crypto-boundary/PLAN.md`
- Next Action: Commit the item-vault-only cutover, then run final installed-binary smoke tests before v0.2.0 release prep.

## Blockers

- None.

## Recent Evidence

- Removed built-in age/v1 stack and v1 surfaces from active product scope.
- Active packages are now CLI/app/config/project/util/vaultjson/vaultmodel.
- Default vault path is `~/.local/share/shelf/vault.json`.
- Local config now points at `/home/han/.local/share/shelf/vault.json`.
- Local backup before cutover: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Installed binary smoke checks passed:
  - `shelf doctor`
  - `shelf secret list --long`
  - `shelf secret export --label ai --format shell | cut -d= -f1`
  - `shelf --help`
- Verification passed:
  - `go test ./...`
  - `go vet ./...`
  - `go-arch-lint check`

## Updated

- 2026-07-10
