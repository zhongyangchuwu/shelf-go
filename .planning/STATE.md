# State

## Current Position

- Version: v0.2.0 local milestone
- Phase: Catppuccin Mocha Panel TUI
- Status: complete
- Active Artifact: `.planning/phases/052-mocha-panel-tui/VERIFICATION.md`
- Next Action: Start a new product milestone only after explicit scope selection.

## Blockers

- None.

## Recent Evidence
- The default CLI version is `v0.2.0`; `shelf --version` reported the local milestone without creating a public tag or release.

- Item-vault-only cutover is implemented and committed.
- Active packages are now CLI/app/cliui/clipboard/config/project/util/vaultjson/vaultmodel.
- `secret copy` now writes a resolved secret to the system clipboard without printing it; WSL tries `clip.exe` then verified `wl-copy` fallback.
- `secret export` and `project export` now support only `shell` and `json` output.
- Transitional implementation `V2` names were removed; persisted schema version fields remain.
- `vaultmodel.Resolver` now owns shared item selector, field, and env-name resolution.
- `vaultmodel.Registry` now owns primary-secret and env-name template-role lookup for item creation and resolution.
- `secret update` atomically changes title, labels, aliases, or a field; secret values remain stdin-only.
- Phase 48 implemented `shelf tui`: an optional Bubble Tea workspace with safe catalog search/filter/detail views, masked create/edit forms, clipboard copy, app-backed create/update/remove actions, terminal preflight, and a 100×24 resize contract.
- Phase 49 implemented `shelf project tui`: a value-free current-worktree binding workspace with safe status diagnostics, candidate picker/filter, add/remove actions, missing-manifest initialization confirmation, terminal preflight, and a 100×24 resize contract.
- Phase 50 implemented `shelf secret audit`: a value-free UTC expiry audit with deterministic health exit codes, secret/project TUI indicators, and project-status annotations; project export/run behavior is unchanged.
- Phase 51 completed an active-code/documentation debt review: vault backups now repair restrictive permissions, custom template create summaries use canonical env roles, stale TUI/command documentation was corrected, and low-value panic/dead-code paths were removed.
- Phase 52 refined both workspaces into persistent Catppuccin Mocha two-panel frames: rounded panels, active blue borders and text markers, retained catalog/binding context during right-pane actions, and no theme configuration surface.
- The proposed personal workflow sequence is scriptable item lifecycle → wide-screen secret workspace TUI → project binding TUI → credential health audit.
- TUI remains optional and keyboard-first; it must never display secret values and must reuse application workflows.
- Named dev/staging/prod environments are not planned for the personal-use path.
- Default vault path is `~/.local/share/shelf/vault.json`.
- Local config points at `/home/han/.local/share/shelf/vault.json`.
- Local backup before cutover: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Latest Phase 52 verification passed:
  - secret and project model panel/focus/value-free tests
  - 120×32 secret and project pseudo-terminal Mocha panel smokes
  - `bash scripts/test.sh` (`go vet`, `go test`, and `go-arch-lint`)
  - `go build ./cmd/shelf`

## Updated

- 2026-07-12
