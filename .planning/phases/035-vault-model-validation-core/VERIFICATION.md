# Verification: Vault Model and Validation Core

## Claims Checked

- The v2 model package exists behind its own internal boundary.
- Built-in `generic_secret` and `api_credential` templates are available.
- Valid built-in-template items pass validation.
- Invalid model states are rejected with focused validation errors.
- Runtime indexes support item ID, exact title, alias, and env binding lookup.
- Existing v1 vault/project/app/CLI behavior still builds and tests.
- Architecture lint accepts the new component boundary.

## Evidence Observed

- `internal/vaultmodel` contains model types, built-in templates, validation, and indexes.
- `internal/vaultmodel/vaultmodel_test.go` covers valid generic secret and API credential examples, validation failures, and index lookups.
- `.go-arch-lint.yml` includes `vaultmodel` as an internal component.
- `internal/cli/project_configure.go` no longer imports `internal/project` directly.

Commands run:

```text
go test ./internal/vaultmodel
go test ./internal/vault ./internal/project ./internal/app ./internal/cli ./internal/vaultmodel
go test ./...
go vet ./...
go-arch-lint check
```

All commands passed.

## Coverage

- `V2-IMPL-01`: Covered by `internal/vaultmodel` model types, built-in templates, registry, and indexes.
- `V2-IMPL-02`: Covered by `internal/vaultmodel` validation and focused tests for duplicate IDs, unknown templates, missing fields, invalid field values, invalid env bindings, duplicate env names, and public-field env bindings.

## Gaps

- v2 JSON codec/repository is not implemented; this is Phase 36.
- Legacy migration is not implemented; this is Phase 37.
- Secret CLI, project manifest v2, and manager cutover are not implemented; they are later phases.
- Item ID generation remains unresolved by design.

## Result

Pass. Phase 35 acceptance criteria are satisfied and work can proceed to Phase 36: JSON Codec and Repository Boundary.
