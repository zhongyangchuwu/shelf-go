# Verification: Project Manifest v2 Bindings

## Claims Checked

- Version 2 project manifests validate concrete item/field/env bindings and reject invalid shapes.
- Version 1 project manifests still validate and resolve as before.
- V2 resolution materializes only explicitly bound secret fields for export/run.
- V2 missing item, missing field, missing value, non-secret field, and duplicate env-name diagnostics are value-free.
- Optional missing v2 bindings warn instead of failing resolution.
- `project status` reports v2 binding health without secret values.
- `project export` renders v2 bindings in env, shell, and JSON formats.
- `project run` injects v2 bound values into child processes and preserves child exit behavior.
- `project add` can bind v2 item fields and derive env names from public env-name fields.
- Architecture rules cover the new project-to-vaultmodel dependency.

## Evidence Observed

Focused tests:

```text
go test ./internal/project ./internal/app ./internal/cli
```

Full checks:

```text
go test ./...
go vet ./...
go-arch-lint check
```

All commands passed.

Test coverage added/updated:

- `internal/project/manifest_test.go`: v2 manifest validation, defaults, duplicate checks, round-trip save/load.
- `internal/project/resolve_v2_test.go`: v2 item/field/env resolution, optional warnings, duplicate env diagnostics, value-free diagnostics.
- `internal/app/project_v2_test.go`: app-level v2 export/status/run behavior over explicit temp v2 JSON vault paths.
- `internal/cli/project_test.go`: CLI-level v2 project add/export/run/status flows.

## Coverage

- `V2-IMPL-06`: Covered by project package tests, app tests, CLI tests, docs, and full verification.
- Value-free diagnostics: Covered by project resolver tests and app/CLI status/export tests using fake secret literals.
- V1 regression: Covered by existing project/export/run tests in `internal/app` and `internal/cli`, plus full package tests.

## Gaps

- `project init` still creates v1 manifests during cutover.
- No public migration command creates v2 manifests from v1 manifests yet.
- V2 repository encryption remains deferred.
- Local manager v2 behavior remains deferred to Phase 40.

## Result

Pass. Phase 39 acceptance criteria are satisfied with documented cutover gaps.
