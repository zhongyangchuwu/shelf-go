# Troubleshooting

## Check local health

```bash
shelf doctor
```

`doctor` checks config resolution, item-vault path, file permissions, JSON loadability, and shell completion state without printing secret values.

## Vault is missing

Run:

```bash
shelf setup
```

Or pass an explicit path:

```bash
shelf --vault ~/.local/share/shelf/vault.json secret list
```

## Vault file is plaintext

This is expected. Shelf currently relies on external encryption/sync tools such as chezmoi. Do not commit or sync the raw JSON vault unless the external tool encrypts it.

## Secret command says item not found

Use metadata-only commands to find the item:

```bash
shelf secret list --long
shelf secret info <selector>
```

Selectors can be item ID, title, or alias.

## `secret export --label` exports fewer values than expected

Items without an explicit env-name field are skipped. Check metadata:

```bash
shelf secret info <selector>
```

Set an env-name field when needed:

```bash
printf %s OPENAI_API_KEY | shelf secret update openai --field credential_env
```

## Project manifest missing

Inside a Git worktree:

```bash
shelf project init
```

## Project run/export fails

Check value-free diagnostics:

```bash
shelf project status
```

Common causes:

- missing item;
- missing field;
- missing secret value;
- duplicate env name;
- binding points at a public field instead of a secret field.

## Completion not installed

Install completions for your shell and make sure `FPATH`/`fpath` contains the directory that holds `_shelf`.
