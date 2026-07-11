# Architecture

Shelf Go is now a local item-vault CLI. Built-in age encryption, the v1 encrypted vault stack, Web manager, and TUI project configure flow are removed from the active product. External tools such as chezmoi own encryption/sync.

## Package layers

```text
cmd/shelf/                    process entry point

internal/cli/                 Cobra command tree, flags, completions, output routing, and no-echo terminal field prompts
internal/cliui/               Lip Gloss-backed human output helpers for titles, summaries, empty states, and tables
internal/clipboard/           System clipboard adapter; uses `clip.exe`, then `wl-copy`, under WSL and github.com/atotto/clipboard elsewhere
internal/app/                 application workflows for secret, project, setup, doctor, and run

internal/vaultmodel/          item/field/template model, validation, template-role lookups, and shared item/field/env resolver
internal/vaultjson/           strict JSON codec/repository for the local item vault
internal/project/             Git project identity, .shelf.json schema/IO/validation, and binding resolution
internal/config/              config and vault path resolution
internal/util/                atomic write and env/shell/JSON binding formatting
```

Dependency direction is enforced by `.go-arch-lint.yml`:

```text
cmd/shelf -> cli
cli -> app, cliui, clipboard, Cobra, x/term
cliui -> lipgloss
clipboard -> github.com/atotto/clipboard
app -> config, project, util, vaultmodel, vaultjson
project -> util, vaultmodel
vaultjson -> vaultmodel, util
config -> yaml
```

## Item vault

`internal/vaultmodel` owns domain semantics: vault schema version, items, fields, templates, profiles, public env-name fields, secret sensitivity, validation, indexes, and shared resolution of item selectors, fields, and env names.

`internal/vaultjson` owns persistence mechanics for the configured JSON file: strict decode, validated encode/save, atomic writes, and backup behavior. The JSON file is plaintext and must be protected externally.

## CLI workflows

`internal/cli` parses command arguments, routes human output through `internal/cliui`, and delegates the explicit `secret copy` side effect to `internal/clipboard`. Reusable behavior belongs in `internal/app`, `internal/project`, `internal/vaultmodel`, or `internal/vaultjson`.

Active command groups:

- `setup`
- `doctor`
- `secret`
- `project`
- `completion`

## Project workflows

`.shelf.json` stores concrete item/field/env bindings. `project status`, `project export`, and `project run` resolve those bindings against the item vault. Project manifests remain value-free.

## Planned, not yet implemented

- Keyboard-first catalog, item-editing, and project-binding TUI flows backed by the existing app workflows.

## Explicitly deferred

- Built-in encryption.
- Automatic v1 migration.
- Web manager.
- Team sharing or hosted sync.
