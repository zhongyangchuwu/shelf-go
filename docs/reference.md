# Shelf Command Reference

Shelf is a CLI-first local secret manager. The configured vault is a JSON item vault and contains secret values. Use external encryption/sync tooling such as chezmoi before committing or syncing it.

## Global flags

```bash
shelf [--config PATH] [--vault PATH] <command>
```

- `--config`: config YAML path. Defaults to `~/.config/shelf/config.yaml` or `SHELF_CONFIG`.
- `--vault`: item vault JSON path. Overrides config and `SHELF_VAULT`.

## `shelf setup`

Creates config and an empty item vault.

```bash
shelf setup [--vault-path PATH] [--force]
```

## `shelf doctor`

Checks config resolution, version, item-vault path, permissions, JSON validity, and shell completion install state without printing values.

```bash
shelf doctor
```


## `shelf secret create`

Creates an item. Secret values enter through stdin when `--stdin` is set.

```bash
shelf secret create <title> [--template generic|api] [--profile PROFILE] [--id ID] [--env NAME] [--stdin]
```

On success, prints a metadata summary with title, id, template, and env name. Secret values are not echoed.

Useful public-field flags for `api_credential` are `--provider`, `--website`, `--expires`, and repeatable `--scope`.

## `shelf secret set`

Sets a field from stdin. Without `--field`, Shelf sets the template primary secret field.

```bash
shelf secret set <selector> [--field FIELD] --stdin
```

## `shelf secret get`

Prints a plaintext field value.

```bash

shelf secret get <selector> [--field FIELD]
shelf secret get --env ENV [--field FIELD]
```

## `shelf secret list`

Prints item metadata without secret values.

```bash
shelf secret list [query] [--long]
```

`--long` prints a titled, aligned table with `ID`, `TITLE`, `TEMPLATE`, `ENV`, and `VALUE` state columns.

## `shelf secret info`

Prints item metadata as JSON. Public fields are shown. Secret fields show only `value_set`.

```bash
shelf secret info <selector>
```

## `shelf secret export`

Materializes values for items that have all requested labels.

```bash
shelf secret export --label LABEL [--label LABEL ...] [--format shell|env|json]
```

Default format is `shell`:

```zsh
source <(shelf secret export --label ai --format shell)
```

Repeated labels use AND matching. Items without an explicit env-name field are skipped. Missing labels, missing secret values for exportable items, and duplicate env names fail with value-free diagnostics.

## `shelf secret rm`

Removes an item after explicit confirmation.

```bash
shelf secret rm <selector> --yes
```

## `shelf project id`

Prints the current Git project identity.

```bash
shelf project id
```

## `shelf project init`

Creates `<git-root>/.shelf.json`.

```bash
shelf project init [--force]
```

## `shelf project add`

Adds a concrete item-field binding to the project manifest.

```bash
shelf project add <item-id-or-title> [--field FIELD] [--env NAME] [--optional]
```

`--field` defaults to the item's primary secret field. `--env` defaults to that field's public env-name field.

## `shelf project rm`

Removes a manifest binding by `item.field` key.

```bash
shelf project rm <item.field>
```

## `shelf project list`

Lists manifest bindings without values.

```bash
shelf project list
```

## `shelf project status`

Checks manifest resolution without printing secret values.

```bash
shelf project status
```

## `shelf project export`

Prints resolved project values in `shell`, `env`, or `json` format.

```bash
shelf project export [--format shell|env|json]
```

## `shelf project run`

Runs a child process with manifest values injected.

```bash
shelf project run [--dry-run] -- <command> [args...]
```

`--dry-run` prints env names that would be injected, never values.
