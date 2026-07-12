# Shelf Go

Shelf Go is a local, CLI-first secret manager for one developer. It stores item-based secrets in a local JSON vault and leaves encryption/sync to your dotfile tooling, such as chezmoi.

The current product is CLI-first: scriptable secret and project workflows remain authoritative, while optional wide-screen workspaces support interactive secret and project-binding management.

## Setup

Create a config file and an empty item vault:

```bash
shelf setup
```

Default paths:

```text
~/.config/shelf/config.yaml
~/.local/share/shelf/vault.json
```

Override paths when needed:

```bash
shelf setup --vault-path ~/.local/share/shelf/vault.json
shelf --vault ~/.local/share/shelf/vault.json secret list
```

The vault JSON contains secret values. Do not commit it unless an external tool such as chezmoi encrypts it first.

## Store and inspect secrets

Secret values enter through stdin, not argv:

`secret create` prints a small summary with title, id, template, and env name. `secret list --long` prints a titled, aligned table without secret values.
```bash
printf %s "$APP_TOKEN" | shelf secret create app-token --env APP_TOKEN --stdin
printf %s "$OPENAI_API_KEY" | shelf secret create openai --profile openai --stdin
```

Inspect metadata without printing secret values:

```bash
shelf secret info openai
shelf secret list --long
```

Print a value only when intentionally needed:

```bash
shelf secret get openai
shelf secret get --env OPENAI_API_KEY
```

Copy a single value for interactive use without printing it:

```bash
shelf secret copy openai
shelf secret copy --env OPENAI_API_KEY
```

Update a field from stdin; omit `--field` to update the primary secret field:

```bash
printf %s "$NEW_TOKEN" | shelf secret update openai
printf %s LLM_API_KEY | shelf secret update openai --field credential_env
```

Update item identity, organization, and public metadata without editing vault JSON:

```bash
shelf secret update openai --title openai-personal --add-label ai --add-alias openai-personal
shelf secret update openai-personal --field provider --value OpenAI
shelf secret update openai-personal --field scopes --value responses --value embeddings
shelf secret update openai-personal --field website --clear-field
```

`--value` is limited to public fields. If it is omitted for a selected field, Shelf reads stdin; on an interactive terminal it prompts without echoing input. Secret values never enter argv.

Load a label group into the current shell without storing values in shell config:

```bash
source <(shelf secret export --label ai --format shell)
```

Repeated labels use AND matching. Items without an explicit env-name field are skipped.

## Interactive secret workspace

On a 100×24-or-larger interactive terminal, open the optional full-screen catalog:

```bash
shelf tui
```

Use `a` to create, `e` to edit the selected item or field, `c` to copy a selected/default field, and `d` to remove an item after confirmation. `/` searches safe catalog metadata; `f` filters by template and labels. The workspace masks secret input, never reveals a secret value, and shows value-free expiry health where configured; `get`, export, and project run remain CLI-first.

## Use secrets in a project

Inside a Git worktree, create a value-free project manifest:

```bash
shelf project init
```

On a 100×24-or-larger interactive terminal, manage value-free bindings in the optional workspace:

```bash
shelf project tui
```

Use `a` to choose a secret field and env name, `d` to remove a binding after confirmation, and `i` to initialize a missing manifest. The workspace shows binding resolution state and diagnostics without values; `project export` and `project run` stay explicit CLI actions.

Add concrete item-field bindings:

```bash
shelf project add openai --env OPENAI_API_KEY
shelf project status
```

Export sourceable shell lines when explicitly wanted:

```bash
shelf project export > .env.local
source .env.local
```

Run a child command with project secrets injected:

```bash
shelf project run -- npm run dev
```

Preview injected env names without values:

```bash
shelf project run --dry-run -- npm run dev
```

`.shelf.json` stores item IDs, field IDs, env names, and required flags only. It must not contain secret values.

## Credential health

Audit expiry metadata without materializing values:

```bash
shelf secret audit --within 30
```

Tracked credentials report `expired`, `due`, `missing`, `invalid`, or `healthy` status. The command exits `3` for expired credentials, `2` for warnings, and `0` when all tracked credentials are healthy. Use `--format json` for value-free script output.

## Health check

```bash
shelf doctor
```

`doctor` validates config resolution, vault permissions, and item-vault loadability without printing secret values.

## Current non-goals

- Built-in age encryption. Use chezmoi or another external tool for encrypted sync.
- v1 vault compatibility or automated migration.
- Web manager. Both secret and project binding TUI workspaces are available; browser and hosted interfaces remain deferred.
- Team sharing, hosted sync, browser autofill, or daemon-based unlock flows.
