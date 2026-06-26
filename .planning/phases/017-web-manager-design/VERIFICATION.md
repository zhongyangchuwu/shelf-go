# Verification: Phase 17 Web Manager Design Contract

## Claims Checked

- Phase 17 has a WebUI design contract covering list/search, add, edit, delete, reveal, copy, hide, tags, and tag editing.
- The visual direction is selected and grounded in Shelf's local vault subject.
- The technical direction keeps `net/http`, single-binary release, embedded local assets, no CDN, and no SPA requirement.
- The safety contract covers token URL cleanup, loopback/token/Host/Origin checks, no-store responses, POST reveal, and no persistent browser storage for secret values.

## Evidence Observed

- `UI-SPEC.md` contains `Visual System`, `UX Requirements`, `API Contract for Phase 18`, `Security Contract`, `Accessibility and Responsiveness`, and `Build and Asset Strategy` sections.
- `UI-SPEC.md` resolves Phase 17 open questions: first-party embedded CSS/JS, left vault rail plus inspection bench, explicit reveal/copy, 60-second auto-hide, and metadata editing without revealing values.
- `PLAN.md` includes required Objective, Scope, Tasks, Acceptance Criteria, Verification, and Risks sections.

## Coverage

- WEB-01: Covered by List and Search plus layout spec.
- WEB-02: Covered by Add Secret, Edit Secret, Delete Secret flows.
- WEB-03: Covered by Reveal, Hide, Copy and Security Contract.
- WEB-04: Covered by API Contract and Security Contract.
- WEB-05: Covered by Build and Asset Strategy.
- WEB-06: Covered by Visual System and Signature Element.
- BOUND-01: Covered by CLI editing boundary in context/spec.
- BOUND-02: Covered by storage non-goal in context/spec.

## Gaps

- Runtime behavior is not proven in Phase 17; Phase 18 must implement and test the contract.

## Result

Pass. Phase 17 design contract is ready for Phase 18 implementation.
