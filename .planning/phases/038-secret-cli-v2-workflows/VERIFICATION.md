# Verification: Secret CLI v2 Workflows

## Claims Checked

- v2 model env names are public fields rather than `Field.Bindings`.
- Built-in profiles include `openai` over `api_credential` and validation rejects unsafe profile defaults.
- v1 migration writes env names to `value_env` public fields without moving secret values into public fields.
- v2 app service can create, set, get, list, info, and remove items.
- v2 CLI writes secret values only through stdin.
- `secret create --profile openai --stdin` applies registry defaults and does not hardcode provider behavior in CLI.
- `secret set --field value_env --stdin` edits env names as public fields.
- `secret get --env ENV` resolves through env-name fields and returns only the requested secret value.
- `secret list --long`, `secret info`, and duplicate env-name ambiguity errors do not print fake secret values.
- Removed legacy paths (`secret add`, `secret edit`, `secret export`) are absent in the v2 secret CLI.
- Architecture boundaries accept the v2 app service dependencies.

## Evidence Observed

- `internal/vaultmodel/vaultmodel_test.go` covers env-name validation, duplicate env names, profile lookup/default validation, and `FieldsByEnv` duplicate lookup.
- `internal/vaultmigrate/vaultmigrate_test.go` covers `value_env` migration, concealed value preservation, old path aliases, metadata preservation, and value-free path mapping.
- `internal/app/secret_v2_test.go` covers v2 app service create/profile/set/get/list/info/remove/default path behavior and value-free ambiguity errors.
- `internal/cli/secret_test.go` covers v2 CLI create/set/get/list/info/rm, stdin-only writes, profile defaults, env selector behavior, value-free list/info/ambiguity output, and removed legacy commands.
- Docs now describe v2 stdin writes, profiles, env-name public fields, and the plaintext Phase 38 v2 JSON cutover caveat.

Commands run:

```text
go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate ./internal/app ./internal/cli
go test ./internal/app -run 'Test.*V2'
go test ./...
go vet ./...
go-arch-lint check
```

All commands passed.

## Coverage

- `V2-IMPL-05`: Covered by v2 model, migration, app service, CLI tests, docs, and architecture lint.
- Value exposure rules: Covered by list/info/ambiguity tests in app and CLI packages using fake secret values.
- Profile behavior: Covered by model registry tests and app/CLI OpenAI profile tests.

## Gaps

- v2 storage remains plaintext JSON in Phase 38 and must not be committed or synced.
- Project manifest v2 export/run resolution is not implemented; Phase 39 owns it.
- Manager v2 adaptation/gating is not implemented; Phase 40 owns it.
- Clipboard copy is deferred.

## Result

Pass. Phase 38 acceptance criteria are satisfied, with documented storage and integration gaps routed to later phases.
