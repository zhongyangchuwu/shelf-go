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

## `shelf tui`

Opens the optional, keyboard-first secret workspace. It requires interactive stdin and stdout plus a terminal at least 100 columns × 24 rows; a smaller terminal shows a resize-required screen.

```bash
shelf tui
```

The catalog, search, filters, detail, help, and confirmation screens are value-free. `a`, `e`, `c`, and `d` create, update, copy, and delete through the existing app and clipboard workflows. Secret-field form input is masked; the workspace has no reveal action. `secret get`, exports, and project execution remain CLI commands.



## `shelf secret create`

Creates an item. Secret values enter through stdin when `--stdin` is set.

```bash
shelf secret create <title> [--template TEMPLATE] [--profile PROFILE] [--id ID] [--env NAME] [--validity VALUE] [--stdin]
```

On success, prints a metadata summary with title, id, template, and primary-field env name. Secret values are not echoed.

Useful public-field flags for `api_credential` are `--provider`, `--website`, `--validity`, and repeatable `--scope`. `--validity` accepts `unknown`, `permanent`, or `YYYY-MM-DD`.

`--template` accepts a vault template ID or alias; built-in aliases are `generic` and `api`.

## `shelf secret update`

Updates an item's title, labels, aliases, or a selected field without editing vault JSON. Without metadata flags or `--field`, it updates the template primary secret field.

```bash
shelf secret update <selector> [--title TITLE] [--add-label LABEL] [--remove-label LABEL] [--add-alias ALIAS] [--remove-alias ALIAS] [--field FIELD] [--value VALUE ... | --clear-field]
```

Use repeatable `--value` only for a public field; it supports string-list fields. Without `--value`, Shelf reads the selected field from stdin. An interactive terminal displays a corresponding no-echo prompt. Secret values never enter argv. `--clear-field` clears a public field.

## `shelf secret get`

Prints a plaintext field value.

```bash

shelf secret get <selector> [--field FIELD]
shelf secret get --env ENV [--field FIELD]
```

## `shelf secret copy`

Copies a plaintext field value to the system clipboard without printing it. This is the interactive alternative to `secret get`; it does not replace stdout-oriented scripting workflows.

```bash
shelf secret copy <selector> [--field FIELD]
shelf secret copy --env ENV [--field FIELD]
```

Clipboard initialization or write failures return an error and do not fall back to printing the value. On WSL, Shelf first uses the Windows clipboard bridge (`clip.exe`) and falls back to `wl-copy` when the bridge fails. On Linux outside WSL, the clipboard backend requires `xclip` or `xsel`.

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

## `shelf env export`

Materializes all explicitly mapped fields for items that have all requested labels.

```bash
shelf env export --label LABEL [--label LABEL ...] [--format shell|json]
```

Default format is `shell`:

```zsh
source <(shelf env export --label ai --format shell)
```

Repeated labels use AND matching. `env_<field>` maps an eligible `string`, `concealed`, or `url` field to an environment name; mappings project both secret and public values. Unmapped fields are not exported. Missing labels, missing mapped-field values, and duplicate env names fail with value-free diagnostics.

## `shelf secret audit`

Reports lifecycle validity from public `validity_<secret-field>` metadata without printing secret values.

```bash
shelf secret audit [--within DAYS] [--format text|json]
```

The default horizon is 30 UTC calendar days. `expired` means the date is before today; `due` covers today through the horizon. Tracked fields can also be `unknown`, `permanent`, `invalid`, or `healthy`; untracked fields are omitted. Exit status is `3` for expired credentials, `2` for due/unknown/invalid fields, and `0` for healthy/permanent or no tracked fields.

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

## `shelf project tui`

Opens the optional, keyboard-first binding workspace for the current Git worktree. It requires interactive stdin and stdout plus a terminal at least 100 columns × 24 rows.

```bash
shelf project tui
```

The workspace lists value-free bindings and their ready/warn/fail status, including safe missing-value and duplicate-env diagnostics. `a` selects a projectable field and env name, `d` removes a binding after confirmation, and `i` initializes a missing manifest after confirmation. It never reads or writes `.shelf.json` directly, reveals values, exports, or runs child processes.

## `shelf project add`

Adds a concrete item-field binding to the project manifest.

```bash
shelf project add <item-id-or-title> [--field FIELD] [--env NAME] [--optional]
```

`--field` defaults to the item's primary secret field. Any stored public or secret `string`, `concealed`, or `url` field is eligible. `--env` uses its `env_<field>` mapping when one exists; otherwise it is required.

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

Checks manifest resolution and applicable credential health without printing secret values. Health uses only public lifecycle metadata; it does not block export or run.

```bash
shelf project status
```

Binding rows include `health:<status>` and a public `expires:<YYYY-MM-DD>` value when the bound field has date-based validity metadata.

## `shelf project export`

Prints resolved project values in `shell` or `json` format.

```bash
shelf project export [--format shell|json]
```


## `shelf project run`

Runs a child process with manifest values injected.

```bash
shelf project run [--dry-run] -- <command> [args...]
```

`--dry-run` prints env names that would be injected, never values.

## `shelf completion`

Generates shell completion for supported shells.

```bash
shelf completion [bash|zsh|fish|powershell]
```

Example zsh install:

```bash
mkdir -p ~/.zfunc
shelf completion zsh > ~/.zfunc/_shelf
```

Ensure `~/.zfunc` is present in `fpath`; `shelf doctor` reports completion install status.
