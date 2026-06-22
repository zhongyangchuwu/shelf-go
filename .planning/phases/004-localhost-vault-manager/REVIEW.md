# Review: Phase 4 Localhost Vault Manager

## Scope Reviewed

- `internal/cli/manager.go`
- `internal/cli/root.go`
- `internal/manager/server.go`
- `internal/manager/server_test.go`
- `internal/cli/manager_test.go`

## Findings

- No duplicate persistence path found. Manager writes call `store.Vault.Update` and mutate through `Store.Set` / `Store.Delete`.
- Metadata list/search endpoint does not include secret values; reveal is a separate explicit endpoint.
- Unsafe routes require token and reject bad Origin; all routes reject invalid Host.
- CLI listener rejects non-loopback addresses before serving.
- Token generation uses 32 bytes from `crypto/rand` and URL-safe base64.

## Fixes Applied

- None after review. Implementation was added with safety controls in place.

## Waivers

- No visual UI audit. Phase 4 acceptance is safety and functional manager MVP, not polished frontend design.
- Clipboard-specific verification waived because explicit reveal satisfies WEB-03; copy UX is optional polish.

## Remaining Risks

- URL token can be exposed in browser history or local logs. Keep manager on-demand and document handling in Phase 5.
- Host/Origin policy is intentionally strict and local-only; future remote access would require a separate security design.
