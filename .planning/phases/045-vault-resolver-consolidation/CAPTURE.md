# Capture: Vault Resolver Consolidation

## Durable Design

`internal/vaultmodel.Resolver` is the canonical owner of:

- selector resolution and ambiguity diagnostics;
- exact item-ID lookup;
- env-name lookup and ambiguity diagnostics;
- primary-secret-field lookup;
- field lookup/default selection;
- env-name lookup for a secret field;
- secret-field completion selector generation.

`Binding` is persisted manifest data. `ResolvedBinding` contains materialized runtime values.

## Performance

`NewResolver` validates direct callers. `NewResolverFromValidatedVault` reuses the model registry/index construction for repository-loaded or otherwise known-valid vaults, avoiding duplicate validation in CLI workflows.

## Next Work

Continue CLI UX only when it creates a clear user-facing gain. Keep current model terminology and the shared resolver as the only active paths.
