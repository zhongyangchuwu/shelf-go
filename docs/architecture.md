# Architecture

Shelf Go is a local-first Go CLI. The durable store is an age-encrypted vault file; project bindings are value-free JSON manifests; the local manager is an on-demand loopback surface over the same vault operations.

## Package layers

```text
cmd/shelf/          process entry point

internal/cli/       Cobra command tree, flags, argument validation, and text rendering
internal/manager/   local manager surface, currently loopback HTTP/Web

internal/app/       runtime, vault construction, and version composition
internal/project/   project identity, .shelf.json schema/IO/validation, and binding resolution
internal/secret/    reusable secret workflows such as editor-based updates

internal/config/     runtime config resolution
internal/vault/      encrypted vault core: model, path grammar, JSON codec, age persistence, locking, diagnostics
internal/exportfmt/  env/shell/JSON export formatting
```

The intended dependency direction is display to feature support to base support:

```text
cmd/shelf -> internal/cli
internal/cli -> app, project, secret, vault, manager, config, exportfmt
internal/manager -> vault, exportfmt
app -> config, vault
project -> vault, exportfmt
secret -> vault
exportfmt -> vault
```

Base packages must not import `internal/cli` or `internal/manager`.

## Command layer

`internal/cli` owns user-facing command shape:

- root command setup and global flags;
- `setup` / `vault` lifecycle commands;
- `manager` local manager command;
- `secret` commands;
- `project` commands;
- `doctor`;
- shell completion.

Command handlers should stay thin: parse flags, call feature/base packages, then render output through Cobra writers. Reusable behavior belongs outside `internal/cli` once it is not purely command presentation.

## Runtime and vault construction

`internal/app` centralizes runtime, vault loading, and version composition:

- `LoadVault(configPath, vaultPath)` resolves config and constructs `*vault.Vault`;
- `LoadRuntime(configPath, vaultPath)` loads a decrypted vault store snapshot;
- `ReadVault(configPath, vaultPath, fn)` runs read-only vault work;
- `UpdateVault(configPath, vaultPath, fn)` locks, loads, mutates, and encrypted-saves through `vault.Vault.Update`;
- `String()` returns the application version string from release ldflags or Go build info.

This keeps command files independent from vault construction and build-info details.

## Vault core and persistence

`internal/vault` owns the secret data model, encrypted vault persistence, and vault diagnostics.

Current file responsibilities:

- `model.go`: `Data`, `Secret`, `Info`, `CurrentVersion`, `NewData`;
- `path.go`: secret path parsing and path-token validation;
- `validate.go`: secret validation and env-name validation;
- `store.go`: decrypted in-memory `Store` snapshot methods;
- `json.go`: strict JSON encode/decode for the plaintext model;
- `age.go`: age recipient parsing, encryption, identity loading, and decryption;
- `vault.go`: vault file format detection and encrypted vault orchestration;
- `io.go`: legacy plaintext store load/save support used by migration tests and compatibility paths;
- `lock.go`: file locking for vault writes;
- `atomicfile.go`: atomic write primitive for vault/config/manifest writes;
- `status.go`: typed status records for vault status/check/doctor diagnostics.

There is intentionally no storage backend interface yet. A second backend should be introduced only after a concrete storage spike proves the need.

## Project workflows

`internal/project` owns Git project identity, `.shelf.json` schema/IO/validation, and manifest-to-env resolution.

Resolution order for env names:

1. manifest entry `env`;
2. secret object's `env`;
3. env name derived from the full secret path.

Prefix and tag manifest entries may expand to multiple secrets and cannot carry `env`. Required missing entries and duplicate env names are diagnostics; commands decide whether diagnostics are fatal.

## Secret edit workflow

`internal/secret` owns reusable editor-based secret updates. It converts a stored secret into editable JSON, invokes the configured editor through the CLI seam, validates the edited object, and ensures the temporary plaintext file is removed on normal exits where possible.

## Export formatting

`internal/exportfmt` formats vault secrets and project bindings as env, shell, or JSON output. It is not UI rendering; manager UI assets stay in `internal/manager`.

## Local manager

`internal/manager` is an on-demand local manager surface. Today it is implemented as loopback HTTP/Web, but the package name is intentionally not vault-only or Web-only so future config/project panels can live behind the same manager concept.

Safety boundaries:

- loopback-only address validation in CLI before server start;
- tokenized URL and strict cookie;
- Host and Origin checks;
- metadata list/search without values;
- explicit reveal for plaintext values.

## Public documentation boundary

Public docs describe current behavior only. Planning state, phase history, and architecture refactor records stay under `.planning/`.
