# Plan: Phase 12 Restore Simplification

## Objective

Remove the restore command added in Phase 10 and keep backup recovery as explicit file operations plus status/doctor verification.

## Scope

In scope:

- Remove `newRestoreCmd`, restore helper, and restore tests.
- Remove restore command registration from `shelf vault`.
- Update README/reference/getting started/security/troubleshooting to remove restore command usage.
- Document `.bak` as a single-slot last-write encrypted backup.
- Update planning state and requirements.

Out of scope:

- Timestamped history.
- Dolt or database-backed storage.
- Backup prune/list/history commands.

## Tasks

1. Remove command and tests.
2. Update public docs to manual recovery.
3. Update planning artifacts to supersede VREC command requirement.
4. Run focused command/doc searches and `go test ./...`.

## Acceptance Criteria

- No `vault restore` command appears in code or public docs.
- Manual recovery docs show backing up current vault, copying `.bak` to active vault, and running `shelf vault status`.
- Docs state `.bak` is overwritten on each vault replacement.
- Full test suite passes.

## Verification

- Search for `vault restore` in code/docs.
- `go test ./...`

## Risks

- Removing a just-added command changes Phase 10 history; planning records must clearly state the supersession.
