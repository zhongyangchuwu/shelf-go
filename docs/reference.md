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

The catalog, search, filters, detail, help, and confirmation screens are value-free. `a`, `e`, `c`, and `d` create, update, copy, and delete through the existing app and clipboard workflows. Create and Edit are single scrollable forms; the create form distinguishes item kind from optional provider presets, and secrets are masked without a reveal action. `secret get`, exports, and project execution remain CLI commands.

In the catalog workspace, `c` or `Ctrl+C` copies the selected field (or the primary secret when no field is selected); `q` quits. In form insert mode and search, `Ctrl+V` reads the system clipboard, and terminal paste shortcuts such as Windows Terminal's `Ctrl+Shift+V` also work. A typed `q` is text in these inputs. Clipboard failures preserve the active input and never display clipboard contents. Search matches all primary and secondary environment names case-insensitively, in addition to the other public catalog metadata.

The secret editor keeps drafts only in memory. Validation or save failures retain every input, including masked secrets, so `Ctrl+S` can retry after the problem is corrected. Saving blocks further editing until the result arrives; resizing the terminal preserves the editor. Successful saves and explicit discards clear editor-held inputs, but do not promise secure erasure of Go string memory.

In insert mode, `Esc` only returns to normal mode. In normal mode, `Esc` requests cancellation and `q` or `Ctrl+C` requests quitting. If the form has unsaved changes, `y` explicitly discards them; `n` or `Esc` keeps editing, and `Enter` does nothing. Unchanged or reverted forms exit without confirmation. Closing or forcibly terminating the process still loses an unsaved draft.

## `shelf secret set`

Sets a value by environment variable name without choosing a template or item title first.

```bash
shelf secret set <ENV> [--target ITEM.FIELD]
```

With interactive stdin, Shelf prompts without echoing the value. Otherwise it reads stdin directly; no `--stdin` flag is needed. Values are never accepted through argv or a `--value` flag. Success prints only a value-free confirmation.

- One existing mapping: updates only that mapped field, preserving the item's ID, title, template, labels, description, and other fields.
- No existing mapping: creates a generic single-secret item titled `<ENV>`, with the normal derived item ID and environment mapping. ID/title collisions fail without adopting or overwriting an unrelated item; use `secret create` with an explicit distinct title/ID in that case.
- Multiple mappings: interactive use offers a numbered target picker before reading the value; a blank choice cancels. Noninteractive use requires `--target ITEM.FIELD`, using an exact stable item ID and an eligible field mapped to that ENV. Titles are not accepted as targets.

```bash
printf %s "$OPENAI_API_KEY" | shelf secret set OPENAI_API_KEY --target item_openai.credential
```

Existing field types and sensitivity are retained: setting a public mapped field does not make it secret. The workflow rechecks the target after reading input and refuses stale mappings or schema changes. Invalid input and failed saves leave the vault unchanged and do not print collected values.

## `shelf secret create`

Creates an item. Secret values enter through stdin when `--stdin` is set.

```bash
shelf secret create <title> [--template TEMPLATE] [--profile PROFILE] [--id ID] [--description TEXT] [--env NAME] [--validity VALUE] [--stdin]
```

On success, prints a metadata summary with title, id, template, and primary-field env name. Secret values are not echoed.

`--description` is public metadata. It is returned by `secret info` and shown by the TUI; never store a secret value there.

Useful public-field flags for `api_credential` are `--provider`, `--website`, `--validity`, and repeatable `--scope`. `--validity` accepts `unknown`, `permanent`, or `YYYY-MM-DD`.

`--template` accepts a vault template ID or alias; built-in aliases are `generic` and `api`.

## `shelf secret edit`

Explicitly opens one existing item's **plaintext values** in an external editor. Use a trusted editor and account for its plugins, backups, swap files, histories, and synchronization. This is a CLI-only exception to stdin-based writes; it does not add a reveal action to the TUI.

```bash
shelf secret edit <selector>
shelf secret edit --env ENV [--target ITEM.FIELD]
```

Selector and `--env` modes are mutually exclusive. Ambiguous environment mappings use the same numbered picker as `secret set`; noninteractive use requires an exact mapped item-ID.field target. The selected field chooses its owning item; the editor opens that whole item. Missing or invalid selections do not create drafts. Selector, env-name and target completion remain value-free.

The first nonblank `VISUAL`, then `EDITOR`, then `vi` chooses the executable and arguments. Single/double quotes and backslash escaping are supported without shell evaluation, variable expansion, substitution, or redirection. The absolute draft filename is appended as one argument. Configure a GUI editor to wait for editing to finish; a launcher that returns immediately cannot communicate when saving is complete.

