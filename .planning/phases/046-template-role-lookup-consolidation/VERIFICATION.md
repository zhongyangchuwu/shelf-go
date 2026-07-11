# Verification: Template Role Lookup Consolidation

## Automated

Passed:

```text
go test ./...
go vet ./...
go-arch-lint check
go build ./cmd/shelf
```

Coverage includes:

- built-in primary-secret and associated env-name role lookup;
- an absent env-name role returning `found=false` without error;
- unknown-template and no-primary-secret diagnostics;
- secret creation with generic, API, and custom vault templates; profile defaults, stdin secret values, env names, and metadata;
- project binding and secret workflow regression coverage.

## Structural Check

`internal/app` and `internal/project` no longer contain template-role scan helpers (`primaryFieldTemplateForTemplate`, `setEnvNameField`, or `findField`). `Registry` is the single implementation of template primary-secret and env-name role lookup.
