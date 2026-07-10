# Verification: Item Vault Plain JSON Cutover

## Claims Checked

- Built-in age/v1 stack is removed from active code.
- Active CLI exposes setup, doctor, secret, project, and completion only.
- Item vault uses configured JSON path directly, without `.v2.json` suffix.
- Local workstation config points at `/home/han/.local/share/shelf/vault.json`.
- Local existing item metadata is readable without printing values.
- Label export env-name check does not include items without explicit env names.
- Full repository verification passes.

## Evidence Observed

- Local backup created at `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Local config rewritten to:
  - `vault_path: /home/han/.local/share/shelf/vault.json`
- Local vault copied from previous item sidecar to:
  - `/home/han/.local/share/shelf/vault.json`
- Local file modes verified as `0o600` for config and vault.
- Installed binary smoke checks passed:
  - `shelf doctor`
  - `shelf secret list --long`
  - `shelf secret export --label ai --format shell | cut -d= -f1`
  - `shelf --help`
- Repository verification passed:
  - `go test ./...`
  - `go vet ./...`
  - `go-arch-lint check`

## Result

Passed. Shelf now operates locally as an item-vault JSON CLI with external encryption expected from chezmoi or equivalent tooling.
