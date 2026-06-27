# Plan: Phase 9 Project Export Shell Default

## Objective

Change `shelf project export` so the no-flag default matches `shelf secret export`: sourceable `shell` output.

## Scope

In scope:

- Change the `project export --format` default from `env` to `shell`.
- Add or update tests for bare `project export` default behavior.
- Update README and docs to recommend explicit source workflows and warn about plaintext env files.

Out of scope:

- New `dotenv` format.
- `project activate`, `project deactivate`, or `project shell`.
- `--out` file writing convenience.
- Vault restore or edit/manager hardening; later phases own those.

## Tasks

1. Update CLI default.
   - Change `newProjectExportCmd` flag default to `shell`.
   - Keep format completions unchanged: `env`, `shell`, `json`.

2. Update tests.
   - Add coverage for bare `shelf project export` emitting `export NAME=value`.
   - Keep explicit `--format env`, `--format shell`, and `--format json` tests passing.

3. Update docs.
   - README quick/core docs should show sourceable project export workflow.
   - Getting started should show `shelf project export > .env.local` and `source .env.local` as the manual current-shell path.
   - Reference should state `project export` defaults to `shell` and that generated files are plaintext.

## Acceptance Criteria

- `shelf project export` without `--format` emits shell export lines.
- Explicit `--format env` still emits bare env lines.
- Explicit `--format json` still emits JSON.
- No `dotenv` format appears in docs or completions.
- Docs warn not to commit generated env files.

## Verification

- Run focused project export tests.
- Run the relevant docs/code search for accidental `dotenv` additions.
- Run `go test ./internal/cli -run 'TestProjectExport'`.

## Risks

- Default output change may surprise pre-release users; acceptable before public release and aligned with `secret export`.
- Redirecting shell output to `.env.local` creates plaintext; docs must make this explicit.
