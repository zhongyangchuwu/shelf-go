# Phase 22 Context: Architecture Repartition Core

## Goal

Cleanly repartition internal packages before v0.1.1 release so command names and package boundaries match Shelf's product concepts: a local manager surface, project workflows, secret workflows, and an encrypted vault core.

## Constraints

- Do not preserve `shelf vault open` as a compatibility alias; the project is pre-release and one feature should have one command name.
- Use `shelf manager` as the manager entrypoint.
- Keep `internal/manager` as the local manager surface package; it is not limited to vault-only features or Web-only UI forever.
- Keep `internal/project` and `internal/secret` as independent feature packages.
- Move project manifest schema/IO/validation into `internal/project` because `.shelf.json` is project-specific.
- Rename `internal/store` to `internal/vault` and merge vault diagnostics plus atomic file writes into that vault core package.
- Move version composition into `internal/app`; keep `internal/app` as runtime/application composition, not a forced home for unrelated feature packages.
- Rename `internal/render` to a clearer export formatting package.
- Defer broad user/developer documentation rewrite to the following docs phase.
- Do not change vault file format, manifest schema, encryption behavior, Web manager behavior, tag semantics, or release automation semantics.

## Decisions

- `shelf manager` replaces `shelf vault open` with no alias.
- `internal/manager` remains the package for the local manager surface.
- `internal/project` absorbs `internal/manifest`.
- `internal/vault` becomes the encrypted vault core package and absorbs old `internal/store`, old vault diagnostics, and atomic write implementation.
- `internal/secret` remains independent and depends on vault core.
- `internal/app` absorbs version composition only.
- `internal/render` is renamed to `internal/exportfmt` because it formats env/shell/JSON export output rather than UI rendering.

## Open Questions

- None.

## Verification Expectations

- `shelf manager` exists and starts the local manager command path.
- `shelf vault open` no longer exists.
- All imports use the new package layout.
- Package tests move with their packages and keep behavior coverage.
- `go test ./...` and `go vet ./...` pass.
- Focused CLI tests cover the manager command cutover.
