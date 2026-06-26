# Plan: Phase 3 Project Workflow Compatibility

## Objective

Deliver regression evidence and any required fixes so `shelf export`, `shelf project`, and `shelf run` operate over encrypted vault storage with the same command semantics they had before encryption.

## Scope

In scope:

- Direct export command compatibility for exact path and prefix flows across shell, env, and JSON formats.
- Project manifest command compatibility for add/list/explain/export over encrypted storage.
- Runtime injection compatibility for `shelf run -- ...` and no-value `shelf run --dry-run -- ...` behavior.
- Regression tests that use real age-encrypted vault files.
- Phase summary, review, verification, capture, and root planning updates after implementation.

Out of scope:

- New manifest schema fields.
- Localhost vault manager work from Phase 4.
- Documentation/release hardening from Phase 5 except capture notes needed by this phase.
- Team sharing, hosted sync, permanent daemon, or direct chezmoi control.

## Tasks

1. Add or tighten direct export regression tests.
   - Cover exact path export in shell/env/JSON.
   - Cover prefix export with sorted or deterministic expectations.
   - Confirm the vault file does not contain exported plaintext values or paths.

2. Add or tighten project workflow regression tests.
   - Confirm `project add` writes only non-secret manifest data.
   - Confirm `project explain` reports paths/env names without leaking secret values.
   - Confirm `project export` exact and prefix flows still intentionally render values in all supported formats.
   - Confirm missing/optional/conflict diagnostics remain value-free.

3. Add or tighten run workflow regression tests.
   - Confirm `run` injects exact and prefix-derived env values from the encrypted vault.
   - Confirm `run --dry-run` prints injection names and override warnings without secret or parent values.
   - Confirm failed resolution prevents child execution.

4. Fix any production gaps found while adding tests.
   - Prefer shared loader use over command-specific storage logic.
   - Keep errors concise and actionable.
   - Do not introduce mocks or plaintext fallbacks.

5. Run review and verification gates.
   - Review changed CLI paths for vault-boundary bypasses and value leaks.
   - Run targeted tests, then package tests, then full `go test ./...`.

6. Close the phase.
   - Write `SUMMARY.md`, `REVIEW.md`, `VERIFICATION.md`, and `CAPTURE.md`.
   - Update `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`, `.planning/PROJECT.md`, and `.planning/STATE.md` for Phase 3 completion and Phase 4 readiness.

## Acceptance Criteria

- CLI-02: `shelf export` exact-path and prefix flows render env, shell, and JSON output from encrypted storage.
- CLI-03: `shelf project` manifest commands resolve paths and prefixes from encrypted storage and `.shelf.json` remains value-free.
- CLI-04: `shelf run -- ...` injects encrypted-vault secrets into child processes, and `shelf run --dry-run -- ...` does not print secret values.
- CLI-05: Regression tests prove existing command semantics survive the storage change.
- TEST-01: CLI compatibility coverage is included with the already completed encrypted load/save, identity error, and migration coverage from Phases 1 and 2.

## Verification

Targeted during implementation:

```bash
go test ./internal/cli -run 'Test(Export|Project|Run)'
```

Phase gate:

```bash
go test ./internal/cli ./internal/store ./internal/render ./internal/manifest
go test ./...
```

Manual/evidence checks:

- Inspect encrypted vault bytes in tests for absence of known secret values and known secret paths.
- Inspect `.shelf.json` written by `project add` for absence of known secret values.
- Inspect `project explain` and `run --dry-run` outputs for absence of known secret and parent env values.

## Risks

- Existing tests may look encrypted only because they pass `--vault`; verify helper behavior before claiming coverage.
- Diagnostics can leak values indirectly if they include raw rendered bindings or parent environment values.
- Export/project export intentionally print values; tests must not confuse expected value output with no-leak diagnostics.
- Prefix-derived env names can collide; compatibility tests must preserve current conflict behavior.
