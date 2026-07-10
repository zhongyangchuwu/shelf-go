# Portable vault guide

Shelf's configured item vault is a normal JSON file. It contains plaintext secret values. For portability, let chezmoi or another external tool encrypt the file before it is committed or synced.

## Files you can sync after review

| File | Safe to sync? | Notes |
| --- | --- | --- |
| `<git-root>/.shelf.json` | Yes | Value-free project manifest with item/field/env bindings. |
| `~/.config/shelf/config.yaml` | Usually | Contains local paths only; machine-specific paths may need adjustment. |

## Files you must protect

| File | Why |
| --- | --- |
| `~/.local/share/shelf/vault.json` or configured vault path | Contains plaintext secret values. Encrypt with chezmoi before committing or syncing. |
| `.env`, `.env.local`, generated env files | Contain plaintext exported values. |
| Shell logs/history | May contain values if commands are misused. |

## First machine

```bash
shelf setup
printf %s "$APP_TOKEN" | shelf secret create app-token --env APP_TOKEN --stdin
shelf doctor
```

If using chezmoi, add the configured vault path through chezmoi's encrypted-file workflow, not as a raw plaintext file.

## Second machine

Restore/decrypt the vault through your external tool, then point Shelf at it:

```bash
shelf setup --vault-path ~/.local/share/shelf/vault.json
shelf doctor
```

## Current-shell workflow

```bash
shelf project export > .env.local
source .env.local
```

Add `.env.local` to `.gitignore` and delete it when it is no longer needed.
