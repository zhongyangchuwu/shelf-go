# Security model

Shelf is a local secret manager. Its main safety guarantee is that the durable vault file is encrypted at rest with age before it is suitable for backup, Git, or chezmoi sync.

## Protected data

Shelf encrypts the vault file before durable persistence. The decrypted in-memory model is a JSON object with `version` and `secrets` fields, but that plaintext model should only exist during command execution.

Encrypted vault files are intended to be portable:

```text
~/.local/share/shelf/vault.age
```

Use this before committing or syncing a vault:

```bash
shelf vault status
shelf doctor
```

## Config is not a secret store

Default config path:

```text
~/.config/shelf/config.yaml
```

Config may contain:

- `vault_path`: path to the encrypted vault.
- `recipients`: public age recipients.
- `identity_paths`: filesystem paths to private age identities.
- `editor`: editor command for `secret edit`.

Config must not contain:

- secret values;
- private age identity contents;
- generated `.env` values.

Private identity files are sensitive even though their paths may appear in config.

## Project manifests are value-free

Project manifests live at `<git-root>/.shelf.json`. They may be committed after review because they store references, not values:

```json
{
  "version": 1,
  "secrets": [
    { "path": "app:token", "env": "APP_TOKEN" },
    { "prefix": "providers/openai", "required": false }
  ]
}
```

`.shelf.json` must not contain `value`, fallback plaintext, shell commands, or templates.

## Plaintext materialization boundaries

These operations intentionally handle plaintext values:

- `shelf secret get` prints a value.
- `shelf secret export` prints values in shell/env/JSON form.
- `shelf project export` prints resolved project values.
- `shelf project run` puts values in a child process environment.
- `shelf secret edit` writes a temporary editor buffer containing the secret object.
- `shelf vault open` can reveal values in the local browser and can update the encrypted vault.

Generated `.env` and `.env.local` files are plaintext artifacts. Do not commit them.

## Migration safety

`vault migrate` reads an old plaintext JSON store, writes an encrypted vault, decrypts and validates the target, then confirms the source was not changed.

```bash
shelf vault migrate --from secrets.json --to vault.age
shelf vault status
```

Shelf preserves the plaintext source. After verification, move, delete, or securely archive it yourself.

## Localhost manager safety

`shelf vault open` starts an on-demand local HTTP manager. It is not a hosted service and not a permanent daemon.

Current safety boundaries:

- listens on loopback addresses only;
- prints a random token in the local URL;
- validates expected Host and Origin for browser requests;
- list/search responses show metadata, not values;
- reveal actions are explicit and return plaintext locally;
- create/update/delete reuse the same vault validation, locking, encrypted-save, and backup rules as CLI writes.

## Non-goals

Shelf does not currently provide team sharing, hosted sync, browser autofill, or a long-running unlock daemon. The current product is for one developer managing local project secrets.
