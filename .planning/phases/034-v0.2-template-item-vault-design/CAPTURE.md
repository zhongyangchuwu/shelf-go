# Capture: v0.2.0 Template Item Vault Design

## Durable Docs Updated

- None outside `.planning/`. This phase changed design/workflow artifacts only; user-facing docs should update when implementation changes behavior.

## Planning Records Updated

- `.planning/PROJECT.md` records v0.2 design validation and the next active implementation direction.
- `.planning/REQUIREMENTS.md` records completed Phase 34 design requirements and new implementation requirements.
- `.planning/ROADMAP.md` marks Phase 34 complete and adds follow-on v0.2 implementation phases.
- `.planning/STATE.md` advances to the next phase.
- Phase-local `SPEC.md`, `STORAGE-DECISION.md`, `MIGRATION.md`, `SUMMARY.md`, `REVIEW.md`, and `VERIFICATION.md` capture the design result.

## Learnings

- `group:key` should be treated as a v1 migration alias, not the v2 product model.
- `api_credential` is useful as a built-in template, but the root model must stay generic.
- Env names belong on secret fields as bindings, not on items or project-level values.
- JSON remains the simplest storage shape for age-wrapped portable vaults; SQLite needs stronger future requirements before adoption.
- Migration should default all v1 records to `generic_secret` to avoid guessing semantics from names.

## Ship Inputs

- Next phase should implement the v2 model package and validation contracts before touching storage, CLI, project manifests, or manager behavior.
- Implementation should preserve the adapter boundary from the design: model/validation first, JSON codec/repository second, legacy migration third, user-facing commands after core contracts are proven.
