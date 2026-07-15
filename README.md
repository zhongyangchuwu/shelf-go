# Shelf Go

Shelf Go is a local, CLI-first secret manager for one developer. It stores item-based secrets in a local JSON vault and leaves encryption/sync to your dotfile tooling, such as chezmoi.

The current product is CLI-first: scriptable secret and project workflows remain authoritative, while optional wide-screen workspaces support interactive secret and project-binding management.

**Current local milestone:** `v0.2.0`. This marks the shipped personal-workflow feature set; it is not a public tag or release.

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
printf %s LLM_API_KEY | shelf secret update openai --field env_credential
```

Update item identity, organization, and public metadata without editing vault JSON:

```bash
shelf secret update openai --title "OpenAI personal" --description "Personal development account" --add-label ai
shelf secret update "OpenAI personal" --field provider --value OpenAI
shelf secret update "OpenAI personal" --field scopes --value responses --value embeddings
shelf secret update "OpenAI personal" --field website --clear-field
```

`--value` is limited to public fields. If it is omitted for a selected field, Shelf reads stdin; on an interactive terminal it prompts without echoing input. Secret values never enter argv.

`description` is public metadata shown by `secret info` and the TUI. Use it for context only; never put a secret value in it.

Load all explicitly mapped fields in a label group into the current shell without storing values in shell config:

```bash
source <(shelf env export --label ai --format shell)
```

Repeated labels use AND matching. `env_<field>` maps that field to a default environment name; mapped secret and public `string`, `concealed`, and `url` fields are exported. Unmapped fields remain metadata.

## Interactive secret workspace

On a 100×24-or-larger interactive terminal, open the optional full-screen catalog:

```bash
shelf tui
```

Use `a` to create, `e` to edit the selected item, `c` to copy a selected/default field, and `d` to remove an item after confirmation. Create and Edit use one scrollable form; secret inputs are masked and never prefilled. `/` searches safe catalog metadata; `f` filters by template and labels. The catalog is title-first and uses a colored symbol for missing, attention, healthy, or unavailable state; Detail shows the reason, fields, and environment mappings. `get`, `env export`, and project run remain CLI-first.

The catalog and right detail/editor remain visible as separate Catppuccin Mocha panels. A blue frame and `›` title marker identify the active pane; focus and catalog status never depend on color alone.

## Use secrets in a project

Inside a Git worktree, create a value-free project manifest:

```bash
shelf project init
```

On a 100×24-or-larger interactive terminal, manage value-free bindings in the optional workspace:

```bash
shelf project tui
```

Use `a` to choose a projectable item field and env name, `d` to remove a binding after confirmation, and `i` to initialize a missing manifest. The workspace uses the same persistent Catppuccin Mocha panel frame and shows binding resolution state and diagnostics without values; `project export` and `project run` stay explicit CLI actions.

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

Run a child command with project fields injected:

```bash
shelf project run -- npm run dev
```

Preview injected env names without values:

```bash
shelf project run --dry-run -- npm run dev
```

`.shelf.json` stores item IDs, field IDs, env names, and required flags only. It must not contain secret values.

## Credential health

Audit lifecycle validity without materializing values:

```bash
shelf secret audit --within 30
```

Tracked fields report `expired`, `due`, `unknown`, `invalid`, `permanent`, or `healthy` status. The command exits `3` for expired credentials, `2` for warnings, and `0` when all tracked fields are healthy or permanent. Use `--format json` for value-free script output.

## Health check

```bash
shelf doctor
```

`doctor` validates config resolution, vault permissions, and item-vault loadability without printing secret values.

## Current non-goals

- Built-in age encryption. Use chezmoi or another external tool for encrypted sync.
- Vault compatibility or automatic migration. See the [manual migration guide](docs/v3-migration.md), including the V3-to-V4 description cutover.
- Web manager. Both secret and project binding TUI workspaces are available; browser and hosted interfaces remain deferred.
- Team sharing, hosted sync, browser autofill, or daemon-based unlock flows.
