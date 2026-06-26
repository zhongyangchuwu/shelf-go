# Capture: Phase 17 Web Manager Design Contract

## Durable Docs Updated

- None. Product documentation updates belong to Phase 21 release hardening after behavior is implemented.

## Planning Records Updated

- `.planning/phases/017-web-manager-design/UI-SPEC.md`
- `.planning/phases/017-web-manager-design/PLAN.md`
- `.planning/phases/017-web-manager-design/SUMMARY.md`
- `.planning/phases/017-web-manager-design/VERIFICATION.md`

## Learnings

- Shelf Web Manager should read as a local vault workbench, not a SaaS admin dashboard.
- First-party embedded CSS/JS is enough for v0.1.1 and preserves the single-binary release constraint.
- Security hardening is part of the UI foundation: token-bearing URLs, GET reveal, missing no-store headers, and silent overwrite semantics must be fixed before expanding Web editing.

## Ship Inputs

- Phase 18 should implement `UI-SPEC.md` before adding tag CLI/project features.
- Phase 18 tests should cover token redirect, no-store headers, POST reveal, no value leakage in list/detail, metadata-only updates, explicit rename semantics, and delete behavior.
