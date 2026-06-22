# Shelf Go

Shelf Go is a fast local secret manager for solo developers. It stores secrets in a portable age-encrypted vault file, keeps project manifests value-free, and preserves scriptable CLI workflows for direct export and `shelf run`.

Core workflow:

```text
Encrypted local vault + value-free project manifests + fast CLI export/run.
```

## Specs

- [Usage spec](docs/usage-spec.md)
- [Data spec](docs/data-spec.md)
- [Roadmap](docs/roadmap.md)

## Current command surface

```bash
shelf init [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
shelf migrate --from <plaintext.json> [--to <vault.age>] [--force]
shelf doctor

shelf secret add [path-or-group]
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>
shelf secret rm <path>

shelf export <path-or-prefix> --format shell|env|json [--all]

shelf project id
shelf project init [--force]
shelf project explain
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project export --format env|shell|json

shelf run -- command args...
shelf run --dry-run -- command args...

shelf manager [--addr 127.0.0.1:0]
shelf completion zsh
```

Global flags:

```bash
--config PATH   Path to config.yaml
--vault PATH    Path to encrypted vault
```

## Storage model

- Default config path: `~/.config/shelf/config.yaml`.
- Default vault path: `~/.local/share/shelf/vault.age`.
- Config contains non-secret settings: vault path, public age recipients, identity file paths, and editor.
- The vault is the encrypted source of truth and is suitable for backup or git/chezmoi sync.
- `.shelf.json` project manifests contain only secret paths, prefixes, env overrides, and required/optional flags. They must not contain values.

## Safety notes

- `shelf secret get`, `shelf export`, `shelf project export`, and `shelf manager` reveal actions intentionally materialize plaintext values.
- Generated `.env` / `.env.local` files contain plaintext values. Do not commit them.
- `shelf migrate` preserves the old plaintext JSON source after successful encrypted migration; delete, move, or archive it manually after verifying the new vault.
- `shelf doctor` reports plaintext-vs-encrypted store format and flags tracked plaintext secret files as unsafe.
- `shelf manager` binds to loopback by default and uses a random token plus Host/Origin checks, but browser reveal actions still show plaintext values locally.

## Status

Implemented:

- Secret CRUD and interactive add/edit flows.
- Age-encrypted vault persistence with encrypted backups and actionable load errors.
- Plaintext-to-vault migration.
- Git/chezmoi safety checks in `shelf doctor`.
- Direct export in shell/env/JSON formats.
- Project manifests and project export.
- `shelf run` runtime injection and value-free dry-run.
- Localhost-only vault manager for metadata search, explicit reveal, create/update, and delete.
