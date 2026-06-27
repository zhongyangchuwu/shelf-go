# Summary: Phase 17 Web Manager Design Contract

## Completed Changes

- Created `UI-SPEC.md` for the v0.1.1 Shelf Web Manager.
- Selected the vault rail + inspection bench layout.
- Selected first-party embedded CSS/JS with TailAdmin/daisyUI as visual references, not runtime dependencies.
- Defined explicit reveal/copy/hide, add/edit/delete, tag editing, and metadata search flows.
- Defined the manager API hardening contract for token URL cleanup, POST reveal, no-store responses, no value list/detail leakage, and no persistent browser storage.
- Recorded implementation plan in `PLAN.md`.

## Files Changed

- `.planning/phases/017-web-manager-design/UI-SPEC.md`
- `.planning/phases/017-web-manager-design/PLAN.md`

## Deviations

- None.

## Evidence

- Phase 17 success criteria in `.planning/ROADMAP.md` were mapped directly to `UI-SPEC.md` sections.
- Security reviewer findings confirmed the spec's hardening direction: token URL redirect, no-store responses, POST reveal, explicit browser-memory clearing, and safe update semantics.

## Unresolved Risks

- Phase 18 must verify the implemented browser behavior; the design contract alone does not prove runtime DOM cleanup or copy behavior.
