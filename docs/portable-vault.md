# Portable vault guide

Shelf's durable source of truth is an age-encrypted vault file. That file is designed to be copied, backed up, or managed by dotfile tools such as chezmoi.

## Files you can sync

These files are value-free or encrypted and can usually be reviewed and synced:

| File | Safe to sync? | Notes |
| --- | --- | --- |
| `~/.local/share/shelf/vault.age` | Yes | Encrypted vault source of truth. Verify with `shelf vault status` before syncing. |
| `~/.local/share/shelf/vault.age.bak` | Yes | Single last-write encrypted backup. Optional; not a history system. |
| `<git-root>/.shelf.json` | Yes | Project manifest with legacy paths or v2 item/field/env bindings only; review before committing. |
| `~/.config/shelf/config.yaml` | Usually | May contain public recipients and local identity paths. Machine-specific paths may not be portable. |

## Files you must not sync

Do not commit or sync these through Git or chezmoi:

| File | Why |
| --- | --- |
| Private age identity files | They decrypt the vault. Their paths may appear in config, but the key material is secret. |
| `.env`, `.env.local`, generated env files | They contain plaintext exported values. |
| Phase 38 v2 JSON vault files (`*.v2.json`) | They contain plaintext secret values until v2 age wrapping is wired in. |
| TUI/editor temp files | Future interactive editing may handle plaintext during editing; cleanup must be best-effort on normal command exit. |

## First machine

Create the vault and check it before syncing:

```bash
shelf setup
shelf vault status
shelf doctor
```

Add secrets and project bindings:

```bash
printf %s "$APP_TOKEN" | shelf secret create app-token --env APP_TOKEN --stdin
shelf project init
shelf project configure
```

Commit or sync only the encrypted vault and value-free project manifest after review.

## Second machine

Copy or provision the private age identity outside Shelf, then point config at the synced vault and identity path:

```bash
shelf vault status
shelf doctor
```

If the identity can decrypt the vault, `vault status` reports the vault as loadable without printing secret values.

## Manual current-shell workflow

To export project bindings into your current shell explicitly:

```bash
shelf project export > .env.local
source .env.local
```

Add `.env.local` to `.gitignore` and delete it when it is no longer needed.

## Last-write backup recovery

When Shelf replaces an existing vault, it copies the previous encrypted vault to `<vault>.bak` first. This is a single last-write backup; every later replacement overwrites it.

Recover manually:

```bash
cp ~/.local/share/shelf/vault.age ~/.local/share/shelf/vault.age.bad
cp ~/.local/share/shelf/vault.age.bak ~/.local/share/shelf/vault.age
shelf vault status
```

If the matching private age identity is lost, Shelf cannot recover the vault or backup.

## Git and merge notes

Encrypted vault files do not merge meaningfully as text. Avoid editing the same synced vault independently on multiple machines before syncing. If a conflict happens, keep both copies, inspect them with the matching identity, and choose the correct encrypted vault file explicitly.
