# Plan: Phase 11 Secret Edit and Manager Safety Hardening

## Objective

Harden plaintext boundaries that already exist in `secret edit` and `vault open` workflows.

## Scope

In scope:

- Explicitly chmod `secret edit` temp files to `0o600`.
- Preserve cleanup on editor and JSON validation failure paths.
- Add tests for manager token/cookie/host safety boundaries.
- Document edit temp-file and manager token limitations.
- Optionally chmod config setup temp files to `0o600` for consistency.

Out of scope:

- Replacing editor workflow.
- Long-running daemon or automatic browser session expiry.
- TLS for localhost manager.
- Removing tokenized manager URLs.

## Tasks

1. Harden secret edit temp file.
   - Call `tmp.Chmod(0o600)` immediately after `CreateTemp`.
   - Add focused tests for permissions and cleanup on failure paths.

2. Pin manager safety behavior.
   - Add tests for localhost host acceptance.
   - Add tests for `X-Shelf-Token` and cookie token paths.
   - Add test that GET without token remains unauthorized.
   - Add test that query-token response sets HttpOnly and SameSite Strict cookie.

3. Update docs.
   - Security model: temp file has restrictive permissions and is cleaned after command exit where possible.
   - Reference/troubleshooting: manager URL token is local but visible in browser history/process surfaces; close manager with Ctrl-C when done.

4. Verify.
   - Run focused secret edit and manager tests.
   - Run `go test ./...`.

## Acceptance Criteria

- `secret edit` plaintext temp files are explicitly `0o600`.
- Failure-path tests prove edit temp files are removed on editor error and invalid JSON.
- Manager safety tests cover header token, cookie token, localhost host, missing-token GET, and cookie attributes.
- Docs state remaining plaintext/token boundaries.

## Verification

- `go test ./internal/cli -run 'TestSecretEdit'`
- `go test ./internal/manager -run 'TestManager'`
- `go test ./...`

## Risks

- Editor temp files cannot be removed after SIGKILL or system crash; docs should say cleanup occurs on normal command exit.
- Manager URL token can remain in local browser history; docs should state this rather than imply stronger secrecy.
