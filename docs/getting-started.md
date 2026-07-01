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

## Add and tag secrets

```bash
shelf secret set app:token sk-example --env APP_TOKEN --tag app --tag local
shelf secret set providers/openai:api_key sk-openai --env OPENAI_API_KEY --tag ai --tag prod
shelf secret info app:token
```

`secret info` prints metadata and `value_set`, not the value. To intentionally print the value:

```bash
shelf secret get app:token
```

List or export by tag when a workflow needs a selected set:

```bash
shelf secret list --tag ai --tag prod
shelf secret export --tag ai --tag prod --format env
```

Repeated `--tag` selectors use AND semantics. `secret list` stays value-free; `secret export` contains plaintext values. Redirected `.env` files must not be committed.

## Use a project manifest

Inside a Git project:

```bash
shelf project init
shelf project add app:token
shelf project status
```

Shelf writes `<git-root>/.shelf.json`. The manifest stores paths, prefixes, tag selectors, env overrides, and required/optional flags; it does not store values.

Add a tag-selected project binding when a project needs every secret matching a tag set:

```bash
shelf project add --tag ai --tag prod --optional
shelf project list
shelf project status
```

Tag project bindings expand during `project status`, `project export`, and `project run`. Prefix and tag entries cannot carry `--env` because they may expand to multiple secrets.

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

The manager binds to loopback, prints a tokenized local URL, and provides a Web console for search, add, edit, rename, delete, reveal, copy, and tag workflows. List/detail responses do not include secret values; reveal/copy actions intentionally handle plaintext locally.

## Next steps

- Read the [security policy](../SECURITY.md) before syncing a vault, opening the local manager, or exporting plaintext values.
- Use the [portable vault guide](portable-vault.md) for Git, chezmoi, second-machine setup, and `.bak` recovery.
- Read [architecture](architecture.md) if you are changing package boundaries or persistence behavior.
- Use [reference](reference.md) for command, config, and manifest details.
- Use [troubleshooting](troubleshooting.md) for common setup and decrypt errors.
