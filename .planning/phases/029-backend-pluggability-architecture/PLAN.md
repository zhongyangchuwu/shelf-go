# Plan: Backend Pluggability Architecture

## Objective

Add gopass as the first non-Shelf vault source and use that work to expose architecture problems before adding any broad backend framework. Separately plan GPG as a crypto backend for Shelf's local vault format.

## Scope

- Project env workflows must support reading secrets from the current Shelf vault and from gopass.
- Shelf's existing age-encrypted JSON vault remains the default.
- The first gopass slice is read-only through `source.Reader`.
- GPG work is planned as a crypto-boundary spike before implementation.

## Non-goals

- No team sharing.
- No hosted sync.
- No manager editing of gopass secrets in the first slice.
- No generalized write backend registry until a second write-capable backend proves the interface.
- No automatic migration between Shelf vault, gopass, age, and GPG in the first slice.

## Phase 29: Backend Model Spike

**Goal:** Decide the smallest config and runtime shape that can select Shelf vault vs gopass without disrupting existing commands.

**Tasks:**
1. Add a backend/source design note covering:
   - `source.Reader` read path;
   - current concrete Shelf write path;
   - config keys for selecting source backend;
   - error and diagnostics shape.
2. Decide gopass integration mode:
   - preferred first spike: CLI adapter around installed `gopass` for lower dependency risk;
   - optional later spike: Go API adapter if CLI process behavior is too brittle.
3. Define path/metadata mapping:
   - how Shelf `group:key` maps to gopass paths;
   - how env names are derived or overridden;
   - whether tags are unsupported, encoded in secret metadata, or ignored in v1.
4. Update architecture lint only for planned package boundaries, not for unimplemented adapters.

**Acceptance:**
- A maintainer can answer: where does backend selection live, what remains Shelf-vault-only, and which commands are expected to work with gopass.
- No production behavior changes required.

## Phase 30: Gopass Read Source MVP

**Goal:** Implement gopass as a read-only `source.Reader` for `project explain`, `project export`, and `project run`.

**Target packages:**
- `internal/adapters/gopass/`
- `internal/config/`
- `internal/app/runtime.go`
- `internal/project/` tests as needed
- `.go-arch-lint.yml`

**Tasks:**
1. Extend config with an explicit source selector, e.g. `source.type: shelfvault|gopass`, while preserving current config as default Shelf vault.
2. Add `internal/adapters/gopass.Reader` implementing `source.Reader`.
3. Wire `app.LoadSecretReader` to select `shelfvault.Reader` or `gopass.Reader`.
4. Keep `LoadVault`, `UpdateVault`, secret CRUD, manager, setup, migration, and status on `shelfvault` only.
5. Add focused tests using a fake command runner or fake gopass client boundary; do not require a real user gopass store in unit tests.
6. Add one optional smoke scenario gated on `gopass` availability if useful, but do not make local tests depend on external secret state.

**Acceptance:**
- Existing Shelf vault project workflows remain unchanged.
- Config can select gopass for project read workflows.
- Missing gopass binary/store produces actionable errors without leaking secret values.
- `./scripts/test.sh` and architecture lint pass.

## Phase 31: Source Metadata and Project Semantics Hardening

**Goal:** Resolve mismatches exposed by gopass: path grammar, tags, env metadata, and diagnostics.

**Tasks:**
1. Decide whether `source.Reader` should return capabilities, e.g. supports tags / supports metadata.
2. Decide how gopass entries map to env names:
   - derived from path only;
   - sidecar metadata;
   - first-line/password plus YAML fields;
   - explicit `.shelf.json` env overrides only.
3. Make project diagnostics provider-aware but still backend-neutral.
4. Add tests for unsupported tag bindings against gopass if tags are not supported.
5. Document the supported gopass subset.

