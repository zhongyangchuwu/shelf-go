# Context: Phase 8 Project Session Design

## Goal

Plan venv-like project session workflows without implementing them yet.

## Commands

Future commands belong under `shelf project` because they read current project `.shelf.json`:

```bash
shelf project activate
shelf project deactivate
shelf project shell
```

## Constraints

- A Go CLI process cannot directly mutate the parent shell environment.
- `project activate` and `project deactivate` require a shell hook/function if the user wants to type those commands directly and change the current shell.
- `project shell` can work without a hook by spawning a child shell with injected env.
- Activation must be reversible: values that existed before activation must be restored, not blindly unset.
- Activation must not leak secret values in preview/dry-run output.
- Repeated activation or project switching must be explicit to avoid mixed secret environments.

## Proposed Semantics

### `shelf project activate`

With a shell hook installed, this activates the current project's `.shelf.json` bindings in the current shell.

Rules:

1. Resolve Git root and `.shelf.json`.
2. Resolve project entries using the encrypted vault.
3. Fail before changing shell state if any required secret is missing, env names conflict, or vault load fails.
4. Save previous env state for each affected env name in shell-local variables or a shell-local state block.
5. Export resolved env values.
6. Set activation metadata such as `SHELF_ACTIVE=1` and `SHELF_ACTIVE_PROJECT=<project-id>`.
7. Refuse activation when another project is active unless a future `--replace` flag is supplied.

### `shelf project deactivate`

With a shell hook installed, this restores the shell state saved by activation.

Rules:

1. If an env var existed before activation, restore its previous value.
2. If an env var was introduced by activation, unset it.
3. Clear Shelf activation metadata.
4. If no activation is present, no-op with a concise message.

### `shelf project shell`

Without requiring a hook, this spawns `$SHELL` with project env injected.

Rules:

1. Reuse project run resolution and conflict behavior.
2. Set a prompt/status marker such as `SHELF_ACTIVE_PROJECT`.
3. Exit restores the parent shell naturally.

## Open Implementation Notes

- The hook should hide shell patch generation from users while keeping commands ergonomic.
- The non-hook fallback should be documented as `shelf project shell`.
- The command hierarchy phase intentionally does not implement these commands.
