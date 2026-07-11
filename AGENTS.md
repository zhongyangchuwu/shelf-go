## Project

**Shelf Go**

Shelf Go is a fast local secret manager for solo developers who want developer-secret workflows without `.env` sprawl or hosted services. The current product is a Go CLI that manages item-based secrets, value-free project bindings, label-scoped exports, and `shelf project run`.

Shelf is CLI-first. Web manager and TUI editing are intentionally deferred. Built-in encryption is intentionally removed for the current personal-use direction; use chezmoi or another external tool to encrypt/sync the vault file.

**Core Value:** A single developer can manage project secrets locally through a simple item vault and scriptable CLI while keeping project manifests and shell config value-free.

### Constraints

- **Plaintext vault**: The configured item vault JSON contains secret values. Do not commit or sync it unless an external tool encrypts it first.
- **External encryption**: Chezmoi or similar tooling owns encryption/sync outside Shelf.
- **Local-first**: Shelf must not require a hosted backend, account, daemon, Web UI, or TUI for core workflows.
- **CLI-first UX**: Commands must stay scriptable and predictable.
- **Non-secret config**: Shelf config and `.shelf.json` project manifests must not contain secret values.
- **No v1 compatibility requirement**: v1 age/jsonvault/store paths are removed from active product scope.

## Technology Stack

- Go 1.26.4 production code in `cmd/shelf/main.go` and `internal/**`.
- JSON item vault managed by `internal/vaultjson` and modeled by `internal/vaultmodel`.
- YAML config parsed by `internal/config/config.go`.
- Cobra command surface in `internal/cli`.
- Cross-platform clipboard adapter in `internal/clipboard`, using the native backend or `clip.exe` on WSL.
- Lip Gloss-backed human CLI output helpers in `internal/cliui`.
- Standard Go `testing` package for tests.

## Configuration

- Default config path: `~/.config/shelf/config.yaml`.
- Default vault path: `~/.local/share/shelf/vault.json`.
- `SHELF_CONFIG` overrides config path.
- `SHELF_VAULT` overrides vault path.
- `EDITOR` is reserved for future interactive flows.

## Current Commands

```text
shelf setup
shelf doctor
shelf secret create <title> [--template generic|api] [--profile PROFILE] [--id ID] [--env NAME] [--stdin]
shelf secret set <selector> [--field FIELD] --stdin
shelf secret get <selector> [--field FIELD]
shelf secret get --env ENV [--field FIELD]
shelf secret copy <selector> [--field FIELD]
shelf secret copy --env ENV [--field FIELD]
shelf secret list [query] [--long]
shelf secret info <selector>
shelf secret export --label LABEL [--label LABEL ...] [--format shell|json]
shelf secret rm <selector> --yes
shelf project id
shelf project init [--force]
shelf project add <item-id-or-title> [--field FIELD] [--env NAME] [--optional]
shelf project rm <item.field>
shelf project list
shelf project status
shelf project export [--format shell|json]
shelf completion <shell>
```

## Architecture

| Component | Responsibility | Files |
| --- | --- | --- |
| Entry point | Construct and execute root command | `cmd/shelf/main.go` |
| CLI | Cobra commands, flags, completions, output routing | `internal/cli` |
| CLI UI | Human output helpers for titles, summaries, empty states, and tables | `internal/cliui` |
| App | Workflow orchestration for setup, doctor, secret, project, run | `internal/app` |
| Clipboard | Copy secret values to the system clipboard without printing them | `internal/clipboard` |
| Item model | Vault/item/field/template/profile model and validation | `internal/vaultmodel` |
| Item repository | Strict JSON codec/repository with atomic writes | `internal/vaultjson` |
| Project | Git identity, `.shelf.json`, item-field binding resolution, child env helpers | `internal/project` |
| Config | Config and vault path resolution | `internal/config` |
| Util | Atomic write and env/shell/JSON binding formatting | `internal/util` |

Dependency direction: `cli -> app/cliui/clipboard`; `cliui -> lipgloss`; `clipboard -> github.com/atotto/clipboard`; `app -> project/vaultmodel/vaultjson/config/util`; `project -> vaultmodel/util`; `vaultjson -> vaultmodel/util`.
## Data Rules

- Secret values enter through stdin only for writes.
- `secret get`, `secret export`, `project export`, and `project run` intentionally materialize values. `secret copy` materializes a value only to the system clipboard and must not print it.
- `.shelf.json` stores concrete `item`, `field`, `env`, optional `required`, and optional `title` only.
- Label-scoped export skips items without an explicit env-name field.

## Style and Verification

- Use `gofmt` on changed Go files.
- Use octal permission literals with `0o` notation.
- Return errors instead of logging/exiting in package code.
- Use `cmd.OutOrStdout()` / `cmd.OutOrStderr()` in Cobra commands.
- Baseline verification: `go test ./...`, `go vet ./...`, `go-arch-lint check`.
