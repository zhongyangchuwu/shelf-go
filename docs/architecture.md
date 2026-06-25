# Architecture

Shelf Go is a local-first Go CLI. The durable store is an age-encrypted vault file; project bindings are value-free JSON manifests; the localhost manager is an on-demand UI over the same vault operations.

## Package layers

```text
cmd/shelf/          process entry point

internal/cli/       Cobra command tree, flags, argument validation, and text rendering
internal/manager/   loopback HTTP vault manager UI

internal/app/       runtime and vault construction helpers
internal/project/   project identity and manifest binding resolution
internal/secret/    reusable secret workflows such as editor-based updates
internal/vault/     vault status/check/doctor diagnostics

internal/atomicfile/ atomic write primitive
internal/config/     runtime config resolution
internal/store/      secret model, path grammar, JSON codec, age vault persistence, locking
internal/manifest/   .shelf.json model, validation, and IO
internal/render/     env/shell/JSON rendering
internal/version/    version string
```

The intended dependency direction is display to feature support to base support:

```text
cmd/shelf -> internal/cli
internal/cli -> app, project, secret, vault, manager, config, store, manifest, render, version
internal/manager -> store, render
app -> config, store
project -> manifest, store, render
secret -> store
vault -> config, store
manifest -> store
render -> store
store -> atomicfile
```

Base packages must not import `internal/cli` or `internal/manager`.

## Command layer

`internal/cli` owns user-facing command shape:

- root command setup and global flags;
- `setup` / `vault` commands;
- `secret` commands;
- `project` commands;
- `doctor`;
- shell completion.

Command handlers should stay thin: parse flags, call feature/base packages, then render output through Cobra writers. Reusable behavior belongs outside `internal/cli` once it is not purely command presentation.

## Runtime and vault construction

`internal/app` centralizes runtime and vault loading:

- `LoadVault(configPath, vaultPath)` resolves config and constructs `*store.Vault`;
- `LoadRuntime(configPath, vaultPath)` loads a decrypted store snapshot;
- `ReadVault(configPath, vaultPath, fn)` runs read-only vault work;
- `UpdateVault(configPath, vaultPath, fn)` locks, loads, mutates, and encrypted-saves through `store.Vault.Update`.

This keeps command files independent from vault construction details.

## Store and persistence

`internal/store` owns the secret data model and encrypted vault persistence.

Current file responsibilities:

- `model.go`: `Data`, `Secret`, `Info`, `CurrentVersion`, `NewData`;
- `path.go`: secret path parsing and path-token validation;
- `validate.go`: secret validation and env-name validation;
- `store.go`: in-memory `Store` methods;
- `json.go`: strict JSON encode/decode for the plaintext model;
- `age.go`: age recipient parsing, encryption, identity loading, and decryption;
- `vault.go`: vault file format detection and encrypted vault orchestration;
- `io.go`: legacy plaintext store load/save support used by migration tests and compatibility paths;
- `lock.go`: file locking for vault writes.

`internal/atomicfile` provides the shared atomic write primitive. Store and vault writes use restrictive permissions, sync, and a single last-write `.bak` backup.

There is intentionally no storage backend interface yet. A second backend should be introduced only after a concrete storage spike proves the need.

## Project binding resolution

`internal/project` owns Git project identity and manifest-to-env resolution.

Resolution order for env names:

1. manifest entry `env`;
2. secret object's `env`;
3. env name derived from the full secret path.

Prefix manifest entries may expand to multiple secrets and cannot carry `env`. Required missing entries and duplicate env names are diagnostics; commands decide whether diagnostics are fatal.

## Secret edit workflow

`internal/secret` owns reusable editor-based secret updates. It converts a stored secret into editable JSON, invokes the configured editor through the CLI seam, validates the edited object, and ensures the temporary plaintext file is removed on normal exits where possible.

## Vault diagnostics

`internal/vault` produces typed status records for config, recipient configuration, vault format, permissions, loadability, Git tracking, and backup checks. `internal/cli` renders those records for `vault status`, `vault check`, and `doctor`.

## Localhost manager

`internal/manager` is an on-demand loopback HTTP UI. It receives a `*store.Vault`, uses the same store validation and encrypted-save path as CLI writes, and does not run as a permanent daemon.

Safety boundaries:

- loopback-only address validation in CLI before server start;
- tokenized URL and strict cookie;
- Host and Origin checks;
- metadata list/search without values;
- explicit reveal for plaintext values.

## Public documentation boundary

Public docs describe current behavior only. Planning state, phase history, and architecture refactor records stay under `.planning/`.
