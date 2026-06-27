# Plan: Phase 21 Script Workflow Consolidation

## Objective

Consolidate install, tag, and release-preparation workflows into reusable scripts under `scripts/` and reduce `justfile` to stable task aliases.

## Scope

- Add Bash scripts under `scripts/` for local install and release preparation.
- Keep release preparation under one script command surface with subcommands for check, snapshot, and tag.
- Update `justfile` to delegate relevant recipes to those scripts.
- Keep script usage self-describing enough for Phase 22 documentation.
- Do not add a generic workflow-check wrapper that only performs shallow checks.
- Do not update release docs/changelog beyond phase artifacts.
- Do not publish/tag a real release in the project repository.

## Tasks

1. Create `scripts/install.sh` preserving current `just install` behavior.
2. Create `scripts/release.sh` with `check`, `snapshot`, and `tag <version>` subcommands.
3. Update `justfile` recipes to call scripts.
4. Run focused script verification and existing product tests.
5. Record summary, verification, and capture artifacts.

## Acceptance Criteria

- `scripts/install.sh` installs `./cmd/shelf`, writes zsh completion, and reports installed paths.
- `scripts/release.sh check` runs the same GoReleaser config check currently exposed by `justfile`.
- `scripts/release.sh snapshot` runs the same GoReleaser snapshot release currently exposed by `justfile`.
- `scripts/release.sh tag <version>` accepts a semantic version-like argument and creates `v<version>` tags through Git.
- Missing or invalid release subcommands and tag arguments fail before creating tags.
- `just install`, `just tag <version>`, `just release-check`, and `just release-snapshot` delegate to scripts.
- Script behavior is testable without writing to the maintainer's real home directory or creating real release tags.

## Verification

- `bash -n scripts/*.sh`
- Focused checks for install script with temporary output paths.
- Focused checks for release script missing/invalid subcommands and tag argument failures.
- Focused tag creation check in a disposable Git repository.
- `go test ./...`
- `go vet ./...`
- GoReleaser config check and snapshot release through `scripts/release.sh`.

## Risks

- Running install directly can modify a developer's real Go bin/completion paths; verification must use temporary overrides.
- Running tag helpers directly can mutate Git refs; verification must use validation-only failure paths or a disposable repo if success behavior is checked.
- Snapshot release is slower and produces ignored `dist/` artifacts; clean up after verification.
