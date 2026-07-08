# Plan: Project Manifest v2 Bindings

## Objective

Add v2 project manifest item/field/env bindings and wire them into project status/export/run while keeping existing v1 project behavior working during cutover.

## Scope

In scope:

- Extend project manifest loading/validation to accept v1 and v2 manifests.
- Add v2 manifest binding model: item ID, field ID, env name, required flag, optional title.
- Add v2 project binding resolver over `vaultmodel`/`vaultjson`.
- Add v2 project app flows for status/export/run and project add/list/rm display where needed.
- Add CLI support for binding v2 items, including `project add --field FIELD <item-selector>` and `--env` override.
- Keep diagnostics value-free.
- Keep existing v1 project tests passing.

Out of scope:

- Public migration command.
- Manager adaptation.
- V2 age encryption.
- Prefix/tag selector support in new v2 manifests beyond migrated concrete bindings.
- Fuzzy item selection.

## Tasks

1. Inspect current project manifest, resolver, app, CLI, and tests.
2. Extend `internal/project` manifest model and validation for version 2 bindings.
3. Implement v2 binding resolution over `vaultmodel.Vault`:
   - item exists;
   - field exists;
   - field is secret-bearing for export/run;
   - env name validates;
   - duplicate env names fail value-free;
   - optional missing bindings warn without failing.
4. Wire app project status/export/run to dispatch v1 vs v2 manifests.
5. Wire project add/list/rm/configure enough for v2 manifest daily usage.
6. Add/adjust tests:
   - project package v2 validation/resolution;
   - app/CLI v2 status/export/run/add;
   - v1 regression coverage.
7. Update docs and planning artifacts.
8. Run verification and commit.

## Acceptance Criteria

- A version 2 `.shelf.json` can bind item/field/env entries without secret values.
- `shelf project status` reports v2 binding health without printing secret values.
- `shelf project export` materializes v2 bound secret fields in env/shell/json formats.
- `shelf project run` injects v2 bound secret fields into a child process and preserves child exit behavior.
- Missing items, missing fields, duplicate env names, and missing required values produce value-free diagnostics.
- Optional missing v2 bindings do not fail export/run/status.
- Existing v1 project export/run/status tests still pass.

## Verification

- `go test ./internal/project`
- `go test ./internal/app ./internal/cli`
- `go test ./...`
- `go vet ./...`
- `go-arch-lint check`

## Risks

- Mixed v1/v2 manifest loading can blur contracts if version dispatch is not explicit.
- V2 values are plaintext JSON until repository encryption lands; export/run must load the same v2 path semantics as Phase 38 secret commands.
- Keeping v1 compatibility while enabling v2 may leave temporary branching in app/project code; keep branches shallow and isolated.
