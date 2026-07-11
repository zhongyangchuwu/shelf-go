# Capture: Scriptable Item Lifecycle

## Durable Design

`secret update` is the single scriptable workflow for changing value-free item identity and organization plus one typed public field:

- title;
- additive/removal label changes;
- additive/removal alias changes;
- a selected public field via repeatable `--value`, or explicit `--clear-field`.

Secret fields remain exclusively on `secret set --stdin`. This preserves the no-secret-in-argv rule while allowing public metadata to be maintained without editing vault JSON.

The app resolves the target once, performs every requested mutation against an in-memory copy, and saves the complete vault through the existing validated atomic repository. Conflicting item selectors, invalid public field values, and secret-field updates fail before persistence.

## Next Work

Phase 48 may consume the same `UpdateSecret` app workflow after establishing a value-free, read-only TUI catalog.
