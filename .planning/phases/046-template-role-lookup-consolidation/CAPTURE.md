# Capture: Template Role Lookup Consolidation

## Durable Design

`vaultmodel.Registry` owns template-role lookup:

- `PrimarySecretField(templateID)` resolves the template field designated for primary-secret writes and default field selection.
- `EnvNameField(templateID, secretFieldID)` resolves the public env-name field associated with a secret field.

`vaultmodel.Resolver` delegates its instance-level role lookups to that registry. Secret creation uses the same registry API when constructing an item, so template role scans no longer exist in `internal/app`.

## Preserved Contracts

- Existing error strings for unknown templates and missing primary-secret roles are retained.
- An absent env-name role remains a non-error lookup result; `secret create --env` translates it to the established value-free command error.
- Vault JSON, CLI flags, output, and secret-materialization behavior are unchanged; valid custom vault templates now remain valid during item creation.

## Next Work

Choose only a clearly bounded local CLI UX improvement or a duplicate-logic refactor with an observable maintenance benefit.
