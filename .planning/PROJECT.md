# Shelf Go

## What This Is

Shelf Go is a personal, local, CLI-first secret manager. It stores item-based secrets in a local JSON vault and relies on external tooling such as chezmoi for encryption/sync.

## Current State

### Validated

- [x] v0.2.0 item model, JSON repository, secret CLI, label export, project item bindings, and item-vault run/export workflows are implemented.
- [x] Built-in age/v1 compatibility, v1 manager, v1 migration/import/export, and project configure TUI were removed for the personal-use CLI direction.
- [x] Local workstation config now points at `/home/han/.local/share/shelf/vault.json`, copied from the previous item sidecar after backup.
- [x] Human-facing secret list/create output is routed through `internal/cliui`.
- [x] `secret copy` provides clipboard-only interactive retrieval while `secret get` remains scriptable stdout output.
- [x] `secret update` maintains item title, labels, aliases, and public fields atomically while keeping secret writes stdin-only.
- [x] `shelf tui` provides a wide-screen secret workspace with safe catalog/filter/detail views, masked forms, clipboard copy, and app-backed mutations.
- [x] `shelf project tui` manages current-worktree bindings with safe resolution diagnostics, picker/form workflows, and app-owned manifest persistence.
- [x] `shelf secret audit` supplies value-free credential expiry signals in CLI, secret TUI, project TUI, and project status without changing export/run enforcement.

### Planned

- [x] Scriptable item lifecycle is complete.
- [x] Wide-screen secret workspace TUI is complete.
- [x] Project binding TUI is complete.
- [x] Credential health audit is complete.
- [x] Personal workflow UX roadmap is complete; a future milestone requires explicit planning.

### Out of Scope

- Built-in encryption; use chezmoi or another external tool.
- v1 compatibility and automatic migration.
- Web manager.
- Team sharing, hosted sync, browser autofill, or daemon unlock flows.
- Public repository/release packaging until a future v1.0.0 milestone.

## Constraints

- The vault JSON contains plaintext secret values; do not commit/sync it unless external tooling encrypts it.
- `.shelf.json` and config must remain value-free.
- Secret values enter through stdin for writes.
- Diagnostics, list/info/status, and completions must not print secret values.

## Package Boundaries

- `internal/cli`: Cobra command surface.
- `internal/cliui`: Lip Gloss-backed human output helpers for titles, summaries, empty states, and tables.
- `internal/tui`: optional Bubble Tea workspace adapter over `app` and `clipboard`; no direct vault, config, or project I/O.
- `internal/clipboard`: system clipboard boundary for the explicit `secret copy` workflow.
- `internal/app`: application workflows.
- `internal/vaultmodel`: item/field/template/profile model, validation, and shared item/field/env resolver.
- `internal/vaultjson`: JSON item-vault codec/repository.
- `internal/project`: Git project identity, manifest schema, binding resolution, child env helpers.
- `internal/config`: config and vault path resolution.
- `internal/util`: atomic write and env/shell/JSON binding formatting.

## Key Decisions

| Decision | Rationale | Outcome |
| --- | --- | --- |
| Use external encryption | User prefers chezmoi encryption and a simpler personal tool. | Removed age/jsonvault/v1 encryption code from active product. |
| Use item vault JSON as canonical file | It is easy to inspect, back up, and encrypt externally. | Default vault path is `~/.local/share/shelf/vault.json`. |
| Drop v1 compatibility | Personal-use project; compatibility burden slows local iteration. | Removed v1 vault, manager, import/migrate/export, project selectors. |
| Keep CLI authoritative before adding TUI | Scriptability and one canonical workflow path must survive interactive UX work. | Existing CLI remains the authority; optional TUI phases start only after scriptable lifecycle parity. |
| Defer public release work | Repository stays private/local until a future v1.0.0 milestone. | Removed v0.2.0 tag/release preparation from active scope. |
| Copy through an explicit command | Interactive secret retrieval should not break stdout-based scripting. | `secret copy` writes only to the clipboard and returns value-free confirmation. |
| Use a shared model resolver | Secret and project workflows must agree on item, field, and env-name semantics. | `vaultmodel.Resolver` is the only active owner of those rules. |
| Treat `.env` as a projection, not the primary model | A developer should maintain values and context once, then select fields for a project or command. | The vault is item-first; project bindings and exports are value-free runtime projections. |
| Add TUI after scriptable lifecycle parity | TUI must be an optional front-end, not the only way to edit or a second persistence path. | Phases 48–49 deliver secret and project binding workflows through existing app and clipboard boundaries. |
| Use expiry metadata as an advisory health signal | Local workflows need timely maintenance signals without hidden runtime enforcement or remote dependencies. | Phase 50 adds value-free audit status; export/run remain unchanged. |
| Keep vault backups at the configured restrictive mode | A `.bak` file contains the prior plaintext vault and is equally sensitive. | Atomic backup writes now repair the requested mode even when the backup already exists. |
| Use a fixed Catppuccin Mocha frame for TUI workspaces | The shared dark palette and visible two-panel frame make catalog/detail context legible without a new theme configuration surface. | Phase 52 applies rounded persistent panels, blue active borders, and text focus markers to secret and project workspaces. |

*Last updated: 2026-07-13 after Catppuccin Mocha panel refinement*
