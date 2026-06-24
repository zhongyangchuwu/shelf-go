# Context: Phase 9 Project Export Shell Default

## Goal

Make `shelf project export` sourceable by default by using the existing `shell` output format, while keeping the command surface small and avoiding a new dotenv format or shell-hook workflow.

## Constraints

- Do not implement `project activate`, `project deactivate`, or `project shell` in this phase.
- Do not add a `dotenv` output format.
- Preserve explicit `--format env`, `--format shell`, and `--format json` behavior.
- Preserve no-value diagnostics for dry-run/explain/status paths.
- Keep changes inside current package boundaries: CLI behavior in `internal/cli`, rendering in `internal/render`, public docs in README/docs.

## Decisions

- `shell` is the default project export format because it produces `export NAME=value` lines that can be redirected and sourced explicitly.
- `env` remains available for tools that need bare `NAME=value` lines.
- `json` remains available for machine-readable use.
- Documentation should recommend explicit source workflows and warn that redirected files contain plaintext.

## Open Questions

None for this phase.

## Verification Expectations

- Focused CLI tests prove bare `shelf project export` now emits shell output.
- Existing explicit format tests continue to pass.
- Docs mention the recommended explicit source workflow and plaintext/git-ignore warning.
