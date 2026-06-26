# Summary: Phase 21 Script Workflow Consolidation

## Completed Changes

- Added reusable Bash workflow scripts under `scripts/`:
  - `scripts/install.sh`
  - `scripts/release-check.sh`
  - `scripts/release-snapshot.sh`
  - `scripts/tag-release.sh`
  - `scripts/check-workflows.sh`
- Updated `justfile` so install, tag, release-check, release-snapshot, and workflow-check are thin script delegations.
- Preserved existing install behavior while making completion output paths overrideable for safe verification.
- Preserved GoReleaser check and snapshot release commands behind scripts.
- Added tag argument validation and duplicate tag detection before tag creation.
- Updated root planning artifacts to mark OPS-01..OPS-03 and Phase 21 complete.

## Files Changed

- `justfile`
- `scripts/install.sh`
- `scripts/release-check.sh`
- `scripts/release-snapshot.sh`
- `scripts/tag-release.sh`
- `scripts/check-workflows.sh`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/021-script-workflow-consolidation/CONTEXT.md`
- `.planning/phases/021-script-workflow-consolidation/PLAN.md`
- `.planning/phases/021-script-workflow-consolidation/SUMMARY.md`
- `.planning/phases/021-script-workflow-consolidation/VERIFICATION.md`
- `.planning/phases/021-script-workflow-consolidation/CAPTURE.md`

## Deviations

- Added `scripts/check-workflows.sh` and a `just workflow-check` alias to keep script validation reusable.
- Ran a snapshot release script during verification even though the plan allowed deferring it; the resulting `dist/` output is ignored and not committed.
- Did not update public developer docs; Phase 22 owns docs and architecture cleanup.

## Evidence

- `./scripts/check-workflows.sh` passed.
- `./scripts/release-check.sh` passed.
- `./scripts/release-snapshot.sh` passed.
- `just --dry-run install release-check release-snapshot tag 0.1.1 workflow-check` showed script delegation.
- Install script verification with temporary `GOBIN` and completion directory created the binary and completion file.
- Tag script verification in a disposable Git repo created `v0.1.1` and rejected the duplicate tag.
- `go test ./...` passed.
- `go vet ./...` passed.

## Unresolved Risks

- Phase 22 still needs to document the new scripts for maintainers.
- Phase 23 still needs final release hardening and release readiness evidence.
