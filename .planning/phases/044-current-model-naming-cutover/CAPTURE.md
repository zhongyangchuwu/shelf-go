# Capture: Current-Model Naming Cutover

## Decisions

- `V2` was removed from active Shelf code because the former v1 stack is gone.
- No aliases or compatibility wrappers remain; all internal callers use current names.
- Persisted `Version` fields remain. Schema versioning is a file-format safety contract, not active v1 compatibility.
- `Binding` represents the persisted project binding. `ResolvedBinding` represents a materialized runtime binding and avoids an ambiguous type name.

## Next Work

Phase 45 consolidates selector, primary-field, field, and env-field rules into `internal/vaultmodel`.
