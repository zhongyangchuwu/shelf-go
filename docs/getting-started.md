# Getting started

This guide creates an encrypted vault, writes one secret, and injects it into a project command.

## Install or run locally

From a checkout:

```bash
go run ./cmd/shelf --help
```

Build a binary:

```bash
go build -o ./bin/shelf ./cmd/shelf
```

Install with Go:

```bash
go install ./cmd/shelf
```

## Create the vault

Run interactive setup:

```bash
shelf setup
```

Or provide the important paths up front:

```bash
shelf setup \
  --vault-path ~/.local/share/shelf/vault.age \
  --identity ~/.config/shelf/identity.txt
```

If no recipient is supplied, Shelf generates or reuses the configured age identity and stores its public recipient in config. The private identity file is sensitive and must not be committed.

Check the result:

```bash
shelf vault status
```

Expected output reports the config path, vault path, configured recipient count, vault format, and loadability without printing secret values.

## Add secrets

```bash
printf %s "$APP_TOKEN" | shelf secret create app-token --env APP_TOKEN --stdin
printf %s "$OPENAI_API_KEY" | shelf secret create openai --profile openai --stdin
shelf secret info openai
```

`secret info` prints public fields and `value_set`, not secret field values. To intentionally print the primary secret field value:

```bash
shelf secret get openai
shelf secret get --env OPENAI_API_KEY
```

Update fields through stdin:

```bash
printf %s "$NEW_TOKEN" | shelf secret set openai --stdin
printf %s LLM_API_KEY | shelf secret set openai --field credential_env --stdin
```

The first v2 secret CLI is intentionally non-interactive. TUI-based creation and editing will handle richer UX later.

## Use a project manifest

Inside a Git project:

```bash
shelf project init
shelf project status
shelf project configure
```

Shelf writes `<git-root>/.shelf.json`. Version 1 manifests store legacy paths, prefixes, tag selectors, env overrides, and required/optional flags. Version 2 manifests store concrete item/field/env bindings. Neither format stores secret values.

When `project status` reports env-file variables that need bindings, run the interactive configurator:

```bash
shelf project configure
```

Project configuration queues value-free manifest changes and applies them only from the review screen. In v1 manifests, tag project bindings expand during `project status`, `project export`, and `project run`; prefix and tag entries cannot carry `--env` because they may expand to multiple secrets. In v2 manifests, bindings are concrete item fields.

Export sourceable shell lines when you want to update your current shell manually:

```bash
shelf project export > .env.local
source .env.local
```

The generated file contains plaintext values. Add it to `.gitignore` and delete it when it is no longer needed.

Inject secrets into a child process:

```bash
shelf project run -- sh -c 'printf "%s\n" "$APP_TOKEN"'
```

Preview without values:

```bash
shelf project run --dry-run -- sh -c 'printf "%s\n" "$APP_TOKEN"'
```

`project run` modifies only the child process environment. It cannot change the parent shell.

## Migrate an old plaintext store

```bash
shelf vault migrate \
  --from ~/.local/share/shelf/secrets.json \
  --to ~/.local/share/shelf/vault.age

shelf vault status
```

After verifying the encrypted vault, move, delete, or securely archive the old plaintext source.

## Recover from the last-write backup

When replacing an existing vault, Shelf keeps one encrypted last-write backup next to it as `vault.age.bak`. To recover manually:

```bash
cp ~/.local/share/shelf/vault.age ~/.local/share/shelf/vault.age.bad
cp ~/.local/share/shelf/vault.age.bak ~/.local/share/shelf/vault.age
shelf vault status
```

The backup is a normal encrypted Shelf vault file. Shelf overwrites this single `.bak` on each later vault replacement, so it is not a history system.

## Open the local manager

```bash
shelf manager
```

The manager binds to loopback, prints a tokenized local URL, and provides a Web console for v1 vault search, add, edit, rename, delete, reveal, copy, and tag workflows. List/detail responses do not include secret values; reveal/copy actions intentionally handle plaintext locally. During the v2 cutover, the manager is gated when the resolved v2 vault exists; use CLI v2 item commands until the v2 manager is adapted.

## Next steps

- Read the [security policy](../SECURITY.md) before syncing a vault, opening the local manager, or exporting plaintext values.
- Use the [portable vault guide](portable-vault.md) for Git, chezmoi, second-machine setup, and `.bak` recovery.
- Read [architecture](architecture.md) if you are changing package boundaries or persistence behavior.
- Use [reference](reference.md) for command, config, and manifest details.
- Use [troubleshooting](troubleshooting.md) for common setup and decrypt errors.
