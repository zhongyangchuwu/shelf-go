# Capture: Phase 23 Documentation and Usage Alignment

## Durable Docs Updated

- `README.md`: primary v0.1.1 flows now include manager editing, direct tag selection, project tag bindings, and script workflows.
- `docs/getting-started.md`: step-by-step guide now includes tag examples and manager editing scope.
- `docs/reference.md`: command reference now matches implemented tag and manager command behavior.
- `docs/troubleshooting.md`: required tag binding and prefix/tag env override troubleshooting added.
- `docs/contributing.md`: install/release scripts and thin `justfile` wrappers documented.
- `docs/architecture.md`: final package layout and dependency direction documented after repartition.

## Durable Decisions

- `shelf manager` is the only local manager entrypoint documented for users.
- Tags use repeatable AND semantics in direct secret commands and project bindings.
- Project tag bindings are value-free manifest selectors and cannot carry `env` overrides.
- The manager is the full-object editing surface; no fine-grained `secret meta` or `secret tag` CLI group is introduced.
- v0.1.1 keeps the age-encrypted JSON vault format and defers SQLite/storage redesign to v0.2.0.
- Release readiness is now unblocked for Phase 24.

## Next Consumer

Phase 24 release hardening should use these docs as the public behavior baseline when preparing the changelog, release checks, and snapshot evidence.
