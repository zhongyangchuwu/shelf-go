# Shelf Go

Shelf Go is a local-first encrypted secret environment manager for solo developers. It stores secrets in a portable age-encrypted vault file, keeps project manifests value-free, and preserves scriptable CLI workflows for explicit secret reads and `shelf project run`.

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

## Store and inspect secrets

Create v2 secret items with templates or profiles. Secret values enter through stdin, not argv:

```bash
printf %s "$APP_TOKEN" | shelf secret create app-token --env APP_TOKEN --stdin
printf %s "$OPENAI_API_KEY" | shelf secret create openai --profile openai --stdin
```

Inspect metadata without printing secret field values:

```bash
shelf secret info openai
shelf secret list --long
```

`secret info` shows public fields such as `credential_env`, `provider`, and `website`, plus `value_set` for secret fields. It does not print secret field values.

Print a value only when you intentionally need plaintext:

```bash
shelf secret get openai
shelf secret get --env OPENAI_API_KEY
```

Update fields from stdin:

```bash
printf %s "$NEW_TOKEN" | shelf secret set openai --stdin
printf %s LLM_API_KEY | shelf secret set openai --field credential_env --stdin
```

`secret add`, `secret edit`, and direct `secret export` are removed from the v2 secret CLI. Full interactive editing is deferred to the later TUI.

Phase 38 v2 secret commands currently use the plaintext v2 JSON repository boundary and write to `<resolved-vault-path>.v2.json` unless `--vault` is supplied. Treat that file as secret data: do not commit or sync it until age encryption is wired into the v2 repository.

## Use secrets in a project

Inside a Git worktree, create a value-free project manifest:

```bash
shelf project init
shelf project status
shelf project configure
```

Shelf writes `.shelf.json` at the Git root. Version 1 manifests store legacy exact paths, prefixes, tag selectors, env overrides, and required/optional flags. Version 2 manifests store concrete v2 item/field/env bindings. Both formats are value-free. Use `shelf project configure` to queue value-free binding changes in the TUI and apply them after review. Advanced commands such as `project add`, `project rm`, and `project list` remain available for manual manifest edits.

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

The manager starts an on-demand loopback Web console for the current v1 encrypted vault. It can search, add, edit, rename, delete, reveal, copy, and tag secret records. List and detail responses stay metadata-only; values reveal only through explicit reveal/copy actions. During the v2 cutover, `shelf manager` refuses to open when the resolved v2 vault file exists; use `shelf secret` and `shelf project` commands for v2 item workflows until the v2 manager is adapted. Stop the process with Ctrl-C when finished.

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

- `shelf secret get`, `shelf project export`, `shelf project run`, and the v1 `shelf manager` can intentionally materialize plaintext values. `shelf secret info` and `shelf secret list` are value-free.
- `.shelf.json` project manifests are value-free and can be committed after review.
- Generated `.env` / `.env.local` files contain plaintext values. Do not commit them.
- Phase 38 v2 JSON vault files (`*.v2.json`) contain secret values in plaintext during the cutover. Do not commit or sync them.
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
