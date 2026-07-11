# Plan: Template Role Lookup Consolidation

## Goal

Make `internal/vaultmodel.Registry` the canonical owner of template-role lookups used by both item construction and vault resolution.

## Scope

- Add registry lookups for a template's primary secret field and an env-name field targeting a secret field.
- Refactor `vaultmodel.Resolver` and app secret creation to use those lookups.
- Delete app-local template role scanning helpers.
- Preserve creation of items based on custom vault templates by relying on the full-vault repository validation.
- Preserve CLI, vault JSON, error, and secret-materialization contracts.

## Acceptance

- A primary-secret lookup and env-name-field lookup have exactly one implementation.
- `CreateSecret`, `SetSecretField`, `GetSecretField`, `secret export`, and project binding resolution retain their existing behavior.
- Missing template roles return value-free errors.
- `secret create --template <custom-template>` saves a valid custom-template item.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
