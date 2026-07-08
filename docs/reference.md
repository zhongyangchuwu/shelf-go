# Reference

## Global flags

```bash
--config PATH   Path to config.yaml
--vault PATH    Path to encrypted vault
```

## Environment variables

| Variable | Purpose |
| --- | --- |
| `SHELF_CONFIG` | Overrides the config file path. |
| `SHELF_VAULT` | Overrides the encrypted vault path. |
| `EDITOR` | Fallback editor for editor-based workflows when config has no editor value. |
| `FPATH` / `fpath` | Used by `shelf doctor` to check zsh completion installation. |

## Default paths

| Path | Default |
| --- | --- |
| Config | `~/.config/shelf/config.yaml` |
| Vault | `~/.local/share/shelf/vault.age` |

Vault path resolution order:

1. `--vault`
2. `SHELF_VAULT`
3. `vault_path` in config
4. default vault path

Config path resolution order:

1. `--config`
2. `SHELF_CONFIG`
3. default config path

## Config file

```yaml
version: 1
vault_path: ~/.local/share/shelf/vault.age
recipients:
  - age1...
identity_paths:
  - ~/.config/shelf/identity.txt
editor: vim
```

`recipients` are public age recipients. `identity_paths` point to private age identity files; the private key material is not stored in config.

## Secret items

The v2 secret model stores template-based items. A template defines fields and sensitivity; a profile supplies default public field values over a template.

Built-in templates:

- `generic_secret` (`generic`): `value` secret field plus public `value_env` env-name field.
- `api_credential` (`api`): `credential` secret field plus public `credential_env`, `provider`, `website`, `expires_at`, and `scopes` fields.

Built-in profiles:

- `openai`: `api_credential` defaults for `provider`, `website`, and `credential_env`.

Secret values are written from stdin. Env names are ordinary public fields such as `value_env` and `credential_env`, not a separate binding object.

```bash
printf %s "$APP_TOKEN" | shelf secret create app-token --env APP_TOKEN --stdin
printf %s "$OPENAI_API_KEY" | shelf secret create openai --profile openai --stdin
printf %s LLM_API_KEY | shelf secret set openai --field credential_env --stdin
```

During the Phase 38 cutover, v2 secret commands use the plaintext v2 JSON repository boundary. If `--vault` is supplied, that path is used directly. Without `--vault`, Shelf writes to `<resolved-vault-path>.v2.json` so the default v2 CLI does not overwrite the existing v1 encrypted vault during development/cutover. Treat that v2 JSON file as secret data and do not commit or sync it until v2 age wrapping is added.

## CLI commands

### `shelf setup`

Creates or reuses config, identity, and encrypted vault state.

```bash
shelf setup [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
```

### `shelf vault init`

Vault-scoped form of setup.

```bash
shelf vault init [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
```

### `shelf vault migrate`

Migrates a legacy plaintext JSON store to an encrypted vault.

```bash
shelf vault migrate --from <plaintext.json> [--to <vault.age>] [--force]
```

The source is preserved after successful migration.

### `shelf vault status` / `shelf vault check`

Reports config path, vault path, recipient configuration, vault format, and loadability without revealing values.

```bash
shelf vault status
shelf vault check
```

### `shelf manager`

Starts the on-demand local manager.

```bash
shelf manager [--addr 127.0.0.1:0]
```

The manager prints a tokenized loopback URL and serves an embedded Web console for search, add, edit, rename, delete, reveal, copy, and tag workflows. List/detail responses are metadata-only; reveal/copy actions intentionally return plaintext values. Treat the URL and revealed values as sensitive local plaintext, and stop the process with Ctrl-C when finished.

### `shelf doctor`

Checks config resolution, version, vault existence/mode/format/loadability, Git tracking safety, and zsh completion state.

```bash
shelf doctor
```

### `shelf secret create`

Creates a v2 secret item. Secret values enter through stdin when `--stdin` is set.

```bash
shelf secret create <title> [--template generic|api] [--profile PROFILE] [--id ID] [--env NAME] [--stdin]
```

Useful public-field flags for `api_credential` are `--provider`, `--website`, `--expires`, and repeatable `--scope`. Explicit flags override profile defaults.

### `shelf secret set`

Sets a field from stdin. Without `--field`, Shelf sets the template primary secret field.

```bash
shelf secret set <selector> [--field FIELD] --stdin
```

