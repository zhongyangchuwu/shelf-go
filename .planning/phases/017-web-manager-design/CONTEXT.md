# Phase 17 Context: Web Manager Design Contract

## Goal

Define the v0.1.1 Web manager design contract before implementation. The contract should make `shelf vault open` the primary editing surface for local secret CRUD, reveal/copy, and tag management while preserving Shelf's local-only safety boundaries and single-binary distribution model.

## Constraints

- WebUI is the primary editing surface for add/edit/delete/tag workflows; CLI should remain compact.
- Do not add `secret meta` or `secret tag` command groups.
- CLI enhancements for v0.1.1 should focus on application workflows: tag-based `secret list`, tag-based `secret export`, and project tag bindings.
- Keep the current age-encrypted JSON vault format for v0.1.1.
- Do not implement or spike SQLite in v0.1.1; defer storage redesign to v0.2.0.
- Keep `net/http` compatibility and existing manager safety boundaries unless a later plan explicitly replaces them.
- Use embedded local assets only. No CDN, hosted frontend, account, backend service, or permanent daemon.
- Avoid SPA/toolchain complexity unless explicitly re-approved. Single-binary Go release remains a hard distribution goal.
- Preserve no-value list/search responses. Secret values may appear only after explicit reveal/copy actions.

## Decisions

- v0.1.1 milestone: improve WebUI editing and tag-based workflows.
- Web implementation direction: Go server-rendered local manager using `net/http`, embedded templates/assets, and a polished console/admin visual system.
- Candidate visual systems/templates:
  - TailAdmin-style Tailwind admin dashboard for full console layout references.
  - daisyUI/Tailwind component language for polished cards, tables, badges, modals, inputs, drawers, and toasts.
  - Pico CSS only if implementation chooses a minimal no-build UI, but it is less suited to a rich console.
- Fine-grained CLI metadata editing is rejected for v0.1.1; existing `secret set` remains the compact CLI editing path.
- Tag selection semantics: multiple tags use AND semantics for v0.1.1.
- Project tag bindings are in scope after secret tag list/export semantics are implemented.
- SQLite/storage redesign is deferred to v0.2.0 and should not consume v0.1.1 planning or implementation capacity.

## Open Questions

- Should Phase 17 choose daisyUI/Tailwind with a build step that commits embedded generated CSS, or a no-build CSS approach with lower visual ceiling?
- Should htmx be used as an embedded enhancement for partial updates, or should the manager use small first-party JavaScript only?
- What exact visual layout should design use: left secret list plus right detail/editor panel, or full-width table plus modal/drawer editor?
- Should reveal and copy be separate actions, or should copy trigger an explicit one-shot reveal internally without showing the value?
- How long should revealed values stay visible before auto-hide, if at all?
- Should the manager support raw JSON value editing, string-only editing, or both?

## Verification Expectations

- Phase 17 should produce a UI/design spec before Phase 18 implementation.
- The spec should include screenshots/mockups or enough layout detail for implementation without reinterpretation.
- The spec should define routes/endpoints or handler responsibilities that preserve existing manager tests.
- The spec should define safety acceptance for token URL cleanup, Host/Origin/token checks, no-store secret responses, explicit reveal/copy, and no persistent browser storage of values.
- The spec should define the chosen asset strategy and how release builds remain single-binary/offline.
- Follow-on implementation phases must verify with focused manager tests and `go test ./...`.
