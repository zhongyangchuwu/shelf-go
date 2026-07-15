# Architecture

Shelf Go is a local item-vault CLI with optional wide-screen secret and project-binding workspaces. Built-in age encryption, the v1 encrypted vault stack, and Web manager are outside the active product; external tools such as chezmoi own encryption and sync.

## Package layers

```text
cmd/shelf/                    process entry point

internal/cli/                 Cobra command tree, flags, completions, output routing, and no-echo terminal field prompts
internal/cliui/               Lip Gloss-backed human output helpers for titles, summaries, empty states, and tables
internal/tui/                 Bubble Tea secret and project-binding workspaces; app/clipboard-backed only
internal/clipboard/           System clipboard adapter; uses `clip.exe`, then `wl-copy`, under WSL and github.com/atotto/clipboard elsewhere
internal/app/                 application workflows for secret, credential health, project, setup, doctor, and run

internal/vaultmodel/          item/field/template model, validation, template-role lookups, and shared item/field/env resolver
internal/vaultjson/           strict JSON codec/repository for the local item vault
internal/project/             Git project identity, .shelf.json schema/IO/validation, and binding resolution
internal/config/              config and vault path resolution
internal/util/                atomic write and env/shell/JSON binding formatting
```

Dependency direction is enforced by `.go-arch-lint.yml`:

```text
cmd/shelf -> cli
cli -> app, cliui, clipboard, tui, Cobra, x/term
cliui -> lipgloss
tui -> app, clipboard, Bubble Tea, Bubbles, Huh, Lip Gloss
clipboard -> github.com/atotto/clipboard
app -> config, project, util, vaultmodel, vaultjson
project -> util, vaultmodel
vaultjson -> vaultmodel, util
config -> yaml
```

## Item vault

`internal/vaultmodel` owns schema version 4: atomic item fields, item descriptions, template-owned field definitions, `env_<field>` environment mappings, `validity_<secret-field>` lifecycle metadata, validation, indexes, and shared item/field/env resolution.

`internal/vaultjson` owns persistence mechanics for the configured JSON file: strict decode, validated encode/save, atomic writes, and backup behavior. The JSON file is plaintext and must be protected externally.

## CLI workflows

`internal/cli` parses command arguments, routes human output through `internal/cliui`, and delegates the explicit `secret copy` side effect to `internal/clipboard`. Reusable behavior belongs in `internal/app`, `internal/project`, `internal/vaultmodel`, or `internal/vaultjson`. `shelf env export` materializes explicitly mapped fields; `shelf secret audit` derives value-free field health only from public lifecycle metadata.

Active command groups:

- `setup`
- `doctor`
- `tui`
- `secret`
- `env`
- `project`
- `completion`

## Project workflows

`.shelf.json` stores concrete item/field/env bindings. `project status`, `project export`, and `project run` resolve those bindings against the item vault. Project manifests remain value-free. `shelf project tui` is an optional app-backed manifest workspace and does not read/write the manifest directly.

## Interactive workspaces

`shelf tui` and `shelf project tui` require an interactive terminal at least 100 columns × 24 rows. They use app workflows (plus clipboard for explicit copy), mask secret input, and never provide a secret-reveal view.

## Explicitly deferred

- Built-in encryption.
- Automatic v1 migration.
- Web manager.
- Team sharing or hosted sync.
