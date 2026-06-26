# Plan: Phase 18 Web Manager Editing Console

## Objective

Implement the Phase 17 Web manager design contract as a polished local editing console and harden the manager API before exposing richer CRUD/reveal/copy workflows.

## Scope

- Replace the minimal inline manager page with a vault rail + inspection bench UI.
- Keep all assets embedded in the Go binary through first-party CSS/JS.
- Add metadata-only detail and update behavior needed by the editor.
- Harden browser/API safety around token URLs, no-store responses, reveal methods, and update semantics.
- Add focused manager tests for behavior and security invariants.
- Do not implement tag CLI/project workflows in this phase.

## Tasks

1. Move manager UI template out of `server.go` into `ui.go`.
2. Implement the first-party CSS/JS workbench from `UI-SPEC.md`.
3. Add `GET /api/secret` metadata detail without values.
4. Change reveal to POST-only JSON body.
5. Redirect query-token GET requests after setting the strict cookie.
6. Add no-store/cache prevention headers across manager responses.
7. Extend PUT `/api/secrets` with `old_path` and optional `value` for safe rename/metadata-only edits.
8. Update manager tests for the new contracts.

## Acceptance Criteria

- Manager page exposes search/list, add, edit/rename, delete, reveal, copy, hide, and tag editing controls.
- List and detail API responses include metadata and `value_set`, never secret values.
- Reveal is POST-only and requires the same token and Origin protections as other unsafe methods.
- Query-token GET requests set the cookie and redirect to a token-free URL.
- Manager responses use `Cache-Control: no-store`.
- Metadata-only PUT preserves the existing secret value and supports explicit rename.
- Existing encrypted vault write/delete behavior still works.

## Verification

- `go test ./internal/manager`
- `go test ./...`
- Workspace Go diagnostics through LSP.

## Risks

- Browser clipboard behavior depends on secure-context/browser support; tests cover server/UI contract, not native clipboard permission prompts.
- Values revealed in DOM remain visible until hide/auto-hide; implementation clears DOM state but a browser memory forensic threat is out of scope for this local manager.
- The handcrafted CSS is intentionally embedded and no-build; future visual expansion may need asset organization.
