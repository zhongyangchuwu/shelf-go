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

- Add separate scripts for install, release config check, snapshot release, and release tag creation instead of one multiplexed script.
- Keep `justfile` recipe names stable and delegate to scripts.
- Make install completion path overrideable for verification so tests do not write to the real home directory.
- Use `go run github.com/goreleaser/goreleaser/v2@latest ...` inside scripts to preserve the existing toolchain behavior without requiring a separate binary install.

## Open Questions

- None.

## Verification Expectations

- Scripts pass Bash syntax checks.
- Install script can run with temporary `GOBIN` and completion directory overrides.
- Tag script rejects missing/invalid versions before creating tags.
- `justfile` recipes are thin delegations to scripts for install, tag, release-check, and release-snapshot.
- Existing Go test suite still passes because script consolidation should not affect product behavior.
