# Verification: Current-Model Naming Cutover

Passed:

```text
go test ./...
go vet ./...
go-arch-lint check
```

Evidence:

- Active app and project APIs, tests, helper names, and filenames no longer use transitional `V2` naming.
- `internal/app/secret.go` and `internal/project/resolve.go` now name the only active model directly.
- `Manifest.Binding` names persisted project bindings; `ResolvedBinding` distinguishes materialized runtime values.
- `Vault.Version`, `Manifest.Version`, and `CurrentVersion` remain as persisted-schema contracts.
- The remaining `GenBashCompletionV2` identifier is Cobra's external API, not Shelf migration terminology.
