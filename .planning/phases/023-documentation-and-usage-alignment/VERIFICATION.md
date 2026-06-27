# Verification: Phase 23 Documentation and Usage Alignment

## Claims Checked

1. User docs describe manager editing, direct tag selection, and project tag bindings.
2. Command reference matches implemented CLI flags for manager, secret tag selection, and project tag bindings.
3. Developer docs describe consolidated install/release scripts instead of old manual release commands.
4. Architecture docs describe final Phase 22 package layout.
5. Docs preserve scope boundaries: no fine-grained CLI metadata edit commands and no SQLite/storage redesign in v0.1.1.

## Evidence Observed

- `README.md`
  - Contains tag examples for `secret list`, `secret export`, and `project add --tag`.
  - Contains `shelf manager` as the manager command and states reveal/copy/edit plaintext boundaries.
  - Contains script workflow examples for `scripts/install.sh` and `scripts/release.sh`.
- `docs/getting-started.md`
  - Contains tag-based list/export and project tag binding examples.
  - States repeated tags use AND semantics and `secret list` is value-free.
  - Describes manager search/add/edit/rename/delete/reveal/copy/tag flows.
- `docs/reference.md`
  - Documents `shelf manager [--addr 127.0.0.1:0]`.
  - Documents `shelf secret list [prefix] [--tag TAG ...]`.
  - Documents `shelf secret export [path-or-prefix] --format shell|env|json [--all] [--tag TAG ...]`.
  - Documents `shelf project add --tag TAG [--tag TAG ...] [--optional]`.
  - Documents `.shelf.json` `tags` selector entries.
- `docs/contributing.md`
  - Documents `./scripts/install.sh`.
  - Documents `./scripts/release.sh check`, `snapshot`, and `tag`.
  - States matching `just` recipes are thin wrappers.
- `docs/architecture.md`
  - Describes `internal/app`, `internal/project`, `internal/secret`, `internal/config`, `internal/vault`, `internal/exportfmt`, `internal/manager`, and dependency direction.
  - States `shelf manager` is the public entrypoint and there is no old alias.
  - States v0.1.1 keeps JSON inside an age-encrypted vault and defers SQLite/storage redesign.
- `go test ./...`
  - Result: passed.
- LSP workspace diagnostics
  - Result: no Go issues found.
- Stale-reference grep over `README.md`, `docs`, `internal`, and `cmd`
  - No active user instruction recommends old internal package paths or old manual GoReleaser commands.
  - Remaining `shelf vault open` and `secret meta` matches are explicit boundary statements.

## Result

Passed. DOC-01 and DOC-02 are satisfied.
