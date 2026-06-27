# Plan: Phase 24 v0.1.1 Release Hardening

## Objective

Make v0.1.1 release-ready by updating the changelog, running the consolidated release checks, and recording final verification evidence.

## Non-Goals

- Do not publish or tag the release in this phase.
- Do not change storage format or add SQLite/storage backend work.
- Do not add new command behavior unless a release check exposes a blocking issue.
- Do not create compatibility aliases for removed pre-release commands.

## Work Items

1. Changelog
   - Add `## 0.1.1 - 2026-06-27` under `Unreleased`.
   - Summarize user-facing manager, tag workflow, project binding, script, docs, and architecture command naming changes.
   - Preserve the existing `0.1.0` notes.

2. Verification
   - Run `go test ./...`.
   - Run `go vet ./...`.
   - Run `./scripts/release.sh check`.
   - Run `./scripts/release.sh snapshot`.
   - Inspect failures and fix only release-blocking issues.

3. Boundary checks
   - Search active code/docs for unsupported `secret meta`, `secret tag`, SQLite implementation claims, and old manager command usage.
   - Confirm docs/planning state still say SQLite/storage redesign is deferred to v0.2.0.

4. Records
   - Write `SUMMARY.md`, `VERIFICATION.md`, and `CAPTURE.md` for Phase 24.
   - Update root planning requirements, roadmap, project, and state.
   - Commit phase hardening work.

## Acceptance Criteria

- Changelog includes v0.1.1 user-facing changes.
- All planned checks pass through consolidated commands where applicable.
- Phase records include exact observed evidence and known gaps.
- `REL-011-01`, `BOUND-01`, and `BOUND-02` are satisfied.
- Working tree is clean after commit.
