# Capture: Phase 7 Vault UX Hardening

## Durable Decisions

- Vault recovery guidance belongs at the CLI layer. Store errors remain precise and command-agnostic; commands translate them into user next steps.
- `shelf vault status` and `shelf doctor` should share diagnostic wording where possible so users do not get conflicting recovery paths.
- Missing recipients are write-blocking but not needed to decrypt an existing vault. Status reports recipient configuration separately from vault loadability.

## Maintainer Notes

- `vaultFormatDetail` owns user-facing guidance for missing, empty, plaintext, unsupported-vault, and unsupported file content states.
- `vaultLoadErrorDetail` owns user-facing guidance for missing identity paths, unreadable identity files, malformed identities, wrong identities, decrypt failures, and invalid decrypted stores.
- Future backup restore or recipient-inspection commands should extend the same guidance helpers rather than adding a second diagnostic vocabulary.

## Verification to Preserve

- Keep status/check tests value-free. They should assert guidance text and absence of regressions without requiring secret values in output.
- Keep doctor tests aligned with status guidance for vault load failures.
