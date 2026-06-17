# Phase 1: Encrypted Vault Core - Context

**Gathered:** 2026-06-16
**Status:** Ready for planning

## Goal

Replace the durable secret-store boundary for existing `shelf secret` workflows with an age-encrypted vault. After decrypt/load and before encrypt/save, commands should continue operating on the existing plaintext `store.Data` model so command semantics remain stable.

## Constraints

- In scope: configuring an encrypted vault path, public age recipients, identity file paths, encrypted load/save, validation after decrypt, encrypted durable backups, and actionable errors for missing identity, wrong identity, corrupt vaults, unsupported vault formats, and legacy plaintext inputs.
- Out of scope: plaintext-to-encrypted migration, git/chezmoi safety checks, `shelf export`, `shelf project`, `shelf run`, localhost vault manager behavior, and full documentation hardening. Those are covered by later roadmap phases.
- Shelf config must remain non-secret. It may contain public age recipients and filesystem paths to identity files, but it must not contain private identity material or secret values.

## Decisions

### Vault Config Contract
- **D-01:** Introduce explicit vault-oriented config fields for encrypted storage, such as vault path, recipients, and identity paths. Downstream agents should choose exact names during planning, but the model should make encrypted vault configuration distinct from the legacy plaintext `data` field.
- **D-02:** Preserve compatibility for Phase 1 by treating existing `data`, `--data`, and `SHELF_DATA` as aliases for the active store path where practical. Existing scriptable command behavior should not break just because the underlying store becomes encrypted.
- **D-03:** Shelf config must remain non-secret. It may contain public age recipients and filesystem paths to identity files, but it must not contain private identity material or secret values.

### Age Recipients and Identities
- **D-04:** Use age as the encryption mechanism for Phase 1.
- **D-05:** Store public recipient strings in config. Store identity file paths or identity discovery inputs in config, not identity contents.
- **D-06:** Error messages should be actionable and name the failing class of problem: missing identity path, unreadable identity file, wrong identity/no matching identity, corrupt vault, unsupported vault format, or invalid decrypted store JSON.

### Vault File Format
- **D-07:** Prefer a Shelf vault envelope/header over raw age-encrypted store JSON. The exact technical shape is left to research/planning, but the file format must give Shelf enough information to identify a Shelf vault and reject unsupported versions cleanly.
- **D-08:** The decrypted payload should preserve the existing store model unless research finds a strong reason not to. The current `version: 1` JSON store remains the in-memory/plaintext domain model after decryption.
- **D-09:** Loading should distinguish at least these cases: missing file/new empty store, legacy plaintext JSON store, unsupported Shelf vault format/version, undecryptable age data, corrupt ciphertext, and decrypted JSON that fails store validation.

### Plaintext Side-File Policy
- **D-10:** Phase 1 must not create durable plaintext side files containing secret values. If `Store.Save` creates a backup next to the vault, that backup must be encrypted as well.
- **D-11:** Existing temp-file and atomic-write behavior should remain, but temp files for vault persistence must contain encrypted bytes, not plaintext JSON.
- **D-12:** `secret edit` editor temp-file hardening is important but not the center of Phase 1. Defer broad editor-temp cleanup unless the selected implementation naturally touches it or a test reveals a durable plaintext-at-rest regression.

### Agent Discretion
The planner may decide the exact config field names, envelope layout, and age Go package integration after research, as long as the decisions above hold. Favor minimal changes at the CLI command layer and a clear storage boundary under `internal/store`.

## Open Questions

None — the target mental model is age-compatible and chezmoi-friendly: a normal encrypted vault file can be moved, backed up, or committed, while Shelf config remains reviewable and does not contain private key material. Use explicit vault concepts in new config even if compatibility aliases remain. Prefer good diagnostics over a barely wrapped ciphertext file.

## Verification Expectations

Based on ROADMAP.md Phase 1 success criteria:

1. A user can configure Shelf to store secrets in an age-encrypted vault file.
2. Shelf config supports vault path, recipients, and identity locations without embedding private identity material.
3. `shelf secret set/get/list/info/edit/rm` can operate on the encrypted vault.
4. Wrong identity, missing identity, corrupt vault, and unsupported format errors are actionable.
5. The plaintext store model remains internal to load/decrypt and save/encrypt boundaries.

## References

**Downstream agents MUST read these before planning or implementing.**

### Phase Scope and Requirements
- `.planning/ROADMAP.md` — Phase 1 goal, requirements, success criteria, key risks, later-phase boundaries.
- `.planning/REQUIREMENTS.md` — VAULT-01 through VAULT-06 and CLI-01 for Phase 1, plus later requirements that must not be pulled into this phase.
- `.planning/PROJECT.md` — Core value, constraints, and brownfield architecture boundaries.

### Existing Store and CLI Semantics
- `docs/data-spec.md` — Current plaintext store shape, path grammar, storage policy, lock/save expectations, edit object format.
- `docs/usage-spec.md` — Current command behavior for `shelf secret` workflows that Phase 1 must preserve.

## Code Context

### Reusable Assets
- `internal/store/model.go`: Existing `Data`, `Secret`, and `NewData` types should remain the plaintext in-memory model after decrypt/load.
- `internal/store/io.go`: Current `Load` and `Store.Save` are the persistence boundary to wrap or split for encryption, validation, encrypted backup, and atomic encrypted writes.
- `internal/store/lock.go`: Existing write lock pattern should continue to serialize mutating secret commands against the active store path.
- `internal/config/config.go`: Existing config precedence and path expansion should be extended instead of reimplemented.
- `internal/cli/root.go`: `loadRuntime` and `loadRuntimeForWrite` are the command integration points for encrypted store loading and writing.
- `internal/cli/init.go`: `shelf init` currently creates a plaintext store and `data` config; Phase 1 needs an encrypted-vault-aware path without breaking existing initialization assumptions.

### Established Patterns
- Commands call `config.Resolve`, then `store.Load`; mutating commands use `loadRuntimeForWrite` to lock before loading.
- Store validation happens after decoding and before saving. Encrypted load should decrypt first, then reuse the same validation expectations.
- Store writes currently use temp file, fsync, optional backup, and rename. Preserve that durability shape while ensuring durable bytes are encrypted.
- User-facing errors are concise and lower-case, but encrypted-vault failures need enough context to be actionable.

### Integration Points
- `internal/config.Config` and `internal/config.Runtime` need enough fields to carry vault path, recipients, and identity paths to the store layer.
- `internal/store.Load` / `Store.Save` likely need new options or a backend/config abstraction so callers do not need to know encryption details.
- Existing `shelf secret set/get/list/info/edit/rm` tests should be reused or extended to prove command semantics survive encrypted storage.

## Deferred

- Plaintext store migration, source preservation, migration next steps, encrypted recovery artifacts, and migration cleanup belong to Phase 2.
- `shelf doctor` encrypted/plaintext reporting and git/chezmoi safety checks belong to Phase 2.
- `shelf export`, `shelf project`, and `shelf run` compatibility and regression hardening belong to Phase 3.
- Localhost vault manager read/write/reveal controls belong to Phase 4.
- User documentation, plaintext export warnings, and release verification belong to Phase 5.

---

*Phase: 001-encrypted-vault-core*
*Context gathered: 2026-06-16*
