# Context: Phase 6 Command Hierarchy Cutover

## Goal

Replace ambiguous pre-release top-level commands with scoped namespaces before public compatibility exists.

## Decisions

- No backward-compatible command aliases. The project is not published, so clean cutover is preferred.
- `shelf setup` owns global onboarding: config path, vault path, age recipient, identity path, and initial encrypted vault creation.
- `shelf vault init` owns explicit vault/config initialization for users who think in vault lifecycle terms.
- `shelf vault migrate` owns plaintext-to-encrypted vault migration.
- `shelf vault open` owns the local vault manager entrypoint.
- `shelf secret export` owns direct path/prefix export from the vault.
- `shelf project run` owns `.shelf.json`-based runtime injection.
- Future `activate`, `deactivate`, and `shell` commands stay under `shelf project`, but are not implemented in this phase.

## Risks

- Command tests and docs can keep stale top-level command names.
- Cobra completion and root command assertions must reflect the new root shape.
- Refactor should not change vault storage, project resolution, rendering, manager security, or child exit behavior.
