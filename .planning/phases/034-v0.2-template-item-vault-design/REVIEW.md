# Review: v0.2.0 Template Item Vault Design

## Scope Reviewed

- `.planning/phases/034-v0.2-template-item-vault-design/CONTEXT.md`
- `.planning/phases/034-v0.2-template-item-vault-design/PLAN.md`
- `.planning/phases/034-v0.2-template-item-vault-design/SPEC.md`
- `.planning/phases/034-v0.2-template-item-vault-design/STORAGE-DECISION.md`
- `.planning/phases/034-v0.2-template-item-vault-design/MIGRATION.md`
- Root planning updates for ROADMAP, REQUIREMENTS, PROJECT, and STATE.

## Findings

- No blocking findings.
- The model keeps secret values out of value-free surfaces and planning artifacts.
- JSON-first storage aligns with portability, age wrapping, and chezmoi/git-safe sync better than SQLite for the current scope.
- Migration defaults to `generic_secret` and does not infer API credential semantics from old paths.

## Fixes Applied

- Added explicit public-vs-secret value exposure rules to the spec.
- Added a storage reconsideration trigger list so SQLite can be revisited with concrete evidence.
- Added migration safety rules requiring validation before replacement and no deletion of the v1 vault.
- Added implementation phase sequencing to the roadmap so execution can continue after this design phase.

## Waivers

- No external library/API documentation was needed because this phase defines internal data contracts and planning state only.
- No source tests were run because no production code was changed in this phase.

## Remaining Risks

- The exact item ID generation strategy must be chosen before migration implementation.
- The manager adaptation sequence may temporarily reduce editing UX unless a later phase explicitly scopes it.
- Alias lifetime after migration remains undecided.
