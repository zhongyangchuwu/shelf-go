# Shelf Go

Shelf Go is a local-first encrypted secret environment manager for solo developers. It stores secrets in a portable age-encrypted vault file, keeps project manifests value-free, and preserves scriptable CLI workflows for direct secret export and `shelf project run`.

Shelf is pre-release software. Planning, roadmap, and phase records live in `.planning/`; public documentation stays focused on current user-facing behavior.

## Why Shelf

- **Encrypted by default:** the durable source of truth is an age-encrypted vault file.
- **Portable:** the vault is a normal file that can be backed up or managed by Git/chezmoi.
- **Project-aware:** projects declare secret paths in `.shelf.json` without storing values.
- **CLI-first:** commands remain predictable for shell scripts, export flows, and child-process injection.
- **Local-only:** no hosted backend, account, or permanent daemon is required.

## Quick start

```bash
shelf setup
shelf vault status

shelf secret set app:token sk-example --env APP_TOKEN
shelf secret get app:token
```

In a Git project:

```bash
shelf project init
shelf project add app:token
shelf project export > .env.local
source .env.local
shelf project run -- sh -c 'printf "%s\n" "$APP_TOKEN"'
```

For a fuller walkthrough, see [Getting started](docs/getting-started.md).

## Core commands

```bash
shelf setup [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
shelf vault init [--vault-path PATH] [--recipient AGE_RECIPIENT] [--identity PATH] [--force]
shelf vault migrate --from <plaintext.json> [--to <vault.age>] [--force]
shelf vault restore --from <backup.age> [--to <vault.age>] [--force]
shelf vault status|check
shelf vault open [--addr 127.0.0.1:0]
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
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project explain
shelf project export [--format shell|env|json]
shelf project run [--dry-run] -- command args...

shelf completion [bash|zsh|fish|powershell]
```

Global flags:

```bash
--config PATH   Path to config.yaml
--vault PATH    Path to encrypted vault
```

## Documentation

- [Getting started](docs/getting-started.md)
- [Security model](docs/security.md)
- [Reference](docs/reference.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Contributing](docs/contributing.md)

## Safety notes

- `shelf secret get`, `shelf secret export`, `shelf project export`, `shelf project run`, `shelf secret edit`, and `shelf vault open` can intentionally materialize plaintext values.
- `.shelf.json` project manifests are value-free and can be committed after review.
- Generated `.env` / `.env.local` files contain plaintext values. Do not commit them.
- `config.yaml` may contain public age recipients and identity file paths. It must not contain private identity contents or secret values.
- `shelf vault migrate` preserves the old plaintext JSON source after successful encrypted migration; delete, move, or securely archive it after verifying the new vault.
- `shelf vault restore` only restores encrypted Shelf vault files and validates them before replacing the target. It requires a configured identity that can decrypt the backup.

## Development

```bash
go test ./...
go build -o ./bin/shelf ./cmd/shelf
```
