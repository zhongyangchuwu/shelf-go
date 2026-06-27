# Plan: Phase 23 Documentation and Usage Alignment

## Objective

Bring public user/developer docs in line with the implemented v0.1.1 command surface and package layout before release hardening.

## Non-Goals

- No code behavior changes.
- No new commands, aliases, or compatibility shims.
- No storage format changes or SQLite discussion beyond explicit deferral.
- No changelog/release readiness work beyond docs needed to make current behavior accurate.

## Work Items

1. User guide flows
   - Update `README.md` and `docs/getting-started.md` to show:
     - manager editing as the full-object editing surface;
     - direct tag filtering for `secret list` / `secret export`;
     - project tag binding examples.
   - Preserve plaintext boundary warnings.

2. Command reference
   - Update `docs/reference.md` for:
     - `secret list --tag` repeatable AND semantics;
     - `secret export --tag` with optional path/prefix and `--all` behavior;
     - `project add --tag` semantics and constraints;
     - `.shelf.json` `tags` entries;
     - manager reveal/write safety behavior.

3. Developer docs
   - Update `docs/contributing.md` for:
     - `scripts/install.sh` and `scripts/release.sh` command surfaces;
     - `justfile` as a thin wrapper;
     - final package layout and architecture-doc link.

4. Architecture docs
   - Ensure `docs/architecture.md` accurately describes:
     - display / feature support / base support layers;
     - `internal/manager` as local manager surface;
     - `internal/vault` as encrypted vault core;
     - `internal/project` owning manifest schema and tag bindings;
     - `internal/exportfmt` as export formatting.

5. Verification and records
   - Grep for stale `shelf vault open` and old package names in active docs/code.
   - Check referenced command options against CLI code.
   - Run `go test ./...` as a regression check after docs-only edits.
   - Write `SUMMARY.md`, `VERIFICATION.md`, and `CAPTURE.md`.
   - Update root planning state, requirements, roadmap, and project status.

## Acceptance Criteria

- DOC-01 and DOC-02 are complete.
- Public docs no longer imply the old manager command or old package split.
- Docs explain tag selection without suggesting unsupported metadata edit subcommands.
- Docs preserve secret-value warnings for export, manager reveal/copy/edit, and generated env files.
- Verification evidence is recorded under this phase.
