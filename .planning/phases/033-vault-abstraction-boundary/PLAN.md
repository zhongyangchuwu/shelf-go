# Plan: Vault Abstraction Boundary

## Objective

Make `internal/vault` the application-facing vault abstraction and make `internal/jsonvault` a concrete encrypted JSON implementation. `internal/app` should not import or name `internal/jsonvault`.

## Decisions

- `internal/vault` owns the repository contract used by application workflows.
- Encryption is part of the vault contract through open options: repositories are opened with recipients and identity paths, but `vault` does not know the encryption algorithm or file format.
- `internal/jsonvault` implements `vault.Opener` and `vault.Repository`.
- `cmd/shelf` is the composition root: it wires `app.New(jsonvault.Opener{})` into `cli.NewRootCmd`.
- `internal/cli` depends on `internal/app` only, not `internal/jsonvault`.
- Package-level app functions should be replaced by methods where CLI calls application behavior.

## Target Graph

```text
cmd/shelf -> cli, app, jsonvault
cli -> app, manager
manager -> app
app -> config, project, secret, vault, importer/gopass, util
project -> vault, util
secret -> vault
jsonvault -> vault, age, util, flock
```

No `app -> jsonvault` and no `cli -> jsonvault`.

## Implementation Steps

1. Add `vault.Options`, `vault.Repository`, `vault.Opener`, and status report types to `internal/vault`.
2. Make `jsonvault.Vault` implement `vault.Repository`; add `jsonvault.Opener`.
3. Change `internal/app` to `type App struct { vaults vault.Opener }` and route runtime/vault operations through methods.
4. Update CLI constructors to accept an `*app.App` service and call methods instead of package-level app functions.
5. Wire the concrete implementation in `cmd/shelf/main.go` with `jsonvault.Opener{}`.
6. Move app status handling to vault-level report types while keeping jsonvault-specific file checks behind the opener/repository boundary.
7. Update architecture docs and arch lint.
8. Run targeted tests and full verification.

## Acceptance

- `internal/app` has no import of `internal/jsonvault`.
- `internal/cli` has no import of `internal/jsonvault`.
- `internal/jsonvault` imports `internal/vault` and implements its repository interface.
- `cmd/shelf` is the only production composition root that wires jsonvault into app/cli.
- Existing CLI behavior, project/run workflows, setup, status, migration, import, and manager tests pass.
- `./scripts/test.sh` passes and arch lint reports no warnings.
