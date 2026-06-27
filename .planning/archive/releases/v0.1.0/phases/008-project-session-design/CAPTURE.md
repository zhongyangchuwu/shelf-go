# Capture: Phase 8 Project Session Design

## Durable Decisions

- Project session commands belong under `shelf project` because they read `.shelf.json` and resolve project bindings.
- `project activate` / `project deactivate` require a shell hook/function to mutate the current shell. The Go CLI alone can only mutate its own process or child processes.
- `project shell` is the no-hook fallback: spawn a child shell with injected env and let process exit restore the parent shell naturally.
- Deactivation must restore previous env values for variables that existed before activation and unset only variables introduced by activation.
- Dry-run/preview output must be value-free and limited to project id, manifest path, env names, overrides, and planned shell actions.
- Repeated activation or project switching should fail by default unless a future explicit replace/refresh path is added.

## Future Implementation Notes

- Reuse existing project manifest resolution and env conflict behavior.
- Generate shell-specific patches from a hook command rather than trying to mutate the parent shell from Go.
- Store activation metadata in shell-local variables or a shell-local state block, not durable project files.
