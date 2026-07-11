# Shelf Go

Shelf Go is a local, CLI-first secret manager for one developer. It stores item-based secrets in a local JSON vault and leaves encryption/sync to your dotfile tooling, such as chezmoi.

The current product focus is the CLI: create secrets, inspect metadata, export label groups, and bind project commands to item fields. Web and TUI surfaces are deferred.

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

Update fields from stdin:

```bash
printf %s "$NEW_TOKEN" | shelf secret set openai --stdin
printf %s LLM_API_KEY | shelf secret set openai --field credential_env --stdin
```

Load a label group into the current shell without storing values in shell config:

```bash
source <(shelf secret export --label ai --format shell)
```

Repeated labels use AND matching. Items without an explicit env-name field are skipped.

## Use secrets in a project

Inside a Git worktree, create a value-free project manifest:

```bash
shelf project init
```

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

## Health check

```bash
shelf doctor
```

`doctor` validates config resolution, vault permissions, and item-vault loadability without printing secret values.

## Current non-goals

- Built-in age encryption. Use chezmoi or another external tool for encrypted sync.
- v1 vault compatibility or automated migration.
- Web manager and TUI editing.
- Team sharing, hosted sync, browser autofill, or daemon-based unlock flows.
