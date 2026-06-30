# Plan: Graph Source Cleanup

## Objective

Clean up the new architecture graph after the vault naming pass by removing the obsolete `source` package and moving project-run orchestration out of the CLI layer.

## Decisions

- Delete `internal/source`; it existed for runtime backend pluggability that was removed by the gopass import pivot.
- Keep project resolution local-vault-first: `internal/project` accepts `*vault.Store` directly for manifest resolution and project entry building.
- Avoid `vault -> project`; the vault domain must not know project manifest concepts.
- Remove `cli -> project` by moving `shelf run` orchestration into `internal/app.ProjectRun`.
- Keep `internal/cli` thin: parse flags/args, pass streams/env to app, return app errors.

## Target Graph

```text
cli -> app, manager
manager -> app
app -> config, project, secret, vault, jsonvault, importer/gopass, util
project -> vault, util
secret -> vault
jsonvault -> vault, age, util, flock
age -> filippo.io/age
```

No `source` node. No `vault -> project`. No `cli -> project`.

## Implementation Steps

1. Move source concepts used by project resolution into project/vault boundaries:
   - `project.ResolveEntries(m, st *vault.Store)`;
   - `project.BuildEntry(st *vault.Store, req)`;
   - `vault.EnvName(path, secret)` for env derivation.
2. Remove `vault.Reader`, `app.LoadSecretReader`, and `internal/source`.
3. Move run command orchestration to `app.ProjectRun` and update CLI to call it.
4. Update `.go-arch-lint.yml` and `docs/architecture.md` to match the graph.
5. Run targeted tests, full script verification, arch lint, and LSP diagnostics.

## Acceptance

- `internal/source` is gone.
- `internal/vault` does not import `internal/project`.
- `internal/cli` does not import `internal/project`.
- Project explain/add/export/run still resolve local vault secrets correctly.
- `shelf run` still preserves child exit codes.
- `./scripts/test.sh` passes and arch lint reports no warnings.
