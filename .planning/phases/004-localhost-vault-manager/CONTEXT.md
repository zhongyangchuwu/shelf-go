# Context: Phase 4 Localhost Vault Manager

## Goal

Add an on-demand localhost-only vault manager that lets a single developer browse/search metadata, intentionally reveal values, and create/update/delete secrets while reusing the existing encrypted vault validation, locking, and save behavior.

## Constraints

- Local-first and CLI-first: the manager is additive and started by a CLI command; core workflows must not require a daemon.
- Bind to loopback by default; do not expose on non-loopback addresses unless a future phase explicitly accepts that risk.
- State-changing HTTP routes need write-safety controls: tokenized access, Origin validation, Host validation, and method restrictions.
- Create/update/delete must reuse `store.Store` validation and `store.Vault.Update`; no parallel persistence path.
- Browsing/search responses may include paths and non-secret metadata only. Secret values are returned only from an explicit reveal endpoint.
- Do not add heavyweight frontend frameworks. Keep the UI and server boring and standard-library based unless required.
- Keep package boundaries: CLI command in `internal/cli`; reusable manager/server behavior can live in a focused package under `internal/manager` if needed.

## Decisions

- Implement a `shelf manager` command that starts an HTTP server bound to loopback with an ephemeral random session token.
- Provide a minimal HTML UI plus JSON endpoints in the same server. The JSON API makes behavior testable without browser automation.
- Use a token carried in the manager URL and required on unsafe routes. The server should also set a same-site session cookie for browser form/API use.
- Use `store.Vault.Read` for list/search/reveal and `store.Vault.Update` for create/update/delete so locking, encrypted save, and backups remain centralized.
- Treat Phase 4 as backend-first usable MVP. UI can be plain HTML/forms/fetch as long as all WEB requirements are satisfied.

## Open Questions

- Whether to auto-open the browser. Default should be no for scriptability; an `--open` flag can be deferred if not needed.
- Exact route names are implementation details, but tests need stable routes.
- Whether copy-to-clipboard is required for MVP. Browser clipboard APIs are awkward in tests; explicit reveal is sufficient unless implementation time allows a copy button in UI.

## Verification Expectations

- Automated tests for loopback binding, token requirement, Origin/Host rejection, search/list metadata without values, explicit reveal, and create/update/delete persistence.
- Automated tests must verify writes go through encrypted vault storage and do not leave known plaintext values in the vault file.
- Run targeted manager/CLI tests, then `go test ./...`.
