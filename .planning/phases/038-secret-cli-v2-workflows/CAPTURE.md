# Capture: Secret CLI v2 Workflows

## Durable Docs Updated

- `README.md` now shows v2 `secret create`, stdin writes, profile usage, env-name field updates, and the Phase 38 plaintext v2 JSON caveat.
- `docs/getting-started.md` now teaches v2 create/get/set workflows instead of old value-argv `secret set`.
- `docs/reference.md` now documents v2 secret items, built-in templates, profiles, command grammar, removed/deferred commands, and the temporary v2 JSON path rule.
- `docs/portable-vault.md` warns not to sync `*.v2.json` during the cutover.
- `docs/architecture.md` records v2 secret item/field workflows and defers full-object editing to a later TUI/wizard.
- `docs/troubleshooting.md` uses v2 secret creation in missing-binding guidance.
- `CHANGELOG.md` records the breaking v2 secret CLI cutover.

## Planning Records Updated

- `.planning/PROJECT.md` records Phase 38 as validated and Phase 39 as the next active direction.
- `.planning/REQUIREMENTS.md` marks `V2-IMPL-05` complete and updates env-binding language to env-name public fields.
- `.planning/ROADMAP.md` marks Phase 38 complete.
- `.planning/STATE.md` advances to Phase 39.
- Phase-local `SUMMARY.md`, `REVIEW.md`, and `VERIFICATION.md` capture implementation and evidence.

## Learnings

- Env names work better as ordinary public fields than as a separate binding subsystem: users can inspect and edit them with the same field mechanics as other public metadata.
- Profiles keep provider defaults out of CLI branching. `openai` can stay a data definition over `api_credential` while `provider` remains a normal public field.
- Stdin-only writes make it easy to test value exposure boundaries and avoid accidental shell history leaks.
- During transition, v2 storage must use a distinct default path so a partially wired plaintext JSON repository cannot overwrite the existing v1 encrypted vault.
- Direct app-service tests catch behavior that CLI smoke tests might only observe indirectly; keep that split for future project/manager v2 work.

## Ship Inputs

- Phase 39 should integrate project manifests with v2 item/field/env resolution and decide how to bridge or migrate existing v1 `.shelf.json` path entries.
- Phase 39 must keep project status/export/run diagnostics value-free for missing items, missing fields, and duplicate env names.
- Phase 40 should adapt or gate the local manager before users can accidentally edit the wrong vault model.
- v2 repository encryption/age wrapping remains a security prerequisite before v2 vault files are safe to sync or commit.
