# Plan: Item Vault Storage and Crypto Boundary

## Objective

Make the item vault release-ready by replacing temporary `*.v2.json` plaintext storage with a canonical encrypted item-vault repository and a setup flow that can choose an existing age identity or generate a new one.

## Product Shape

### Setup UX

`shelf setup` should guide the user through encryption identity selection:

1. Detect configured age identities and likely default identity paths.
2. Offer choices:
   - use an existing identity file;
   - generate a new age identity;
   - provide a recipient manually for advanced/non-interactive use.
3. Write config with explicit recipient and identity path values.
4. Initialize the canonical encrypted item vault only after config and key material are valid.
5. Print paths and safety guidance without printing private key material or secret values.

This mirrors the gopass setup pattern at a high level: a setup wizard chooses encryption material before initializing the store, with flags available for non-interactive users.

### Naming

Replace user-facing temporary naming:

- Stop documenting `v2 vault` as the product name.
- Use `item vault` or `Shelf vault` for the new model.
- Stop defaulting to `<vault>.v2.json` for the canonical release path.
- Keep any legacy/cutover sidecar support explicit and temporary.

### Crypto Boundary

Encryption should wrap repository bytes, not leak into the item model:

```text
vaultmodel.Vault        item/domain model
vaultjson.Codec         strict JSON encode/decode of item vault plaintext bytes
vaultcrypto/age layer   seal/open byte payloads with age recipients/identities
repository/provider     load/save encrypted item vault file atomically
```

The model and codec should remain usable in tests with fake plaintext bytes. Production load/save should use encrypted bytes.

## Work Packages

### 1. Storage contract and naming cutover

- Introduce a v2/item-vault repository/provider abstraction before adding encryption.
- Decide canonical item-vault file path and extension.
- Replace app-layer direct construction of `vaultjson.Repository` with the provider abstraction.
- Preserve explicit paths for tests and advanced use.
- Keep legacy v1 vault loading behavior unchanged.

### 2. Age encryption wrapper

- Add an age-backed item-vault repository that composes:
  - config-resolved recipients/identities;
  - `vaultjson.Codec` for plaintext model bytes;
  - atomic encrypted file writes and backups.
- Ensure plaintext JSON temp files are not written by production repository paths.
- Keep backup files encrypted.
- Validate decrypted item vault before returning it.

### 3. Setup flow

- Extend `shelf setup` to support interactive existing-key vs generate-key flow.
- Keep non-interactive flags for scripts.
- Avoid modifying private identity files unless the user selects generation.
- Make output value-free and key-safe.

### 4. CLI cutover

- Update `secret` item workflows and project-vault loading to use the encrypted item-vault provider.
- Remove ordinary default use of `*.v2.json`.
- Keep manager gate against the canonical item-vault path.
- Keep legacy v1 commands/paths only where explicitly required for cutover compatibility.

### 5. Docs and verification

- Update README, SECURITY, portable-vault, reference, architecture, troubleshooting, and changelog.
- Add tests for encrypted save/load, no plaintext temp persistence, setup key choice behavior, explicit path behavior, and manager gate path behavior.
- Verify with focused package tests, full tests, vet, and architecture lint.

## Non-Goals

- No automatic v1-to-item-vault migration command.
- No Web manager adaptation.
- No TUI editing flow.
- No project workflow expansion beyond keeping existing v2 bindings working.
- No team sharing, hosted sync, daemon, or shell hook activation.

## Acceptance Criteria

- A newly initialized item vault is encrypted at rest by default.
- Setup can choose existing age identity or generate a new one interactively.
- Secret CLI item workflows no longer require or create plaintext `*.v2.json` as the default storage path.
- Encryption code is outside `vaultmodel` and outside JSON model semantics.
- App code depends on an item-vault provider abstraction rather than direct JSON repository construction.
- Documentation uses release-quality naming and clearly distinguishes legacy v1 vaults from the item vault only where needed.
- Full verification passes: `go test ./...`, `go vet ./...`, `go-arch-lint check`.

## Follow-On Phases

After this phase:

1. CLI/model/file polish for release hardening.
2. Project workflow refinement only after item-vault storage is stable.
3. Web manager item-vault design and implementation.
4. TUI/full-object editing design and implementation.
