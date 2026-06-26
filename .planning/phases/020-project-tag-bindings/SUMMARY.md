# Summary: Phase 20 Project Tag Bindings

## Completed Changes

- Extended project manifest entries with value-free `tags` selectors.
- Enforced selector mutual exclusion across `path`, `prefix`, and `tags` entries.
- Added tag entry validation for invalid tags, duplicate tags, duplicate tag entries, and `env` misuse.
- Updated project resolution to expand tag entries through `Store.ListByTags` using AND semantics.
- Added `shelf project add --tag` as a repeatable tag selector.
- Updated `project list`, `project rm`, `project explain`, `project export`, and `project run` resolution paths to handle tag entries.
- Added focused manifest and project command tests.

## Files Changed

- `internal/manifest/manifest.go`
- `internal/manifest/validate.go`
- `internal/manifest/manifest_test.go`
- `internal/manifest/tag_test.go`
- `internal/project/resolve.go`
- `internal/cli/project.go`
- `internal/cli/project_tag_test.go`
- `.planning/phases/020-project-tag-bindings/CONTEXT.md`
- `.planning/phases/020-project-tag-bindings/PLAN.md`

## Deviations

- None.

## Evidence

- `go test ./internal/manifest` passed.
- `go test ./internal/project` passed.
- `go test ./internal/cli -run 'TestProject.*Tag|TestProjectAdd|TestProjectList|TestProjectExport|TestProjectExplain'` passed.
- `go test ./...` passed.
- LSP workspace diagnostics reported no Go issues.

## Unresolved Risks

- User-facing documentation for dynamic tag bindings is not written yet; Phase 21 owns docs/release hardening.
