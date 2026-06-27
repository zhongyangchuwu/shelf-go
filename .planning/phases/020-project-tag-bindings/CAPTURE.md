# Capture: Phase 20 Project Tag Bindings

## Durable Docs Updated

- None. User-facing docs are deferred to Phase 21 release hardening.

## Planning Records Updated

- `.planning/phases/020-project-tag-bindings/CONTEXT.md`
- `.planning/phases/020-project-tag-bindings/PLAN.md`
- `.planning/phases/020-project-tag-bindings/SUMMARY.md`
- `.planning/phases/020-project-tag-bindings/VERIFICATION.md`

## Learnings

- Tag bindings fit the existing manifest model when treated as a third selector form beside path and prefix.
- Comma-joined tag keys are safe because tag tokens reject commas and spaces through existing store token validation.
- Reusing `Store.ListByTags` kept Phase 19 and Phase 20 semantics aligned.

## Ship Inputs

- Phase 21 docs should show `.shelf.json` tag entries as value-free selectors.
- Phase 21 docs should explain that tag project bindings are dynamic and should be audited with `shelf project explain`.
