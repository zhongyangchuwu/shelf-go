# Summary: v0.2.0 Template Item Vault Design

## Completed Changes

- Completed the v0.2 design phase for a template-based item/field vault model.
- Added a concrete model spec covering `Vault`, `Item`, `Field`, `Binding`, `Template`, built-in templates, validation rules, value exposure rules, selectors, and project manifest v2 shape.
- Added a storage decision selecting JSON as the initial canonical v2 storage behind repository/codec interfaces while deferring SQLite.
- Added a migration plan from v1 flat secrets and project manifests to v2 items and field bindings.
- Updated root planning artifacts to make Phase 34 complete and define the next implementation phases.

## Files Changed

- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/034-v0.2-template-item-vault-design/CONTEXT.md`
- `.planning/phases/034-v0.2-template-item-vault-design/PLAN.md`
- `.planning/phases/034-v0.2-template-item-vault-design/SPEC.md`
- `.planning/phases/034-v0.2-template-item-vault-design/STORAGE-DECISION.md`
- `.planning/phases/034-v0.2-template-item-vault-design/MIGRATION.md`
- `.planning/phases/034-v0.2-template-item-vault-design/REVIEW.md`
- `.planning/phases/034-v0.2-template-item-vault-design/VERIFICATION.md`
- `.planning/phases/034-v0.2-template-item-vault-design/CAPTURE.md`

## Deviations

- No production code was changed. This matches the phase non-goal.
- The design adds `aliases` as a migration aid in the item spec. The plan had only labels/notes; aliases are included to preserve old v1 paths without keeping `group:key` as the product model.
- The design defines follow-on implementation phases because the roadmap had only the design phase.

## Evidence

- Phase artifacts include required OMP sections.
- The spec covers all V2-MODEL requirements.
- The storage decision covers all V2-STORE requirements.
- The migration plan covers all V2-MIG requirements.
- `git status --short` was clean before the phase started.

## Unresolved Risks

- Exact item ID format and deterministic migration ID strategy remain for the model implementation phase.
- The alias lifetime policy remains for the migration implementation phase.
- Manager adaptation is sequenced later and may need a temporary compatibility plan.
