# Plan: Credential Health Audit

## Goal

Turn existing item metadata into actionable local maintenance signals before credentials break a developer workflow.

## Scope

- Add a value-free audit for expired and soon-expiring credentials using existing expiry metadata.
- Report item identity, field, expiry date, and status; support a configurable warning horizon.
- Surface the same warnings in the TUI catalog and current-project status where applicable.
- Define deterministic exit behavior suitable for local scripts without exposing values.

## Non-Goals

- No automatic key rotation, issuer API integration, remote monitoring, background daemon, team notification, or claims of leak detection.

## Acceptance

- Expired, upcoming, absent, and malformed expiry states have explicit behavior.
- Audit output and TUI warnings remain value-free.
- Existing project run/export semantics remain unchanged unless a future user-approved enforcement policy is added.
- `go test ./...`, `go vet ./...`, and `go-arch-lint check` pass.
