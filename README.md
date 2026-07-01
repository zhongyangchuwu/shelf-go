# Shelf Go

Shelf Go is a local-first encrypted secret environment manager for solo developers. It stores secrets in a portable age-encrypted vault file, keeps project manifests value-free, and preserves scriptable CLI workflows for direct secret export and `shelf project run`.

Shelf is pre-release software. Planning, roadmap, and phase records live in `.planning/`; public documentation stays focused on current user-facing behavior.

## Why Shelf

Developer secrets often end up split across `.env` files, shell profiles, password managers, and hosted secret tools. Shelf is for the solo developer case where you want:

- **Encrypted by default:** the durable source of truth is an age-encrypted vault file.
- **Portable:** the vault is a normal file that can be backed up or managed by Git/chezmoi.
- **Project-aware:** projects declare secret paths in `.shelf.json` without storing values.
- **CLI-first:** commands remain predictable for shell scripts, export flows, and child-process injection.
- **Local-only:** no hosted backend, account, or permanent daemon is required.

## Install

From source:

```bash
go install github.com/zhongyangchuwu/shelf-go/cmd/shelf@latest
```

From a checkout:

```bash
go build -o ./bin/shelf ./cmd/shelf
```
GitHub release binaries for Linux, macOS, and Windows are published with checksums starting in `v0.1.0`.

## Initialize an encrypted vault

Run setup once on a machine:

```bash
shelf setup
shelf vault status
```

Shelf creates or reuses an age identity, writes config under `~/.config/shelf/`, and stores encrypted vault data at `~/.local/share/shelf/vault.age` by default.

Use explicit paths when you want chezmoi- or backup-friendly layout control:

```bash
shelf setup \
  --vault-path ~/.local/share/shelf/vault.age \
  --identity ~/.config/shelf/identity.txt
```

The private age identity decrypts the vault. Do not commit it.

## Store, tag, and inspect secrets

Add secrets with optional environment names, descriptions, and tags:

```bash
shelf secret set app:token sk-example --env APP_TOKEN --tag app --tag local
shelf secret set providers/openai:api_key sk-openai --env OPENAI_API_KEY --tag ai --tag prod
```

Inspect metadata without printing values:

```bash
shelf secret info app:token
shelf secret list app
shelf secret list --tag ai --tag prod
```

Repeated `--tag` selectors use AND semantics. `secret list` prints paths only; it does not print secret values.

Print a value only when you intentionally need plaintext:

```bash
shelf secret get app:token
```

Export a single secret, prefix, or tag-selected set for scripts:

```bash
shelf secret export app:token --format shell
shelf secret export app --format env --all
shelf secret export --tag ai --tag prod --format env
```

Export output contains plaintext. Do not commit redirected `.env` files. Use `shelf manager` for full-object editing when you need to rename a path or change env, description, tags, or values comfortably.

## Use secrets in a project

Inside a Git worktree, create a value-free project manifest:

```bash
shelf project init
shelf project add app:token
shelf project add --tag ai --tag prod --optional
shelf project list
shelf project status
```

Shelf writes `.shelf.json` at the Git root. It stores exact paths, prefixes, tag selectors, env overrides, and required/optional flags, but never secret values. Tag selectors expand at export/run time and use the same AND semantics as direct secret commands.

Export sourceable shell lines when you explicitly want a file or current-shell workflow:

```bash
shelf project export > .env.local
source .env.local
```

Run a child command with project secrets injected without changing your parent shell:

```bash
shelf project run -- npm run dev
```

Preview injected env names without values:

```bash
shelf project run --dry-run -- npm run dev
```

## Open the local manager

```bash
shelf manager
```

The manager starts an on-demand loopback Web console. It can search, add, edit, rename, delete, reveal, copy, and tag secret records. List and detail responses stay metadata-only; values reveal only through explicit reveal/copy actions. Stop the process with Ctrl-C when finished.

## Portability and recovery

The encrypted vault file can be copied or synced after verification:

```bash
shelf vault status
shelf doctor
```

When Shelf replaces an existing vault it keeps one encrypted last-write backup at `<vault>.bak`. Recover manually by copying the backup over the active vault, then verifying with `shelf vault status`.

Legacy plaintext stores can be migrated:

```bash
shelf vault migrate --from ~/.local/share/shelf/secrets.json --to ~/.local/share/shelf/vault.age
shelf vault status
```

After verifying the encrypted vault, move, delete, or securely archive the plaintext source.

## Documentation

- [Getting started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [Security policy](SECURITY.md)
- [Portable vault guide](docs/portable-vault.md)
- [Reference](docs/reference.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Contributing](docs/contributing.md)
- [Changelog](CHANGELOG.md)

## Safety notes

- `shelf secret get`, `shelf secret export`, `shelf project export`, `shelf project run`, `shelf secret edit`, and `shelf manager` can intentionally materialize plaintext values.
- `.shelf.json` project manifests are value-free and can be committed after review.
- Generated `.env` / `.env.local` files contain plaintext values. Do not commit them.
- `config.yaml` may contain public age recipients and identity file paths. It must not contain private identity contents or secret values.
- If all matching private age identities are lost, Shelf cannot recover the encrypted vault or its `.bak` backup.

## Development

```bash
./scripts/test.sh
go build -o ./bin/shelf ./cmd/shelf
```

Reusable install/release workflows live under `scripts/`; `justfile` keeps one local verification entrypoint plus thin install/tag wrappers:

```bash
./scripts/install.sh
./scripts/release.sh check
./scripts/release.sh snapshot
./scripts/release.sh tag 0.1.1
```
