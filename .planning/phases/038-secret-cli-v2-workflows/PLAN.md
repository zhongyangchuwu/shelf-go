# Plan: Secret CLI v2 Workflows

## Objective

Redesign the secret CLI around the v2 template-based item/field model in one clean breaking cutover, prioritizing script safety, low conceptual load, and value-free metadata output.

## Scope

- Replace the old v1 `group:key`-oriented secret command interface where it conflicts with the v2 model.
- Delete obsolete command paths, helpers, tests, and docs instead of carrying compatibility shims.
- Refactor the v2 model so env names are public fields such as `credential_env`, not `Field.Bindings`.
- Add template metadata for primary secret fields and env-name fields.
- Add profile definitions/registry as defaults over templates; provider/source remains a normal public field.
- Add app/service layer operations for v2 item create, set field from stdin, get field, list, info, and remove.
- Add CLI wiring for `secret create`, `secret set`, `secret get`, `secret list`, `secret info`, and `secret rm` as the canonical first-version interface.
- Implement selector resolution by exact item ID, exact title, exact alias, and env-name field lookup.
- Support `generic_secret`, `api_credential`, and `openai` profile behavior through registries, not CLI conditionals.
- Keep `list --long`, `info`, and ambiguity errors value-free.
- Add tests for secret-materializing and value-free command behavior.

## Non-Goals

- No TUI or interactive wizard.
- No project manifest v2 export/run runtime yet; that is Phase 39.
- No manager adaptation; that is Phase 40.
- No hosted/team sharing/password-manager expansion.

## Tasks

1. Refactor `vaultmodel` from env bindings to env-name public fields and profile definitions.
2. Update `vaultjson` and `vaultmigrate` for the revised v2 model.
3. Inspect current secret CLI/app boundaries and identify v1 concepts to remove.
4. Implement v2 service operations behind app/domain boundaries.
5. Replace CLI commands/flags for create/set/get/list/info/rm.
6. Delete obsolete v1 command code/tests/docs that conflict with the redesigned interface.
7. Add tests for selector behavior, stdin value input, value-free list/info, explicit get output, profile defaults, env-name fields, and removed legacy paths.
8. Run focused and project-wide verification.
9. Write summary, review, verification, capture, update root planning docs, and commit.

## Acceptance Criteria

- User can create `generic_secret` and `api_credential` v2 items non-interactively.
- User can create an API credential from `--profile openai`; defaults come from profile registry data, not CLI/provider hardcoding.
- User can set a secret field from stdin without placing values in argv.
- User can set env-name fields as ordinary public fields.
- User can get a field explicitly, with primary secret field as the default.
- `list --long` and `info` expose public metadata and value-set state without printing secret field values.
- Selector ambiguity and duplicate env-name lookup fail with value-free candidates.
- Obsolete v1 secret CLI paths are removed or intentionally replaced by tested v2 behavior; rollback is via git, not compatibility shims.

## Verification

- Focused tests for v2 model/profile validation, migration adaptation, secret service, and CLI commands.
- `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate ./internal/secret ./internal/app ./internal/cli`
- `go test ./...`
- `go vet ./...`
- `go-arch-lint check`

## Risks

- A clean breaking cutover can remove useful muscle memory; mitigate with a coherent command grammar, tests, docs, and a single commit boundary for rollback.
- Clipboard support can become platform-specific; avoid adding it without a small, testable boundary.
- Selector query matching can become surprising; keep exact ID/title/env precedence strict and ambiguous query failures value-free.

## Design Reference

- `.planning/phases/038-secret-cli-v2-workflows/CLI-SPEC.md`
