# Shelf Go Usage Spec

## Purpose

Shelf Go is a local-first secret manager for solo developers. It keeps secrets in an age-encrypted vault file and provides fast CLI workflows for direct export, project manifests, and runtime injection.

Core rules:

- Keep commands scriptable and predictable.
- Keep config and project manifests value-free.
- Encrypt the durable secret source of truth.
- Reveal values only through explicit value-producing commands or manager actions.

## Configuration and vaults

Global flags:

```bash
shelf --config ~/.config/shelf/config.yaml --vault ~/.local/share/shelf/vault.age <command>
```

Defaults:

```text
config: ~/.config/shelf/config.yaml
vault:  ~/.local/share/shelf/vault.age
```

Environment overrides:

```text
SHELF_CONFIG  config file path
SHELF_VAULT   vault file path
EDITOR        fallback editor for `secret edit`
```

Config YAML:

```yaml
version: 1
vault_path: ~/.local/share/shelf/vault.age
recipients:
  - age1...
identity_paths:
  - ~/.config/shelf/identity.txt
editor: vim
```

Rules:

- `recipients` are public age recipients used to encrypt the vault.
- `identity_paths` are paths to private age identity files used to decrypt the vault; the config stores paths, not private key material.
- The identity files themselves are sensitive and must not be committed.
- The encrypted vault file is portable and can be backed up or managed by chezmoi/Git.
- Config is non-secret if it contains only public recipients and identity paths.

## Initialization

```bash
shelf init
shelf init --vault-path ~/.local/share/shelf/vault.age --recipient age1... --identity ~/.config/shelf/identity.txt
```

Behavior:

- Creates config if needed.
- Creates or preserves the encrypted vault.
- Does not overwrite an existing vault when `--force` rewrites config.

## Secret paths

Secret paths have the form `group_path:key`.

Examples:

```text
providers/openrouter/accounts/personal:api_key
providers/openai/accounts/work:api_key
github/accounts/personal:token
```

Rules:

- Exactly one `:` separator.
- `group_path` is the namespace before `:` and may contain `/`.
- `key` is the leaf after `:` and must not contain `/`.
- Allowed path tokens are letters, numbers, `_`, `-`, and `.`.

## Secret commands

```bash
shelf secret add [path-or-group]
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>
shelf secret rm <path>
```

Value behavior:

- `secret set` JSON-parses `<value>` when possible; otherwise it stores a string.
- `secret get` prints the plaintext value by design.
- `secret list` prints paths only.
- `secret info` prints metadata and `value_set`, not the value.
- `secret edit` opens the full object in `$EDITOR`; the edit buffer contains plaintext while the editor is open.

Examples:

```bash
shelf secret set app:port 34222
shelf secret set flags:enabled true
shelf secret set app:options '{"debug":false}'
shelf secret set providers/openrouter/accounts/personal:api_key sk-xxx --env OPENROUTER_API_KEY
```

## Migration

```bash
shelf migrate --from ~/.local/share/shelf/secrets.json --to ~/.local/share/shelf/vault.age
```

Behavior:

- Reads the plaintext JSON source.
- Writes the encrypted target vault.
- Decrypts and validates the target before reporting success.
- Leaves the plaintext source unchanged.
- Refuses to replace an existing target unless `--force` is supplied.
- Refuses plaintext JSON as the target path.

Cleanup requirement:

- After confirming the new vault and config, move, delete, or securely archive the old plaintext source.
- Plaintext migration sources are not safe to commit or sync.

## Direct export

```bash
shelf export <path-or-prefix> --format shell|env|json [--all]
```

Formats:

```text
shell  export KEY=value lines
env    KEY=value lines
json   JSON object
```

Rules:

- Exact paths and prefixes are supported.
- By default, prefix export includes secrets with an explicit `env` field.
- `--all` includes secrets without `env` by deriving names from paths.
- `shell` output is intended for `eval "$(shelf export ... --format shell)"`.
- Export output contains plaintext values by design.
- Redirected env files such as `.env.local` contain plaintext values and must be gitignored.

## Project manifests

Project manifests live at `<git-root>/.shelf.json`.

```json
{
  "version": 1,
  "secrets": [
    {
      "path": "providers/openai/accounts/personal:api_key",
      "env": "OPENAI_API_KEY",
      "required": true
    },
    {
      "prefix": "providers/openrouter/accounts/personal",
      "required": false
    }
  ]
}
```

Rules:

- `.shelf.json` is a value-free manifest of intent.
- It may contain exact `path`, `prefix`, env override, and required/optional flags.
- It must not contain `value`, fallback plaintext, shell commands, or templates.
- It can be committed when reviewed as value-free config.

Commands:

```bash
shelf project id
shelf project init [--force]
shelf project explain
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project export --format env|shell|json
```

`project explain` prints project identity, manifest path, env names, and missing/conflict diagnostics. It does not print values.

`project export` resolves `.shelf.json`, reads encrypted vault values, and prints plaintext env bindings. Use generated output carefully:

```bash
shelf project export --format env > .env.local
```

`.env.local` contains plaintext values and must not be committed.

## Runtime injection

```bash
shelf run -- command args...
shelf run --dry-run -- command args...
```

Behavior:

1. Find Git root.
2. Load `<git-root>/.shelf.json`.
3. Resolve path/prefix entries.
4. Read values from the encrypted vault.
5. Compute env names.
6. Fail on required missing secrets or env conflicts.
7. Execute the child command with injected environment.

Rules:

- `run` injects values into the child process only; it does not mutate the parent shell.
- Shelf-resolved env vars override parent env vars.
- `run --dry-run` prints injected env names and override warnings, not values.
- Child process output may print values if the child command prints them.

## Doctor

```bash
shelf doctor
```

Checks:

- Config resolution.
- Version.
- Vault existence, permissions, format, and loadability.
- Plaintext JSON vs encrypted `shelf-vault/v1` format.
- Ordinary Git tracking state.
- zsh completion installation.

Safety behavior:

- Tracked plaintext secret stores are failures.
- Tracked encrypted vault files are reported as safe encrypted vaults.
- Chezmoi integration is not required; chezmoi can manage the encrypted vault file as an ordinary file.

## Localhost vault manager

```bash
shelf manager
shelf manager --addr 127.0.0.1:0
```

Behavior:

- Starts an on-demand HTTP manager bound to loopback.
- Prints a local URL containing a random token.
- Provides metadata search/browse, explicit reveal, create/update, and delete.
- Uses the same encrypted vault load/update path as CLI writes.
- Requires no hosted backend and no permanent daemon.

Safety controls:

- Non-loopback listen addresses are rejected.
- Requests require a manager token.
- Unsafe write methods require a valid Origin.
- Requests must use the expected Host.
- List/search responses include non-secret metadata only.
- Reveal actions intentionally return plaintext values.

Browser warnings:

- The tokenized URL can appear in local browser history.
- Revealed values are visible in the browser process and screen.
- Treat browser reveal like `secret get`: intentional plaintext access.

## Non-goals for v1

- Team sharing.
- Hosted secret service.
- Permanent daemon.
- Browser extension or autofill.
- Direct chezmoi control.
- Plain `.env` as source of truth.
- Password-only encryption.
- Multiple vault profiles.
