# Security Policy

Shelf is a local secret manager. Its main safety guarantee is that the durable vault file is encrypted at rest with age before it is suitable for backup, Git, or chezmoi sync.

## Reporting vulnerabilities

Shelf is pre-release software. Please report suspected security issues privately to the project maintainer before opening a public issue. Include:

- affected command or workflow;
- local platform and Shelf version or commit;
- whether plaintext secret values may have been exposed;
- reproduction steps that avoid including real secrets.

Do not include real secret values, private age identities, vault files, or generated `.env` files in reports.

## Supported versions

Security support currently tracks the latest commit on the main development branch until the first public release is tagged.

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
- `editor`: preferred editor for legacy or future interactive flows; v2 secret writes currently enter through stdin.

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
- `shelf project export` prints resolved project values. Its default `shell` output is intended for explicit workflows such as `shelf project export > .env.local` followed by `source .env.local`.
- `shelf project run` puts values in a child process environment.
- `shelf secret set --stdin` and `shelf secret create --stdin` read plaintext values from stdin.
- `shelf manager` can reveal values in the local browser for v1 vaults and can update the encrypted v1 vault. Its tokenized local URL may remain visible in browser history or local process/UI surfaces. During the v2 cutover it is gated when a valid v2 vault exists.

Generated `.env` and `.env.local` files are plaintext artifacts. Add them to `.gitignore`, do not commit them, and delete them when they are no longer needed.

Cutover v2 JSON vault files (`*.v2.json`) are plaintext secret stores until v2 age wrapping is implemented. Treat them like generated secret material: keep permissions restrictive, do not commit or sync them, and prefer the encrypted v1 vault for portable storage until v2 encryption is wired in.

## Migration safety

`vault migrate` reads an old plaintext JSON store, writes an encrypted vault, decrypts and validates the target, then confirms the source was not changed.

```bash
shelf vault migrate --from secrets.json --to vault.age
shelf vault status
```

Shelf preserves the plaintext source. After verification, move, delete, or securely archive it yourself.

## Backup recovery safety

Shelf replacement writes preserve the previous encrypted vault as `<vault>.bak`. This is a single last-write backup, not a history system; each later replacement overwrites the previous `.bak`.

To recover manually:

```bash
cp vault.age vault.age.bad
cp vault.age.bak vault.age
shelf vault status
```

The backup is still encrypted. A configured identity must be able to decrypt it. If all matching private identities are lost, the encrypted vault and backup cannot be recovered by Shelf.

## Localhost manager safety

`shelf manager` starts an on-demand local HTTP manager for v1 vaults. It is not a hosted service and not a permanent daemon. During the v2 cutover it refuses to open when the resolved v2 vault exists.

Current safety boundaries:

- listens on loopback addresses only;
- prints a random token in the local URL;
- validates expected Host and Origin for browser requests;
- list/search responses show metadata, not values;
- reveal actions are explicit and return plaintext locally;
- create/update/delete reuse the same vault validation, locking, encrypted-save, and backup rules as CLI writes;
- the manager has no TLS because it is loopback-only; close it with Ctrl-C when finished.

## Non-goals

Shelf does not currently provide team sharing, hosted sync, browser autofill, or a long-running unlock daemon. The current product is for one developer managing local project secrets.
