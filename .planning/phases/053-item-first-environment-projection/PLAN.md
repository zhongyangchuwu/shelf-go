# Plan: Item-First Environment Projection V3

## Scope

Implement the approved V3 contract in `SPEC.md` across vault validation/resolution, app workflows, Cobra commands, project bindings, audit, TUI safe summaries, and tests. Do not modify the configured user vault.

## Execution

1. Replace V2 field instances with atomic ID/value storage; move schema metadata into templates and replace `env_for` roles with validated `env_` / `validity_` conventions.
2. Update resolver/index/app creation and updates to use template definitions and convention-derived field relationships.
3. Replace secret-only label export with `env export`, projecting all mapped eligible public and secret fields. Permit project binding/export/run for the same field set.
4. Replace date-only primary expiry handling with per-tracked-field validity values and item-first TUI summaries/detail grouping.
5. Update CLI/TUI/doc contracts and add V3 regression and end-to-end coverage.

## Verification

- Model tests: V3 codec rejection/validation, convention targets, validity state rules, resolver ENV indexing.
- App/CLI tests: all-field public+secret env export, command cutover, duplicate/missing failure paths, project public binding, lifecycle exits.
- TUI tests: item aggregate summaries, grouped structural metadata, value-free rendering.
- Full repository test, build, V3 temporary-vault CLI and pseudo-terminal smoke.

## User-Managed Migration

After code verification, the user manually migrates their plaintext vault from V2 to V3 using a value-free mapping checklist. The implementation provides a migration guide but no automatic migration path.