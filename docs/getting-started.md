# Getting started

## 1. Set up Shelf

```bash
shelf setup
shelf doctor
```

Default files:

```text
~/.config/shelf/config.yaml
~/.local/share/shelf/vault.json
```

The vault JSON contains plaintext secret values. Encrypt it with chezmoi or another external tool before committing or syncing.

## 2. Create a secret

```bash
printf %s "$OPENAI_API_KEY" | shelf secret create openai --profile openai --stdin
shelf secret list --long
shelf secret info openai
```

## 3. Read or update a secret intentionally

```bash
shelf secret get openai
printf %s "$NEW_OPENAI_API_KEY" | shelf secret update openai
```

## 4. Export a label group for shell startup

```zsh
source <(shelf env export --label ai --format shell)
```

This materializes each explicitly mapped field at shell startup but keeps values out of `.zshrc`.

## 5. Bind a project

Inside a Git worktree:

```bash
shelf project init
shelf project add openai --env OPENAI_API_KEY
shelf project status
```

Run a child process with injected values:

```bash
shelf project run -- npm run dev
```

Preview env names without values:

```bash
shelf project run --dry-run -- npm run dev
```

## 6. Audit credential lifecycle validity

```bash
shelf secret audit --within 30
```

The audit is value-free: exit `3` means expired, `2` means due, unknown, or invalid validity, and `0` means healthy, permanent, or no tracked fields.
