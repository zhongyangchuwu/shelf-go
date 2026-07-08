# Context: Project Manifest v2 Bindings

## Goal

Implement project manifest v2 item/field/env bindings so `shelf project status`, `shelf project export`, and `shelf project run` can resolve v2 vault items without storing secret values in `.shelf.json`.

## Constraints

- Do not store secret values in `.shelf.json`.
- Do not print secret values in `project status`, diagnostics, completions, configure views, or ambiguity errors.
- Preserve existing v1 project manifest support while project migration/cutover is underway; v1 export/run tests remain valid for current encrypted vault workflows.
- v2 secret storage is still Phase 38 plaintext JSON at `<resolved-vault-path>.v2.json` unless `--vault` is explicit. Project v2 resolution must use the same v2 path helper behavior as v2 secret commands.
- `project export` and `project run` may materialize secret values only as their explicit purpose.
- Prefer boring manifest JSON with strict validation and no fuzzy selection in first pass.
- Do not implement a public migration command in this phase unless required for export/run behavior.
- Do not adapt the local manager in this phase; Phase 40 owns manager behavior.

## Decisions

- A v2 project manifest uses version `2` with concrete bindings: item ID, field ID, env name, required flag, and optional title/review metadata.
- Binding resolution is concrete, not prefix/tag based. Prefix/tag expansion remains migration-adapter behavior from Phase 37.
- `project status` supports both v1 and v2 manifests; v2 diagnostics name item IDs/titles/field IDs/env names only.
- `project export` and `project run` support both v1 and v2 manifests; v2 resolution loads the v2 vault repository and reads only the explicitly bound field values.
- `project add` in v2 mode binds a v2 item/field. It defaults to the item's primary secret field and derives the env name from the field's linked env-name public field unless `--env` overrides it.
- Duplicate env names in a v2 project manifest are a failure at resolution time, not globally forbidden in the vault.

## Open Questions

- Whether `project init` should immediately create version 2 manifests or keep version 1 until an explicit cutover command exists. Defaulting new manifests to v2 is preferred for v0.2, but existing v1 tests and workflows still need compatibility during transition.
- Whether to add a dedicated `project add --field FIELD <selector>` flag now or defer field binding beyond the primary field. Phase 39 likely needs it for correctness.

## Verification Expectations

- Focused project package tests for v2 manifest validation and v2 resolution diagnostics.
- Focused app/CLI tests for v2 `project add`, `project status`, `project export`, and `project run` using explicit temp v2 JSON vault paths.
- Regression tests proving v1 project export/run still pass.
- Full verification: `go test ./...`, `go vet ./...`, `go-arch-lint check`.
