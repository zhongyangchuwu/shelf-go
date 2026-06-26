# Summary: Phase 11 Secret Edit and Manager Safety Hardening

## Completed Changes

- Added explicit `0o600` permission hardening for `shelf secret edit` plaintext temporary files.
- Added `0o600` permission hardening for setup config temp files for consistency.
- Added tests proving `secret edit` temp files are restrictive and cleaned after editor failure and invalid JSON failure paths.
- Tightened manager Host validation to allow localhost and loopback aliases while still rejecting non-loopback hosts.
- Added manager tests for localhost host acceptance, `X-Shelf-Token`, cookie token transport, missing-token GET rejection, and HttpOnly/SameSite Strict session cookie attributes.
- Updated security, reference, and troubleshooting docs with edit temp-file cleanup/permission behavior and manager token/localhost boundaries.

## Files Changed

- `internal/cli/secret.go`
- `internal/cli/secret_test.go`
- `internal/cli/init.go`
- `internal/manager/server.go`
- `internal/manager/server_test.go`
- `docs/security.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `.planning/phases/011-edit-manager-safety/CONTEXT.md`
- `.planning/phases/011-edit-manager-safety/PLAN.md`

## Deviations

- Manager hardening included a small Host validation fix because the planned localhost-host test exposed that `localhost:port` was rejected when the listener host was `127.0.0.1`.

## Evidence

- `go test ./internal/cli -run TestSecretEdit` passed.
- `go test ./internal/manager -run TestManager` passed.
- `go test ./...` passed.
- Safety docs search found temp-file and token/browser-history guidance in public docs and Phase 11 artifacts.

## Unresolved Risks

- `secret edit` temp files can still remain after SIGKILL, machine crash, or external editor copying buffers elsewhere.
- Manager URL tokens can still appear in local browser history; docs now state this explicitly.
