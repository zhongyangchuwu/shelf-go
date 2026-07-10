# Plan: Item Vault Plain JSON Cutover

## Objective

Make v0.2.0 a clean personal-use item-vault CLI by deleting built-in age/v1 code and using a configured plaintext JSON item vault that users can encrypt externally with chezmoi.

## Work Packages

### 1. Remove v1 and age code

- Delete `internal/age`, `internal/jsonvault`, `internal/vault`, `internal/secret`, `internal/manager`, and `internal/importer/gopass`.
- Delete v1 app workflows: import, migrate, manager, v1 export, v1 secret CRUD, v1 runtime, v1 secret service.
- Delete v1 CLI commands: `vault`, manager, import, migrate, v1 export, and project configure TUI.
- Remove age/flock/bubbletea dependencies and architecture-lint components.

### 2. Keep item-vault CLI working

- Keep `secret` item/field workflows over `vaultjson` + `vaultmodel`.
- Make `setup` create config and an empty item vault JSON.
- Make `doctor` validate config, permissions, and vault loadability.
- Use the configured vault path directly; stop appending `.v2.json`.

### 3. Convert project to item bindings only

- `.shelf.json` stores `version` plus concrete `bindings` only.
- Keep `project id/init/status/add/rm/list/export/run`.
- Remove path/prefix/tag selector manifests and project configure TUI.
- Keep project diagnostics value-free.

### 4. Local workstation cutover

- Back up current config and sidecar item vault.
- Copy the local item vault to `~/.local/share/shelf/vault.json`.
- Point config at that path.
- Install updated `shelf`.
- Verify `shelf doctor`, `shelf secret list --long`, and label export env names without printing values.

### 5. Docs and planning

- Update active docs to describe plaintext item vault + external encryption.
- Update planning records so Web/TUI/project expansion stays after CLI/model/file cleanup.

## Non-Goals

- Built-in encryption.
- v1 compatibility.
- Automatic migration.
- Web manager.
- TUI editing.
- Project configure TUI.
- Hosted/team workflows.

## Acceptance Criteria

- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
- `go.mod` no longer requires `filippo.io/age`, `github.com/gofrs/flock`, or Bubble Tea UI dependencies.
- `shelf --help` exposes only active commands.
- Local installed `shelf` reads `/home/han/.local/share/shelf/vault.json` and lists existing item metadata without values.
- `shelf secret export --label ai --format shell | cut -d= -f1` shows only expected env names.
