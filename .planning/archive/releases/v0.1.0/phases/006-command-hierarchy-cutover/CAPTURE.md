# Capture: Phase 6 Command Hierarchy Cutover

## Durable Docs Updated

- `README.md`
- `docs/usage-spec.md`
- `docs/data-spec.md`

## Planning Records Updated

- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/006-command-hierarchy-cutover/CONTEXT.md`
- `.planning/phases/006-command-hierarchy-cutover/PLAN.md`
- `.planning/phases/006-command-hierarchy-cutover/SUMMARY.md`
- `.planning/phases/006-command-hierarchy-cutover/VERIFICATION.md`
- `.planning/phases/008-project-session-design/CONTEXT.md`
- `.planning/phases/008-project-session-design/PLAN.md`

## Learnings

- The command tree is clearer when scope is explicit: `setup` for global onboarding, `vault` for vault lifecycle, `secret` for direct secret records, and `project` for `.shelf.json` workflows.
- Because the project is unpublished, removing ambiguous top-level commands is cleaner than keeping aliases.
- `vault status` provides immediate UX value and validates the new vault namespace without revealing values.

## Ship Inputs

- `go test ./...` passed after command hierarchy and vault status changes.
- Old top-level commands are intentionally absent.