**Acceptance:**
- Project bindings fail predictably when users request unsupported backend features.
- No silent missing-secret success for unsupported tag/prefix behavior.
- Docs describe exact gopass path and metadata conventions.

## Phase 32: Shelf Vault Crypto Port Spike

**Goal:** Separate Shelf local vault encryption from Shelf vault storage so GPG can be evaluated without changing command semantics.

**Target packages:**
- `internal/crypto/age/`
- new `internal/crypto/gpg/` or `internal/crypto/` interface package
- `internal/adapters/shelfvault/`

**Tasks:**
1. Introduce the smallest crypto interface required by `shelfvault`:
   - encrypt bytes for configured recipients;
   - decrypt bytes with configured identities/keyring;
   - maybe key/recipient validation.
2. Keep `internal/crypto/age` as the first implementation.
3. Prototype `internal/crypto/gpg` behind that interface using `gpg` command execution with non-interactive flags.
4. Decide file format framing before saving anything:
   - keep `shelf-vault/v1` age-only and add `shelf-vault/v2` with crypto marker; or
   - add a separate `shelf-vault-gpg/v1` header.
5. Define failure modes for missing keys, locked keyring, pinentry, and batch mode.

**Acceptance:**
- Age vault tests still pass through the crypto interface.
- GPG prototype tests use a fake runner unless a controlled temporary GNUPGHOME smoke test is explicitly added.
- No existing age vault is written with ambiguous framing.

## Phase 33: GPG Shelf Vault MVP

**Goal:** Add GPG encryption as an optional Shelf-local vault crypto backend after the crypto port and file-format decision are verified.

**Tasks:**
1. Extend config with `vault.crypto: age|gpg` or equivalent.
2. Wire `shelfvault.NewVault` to choose age or GPG crypto based on config.
3. Add setup/init UX for GPG recipients separately from age identity generation.
4. Add status/doctor checks for GPG availability, recipient/key presence, and non-interactive decrypt behavior.
5. Add migration only if explicitly scoped after MVP; otherwise document that GPG requires a new vault file.

**Acceptance:**
- Age default behavior remains unchanged.
- GPG vault can save/load in controlled tests.
- Doctor reports actionable GPG setup failures.
- Existing scripts and architecture lint pass.

## Phase 34: Architecture Review and Cut Lines

**Goal:** Audit what the gopass and GPG work proved about package boundaries.

**Tasks:**
1. Review whether `internal/app` is too aware of concrete adapters.
2. Decide whether read-source selection needs a small factory package.
3. Decide whether Shelf write workflows need a repository port or should remain concrete.
4. Decide whether `internal/util` has grown enough to split atomic write or rendering back out.
5. Update `docs/architecture.md` and `.go-arch-lint.yml` to match implemented facts.

**Acceptance:**
- Architecture docs distinguish source adapters, Shelf vault repository, and crypto backends.
- No package imports contradict the intended dependency direction.
- Review records concrete follow-up decisions, not vague cleanup tasks.

## Verification Gates

Every implementation phase must run:

```text
go test ./internal/source ./internal/project ./internal/app ./internal/adapters/shelfvault ./internal/crypto/age
./scripts/test.sh
```

Additional gates by phase:

- Gopass adapter: unit tests for get/list/prefix/tag failure semantics; optional smoke test only when `gopass` is installed and an isolated store can be created.
- GPG crypto: fake-runner tests for exact command invocation and error mapping; optional controlled `GNUPGHOME` integration test only if deterministic in CI/local.
- Architecture updates: `go-arch-lint check --arch-file .go-arch-lint.yml --project-path ./` through `./scripts/test.sh`.

## Parallelization

- Gopass adapter implementation and GPG crypto spike can start after Phase 29 decisions, but they should not both modify config schema concurrently without a shared config contract.
- Project diagnostics/metadata hardening depends on the gopass MVP exposing real mismatches.
- Architecture review must wait until both a second source and a second crypto implementation exist or have been rejected with evidence.
