# Summary: Manager Adaptation and Cutover

## Completed Changes

- Added an explicit v2 gate before `App.OpenManager` opens the v1 manager service.
- The gate resolves the v2 vault path using the existing Phase 38 behavior:
  - explicit `--vault`: exact path;
  - no explicit `--vault`: `<resolved-vault-path>.v2.json`.
- The gate attempts to load a valid v2 JSON vault at that path. If it loads, `shelf manager` fails with a value-free message explaining that the local manager is not available for v2 vaults yet and that users should use `shelf secret` and `shelf project` commands.
- Existing v1 manager behavior remains available when no valid v2 vault is present.
- `internal/manager` server/API/UI code remains unchanged; the gate lives in `internal/app` before the v1 manager service is constructed.
- Added app and CLI tests for v2 gating and v1 preservation.
- Updated docs and planning records.

## Files Changed

- `internal/app/manager.go`
- `internal/app/manager_test.go`
- `internal/cli/manager_test.go`
- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `docs/architecture.md`
- `docs/portable-vault.md`
- `CHANGELOG.md`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/040-manager-adaptation-cutover/CONTEXT.md`
- `.planning/phases/040-manager-adaptation-cutover/PLAN.md`

## Decision

Phase 40 chose gating over full adaptation.

Full adaptation would require a new v2 item/field editor for templates, public env-name fields, secret fields, field sensitivity, and explicit reveal/copy semantics. Reusing the v1 path/env/tag manager would misrepresent v2 items and risk edits to stale or mismatched data.

## Evidence

- `go test ./internal/app ./internal/cli ./internal/manager` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go-arch-lint check` passed.
- Tester agent reported focused app/CLI/manager tests passed.

## Remaining Work

- A real v2 manager UI remains future work.
- V2 JSON vault files remain plaintext during cutover and are not safe to sync or commit.
- v0.2.0 track is ready for milestone audit/release planning.
