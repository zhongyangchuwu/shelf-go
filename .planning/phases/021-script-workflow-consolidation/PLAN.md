# Plan: Phase 21 Script Workflow Consolidation

## Objective

Consolidate install, tag, and release-preparation workflows into reusable scripts under `scripts/` and reduce `justfile` to stable task aliases.

## Scope

- Add Bash scripts under `scripts/` for local install, release config check, snapshot release, and release tag creation.
- Update `justfile` to delegate relevant recipes to those scripts.
- Add lightweight script checks that validate script syntax and argument behavior.
- Keep script usage self-describing enough for Phase 22 documentation.
- Do not update release docs/changelog beyond phase artifacts.
- Do not perform release hardening or publish/tag a real release.

## Tasks

1. Create `scripts/install.sh` preserving current `just install` behavior.
2. Create release helper scripts for GoReleaser check and snapshot release.
3. Create a tag helper script with usage validation and stable `v<version>` behavior.
4. Update `justfile` recipes to call scripts.
5. Add script-focused checks for syntax and safe argument validation.
6. Run focused script verification and existing product tests.
7. Record summary, verification, and capture artifacts.

## Acceptance Criteria

- `scripts/install.sh` installs `./cmd/shelf`, writes zsh completion, and reports installed paths.
- Release helper scripts run the same GoReleaser commands currently exposed by `justfile`.
- Tag helper script accepts a single semantic version-like argument and creates `v<version>` tags through Git.
- Missing or invalid tag arguments fail before creating tags.
- `just install`, `just tag <version>`, `just release-check`, and `just release-snapshot` delegate to scripts.
- Script behavior is testable without writing to the maintainer's real home directory or creating real release tags.

## Verification

- `bash -n scripts/*.sh`
- Focused checks for install script with temporary output paths.
- Focused checks for tag script missing/invalid argument failures.
- `go test ./...`
- `go vet ./...`
- GoReleaser config check through the new script.

## Risks

- Running install directly can modify a developer's real Go bin/completion paths; verification must use temporary overrides.
- Running tag helpers directly can mutate Git refs; verification must use validation-only failure paths or a disposable repo if success behavior is checked.
- Snapshot release is slower and may produce dist artifacts; keep it as a Phase 23 release-hardening gate unless needed for script wiring confidence.
