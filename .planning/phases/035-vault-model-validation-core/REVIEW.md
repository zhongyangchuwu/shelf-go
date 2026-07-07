# Review: Vault Model and Validation Core

## Scope Reviewed

- `internal/vaultmodel/model.go`
- `internal/vaultmodel/templates.go`
- `internal/vaultmodel/validate.go`
- `internal/vaultmodel/index.go`
- `internal/vaultmodel/vaultmodel_test.go`
- `.go-arch-lint.yml`
- `internal/app/project.go`
- `internal/cli/project_configure.go`
- `internal/cli/project_test.go`

## Findings

- No blocking findings.
- `internal/vaultmodel` has no dependency on CLI/app/project/manager/v1 vault packages.
- Validation is pure and does not perform I/O or mutate persisted state.
- Built-in template metadata is enforced for known template fields.
- Secret values only appear as fake test fixtures; value-free runtime behavior is unchanged because no CLI path consumes v2 yet.
- Architecture lint initially failed because the new package was not registered and CLI imported `internal/project`; both were fixed.

## Fixes Applied

- Registered `vaultmodel` in `.go-arch-lint.yml`.
- Removed the CLI-level `internal/project` dependency by adding `ProjectConfigureEnvFile` and `ProjectConfigureEnvVar` to `internal/app` and projecting project env-file details there.
- Aligned built-in API credential template purposes with the model constants for provider and scopes.
- Removed duplicate test/template files created during parallel test authoring and kept one canonical registry/test surface.

## Waivers

- Custom template loading is not implemented; it is explicitly deferred by Phase 35 scope.
- Unknown fields on built-in templates are not rejected yet; the phase acceptance criteria focus on required built-in fields and metadata mismatches for known fields.

## Remaining Risks

- Future JSON decoding must decide whether unknown built-in fields are preserved, rejected, or scoped to custom templates.
- Future migration must define alias lifetime and deterministic migrated item IDs.
