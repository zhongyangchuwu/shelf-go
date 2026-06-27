# Phase 17 Research: Web Manager Design Options

## Question

What WebUI approach should Shelf Go use for v0.1.1 so `shelf vault open` becomes a polished local secret editing console while preserving local-only security and single-binary distribution?

## Sources Checked

- Current manager code: `internal/manager/server.go`
- Current CLI manager startup: `internal/cli/manager.go`
- Current store model: `internal/store/model.go`, `internal/store/store.go`
- Current secret edit model: `internal/secret/edit.go`
- Context7: daisyUI docs (`/websites/daisyui` resolved)
- Context7: Pico CSS docs (`/websites/picocss` resolved)
- Context7: Tailwind CSS docs (`/tailwindlabs/tailwindcss.com` resolved)
- Prior research agents: `agent://WebUIOptions`, `agent://SecretCapabilities`, `agent://StorageDecision`
- Web search: TailAdmin, Asteria Dash, DaisyUI Nexus, Flowbite Admin dashboard references

## Findings

- Current manager already exposes loopback-only, token-protected, Host/Origin-checked HTTP routes and JSON APIs for list, reveal, write, and delete.
- Current UI is a minimal inline template with search/list and alert-based reveal only; CRUD/tag/copy UX is not exposed even though server APIs partly exist.
- Store model already supports value, env, description, and tags, so WebUI CRUD/tag editing can work over the existing vault model without storage changes.
- `net/http` plus embedded templates/assets is the lowest-risk implementation path because it preserves current tests and single-binary release behavior.
- daisyUI/Tailwind is a strong visual direction for a polished console: cards, tables, badges, modals, drawers, inputs, and toasts map directly to Shelf manager needs.
- TailAdmin-style templates are useful as design references for sidebar/table/form/modal patterns, but should be heavily trimmed instead of imported wholesale.
- Pico CSS is useful for a minimal no-build UI, but likely too limited for the desired console-level editing experience.
- React/Vite SPA and broad Go web frameworks add build/runtime/dependency surface without being necessary for local CRUD and reveal/copy workflows.
- SQLite is not needed for v0.1.1 WebUI/tag work and is deferred to v0.2.0.

## Tradeoffs

- **daisyUI/Tailwind embedded build:** Best polish and component vocabulary; adds frontend build/precompiled CSS management. Acceptable if generated assets are committed or reproducibly generated and embedded.
- **TailAdmin template reference:** Fast visual inspiration; too broad to import directly without bloat. Use as design reference, not product dependency.
- **Pico CSS:** Simplest asset path; less distinctive and less suited to drawer/modal/tag-heavy console UI.
- **htmx:** Good optional enhancement for server-rendered partials; adds third-party JS but avoids SPA state duplication. Should be embedded if used.
- **first-party JS only:** Lowest dependency; enough for reveal/copy/hide/toast. More manual DOM work for inline edit/search refresh.
- **SPA:** Highest polish potential but worst fit for single-binary local secret manager in v0.1.1.

## Confidence

High for keeping `net/http` and embedded local assets. Medium for choosing daisyUI/Tailwind because final visual design still needs a dedicated UI spec/mockup. High for deferring SQLite because v0.1.1 goals do not require a storage model change.
