# Phase 20 Context: Project Tag Bindings

## Goal

Let project manifests bind tag-selected secret sets without storing secret values, so `project explain`, `project export`, and `project run` can use the same AND tag semantics introduced for direct secret commands.

## Constraints

- `.shelf.json` must remain value-free.
- Manifest entry selectors must be mutually exclusive: exactly one of `path`, `prefix`, or `tags`.
- Tag entries cannot carry `env`; they expand to multiple secrets and use each secret's explicit or derived env name.
- Multiple tags use AND semantics.
- Required tag entries fail when no matching secrets exist; optional tag entries warn and skip.
- Reuse `store.ListByTags` / `store.HasTags`; do not create a second selector implementation.
- Keep CLI compact: add `project add --tag`, not a new project command group.

## Decisions

- Manifest schema uses `"tags": ["ai", "prod"]` for tag entries.
- `Entry.Key()` returns comma-joined tags for tag entries so existing remove/completion/list plumbing has a stable value-free identifier.
- `project add --tag <tag>` is repeatable and uses zero positional args; path/prefix entries keep the existing positional form.
- `project rm <tag-list>` removes tag entries by their comma-joined key.
- `project list` renders `tag    ai,prod (required)` for tag entries.
- `project explain` diagnostics identify tag selectors as `ai,prod (tags)` when empty.

## Open Questions

- None.

## Verification Expectations

- Manifest validation tests cover mutual exclusion, duplicate tags entries, invalid tags, and env rejection.
- Project command tests cover add/list/rm, export, explain, optional empty tags, required empty tags, env conflicts, and value-free manifests.
- `go test ./internal/manifest ./internal/project ./internal/cli` and `go test ./...` pass.
