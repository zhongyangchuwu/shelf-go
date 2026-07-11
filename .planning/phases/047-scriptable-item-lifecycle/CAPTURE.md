# Capture: Scriptable Item Lifecycle

## Durable Design

`secret update` is the single scriptable workflow for changing item identity, organization, or a field:

- title;
- additive/removal label changes;
- additive/removal alias changes;
- a selected public field via repeatable `--value`, explicit `--clear-field`, or stdin;
- a selected secret field through stdin only.

Field writes default to the template primary secret field when no other update input is present. Interactive terminal input is masked; piped stdin stays scriptable. Secret values never enter argv.

The app resolves the target once, performs every requested mutation against an in-memory copy, and saves the complete vault through the existing validated atomic repository. Conflicting item selectors, invalid public field values, and secret field values passed by `--value` fail before persistence.

## Next Work

Phase 48 may consume the same `UpdateSecret` app workflow after establishing a value-free, read-only TUI catalog.
