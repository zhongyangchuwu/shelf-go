# Phase 28 Summary

## Completed
- Added `.go-arch-lint.yml` using a light adapter/app/domain boundary for current internal packages.
- Moved project command orchestration into `internal/app/project.go`; `internal/cli/project.go` now keeps command construction, flags, completions, and output routing.
- Moved secret command use cases into `internal/app/secret.go` and interactive add workflow into `internal/secret/add.go`.
- Removed production `internal/manager` imports of `vault` and `exportfmt`; manager routes now call `app.ManagerService`.
- Moved config-aware vault status/doctor orchestration into `internal/app/status.go`; `internal/vault/status.go` is config-free.
- Removed `internal/cli` imports of `config`, `vault`, `secret`, and `exportfmt`.

## Final Boundary
```text
cli -> app, project, manager
manager -> app
app -> config, vault, project, secret, exportfmt, manager
project -> vault, exportfmt
secret -> vault
exportfmt -> vault
vault -> external age/flock only
config -> external yaml only
```

## Behavior
No intended user-visible command, flag, output format, vault format, or manifest schema changes.
