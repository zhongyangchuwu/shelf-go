# UI Specification: Project Binding TUI

**Status:** Implemented and verified.

## Product Target

`shelf project tui` is an optional, wide-screen workspace for the current Git worktree's value-free `.shelf.json` bindings. It accelerates project binding selection without becoming an environment-file editor or a value-reveal path.

## Entry and Scope

- The command requires interactive stdin and stdout and the existing 100×24 terminal minimum.
- It uses the same alternate-screen lifecycle and resize-required view as `shelf tui`.
- It calls only app workflows. It never parses or writes `.shelf.json`, reads the vault directly, or materializes binding values.
- `project run` and `project export` remain explicit CLI commands.

## Workspace Layout

```text
Shelf · Project bindings   my-repository   2 bindings
────────────────────────────────────────────────────────────────────────
> Bindings                         Detail

› openai.credential   OPENAI_API_KEY      openai.credential
  github.credential   GITHUB_TOKEN        env:      OPENAI_API_KEY
                                         required: yes
                                         status:   ready
────────────────────────────────────────────────────────────────────────
a add  d remove  i initialize  / filter picker  r reload  ? help  q quit
```

- Left pane: manifest bindings with item.field, env name, required/optional state, and `ready`/`warn`/`fail` status.
- Right pane: the selected binding's value-free detail and safe diagnostic text.
- A missing manifest shows an empty workspace with `i initialize`; initialization requires `enter` confirmation and never overwrites an existing manifest.

## Actions

| Key | Action | Safety and behavior |
| --- | --- | --- |
| `a` | Add binding | Opens a picker of vault secret fields, then a form with the field's default env name, an explicit override, and optional toggle. Save calls `app.ProjectAdd`. |
| `d` | Remove binding | Shows `item.field`, env name, and required state. `enter` calls `app.ProjectRm`; `esc` cancels. |
| `i` | Initialize manifest | Available only without `.shelf.json`; `enter` calls `app.ProjectInit(false)`. |
| `/` | Filter picker | Case-insensitive match on candidate item title, item ID, field, or env name. |
| `r` | Reload | Reloads the safe workspace DTO. |
| `?` | Help | Shows the full keymap without values. |

`j`/`k` moves in the active binding list or candidate picker. `esc` cancels an active picker, form, or confirmation. `q` and Ctrl+C restore the terminal and exit.

## Safe Data Contract

The app supplies value-free DTOs only:

- workspace: project ID/root, manifest presence, and each binding's item, title, field, env, required state, resolution status, and safe diagnostic;
- candidates: item ID/title, secret field ID, and default env name;
- action status: fixed success text.

Secret values, `ResolvedBinding.Value`, child environment strings, project export output, and app error text are never stored or rendered by the TUI. Error views use fixed value-free copy.

## Verification

- App tests prove workspace/candidate DTOs contain no secret values and report ready, missing, optional, and duplicate-env states.
- TUI model tests cover missing-manifest initialization, candidate filtering, add/remove workflow dispatch, safe error/detail/confirmation rendering, and resize behavior.
- CLI tests prove `shelf project tui` rejects non-interactive streams.
- Pseudo-terminal smokes verify wide and resize-required lifecycle cleanup.
