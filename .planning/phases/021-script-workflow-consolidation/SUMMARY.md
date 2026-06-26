# Summary: Phase 21 Script Workflow Consolidation

## Completed Changes

- Added reusable Bash workflow scripts under `scripts/`:
  - `scripts/install.sh`
  - `scripts/release.sh`
- Updated `justfile` so install, tag, release-check, and release-snapshot are thin script delegations.
- Preserved existing install behavior while making completion output paths overrideable for safe verification.
- Preserved GoReleaser check and snapshot release commands behind one release script command surface.
- Added release tag argument validation and duplicate tag detection before tag creation.
- Removed the weak generic workflow-check wrapper; verification now exercises concrete script behavior directly.
- Updated root planning artifacts to mark OPS-01..OPS-03 and Phase 21 complete.

## Files Changed

- `justfile`
- `scripts/install.sh`
- `scripts/release.sh`
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

- Collapsed the release/tag helpers into one `scripts/release.sh` after review showed separate `release-check`, `release-snapshot`, and `tag-release` scripts were unnecessary command surfaces.
- Removed `scripts/check-workflows.sh`; it mostly checked shell syntax and did not prove meaningful workflow capability.
- Ran a snapshot release script during verification even though the plan allowed deferring it; the resulting `dist/` output is ignored and not committed.
- Did not update public developer docs; Phase 22 owns docs and architecture cleanup.

## Evidence

- `bash -n scripts/*.sh` passed.
- `./scripts/release.sh --help` passed.
- `./scripts/release.sh` rejected a missing subcommand.
- `./scripts/release.sh tag v0.1.1` rejected a leading-`v` version.
- `./scripts/release.sh unexpected` rejected an unknown subcommand.
- `./scripts/release.sh check` passed.
- `./scripts/release.sh snapshot` passed.
- `just --dry-run install release-check release-snapshot tag 0.1.1` showed script delegation.
- Install script verification with temporary `GOBIN` and completion directory created the binary and completion file.
- Release tag verification in a disposable Git repo created `v0.1.1` and rejected the duplicate tag.
- `go test ./...` passed.
- `go vet ./...` passed.

## Unresolved Risks

- Phase 22 still needs to document the new scripts for maintainers.
- Phase 23 still needs final release hardening and release readiness evidence.
