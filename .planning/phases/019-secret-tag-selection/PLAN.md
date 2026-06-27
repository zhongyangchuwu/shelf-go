# Plan: Phase 19 Secret Tag Selection

## Objective

Add tag-based selection to compact direct secret CLI workflows without expanding the edit command surface.

## Scope

- Add reusable tag selector helpers in the store layer.
- Add `--tag` filtering to `shelf secret list`.
- Add `--tag` filtering to `shelf secret export`.
- Preserve existing path/prefix export behavior and `--all` semantics.
- Do not add project manifest tag bindings; Phase 20 owns that.

## Tasks

1. Implement AND-semantics tag matching over `store.Secret`.
2. Add deterministic `Store.ListByTags(prefix, tags)` behavior.
3. Wire `secret list [prefix] --tag <tag>` through the new helper.
4. Wire `secret export [path-or-prefix] --tag <tag>` so tag-only export is allowed and path/prefix plus tags composes.
5. Add focused tests for list/export tag filtering, AND semantics, default env-only filtering, and `--all`.

## Acceptance Criteria

- `shelf secret list --tag ai` prints only matching paths and never values.
- Repeated `--tag` flags use AND semantics.
- `shelf secret export --tag ai --format env` exports matching env-bound secrets.
- `shelf secret export --tag ai --all` includes matching secrets without explicit env using derived env names.
- Existing exact path and prefix export behavior is unchanged.

## Verification

- `go test ./internal/store`
- `go test ./internal/cli -run 'TestSecret.*Tag|TestExport.*Tag|TestExportPrefix|TestSecretSetGetListInfoExport'`
- `go test ./...`

## Risks

- Tag-only export makes the positional path optional; argument validation must reject calls with neither path nor tag.
- Future project tag bindings must reuse the same AND semantics to avoid two tag selector definitions.
