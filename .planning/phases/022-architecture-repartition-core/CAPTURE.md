# Capture: Phase 22 Architecture Repartition Core

## Durable Docs Updated

- `docs/architecture.md` now describes the repartitioned internal package layout.
- `docs/contributing.md` project layout now matches current internal packages.
- `README.md`, `docs/getting-started.md`, `docs/reference.md`, and `docs/troubleshooting.md` were minimally updated from `shelf vault open` to `shelf manager` to avoid stale command references.

## Planning Records Updated

- `.planning/ROADMAP.md` splits Phase 22 architecture, Phase 23 docs, and Phase 24 release hardening.
- `.planning/REQUIREMENTS.md` maps ARCH-01..ARCH-02 to Phase 22 and DOC-01..DOC-02 to Phase 23.
- `.planning/PROJECT.md` records the revised package layout and manager entrypoint decision.
- `.planning/STATE.md` advances to Phase 23.
- Phase 22 records added under `.planning/phases/022-architecture-repartition-core/`.

## Learnings

- Go directory nesting does not create nested packages; moving feature packages under `app/` would only lengthen import paths without giving parent package semantics.
- `internal/project` is the right home for `.shelf.json` because the manifest schema is project-specific.
- `internal/vault` is the right home for the encrypted vault core; the previous split between `store`, `vault`, and `atomicfile` made the core harder to name.
- `internal/exportfmt` is clearer than `render` because the package formats env/shell/JSON export output, not UI rendering.
- Pre-release command cleanup should prefer one canonical command over compatibility aliases.

## Ship Inputs

- Phase 23 should complete user/developer docs for `shelf manager`, tag workflows, scripts, and final architecture.
- Phase 24 should run release hardening with `scripts/release.sh check` and `scripts/release.sh snapshot`.
