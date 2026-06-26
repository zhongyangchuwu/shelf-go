# Phase 21 Context: Script Workflow Consolidation

## Goal

Move common install, tag, and release-preparation workflows out of inline `justfile` recipes and remembered manual commands into reusable Bash scripts under `scripts/`, while keeping `justfile` as a thin task runner.

## Constraints

- Do not begin v0.1.1 release hardening in this phase.
- Do not change product behavior, CLI commands, Web manager code, secret/tag/project behavior, or vault storage format.
- Keep scripts boring Bash with `set -euo pipefail` and clear usage failures.
- Preserve current local install behavior: install `./cmd/shelf`, generate zsh completion, and print installed paths.
- Preserve current release-prep checks: GoReleaser config check and snapshot release.
- Preserve current tag behavior: create `v<version>` tags and remind maintainers to reinstall if they want embedded tag-derived versions locally.
- Leave broad user/developer documentation updates for Phase 22; Phase 21 script usage should be self-describing.

## Decisions

- Keep install as `scripts/install.sh` because it is a local developer setup workflow, separate from release management.
- Use one release workflow command surface, `scripts/release.sh`, with subcommands for `check`, `snapshot`, and `tag <version>` because those tasks are used together during release preparation.
- Do not keep a generic `workflow-check` script; it adds a weak command surface and does not test meaningful release capability.
- Keep `justfile` recipe names stable and delegate to scripts.
- Make install completion path overrideable for verification so tests do not write to the real home directory.
- Use `go run github.com/goreleaser/goreleaser/v2@latest ...` inside `scripts/release.sh` to preserve the existing toolchain behavior without requiring a separate binary install.

## Open Questions

- None.

## Verification Expectations

- Scripts pass Bash syntax checks.
- Install script can run with temporary `GOBIN` and completion directory overrides.
- Release script rejects missing/invalid subcommands and invalid tag versions before creating tags.
- Release script can create a tag in a disposable Git repository and reject duplicates.
- Release script can run GoReleaser check and snapshot flows.
- `justfile` recipes are thin delegations to scripts for install, tag, release-check, and release-snapshot.
- Existing Go test suite still passes because script consolidation should not affect product behavior.
