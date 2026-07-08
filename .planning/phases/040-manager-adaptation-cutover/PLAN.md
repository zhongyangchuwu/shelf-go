# Plan: Manager Adaptation and Cutover

## Goal

Gate the local manager when the resolved v2 vault is present and loadable so Shelf does not open a v1 path-shaped editor over a v2 item/field workflow.

## Design

Add the gate in `internal/app.OpenManager` before `LoadVault`:

1. Resolve the v2 vault path with existing Phase 38 behavior by reusing `v2VaultPath(configPathFlag, vaultPathFlag)`.
2. If a loadable v2 vault exists at that path, return a value-free error:
   - local manager is not available for v2 vaults yet;
   - include the v2 vault path;
   - tell users to use v2 CLI secret/project commands for now.
3. If no loadable v2 vault exists, continue the existing v1 manager flow unchanged.
4. If stat fails for any reason other than not-exist, return the stat error with context.

This keeps `internal/manager` unchanged. The manager server remains a v1 service surface until a later v2 item editor exists.

## Implementation Tasks

1. Add a small app helper to detect the resolved v2 vault file.
2. Call the helper from `App.OpenManager` before `LoadVault`.
3. Add app tests:
   - v2 vault file present -> `OpenManager` fails with value-free gate error;
   - no v2 vault file -> existing v1 manager open path still succeeds.
4. Add CLI test for `shelf manager` over an explicit v2 vault path or default v2 sidecar path that verifies the command fails before serving.
5. Update docs and changelog:
   - README/getting-started/reference/architecture/troubleshooting as needed;
   - portable vault notes if relevant;
   - changelog unreleased entry.
6. Run focused and full verification.
7. Complete phase artifacts and update project state.
8. Commit with Conventional-style message.

## Test Plan

Focused:

```bash
go test ./internal/app ./internal/cli ./internal/manager
```

Full:

```bash
go test ./...
go vet ./...
go-arch-lint check
```

## Acceptance Checks

- V2 gate triggers before v1 vault open.
- Existing v1 manager tests and server security tests still pass.
- Gate error is value-free.
- Docs describe the temporary v2 manager gate and CLI fallback.
