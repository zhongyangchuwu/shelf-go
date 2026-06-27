# Capture: Phase 19 Secret Tag Selection

## Durable Docs Updated

- None. User-facing documentation is deferred to Phase 21 release hardening after project tag bindings land.

## Planning Records Updated

- `.planning/phases/019-secret-tag-selection/CONTEXT.md`
- `.planning/phases/019-secret-tag-selection/PLAN.md`
- `.planning/phases/019-secret-tag-selection/SUMMARY.md`
- `.planning/phases/019-secret-tag-selection/VERIFICATION.md`

## Learnings

- Tag selection belongs in `internal/store` because both direct secret commands and Phase 20 project bindings need identical AND semantics.
- Making `secret export` accept `[path-or-prefix]` keeps existing behavior while allowing tag-only application flows.
- Existing `--all` semantics compose cleanly with tag filtering when selection happens before env filtering.

## Ship Inputs

- Phase 20 should reuse `Store.ListByTags` and `HasTags` for manifest tag bindings.
- Phase 21 docs should explain repeatable `--tag` as AND matching.