Shelf creates `item.json` under the OS temporary directory, closes its writer, then starts the editor with the command's terminal I/O. On Unix/WSL the draft directory and file are restricted to `0o700` and `0o600`. On native Windows, use a user-private `TEMP`/`TMP` directory: Unix permission bits do not configure Windows ACLs. Shelf never opens `vault.json` directly in the editor. On WSL, prefer a Linux-side editor and a Linux temporary directory; custom temporary locations should not be in a project or synchronized folder.

The editable JSON has these properties:

- `id` and `template` are readonly context. Timestamps, field types and sensitivity cannot be changed through the document.
- `title` and `description` are strings; `labels` is a string array. Omitted properties preserve their exact original values. Changed organization follows the existing normalization rules; `""` clears description and `[]` clears labels.
- `fields` maps canonical schema field IDs to actual typed JSON values, including secrets. Set fields are prefilled; unset fields are omitted. Omitted entries preserve values, and schema-defined optional fields may be added. For JSON fields, literal `null` is a value, not a deletion instruction.
- `clear_fields` is an explicit array of public field IDs to clear. To clear `website`, remove its entry from `fields` and add `"website"` to `clear_fields`. A field cannot be both set and cleared. Secret fields cannot be cleared this way.

Unknown properties/fields, duplicate object keys, trailing JSON, invalid types and invalid mappings are rejected without echoing the document. Unchanged values, JSON formatting or key-order changes do not rewrite storage. Shelf reloads the pinned vault path before saving, rejects changes to the selected item or owning schema, and preserves unrelated concurrent changes. This is optimistic pre-save checking, not a global writer lock.

After a validation failure, an interactive terminal offers `e`/`edit` to reopen the same draft or `d`/`discard` to leave storage unchanged. Save failures additionally offer `r`/`retry` to save that draft without reopening. Blank input or EOF cancels with an error; noninteractive failures return immediately. Editor launch/nonzero-exit failures and conflicts abort without saving, and conflicts never automatically rebase or force an overwrite.

Normal returns remove Shelf's private draft directory, including files the editor created inside it. Unix termination manages the attached editor process group; native Windows uses a Job Object. If termination cannot be confirmed within bounded waits, Shelf reports and retains the private draft directory instead of deleting a potentially active file. Cleanup failures also report the directory for manual recovery/removal. Forced termination, crashes, detached processes, existing GUI editor services, and editor files/history outside the directory cannot be guaranteed away. Cleanup removes files; it does not promise secure erasure. Shelf's own status and diagnostics do not echo document values, but the explicitly launched editor can display them.

## `shelf secret update`

Updates an item's title, description, labels, or a selected field without editing vault JSON. Without metadata flags or `--field`, it updates the template primary secret field.

```bash
shelf secret update <selector> [--title TITLE] [--description TEXT] [--add-label LABEL] [--remove-label LABEL] [--field FIELD] [--value VALUE ... | --clear-field]
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

Clipboard initialization or write failures return an error and do not fall back to printing the value. On WSL, Shelf writes Unicode text through `clip.exe`, falling back to `wl-copy`; TUI clipboard reads use Windows PowerShell, falling back to `wl-paste`. The Windows executables can be found at their standard `/mnt/c/Windows/System32` paths even when absent from `PATH`. On Linux outside WSL, the native clipboard backend requires `xclip` or `xsel`.

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

Generates shell completion for supported shells. Dynamic candidates include unambiguous secret IDs/titles, mapped `--env` names, `--field` names for the selected item, labels, and compatible template/profile IDs and aliases. Project field completion excludes metadata fields that cannot be bound. Completion reads one vault snapshot, never prints secret values, and does not create missing config or vault files.

`secret set` also completes ambiguous environment names because they can be selected explicitly; its `--target` candidates are limited to concrete fields mapped to the chosen ENV.

```bash
shelf completion [bash|zsh|fish|powershell]
```

Example zsh install:

```bash
mkdir -p ~/.zfunc
shelf completion zsh > ~/.zfunc/_shelf
```

Add this to `~/.zshrc`, placing the `fpath` assignment before any existing `compinit` call:

```zsh
fpath=(~/.zfunc $fpath)
autoload -Uz compinit
compinit
```

For Bash, add this to `~/.bashrc`:

```bash
source <(shelf completion bash)
```

Open a new shell after installing completion. After upgrading Shelf, regenerate the zsh completion file and restart the shell. `shelf doctor` reports completion install status.
