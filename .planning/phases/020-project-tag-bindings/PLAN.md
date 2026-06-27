# Plan: Phase 20 Project Tag Bindings

## Objective

Add project manifest tag bindings that expand to secret sets during explain/export/run while keeping `.shelf.json` value-free and CLI compact.

## Scope

- Extend manifest entries with `tags` selector.
- Validate path/prefix/tags mutual exclusion and tag syntax.
- Resolve tag entries through existing store tag matching.
- Wire `project add --tag` and render/list/remove support.
- Cover project export/explain diagnostics and conflicts.
- Do not implement docs/changelog release polish; Phase 21 owns that.

## Tasks

1. Extend `manifest.Entry` with `Tags []string` and tag-aware `Key`/`IsTag` helpers.
2. Update manifest validation/add/remove/find behavior for tag entries.
3. Update project resolver for tag entries using `Store.ListByTags`.
4. Update `project add`, `project list`, and completion/removal handling.
5. Add manifest/project tests.
6. Run focused and full verification.

## Acceptance Criteria

- `.shelf.json` supports value-free `{"tags":[...]}` entries.
- Entry selectors are mutually exclusive: path, prefix, or tags.
- `shelf project add --tag ai --tag prod` records a tag entry.
- `project list` displays tag entries.
- `project explain/export/run` expand tag entries using AND semantics and report empty/conflicting bindings clearly.
- No secret values are written to `.shelf.json`.

## Verification

- `go test ./internal/manifest`
- `go test ./internal/project`
- `go test ./internal/cli -run 'TestProject.*Tag|TestProjectExport|TestProjectExplain'`
- `go test ./...`

## Risks

- Comma-joined tag keys must stay stable and unambiguous; tags already reject commas through `store.IsPathToken`.
- Dynamic tag bindings can change project env output as tags change; `project explain` must make expansion visible.
