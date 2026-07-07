# Verification: Legacy Migration Adapter

## Claims Checked

- v1 flat secrets convert to v2 `generic_secret` items.
- Converted items preserve values, env bindings, descriptions, tags, aliases, and deterministic ordering.
- API credential templates are not inferred automatically.
- Invalid v1 vault data fails conversion.
- v1 project manifest path entries convert to value-free item/field/env bindings.
- Prefix and tag entries expand deterministically against the migration mapping.
- Empty prefix/tag selectors produce value-free warnings.
- Converted v2 project manifests validate and reject invalid env names and duplicates.
- Existing packages still pass tests and architecture lint.

## Evidence Observed

- `internal/vaultmigrate/vaultmigrate_test.go` covers conversion and validation contracts with fake secret values only.
- `ConvertVault` calls `vault.ValidateSecret`, `vaultmodel.Validate`, and creates deterministic item IDs from hashed paths.
- `ConvertProjectManifest` calls `project.Validate` and validates converted v2 bindings before returning.
- `.go-arch-lint.yml` registers `vaultmigrate` with dependencies on `project`, `vault`, and `vaultmodel`.

Commands run:

```text
go test ./internal/vaultmigrate
go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate
go test ./...
go vet ./...
go-arch-lint check
```

All commands passed.

## Coverage

- `V2-IMPL-04`: Covered by `internal/vaultmigrate` conversion logic and focused tests for value/metadata preservation, project manifest binding conversion, validation, and warning behavior.

## Gaps

- Migration is not exposed through CLI yet.
- Converted project manifests are not used by project export/run yet.
- v2 storage encryption remains deferred.

## Result

Pass. Phase 37 acceptance criteria are satisfied and work can proceed to Phase 38: Secret CLI v2 Workflows.
