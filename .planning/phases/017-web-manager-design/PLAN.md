# Plan: Phase 17 Web Manager Design Contract

## Objective

Complete a design and implementation contract for the v0.1.1 Web manager before feature implementation, using the downloaded `frontend-design.md` guidance while preserving Shelf's local-only secret safety model.

## Scope

- Define Web manager visual identity, layout, copy, and interaction model.
- Define API and browser-safety contracts for Phase 18.
- Resolve Phase 17 open questions from `CONTEXT.md`.
- Do not implement code as part of this phase's acceptance; implementation belongs to Phase 18.

## Tasks

1. Read active v0.1.1 requirements, roadmap, phase context, research, current manager code, and `frontend-design.md`.
2. Produce `UI-SPEC.md` with visual tokens, layout, signature element, interaction flows, API contract, security contract, copy rules, accessibility, and asset strategy.
3. Review the spec against `WEB-01..WEB-06`, `BOUND-01`, and `BOUND-02`.
4. Update root planning state to move Phase 17 to complete and Phase 18 to executing/planning for implementation.

## Acceptance Criteria

- `UI-SPEC.md` describes list/search, add, edit, delete, reveal, copy, hide, tag filtering, and tag editing flows.
- `UI-SPEC.md` selects a distinctive visual direction grounded in Shelf's local vault subject and avoids generic SaaS admin defaults.
- `UI-SPEC.md` keeps `net/http`, single-binary distribution, embedded local assets, no CDN, no SPA, and no permanent daemon.
- `UI-SPEC.md` covers token URL cleanup, loopback/token/Host/Origin checks, `Cache-Control: no-store`, POST reveal, and no persistent browser storage for secret values.
- Phase 17 artifacts satisfy the phase artifact contracts.

## Verification

- Read `UI-SPEC.md` against the success criteria in `.planning/ROADMAP.md` Phase 17.
- Confirm root planning links point to the active phase directory.
- Confirm no v0.1.1 SQLite/storage work is introduced.

## Risks

- The chosen no-build CSS/JS approach may take more handcrafted CSS than a Tailwind build, but it preserves the single-binary/offline constraint.
- Rich client behavior may retain revealed values in memory; Phase 18 must explicitly clear DOM and JS state on hide and auto-hide.
- Existing manager API allows GET reveal and token-bearing URLs; Phase 18 must harden these before relying on the UI.
