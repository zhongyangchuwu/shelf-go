# Summary: Phase 9 Project Export Shell Default

## Completed Changes

- Changed `shelf project export` default output from `env` to existing `shell` format.
- Kept explicit `--format env`, `--format shell`, and `--format json` behavior.
- Added regression coverage for bare `shelf project export` emitting sourceable `export NAME=value` lines.
- Updated README, getting started, reference, and security docs to recommend explicit export/source workflows and warn about plaintext env files.

## Files Changed

- `internal/cli/project.go`
- `internal/cli/project_test.go`
- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/security.md`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/009-project-export-shell-default/CONTEXT.md`
- `.planning/phases/009-project-export-shell-default/PLAN.md`

## Deviations

- None. No `dotenv` format, `--out` option, hook command, or shell wrapper was added.

## Evidence

- `go test ./internal/cli -run TestProjectExport` passed.
- `go test ./...` passed.
- Search for `dotenv` under `internal`, `README.md`, and `docs` found no matches.

## Unresolved Risks

- Redirected project export files contain plaintext values. Docs now warn to add them to `.gitignore` and delete them when no longer needed.
