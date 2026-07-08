# Context: Manager Adaptation and Cutover

## Phase

Phase 40: Manager Adaptation and Cutover

## Requirement

- `V2-IMPL-07`: The local manager is adapted or explicitly gated for v2 items after CLI/core behavior is stable, preserving value-free list/search and explicit reveal/copy semantics.

## Current State

- Phase 38 moved primary `secret` CLI workflows to v2 item/field JSON storage.
- Phase 39 moved project status/export/run to v2 project bindings while preserving v1 manifests.
- The local manager still opens a v1 encrypted vault through `App.OpenManager -> LoadVault -> NewSecretService`.
- The manager service exposes v1 path/env/description/tag records:
  - list/search metadata without values;
  - detail metadata without values;
  - explicit reveal endpoint;
  - create/update/delete over v1 `group:key` secrets.
- Manager UI labels and forms are v1-shaped:
  - `Path` input such as `providers/openai:token`;
  - `Env name`, `Description`, `Tags`, raw `Value` textarea.
- V2 vault files use Phase 38 path behavior:
  - explicit `--vault`: exact path;
  - no explicit `--vault`: `<resolved-vault-path>.v2.json`.

## Product Decision

Use explicit gating for Phase 40, not full manager adaptation.

Rationale:

- Full manager adaptation requires a different object editor for template items/fields, public env-name fields, secret fields, profile defaults, and field-sensitive reveal/copy. Reusing the v1 path editor would corrupt the v2 product model.
- A gate satisfies `V2-IMPL-07` by preventing users from opening a v1-shaped editor when the active vault has v2 item data.
- CLI v2 secret/project workflows are already stable enough for local use; manager UX can be rebuilt as a later dedicated v2 item editor.

## Safety Requirements

- Do not silently open or mutate the v1 encrypted vault when a matching valid v2 vault is present.
- Do not expose v2 secret values in manager gating errors, status messages, docs, or tests.
- Preserve the existing v1 manager behavior when no valid v2 vault is present.
- Preserve loopback/token/no-store/security behavior in `internal/manager`.
- Keep manager package independent from `internal/app`, `internal/vaultjson`, and `internal/vaultmodel`; gating belongs in `internal/app` before opening the manager runtime.

## Scope

In scope:

- Detect whether a valid v2 vault exists for the resolved manager vault path.
- Return a clear value-free error from `shelf manager` when the resolved v2 vault exists.
- Add app/CLI tests that prove the gate and v1 preservation.
- Update docs to describe the temporary manager gate.
- Mark Phase 40 and `V2-IMPL-07` complete once verified.

Out of scope:

- Building a v2 manager UI.
- Adding v2 manager API endpoints.
- Public v1-to-v2 migration command.
- Age-wrapping v2 JSON vaults.
- Changing `secret` or project CLI behavior.

## Acceptance Criteria

- With no resolved valid v2 vault, `shelf manager` keeps using the existing v1 manager path.
- With a resolved valid v2 vault, `shelf manager` fails before opening a v1 manager service.
- Gate errors mention the v2 vault path and explain that the local manager is not yet available for v2 vaults.
- Gate errors contain no secret values, item fields, or env values.
- Manager server tests still pass.
- App/CLI tests cover v2 gating and v1 manager preservation.
- User docs state that v2 users should use CLI workflows until the v2 manager is adapted.
