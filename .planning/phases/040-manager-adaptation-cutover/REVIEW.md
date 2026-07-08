# Review: Manager Adaptation and Cutover

## Scope Reviewed

- `internal/app/manager.go`
- `internal/app/manager_test.go`
- `internal/cli/manager_test.go`
- Existing `internal/manager` server tests
- User docs and planning updates

## Findings

- No blocking findings.
- The gate is placed at the correct boundary: `internal/app` before `LoadVault` and before constructing a v1 `SecretService`.
- `internal/manager` remains unchanged and continues to enforce loopback, token, no-store, Host, Origin, metadata-only list/detail, and explicit reveal semantics.
- The error is value-free: it includes the resolved v2 vault path and CLI fallback guidance, not item data, field data, env-file contents, or secret values.
- Existing v1 manager behavior is preserved when no valid v2 vault is present.
- The gate loads the candidate v2 JSON through `vaultjson.Repository`, avoiding false positives for arbitrary sidecar files.

## Tradeoffs

- Full v2 manager adaptation is deferred. This is intentional: a correct v2 editor needs item/template/field UX, not the existing path/env/tag form.
- Invalid files at the v2 path do not trigger the v2 gate; they fall through to v1 behavior. This avoids blocking v1 users because of unrelated or corrupt files, while valid v2 vaults are protected.
- The v2 vault path appears in the error. This is acceptable because paths are operational metadata and needed to explain why the gate triggered; no values are exposed.

## Verification Needed

- Focused app/CLI/manager tests.
- Full Go tests.
- `go vet ./...`.
- Architecture lint.

All checks passed in `VERIFICATION.md`.