### `shelf secret get`

Prints a plaintext field value. Without `--field`, Shelf prints the template primary secret field.

```bash
shelf secret get <selector> [--field FIELD]
shelf secret get --env ENV [--field FIELD]
```

`--env` resolves through env-name fields such as `credential_env`. Duplicate env-name matches fail with a value-free ambiguity error.

### `shelf secret list`

Prints v2 secret items without secret values.

```bash
shelf secret list [query] [--long]
```

`--long` prints item ID, title, template, primary env name, and value-set state.

### `shelf secret info`

Prints item metadata as JSON. Public field values are shown. Secret field values are not printed; only `value_set` is shown.

```bash
shelf secret info <selector>
```

### `shelf secret rm`

Removes a v2 secret item after explicit confirmation.

```bash
shelf secret rm <selector> --yes
```

Removed or deferred from the v2 secret CLI: `secret add`, `secret edit`, direct `secret export`, `secret copy`, and env binding subcommands.

### `shelf project id`

Prints the current Git project identity.

```bash
shelf project id
```

### `shelf project init`

Creates `<git-root>/.shelf.json`.

```bash
shelf project init [--force]
```

### `shelf project configure`

Opens a TUI for project secret bindings. The current flow lets you queue env-file variable bindings, add existing vault secrets, remove existing bindings, review pending changes, and apply them to `.shelf.json`. It does not import env-file values or print secret values.

```bash
shelf project configure
```

### `shelf project add`

Adds a binding to the project manifest. Version 1 manifests accept exact paths, prefixes, or tag selectors. Version 2 manifests accept v2 item selectors and bind concrete item fields.

```bash
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project add --tag TAG [--tag TAG ...] [--optional]
shelf project add <item-id-or-title> [--field FIELD] [--env NAME] [--optional]
```

For version 1 manifests, `--env` is valid only for exact path entries. Prefix and tag entries may expand to multiple secrets and cannot carry an env override. For version 2 manifests, `--field` defaults to the item's primary secret field and `--env` defaults to that field's public env-name field.

### `shelf project rm`

Removes a manifest entry by exact path, prefix, comma-joined tag selector key, or v2 `item.field` key.

```bash
shelf project rm <manifest-entry-key>
```

### `shelf project list`

Lists manifest entries without resolving values.

```bash
shelf project list
```

### `shelf project status`

Shows project identity, manifest path, resolved env names, missing entries, conflicts, and local env-file summaries without printing values.

```bash
shelf project status
```

### `shelf project export`

Prints resolved project env bindings with plaintext values. The default format is `shell`, which prints sourceable `export NAME=value` lines.

```bash
shelf project export [--format shell|env|json]
```

Use `--format env` for bare `NAME=value` lines and `--format json` for machine-readable output. Redirected files contain plaintext values and must not be committed.

### `shelf project run`

Runs a child process with resolved project secrets injected into its environment.

```bash
shelf project run [--dry-run] -- command args...
```

`--dry-run` prints injected env names and override warnings without values and does not execute the child command.

### `shelf completion`

Generates shell completion scripts.

```bash
shelf completion [bash|zsh|fish|powershell]
```

## Project manifest

Path: `<git-root>/.shelf.json`

Version 1 manifests bind legacy `group:key` paths, prefixes, or tag selectors:

```json
{
  "version": 1,
  "secrets": [
    {
      "path": "app:token",
      "env": "APP_TOKEN",
      "required": true
    },
    {
      "prefix": "providers/openai",
      "required": false
    },
    {
      "tags": ["ai", "prod"],
      "required": false
    }
  ]
}
```

Version 2 manifests bind concrete v2 item fields:

```json
{
  "version": 2,
  "bindings": [
    {
      "item": "item_github",
      "field": "value",
      "env": "GITHUB_TOKEN",
      "required": true
    }
  ]
}
```

Rules:

- `version` is required. `1` uses `secrets`; `2` uses `bindings`.
- Version 1 entries have exactly one selector: `path`, `prefix`, or `tags`.
- Version 1 `env` is a project-level env override for exact path entries only.
- Version 2 bindings name `item`, `field`, and `env` explicitly.
- `required` defaults to `true` in both versions.
- Duplicate v1 path/prefix/tag entries and duplicate v2 item/field bindings are rejected.
- Secret values are prohibited.
- Duplicate resolved env names are a runtime conflict.
