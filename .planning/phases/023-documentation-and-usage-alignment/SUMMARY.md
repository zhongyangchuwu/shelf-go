# Summary: Phase 23 Documentation and Usage Alignment

## Completed Changes

- Updated user-facing docs for v0.1.1 manager editing, direct tag selection, and project tag bindings.
- Updated command reference for:
  - `shelf manager` Web console behavior and plaintext boundaries;
  - `shelf secret list [prefix] --tag` repeatable AND semantics;
  - `shelf secret export [path-or-prefix] --tag` selector requirements and `--all` behavior;
  - `shelf project add --tag` and tag selector constraints;
  - `.shelf.json` `tags` entries.
- Updated troubleshooting for required tag bindings and prefix/tag env-override limits.
- Updated contributing docs for `scripts/install.sh`, `scripts/release.sh`, and `justfile` as a thin wrapper.
- Refined architecture docs around display / feature support / base support layers, final package layout, manager safety boundaries, and SQLite deferral.
- Updated planning state to mark Phase 23 complete and Phase 24 next.

## Files Changed

- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `docs/contributing.md`
- `docs/architecture.md`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/023-documentation-and-usage-alignment/*`

## Decisions Preserved

- No documentation for fine-grained `secret meta` or `secret tag` command groups.
- `shelf manager` remains the only documented manager entrypoint.
- Storage remains documented as age-encrypted JSON; SQLite/storage redesign stays deferred to v0.2.0.
- Changelog and release snapshot readiness remain Phase 24 responsibilities.

## Evidence

- `go test ./...` passed.
- LSP workspace diagnostics reported no Go issues.
- Grep verification found no active docs recommending old package paths or manual GoReleaser commands.
- Grep matches for `shelf vault open` and `secret meta` were intentional boundary statements, not active usage instructions.

## Next Phase

Phase 24: v0.1.1 Release Hardening.
