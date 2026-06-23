# Shelf Go

Shelf Go is a local-first encrypted secret environment manager for solo developers. It stores secrets in a portable age-encrypted vault file, keeps project manifests value-free, and preserves scriptable CLI workflows for direct secret export and `shelf project run`.

Core workflow:

```text
Encrypted local vault + value-free project manifests + scoped CLI export/run.
```

## Specs

- [Usage spec](docs/usage-spec.md)
- [Data spec](docs/data-spec.md)
- [Roadmap](docs/roadmap.md)

## Current command surface

```bash
shelf setup [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
shelf vault init [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
shelf vault migrate --from <plaintext.json> [--to <vault.age>] [--force]
shelf vault status|check
shelf doctor

shelf secret add [path-or-group]
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>
shelf secret rm <path>

shelf secret export <path-or-prefix> --format shell|env|json [--all]

shelf project id
shelf project init [--force]
shelf project explain
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project export --format env|shell|json

shelf project run -- command args...
shelf project run --dry-run -- command args...

shelf vault open [--addr 127.0.0.1:0]
shelf completion zsh
```

Global flags:

```bash
--config PATH   Path to config.yaml
--vault PATH    Path to encrypted vault
```

First-run flow:

```bash
shelf setup --recipient age1... --identity ~/.config/shelf/identity.txt
shelf vault status
```

Use `shelf vault init` for first-time vault initialization. Use `shelf vault init --force ...` to rewrite an existing config during repair. Use `shelf vault status` or `shelf vault check` before committing/syncing a vault file; both report config path, vault path, format, recipient count, and decrypt/load status without revealing values.

Plaintext migration flow:

```bash
shelf vault migrate --from ~/.local/share/shelf/secrets.json --to ~/.local/share/shelf/vault.age
shelf vault status
```

After migration succeeds, update config to the encrypted vault path if needed, then move, delete, or securely archive the old plaintext JSON source.

## Storage model

- Default config path: `~/.config/shelf/config.yaml`.
- Default vault path: `~/.local/share/shelf/vault.age`.
- Config contains non-secret settings: vault path, public age recipients, identity file paths, and editor.
- The vault is the encrypted source of truth and is suitable for backup or git/chezmoi sync.
- `.shelf.json` project manifests contain only secret paths, prefixes, env overrides, and required/optional flags. They must not contain values.

## Safety notes

- `shelf secret get`, `shelf secret export`, `shelf project export`, and `shelf vault open` reveal actions intentionally materialize plaintext values.
- Generated `.env` / `.env.local` files contain plaintext values. Do not commit them.
- `shelf vault migrate` preserves the old plaintext JSON source after successful encrypted migration; delete, move, or archive it manually after verifying the new vault.
- `shelf doctor` reports plaintext-vs-encrypted store format and flags tracked plaintext secret files as unsafe.
- `shelf vault open` binds to loopback by default and uses a random token plus Host/Origin checks, but browser reveal actions still show plaintext values locally.

## Status

Implemented:

- Secret CRUD and interactive add/edit flows.
- Age-encrypted vault persistence with encrypted backups and actionable load errors.
- Plaintext-to-vault migration.
- Git/chezmoi safety checks in `shelf doctor`.
- Direct export in shell/env/JSON formats.
- Project manifests and project export.
- `shelf project run` runtime injection and value-free dry-run.
- Localhost-only vault manager for metadata search, explicit reveal, create/update, and delete.
