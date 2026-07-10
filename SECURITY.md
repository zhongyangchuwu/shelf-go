# Security Policy

Shelf Go currently stores secrets in a local JSON item vault. The vault file contains plaintext secret values. Protect it with filesystem permissions and encrypt it with external tooling, such as chezmoi, before committing or syncing.

## Protected data

Protect:

- the configured vault JSON file;
- generated `.env` / `.env.local` files;
- shell histories or logs that include exported values;
- process environments created by `shelf project run`.

Check local health:

```bash
shelf doctor
```

## Config is not a secret store

Default config path:

```text
~/.config/shelf/config.yaml
```

Config may contain:

- `vault_path`: path to the local item vault JSON;
- `editor`: reserved preference for future interactive flows.

Config must not contain:

- secret values;
- generated `.env` values;
- private key material for external encryption tools.

## Project manifests are value-free

Project manifests live at `<git-root>/.shelf.json`. They may be committed after review because they store item/field/env references, not values:

```json
{
  "version": 1,
  "bindings": [
    { "item": "item_openai", "field": "value", "env": "OPENAI_API_KEY" }
  ]
}
```

`.shelf.json` must not contain values, fallback plaintext, shell commands, or generated env files.

## Plaintext materialization boundaries

These operations intentionally handle plaintext values:

- `shelf secret get` prints a value.
- `shelf secret export` prints values in shell/env/JSON form.
- `shelf project export` prints resolved project values.
- `shelf project run` puts values in a child process environment.
- `shelf secret create --stdin` and `shelf secret set --stdin` read plaintext values from stdin.

Generated `.env` and `.env.local` files are plaintext artifacts. Add them to `.gitignore`, do not commit them unless encrypted by an external tool, and delete them when no longer needed.

## Non-goals

Shelf does not currently provide built-in encryption, team sharing, hosted sync, browser autofill, a Web manager, a TUI editor, or a long-running unlock daemon. The current product is for one developer managing local project secrets with external encryption/sync tooling.
