# Phase 18 Context: Web Manager Editing Console

## Goal

Implement the v0.1.1 Web manager editing console from Phase 17 so `shelf vault open` exposes add, edit, rename, delete, reveal, copy, hide, metadata, and tag workflows over the local encrypted vault.

## Constraints

- Preserve `net/http`; do not add a Go web framework.
- Preserve single-binary/offline distribution; no CDN or npm build is required for v0.1.1.
- Preserve loopback Host validation, token validation, and Origin validation for unsafe methods.
- Remove token-bearing URLs from the visible URL through redirect after setting the strict HttpOnly cookie.
- Use `Cache-Control: no-store` for app and API responses.
- Reveal values only through explicit POST actions.
- Never include secret values in list or metadata/detail responses.
- Preserve metadata-only updates without forcing reveal or value replacement.
- Do not implement tag CLI/project bindings in this phase.

## Decisions

- Move the inline HTML/CSS/JS template out of `server.go` into `ui.go` to keep request handling readable.
- Use first-party embedded CSS and JavaScript, guided by `frontend-design.md` and the Phase 17 UI spec.
- Add `GET /api/secret?path=...` for metadata-only detail reads.
- Change reveal to POST-only with JSON body.
- Extend PUT `/api/secrets` with `old_path` and optional `value` so rename and metadata-only updates are explicit and safe.

## Open Questions

- None for this phase.

## Verification Expectations

- Focused manager tests prove token redirect, no-store responses, POST-only reveal, no value leakage in list/detail responses, metadata-only rename/update, encrypted vault writes, and delete behavior.
- Workspace diagnostics have no Go issues.
- `go test ./internal/manager` passes.
- `go test ./...` passes.
