# Capture: Phase 21 Script Workflow Consolidation

## Durable Docs Updated

- None. Public documentation is intentionally deferred to Phase 22, which will document the script workflows alongside Web manager and architecture cleanup.

## Planning Records Updated

- `.planning/PROJECT.md` marks script workflow consolidation as validated.
- `.planning/REQUIREMENTS.md` marks OPS-01..OPS-03 complete.
- `.planning/ROADMAP.md` marks Phase 21 complete and links the Phase 21 plan.
- `.planning/STATE.md` advances to Phase 22.
- Phase 21 records added under `.planning/phases/021-script-workflow-consolidation/`.

## Learnings

- `justfile` can stay useful as a task index while scripts carry workflow logic.
- Install verification needs overrideable paths to avoid writing completions into the maintainer's real home directory.
- Tag workflow verification is safest in a disposable Git repository so it does not mutate real release refs.
- Release preparation should be one script command surface because check, snapshot, and tag are used together.
- Generic script-check wrappers are weak unless they exercise real workflow behavior; direct verification of concrete script commands is clearer.
- GoReleaser snapshot verification remains fast enough for local release-prep confidence.

## Ship Inputs

- Release notes can mention maintenance workflow cleanup only if desired; this is primarily maintainer-facing.
- Phase 22 should document `scripts/install.sh` and `scripts/release.sh` in contributor/developer docs.
- Phase 23 can call `scripts/release.sh check` and `scripts/release.sh snapshot` for release readiness checks.
