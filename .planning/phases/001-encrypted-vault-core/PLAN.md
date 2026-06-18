# Phase 1 Plan: Encrypted Vault Core

## Objective

Existing `shelf secret` workflows read and write an age-encrypted vault file while keeping the current `store.Data` model as the plaintext in-memory boundary after decrypt/load and before validate/encrypt/save.

## Scope

- Add vault-first config fields for vault path, age recipients, and age identity paths.
- Intentionally remove plaintext-era `data`, `--data`, and `SHELF_DATA` compatibility for Phase 1 per breaking-change direction.
- Add a Shelf vault envelope/header so Shelf distinguishes missing/new store, legacy plaintext JSON, unsupported vault versions, corrupt ciphertext, no matching identity, and invalid decrypted JSON.
- Encrypt durable vault writes, temp files, and backups in vault mode.
- Keep `shelf secret add/set/get/list/info/edit/rm` command semantics stable over encrypted storage.

## Non-goals

- Plaintext-to-encrypted migration flow.
- `shelf doctor` encrypted/plaintext reporting and git safety checks.
- `shelf export`, `shelf project`, and `shelf run` regression hardening beyond preserving shared load behavior.
- Localhost vault manager.
- Full user documentation hardening.
- Broad `secret edit` editor temp-file hardening, except no plaintext may be written by vault persistence itself.

## Inputs Read

- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/001-encrypted-vault-core/CONTEXT.md`
- `.planning/phases/001-encrypted-vault-core/DISCUSSION-LOG.md`
- `internal/store/io.go`, `model.go`, `lock.go`, `path.go`, `validate.go`, `io_test.go`
- `internal/config/config.go`
- `internal/cli/root.go`, `init.go`, `secret.go`, `secret_test.go`
- Official age Go API evidence: `filippo.io/age` exposes `Encrypt`, `Decrypt`, `ParseX25519Recipient`, `ParseIdentities`, and `NoIdentityMatchError`.

## Tasks

### 1. Add vault config contract

Affected files:

- `internal/config/config.go`
- `internal/cli/root.go`
- config/init tests as needed

Steps:

1. Extend `config.Config` with vault path, age recipients, and age identity paths.
2. Extend `config.Runtime` with the resolved vault path, recipients, and identity paths.
3. Use vault-first path precedence: `--vault` -> `SHELF_VAULT` -> config `vault_path` -> default vault path.
4. Expand vault path and identity paths relative to config file directory when configured in YAML.
5. Keep private identity contents out of config; store only paths.

Acceptance:

- Vault config can resolve a vault path plus recipient and identity path lists.
- Plaintext-era `--data`, `SHELF_DATA`, and config `data` are not part of the Phase 1 interface.

### 2. Add encrypted store load/save boundary

Affected files:

- `internal/store/io.go`
- new `internal/store/vault.go` if clearer
- `internal/store/io_test.go` or new focused store tests

Steps:

1. Add a dedicated vault persistence type for encrypted load/save inputs.
2. Keep `store.Store` as the plaintext in-memory model.
3. Add vault-aware load and save methods used by CLI runtime helpers.
4. Use a Shelf header before age ciphertext, e.g. `shelf-vault/v1\n`.
5. On load:
   - missing or empty file returns `NewData()`;
   - vault header decrypts age payload, strict-decodes JSON, validates store model;
   - legacy plaintext JSON is rejected in vault mode with an actionable error;
   - unsupported Shelf vault header/version is rejected cleanly;
   - age no-identity-match and parse/read failures are translated into actionable errors.
6. On save:
   - validate store model first;
   - marshal plaintext only in memory;
   - encrypt bytes before any durable write;
   - copy existing encrypted vault bytes to `.bak` when replacing;
   - write only encrypted bytes to temp file before rename.

Acceptance:

- No vault-mode `.bak` or temp persistence contains plaintext secret JSON.
- Plaintext `store.Load` / `store.Save(path, st)` remains available only for internal legacy/plaintext paths.
- Load errors distinguish at least legacy plaintext, unsupported vault, unreadable/missing identity, wrong identity, corrupt ciphertext, and invalid decrypted JSON.

### 3. Wire CLI runtime to vault boundary

Affected files:

- `internal/cli/root.go`
- `internal/cli/init.go` only if needed for Phase 1 smoke path
- `internal/cli/secret_test.go`

Steps:

1. Change CLI runtime helpers to construct `store.Vault` from runtime config.
2. Ensure returned `*store.Store` still exposes plaintext `Data` and existing methods.
3. Persist mutating commands through `Vault.Update` / `Vault.Save`, not `Store.Save`.
4. Preserve locking on the vault path.
5. Avoid changing command-specific logic unless a command currently bypasses the store boundary.

Acceptance:

- `shelf secret set/get/list/info/edit/rm` work through encrypted vault config.
- Existing tests use the new vault-first interface unless explicitly exercising internal plaintext helpers.

### 4. Add focused behavior tests

Affected files:

- `internal/config/config_test.go` if needed
- `internal/store/*_test.go`
- `internal/cli/secret_test.go`

Required checks:

1. Vault save writes bytes that do not contain plaintext secret values, and load decrypts the same data.
2. Vault replacement backup is encrypted and loadable.
3. Vault mode rejects legacy plaintext JSON with an actionable error.
4. Wrong identity or missing identity produces actionable errors.
5. At least one CLI secret write/read flow works against vault config.

Acceptance:

- Tests assert behavior and invariants, not incidental helper calls.
- Tests use real age identities generated in test code, not mocks of the encryption boundary.

## Verification Steps

- Run focused tests covering changed packages, at minimum:
  - `go test ./internal/config ./internal/store ./internal/cli`
- If focused tests pass and changes are broad, run:
  - `go test ./...`

## Risks and Escalation Points

- Adding `filippo.io/age` changes dependencies; use the maintained package API and keep the wrapper thin.
- CLI default behavior could accidentally flip plaintext stores into encrypted mode. Vault mode should require recipients or identities configured; legacy plaintext remains compatible until migration phase.
- `secret edit` still writes a plaintext editor temp file by design; this is deferred in context D-12 and must not be confused with vault persistence temp files.
- `doctor` may validate encrypted vault loading, but full git/chezmoi safety classification remains Phase 2 scope.
