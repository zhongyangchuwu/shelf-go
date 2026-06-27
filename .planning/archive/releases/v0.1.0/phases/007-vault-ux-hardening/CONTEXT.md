# Context: Phase 7 Vault UX Hardening

## Goal

Improve vault-specific status, check, and doctor diagnostics without changing the encrypted storage contract or revealing secret values.

## Requirements

- VUX-01: `shelf vault status` / `shelf vault check` reports config path, vault path, file format, loadability, and safe next steps without values.
- VUX-02: Missing recipients, missing identities, plaintext legacy stores, unsupported vault formats, and undecryptable vaults produce concise recovery guidance.
- VUX-03: Docs explain first-run setup, vault init, vault migrate, vault status/check, vault open, and plaintext cleanup.
- VUX-04: Verification covers encrypted load/save, migration, status/check behavior, doctor behavior, and manager write safety under the new hierarchy.

## Current State

Phase 6 already introduced `shelf vault status` / `check` and basic plaintext/encrypted format reporting. The remaining gap is richer, consistent recovery guidance across vault status and doctor for configuration and decrypt failures.

## Constraints

- Do not change vault file format, age encryption behavior, config shape, or secret-store JSON model.
- Do not print secret values or decrypted secret paths in status/doctor diagnostics.
- Keep command hierarchy canonical: `shelf vault init`, `shelf vault migrate`, `shelf vault status`/`check`, `shelf vault open`.
- Prefer small CLI-layer helpers over moving persistence behavior out of `internal/store`.
