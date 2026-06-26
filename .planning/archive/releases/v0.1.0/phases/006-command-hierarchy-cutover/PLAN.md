# Plan: Phase 6 Command Hierarchy Cutover

## Requirements

CMD-01 through CMD-08.

## Tasks

1. Command tree refactor
   - Rename top-level init constructor behavior to `setup`.
   - Add `vault` command group with `init`, `migrate`, and `open`.
   - Add `secret export` using current direct export implementation.
   - Move current `run` command under `project run`.
   - Remove top-level `init`, `migrate`, `export`, `run`, and `manager` registrations.

2. Test update
   - Update init tests to `setup` and `vault init`.
   - Update migration tests to `vault migrate`.
   - Update export tests to `secret export`.
   - Update run tests to `project run`.
   - Update manager/root command tests for `vault open` and absence of old top-level commands.

3. Documentation update
   - Update README command surface and safety notes.
   - Update usage spec command sections.
   - Keep future `project activate/deactivate/shell` out of implementation docs except as roadmap/planning scope.

4. Verification
   - Run focused CLI tests for init/setup, vault, secret export, project run, manager/open, and completion.
   - Run `go test ./...`.

## Acceptance Criteria

- New canonical commands work.
- Old top-level commands are absent.
- Behavior remains unchanged except command paths.
- Docs no longer present old top-level commands as current.
