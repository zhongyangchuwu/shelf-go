# Shelf Go

Shelf Go is a rewrite of the Python Shelf MVP as a smaller, faster local secret manager.

MVP focus:

```text
Fast local secret management and direct export.
```

Project-aware workflows are intentionally deferred until the secret-manager core is useful and stable.

## Specs

- [Usage spec](docs/usage-spec.md)
- [Data spec](docs/data-spec.md)

## Planned MVP command surface

```bash
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>

shelf export <path-or-prefix> --format shell|env|json

shelf project id
```

## Status

Design/spec bootstrap only. Implementation has not started.
