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

## Add a secret

```bash
shelf secret set app:token sk-example --env APP_TOKEN
shelf secret info app:token
```

`secret info` prints metadata and `value_set`, not the value. To intentionally print the value:

```bash
shelf secret get app:token
```

## Export directly

```bash
shelf secret export app:token --format shell
shelf secret export app:token --format env
shelf secret export app:token --format json
```

Export output contains plaintext values. Redirected `.env` files must not be committed.

## Use a project manifest

Inside a Git project:

```bash
shelf project init
shelf project add app:token
shelf project explain
```

Shelf writes `<git-root>/.shelf.json`. The manifest stores paths, prefixes, env overrides, and required/optional flags; it does not store values.

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

## Open the local manager

```bash
shelf vault open
```

The manager binds to loopback, prints a tokenized local URL, and supports metadata browsing plus explicit reveal and write actions. Browser reveal and edit actions handle plaintext locally.

## Next steps

- Read the [security model](security.md) before syncing a vault with Git or chezmoi.
- Use [reference](reference.md) for command, config, and manifest details.
- Use [troubleshooting](troubleshooting.md) for common setup and decrypt errors.
