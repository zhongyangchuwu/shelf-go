# Troubleshooting

## `vault recipients` is missing

`vault status` or `doctor` can fail when no age recipients are configured.

Fix by updating config through vault init:

```bash
shelf vault init --force --recipient AGE_RECIPIENT --identity ~/.config/shelf/identity.txt
```

If no recipient is supplied during setup/init, Shelf can generate or reuse an X25519 identity and derive its recipient.

## Identity path is missing or unreadable

Symptoms mention `identity_paths`, `read age identity`, or file permissions.

Check config:

```bash
cat ~/.config/shelf/config.yaml
```

Then fix the path or regenerate setup:

```bash
shelf vault init --force --identity ~/.config/shelf/identity.txt
shelf vault status
```

Do not commit the identity file.

## Vault cannot decrypt

Common causes:

- configured identity does not match the vault recipient;
- identity file was moved or replaced;
- vault file is corrupt;
- config points at the wrong vault.

Inspect without revealing values:

```bash
shelf vault status
shelf doctor
```

Restore the matching identity or recover manually from the last-write encrypted backup:

```bash
cp ~/.local/share/shelf/vault.age ~/.local/share/shelf/vault.age.bad
cp ~/.local/share/shelf/vault.age.bak ~/.local/share/shelf/vault.age
shelf vault status
```

The `.bak` file is a single last-write backup and is overwritten on each later vault replacement. If the identity that can decrypt both the active vault and backup is lost, Shelf cannot recover the encrypted contents. Keep private age identities backed up outside Shelf.

## Plaintext JSON store detected

Shelf treats plaintext stores as unsafe for the active vault path.

Migrate to an encrypted vault:

```bash
shelf vault migrate --from ~/.local/share/shelf/secrets.json --to ~/.local/share/shelf/vault.age
shelf vault status
```

After verification, move, delete, or securely archive the plaintext source.

## `.shelf.json not found`

Project commands need a Git worktree and a manifest at the Git root.

```bash
shelf project init
shelf project add app:token
shelf project explain
```

## Required secret is missing

`project explain`, `project export`, and `project run` fail when a required manifest entry is absent from the vault.

Add the secret or mark the entry optional:

```bash
shelf secret set app:token sk-example --env APP_TOKEN
shelf project add app:token --optional
```

## Env name conflict

Two project entries resolved to the same env name. Resolution order is project `env`, then secret `env`, then derived name.

Fix by using explicit env overrides on exact path entries:

```bash
shelf project rm app:token
shelf project add app:token --env APP_TOKEN
```

Prefix entries cannot use `--env` because they may expand to multiple secrets.

## `project run` did not change my shell

This is expected. A child CLI process cannot mutate the parent shell environment.

Use `project run` to execute a command with injected values:

```bash
shelf project run -- npm run dev
```

Use `project export` when you explicitly want plaintext env output:

```bash
shelf project export --format shell
```

## Vault manager address rejected

`shelf vault open` only accepts loopback addresses.

Use the default random local port:

```bash
shelf vault open
```

Or provide an explicit loopback address:

```bash
shelf vault open --addr 127.0.0.1:8080
```

The printed URL contains a local access token and can remain in browser history. Treat it as sensitive local plaintext and stop the manager with Ctrl-C when finished.

## Shell completion is not installed

Generate completion for your shell:

```bash
shelf completion bash
shelf completion zsh
shelf completion fish
shelf completion powershell
```

The `just install` task installs zsh completion to `~/.zfunc/_shelf`.
