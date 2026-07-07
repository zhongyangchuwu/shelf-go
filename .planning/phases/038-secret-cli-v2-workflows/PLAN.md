# Plan: Secret CLI v2 Workflows

## Objective

Implement common non-interactive v2 secret CLI workflows over template-based items and fields, prioritizing script safety and value-free metadata output.

## Scope

- Decide transition shape for v2 secret commands relative to existing v1 secret commands.
- Add app/service layer operations for v2 item create, set field from stdin, get field, list, and info.
- Add CLI wiring for selected workflows after repository path selection is clear.
- Implement selector resolution by item ID, exact title, env binding, and unambiguous query.
- Support `generic_secret` and `api_credential` creation with field metadata.
- Keep `list --long` and `info` value-free.
- Add tests for secret-materializing and value-free command behavior.

## Non-Goals

- No TUI or interactive wizard.
- No project manifest v2 export/run runtime yet; that is Phase 39.
- No manager adaptation; that is Phase 40.
- No hosted/team sharing/password-manager expansion.

## Tasks

1. Inspect current secret CLI/app boundaries and v2 repository APIs.
2. Decide command compatibility strategy for v2 workflows.
3. Implement v2 service operations behind app/domain boundaries.
4. Add CLI commands/flags for create/set/get/list/info and copy if clipboard scope is resolved.
5. Add tests for selector behavior, stdin value input, value-free list/info, and explicit get output.
6. Run focused and project-wide verification.
7. Write summary, review, verification, capture, update root planning docs, and commit.

## Acceptance Criteria

- User can create `generic_secret` and `api_credential` v2 items non-interactively.
- User can set a secret field from stdin without placing values in argv.
- User can get a secret field explicitly.
- `list --long` and `info` expose public metadata and value-set state without printing secret field values.
- Selector ambiguity fails with value-free candidates.
- Existing v1 workflows either remain passing or have an explicitly tested cutover path.

## Verification

- Focused tests for v2 secret service and CLI commands.
- `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate ./internal/secret ./internal/app ./internal/cli`
- `go test ./...`
- `go vet ./...`
- `go-arch-lint check`

## Risks

- Replacing existing `secret` commands too early could break current v1 tests and workflows.
- Clipboard support can become platform-specific; avoid adding it without a small, testable boundary.
- Selector query matching can become surprising; keep exact ID/title/env precedence strict and ambiguous query failures value-free.
