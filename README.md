# Shelf Go

Shelf Go is a rewrite of the Python Shelf MVP as a smaller, faster local secret manager.

MVP focus:

```text
Fast local secret management and direct export.
```

Project-aware workflows are being added incrementally after the secret-manager core stabilized.

## Specs

- [Usage spec](docs/usage-spec.md)
- [Data spec](docs/data-spec.md)
- [Roadmap](docs/roadmap.md)

## Current command surface

```bash
shelf secret add [path-or-group]
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>
shelf secret rm <path>

shelf export <path-or-prefix> --format shell|env|json

shelf init
shelf doctor
shelf completion zsh

shelf project id
shelf project init
shelf project explain
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project export --format env|shell|json

shelf run -- command args...
shelf run --dry-run -- command args...
```

## Status

- v0.1 foundation implemented (store, secret CRUD, export, init, completion, project id).
- Post-v0.1 hardening implemented (write-side locking, `doctor`).
- v0.2 foundation implemented (`.shelf.json`, `project init`, `project explain`).
- v0.3 project binding management implemented (`project add/rm/list/export`).
- v0.4 runtime injection implemented (`run`, `run --dry-run`).
- Secret interactive add implemented (`secret add`).

