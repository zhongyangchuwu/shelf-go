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
| `EDITOR` | Fallback editor for `shelf secret edit` when config has no editor. |
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

## Secret paths

Secret paths use this shape:

```text
<group_path>:<key>
```

Examples:

```text
app:token
providers/openai/accounts/personal:api_key
github/accounts/personal:token
```

Rules:

- exactly one `:` separator;
- non-empty group path;
- non-empty key;
- group path segments are separated by `/`;
- key must not contain `/` or `:`.

## Secret values

`secret set` JSON-parses the value when possible. Otherwise it stores a string.

```bash
shelf secret set app:port 34222
shelf secret set flags:enabled true
shelf secret set app:options '{"debug":false}'
shelf secret set app:token sk-example
```

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

### `shelf secret add`

Interactively creates a secret. Reads the value through hidden terminal input.

```bash
shelf secret add [path-or-group]
```

### `shelf secret set`

Creates or replaces a secret.

```bash
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
```

### `shelf secret get`

Prints a plaintext value.

```bash
shelf secret get <path>
```

### `shelf secret list`

Prints matching secret paths only.

```bash
shelf secret list [prefix] [--tag TAG ...]
```

Repeated `--tag` selectors use AND semantics. With a prefix, the tag filter is applied inside that prefix.

### `shelf secret info`

Prints non-secret metadata as JSON.

```bash
shelf secret info <path>
```

### `shelf secret edit`

Opens the full secret object as JSON in `$EDITOR` or configured editor.

```bash
shelf secret edit <path>
```

The editor buffer contains plaintext while editing. Shelf creates the temporary file with restrictive permissions and removes it on normal command exit where possible.

### `shelf secret rm`

Removes a secret.

```bash
shelf secret rm <path>
```

### `shelf secret export`

Exports an exact path, prefix, or tag-selected set in shell, env, or JSON format.

```bash
shelf secret export [path-or-prefix] --format shell|env|json [--all] [--tag TAG ...]
```

At least one path, prefix, or `--tag` selector is required. Repeated `--tag` selectors use AND semantics. For prefix or tag-set export, the default includes only secrets with explicit `env`; `--all` also derives env names for secrets without `env`.

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

### `shelf project add`

Adds an exact path, prefix, or tag selector to the project manifest.

```bash
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project add --tag TAG [--tag TAG ...] [--optional]
```

`--env` is valid only for exact path entries. Prefix and tag entries may expand to multiple secrets and cannot carry an env override.

### `shelf project rm`

Removes a manifest entry by exact path, prefix, or comma-joined tag selector key.

```bash
shelf project rm <path-or-prefix-or-tag-key>
```

### `shelf project list`

Lists manifest entries without resolving values.

```bash
shelf project list
```

### `shelf project explain`

Shows project identity, manifest path, resolved env names, missing entries, and conflicts without printing values.

```bash
shelf project explain
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

Rules:

- `version` is required and fixed at `1`.
- `secrets` is an array.
- Each entry has exactly one selector: `path`, `prefix`, or `tags`.
- `env` is a project-level env override for exact path entries only.
- `required` defaults to `true`.
- Duplicate path, prefix, or identical tag-selector entries are rejected.
- Secret values are prohibited.

Env name resolution order:

1. manifest entry `env`;
2. secret object's `env`;
3. derived from the full secret path.

Tag entries expand to every secret that has all listed tags. If a required tag entry matches no secrets, project explain/export/run reports a missing required entry.

Two entries resolving to the same env name are a conflict.
