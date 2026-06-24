# Context: Phase 11 Secret Edit and Manager Safety Hardening

## Goal

Reduce plaintext exposure in interactive secret editing and clarify localhost manager boundaries without adding a daemon, hook workflow, or complex UI.

## Constraints

- Keep `shelf secret edit` behavior intact: edit one JSON object in the configured editor and validate before saving.
- Do not introduce a permanent daemon, remote manager access, TLS setup, or browser-extension behavior.
- Manager must remain loopback-only and token-gated.
- Prefer low-complexity hardening and explicit docs over larger UX changes.

## Decisions

- Add explicit `0o600` permission hardening for the plaintext `secret edit` temporary file.
- Preserve temp-file cleanup via existing `defer os.Remove` and add tests for failure-path cleanup.
- Add manager tests for existing token paths and cookie attributes so safety boundaries stay pinned.
- Document manager token-in-URL and no-TLS loopback risks instead of adding a session/daemon redesign.
- Also harden setup config temp-file permissions to `0o600` for consistency, even though config is value-free.

## Open Questions

None for this phase.

## Verification Expectations

- Tests prove `secret edit` temp file is restrictive and cleaned when editor or JSON validation fails.
- Tests prove manager accepts header/cookie token paths, allows localhost host, rejects missing-token GET, and sets HttpOnly/SameSite Strict cookie.
- Docs state edit temp-file cleanup/permissions and manager token/localhost boundaries.
