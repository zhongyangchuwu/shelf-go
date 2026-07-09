# Capture: v0.2.0 Milestone Audit and Migration Scope Reset

## Durable Knowledge

- Automatic v1-to-v2 migration is not part of the current v0.2.0 release scope.
- The removed `internal/vaultmigrate` package had no production consumers and represented premature automation for a high-risk model conversion.
- Future migration automation needs explicit backup, dry-run, user review, rollback, and no-secret-log requirements before code is added.
- V2 JSON sidecars remain plaintext secret data until v2 age wrapping exists.
- The current v2 storage boundary is concrete (`vaultjson.Repository` constructed by `internal/app`); a future encryption/storage phase should add a v2 provider/repository abstraction before widening backend support.
- `internal/secret` is legacy v1 editing support. New v2 editing should not be added there without renaming or splitting.

## Docs Updated

- `SECURITY.md`: replaced stale `shelf vault open` / `secret edit` plaintext-boundary text with `shelf manager`, stdin writes, v2 manager gating, and `*.v2.json` plaintext warnings.
- `docs/architecture.md`: added v0.2 package map, v2 model/repository section, concrete v2 storage-boundary note, and legacy `internal/secret` scope.
- `docs/contributing.md`: updated project layout for current packages.
- `CHANGELOG.md`: records removal of the internal migration adapter from release scope.
- `AGENTS.md`: updated stale command references for project status/run and v2 stdin writes.

## Planning Records Updated

- `.planning/PROJECT.md`: records migration automation deferral and current package boundaries.
- `.planning/REQUIREMENTS.md`: marks `V2-IMPL-04` deferred and adds future `V2-MIG-AUTO-01`.
- `.planning/ROADMAP.md`: adds Phase 41 and marks Phase 37 retired by audit.
- `.planning/STATE.md`: moves the workflow to v0.2.0 audit and records the migration scope reset.
- `.planning/phases/041-v0.2-milestone-audit/REVIEW.md`: records audit findings and release assessment.

## Next Decisions

1. Decide whether v0.2.0 ships as a cutover/preview with plaintext v2 warnings or waits for v2 age wrapping.
2. If safe portable vault is required, plan the v2 age-wrapping phase next.
3. If migration automation returns later, start with UX/safety design instead of code conversion.
