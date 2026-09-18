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

Selectors can be item ID or title.

## `env export --label` exports fewer fields than expected

Only fields with a non-empty `env_<field>` mapping are exported. Check the item metadata:

```bash
shelf secret info <selector>
```

Set a mapping when needed:

```bash
printf %s OPENAI_API_KEY | shelf secret update openai --field env_credential
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
- missing mapped field value;
- duplicate env name;
- binding points at a non-projectable field type (lists, dates, or JSON).

## Credential audit reports warnings or expired credentials

Run the value-free audit to inspect public lifecycle metadata:

```bash
shelf secret audit --within 30
```

Exit `3` means at least one tracked field is expired. Exit `2` means a tracked field is due within the horizon, unknown, or invalid. `permanent` and healthy values exit `0`. The audit never prints secret values and does not block `project export` or `project run`.

## Completion not installed

Install completions for your shell and make sure `FPATH`/`fpath` contains the directory that holds `_shelf`.
