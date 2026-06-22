# Plan: Phase 4 Localhost Vault Manager

## Objective

Deliver a secure localhost-only manager MVP for encrypted vaults: browse/search metadata, intentionally reveal values, and create/update/delete secrets through write-safe routes that reuse existing vault validation and encrypted persistence.

## Scope

In scope:

- New CLI command to start the manager on loopback without a permanent daemon.
- Minimal standard-library HTTP server and plain HTML UI.
- JSON endpoints for list/search, reveal, create/update, and delete.
- Token, Host, and Origin checks for manager access and state-changing routes.
- Tests for route behavior, write safety, encrypted persistence, and no-value metadata responses.

Out of scope:

- Team sharing, accounts, hosted backend, or remote access.
- Rich frontend framework or visual polish beyond usable plain HTML.
- Browser extension/autofill.
- Release documentation hardening from Phase 5.
- Direct chezmoi integration.

## Tasks

1. Add manager package.
   - Define `internal/manager.Server` or equivalent with vault, token, allowed host, and `http.Handler`.
   - Implement route dispatch without global mutable state.
   - Render minimal HTML for browse/search/reveal/edit actions.

2. Implement read endpoints.
   - `GET /` returns the UI shell.
   - `GET /api/secrets?q=` returns path/env/description/tags/value_set metadata only.
   - `GET /api/secrets/{path}/reveal` returns the value only when token and safe request checks pass.

3. Implement write endpoints.
   - Create/update route parses path, value, env, description, tags, and force semantics.
   - Delete route removes a path.
   - All writes call `vault.Update` and `Store.Set`/`Store.Delete`.

4. Implement safety controls.
   - Generate high-entropy session token for the CLI command.
   - Require token on API requests, with cookie support for browser use.
   - Reject unsafe methods with bad/missing Origin.
   - Reject requests with Host not matching the listener host.
   - Bind CLI listener to loopback by default.

5. Add CLI command.
   - Register `shelf manager` in root command.
   - Flags: `--addr` default `127.0.0.1:0`; optionally `--no-url` only if needed by tests.
   - Print manager URL including token after listener starts.
   - Serve until interrupted.

6. Add tests.
   - Unit-test manager handler with real encrypted vault setup.
   - Test list/search metadata excludes values.
   - Test reveal requires token and returns value only explicitly.
   - Test create/update/delete mutate encrypted vault correctly and vault bytes stay value-free.
   - Test unsafe writes reject missing/bad Origin and missing/bad token.
   - Test Host validation rejects unexpected hosts.
   - Test CLI command exposes manager command registration and default loopback behavior if feasible without hanging.

7. Review, verify, and close phase.
   - Run targeted manager tests.
   - Run `go test ./...`.
   - Write SUMMARY, REVIEW, VERIFICATION, CAPTURE.
   - Update roadmap, requirements, project, and state for Phase 4 completion.

## Acceptance Criteria

- WEB-01: A CLI command starts a loopback-only local vault manager.
- WEB-02: The manager supports search and browsing of paths and non-secret metadata.
- WEB-03: Secret values are revealed only through intentional reveal requests.
- WEB-04: Create, update, and delete actions work from the manager.
- WEB-05: Writes use existing validation, locking, encrypted-save, and backup behavior.
- WEB-06: The manager binds to loopback by default and uses token, Origin, and Host controls for state-changing requests.
- WEB-07: The manager is on-demand and requires no permanent daemon or hosted service.
- TEST-02: Automated verification covers write protections and no-plaintext side files for representative edit flows.

## Verification

Targeted:

```bash
go test ./internal/manager ./internal/cli -run 'TestManager|TestNewRootCmdIncludesManager'
```

Gate:

```bash
go test ./...
```

Evidence checks:

- HTTP metadata response excludes known values.
- Explicit reveal response includes value only with valid token.
- Bad token / bad Origin / bad Host requests fail.
- Create/update/delete survive decrypt/reload and encrypted vault bytes do not contain known values or paths.

## Risks

- Localhost is not a full trust boundary; malicious web pages can hit loopback, so CSRF and Origin checks are required.
- URL tokens can leak through browser history or logs; keep them random, short-lived, and local-session scoped.
- UI edit logic can drift from CLI validation if it bypasses `Store.Set` or `Store.Delete`.
- Rich UI work can consume scope without improving safety; keep HTML minimal and focus on route contracts.
