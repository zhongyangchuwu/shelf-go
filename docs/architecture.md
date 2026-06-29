# Architecture

Shelf Go is a local-first Go CLI. The default durable store is an age-encrypted Shelf vault file; project bindings are value-free JSON manifests; the local manager is an on-demand loopback surface over the same application services.

## Package layers

```text
cmd/shelf/                    process entry point

internal/cli/                 Cobra command tree, flags, argument validation, and text rendering
internal/manager/             local manager surface, currently loopback HTTP/Web

internal/app/                 runtime construction, workflow orchestration, and version composition
internal/project/             project identity, .shelf.json schema/IO/validation, and binding resolution
internal/secret/              reusable secret workflows such as editor-based updates

internal/source/              backend-neutral secret reader contract
internal/adapters/shelfvault/ current local encrypted Shelf vault adapter and repository
internal/adapters/gopass/     read-only gopass source adapter for project workflows
internal/crypto/age/          age encryption, decryption, and identity helpers
internal/config/              runtime config resolution
internal/util/                small shared primitives: atomic write and env/shell/JSON binding formatting
```

The intended dependency direction is local surface to workflow to kernel/support, enforced by `.go-arch-lint.yml`:

```text
Surface:
  cmd/shelf -> internal/cli
  internal/cli -> app, project, manager
  internal/manager -> app

Workflow:
  app -> config, project, secret, source, shelfvault, gopass, crypto/age, util
  project -> source, util
  secret -> shelfvault

Adapters/support:
  adapters/shelfvault -> source, crypto/age, util, flock
  adapters/gopass -> source
  crypto/age -> filippo.io/age
  source -> util
  config -> YAML
  util -> standard library
```

Base packages must not import `internal/cli` or `internal/manager`. Feature packages should expose concrete functions and data types, not speculative backend interfaces.

## Command layer

`internal/cli` owns user-facing command shape:

- root command setup and global flags;
- `setup` / `vault` lifecycle commands;
- `manager` local manager command;
- `secret` commands;
- `project` commands;
- `doctor`;
- shell completion.

Command handlers should stay thin: parse flags, call feature/base packages, then render output through Cobra writers. The package is broad because Cobra command construction is broad; splitting it before a concrete repeated command subsystem appears would mostly duplicate shared completion, diagnostic, runtime flag, and process-exit helpers.

## Runtime and vault construction

`internal/app` centralizes runtime, Shelf vault loading, source selection, and version composition:

- `LoadVault(configPath, vaultPath)` resolves config and constructs `*shelfvault.Vault`;
- `LoadRuntime(configPath, vaultPath)` loads a decrypted Shelf vault store snapshot;
- `LoadSecretReader(configPath, vaultPath)` selects the configured read source for project workflows;
- `ReadVault(configPath, vaultPath, fn)` runs read-only Shelf vault work;
- `UpdateVault(configPath, vaultPath, fn)` locks, loads, mutates, and encrypted-saves through `shelfvault.Vault.Update`;
- `String()` returns the application version string from release ldflags or Go build info.

Only project workflows use non-Shelf sources today. Secret CRUD, manager editing, setup, status, and migration remain concrete Shelf vault workflows.

## Source boundary

`internal/source` defines the read-side contract for project env resolution. `source.Reader` exposes only exact lookup, prefix listing, and tag listing; it returns backend-neutral `source.Secret` values with string material plus optional env/description/tag metadata. This package must stay provider-neutral and must not import concrete backends.

Implemented source adapters:

- `internal/adapters/shelfvault.Reader`: adapts the local Shelf vault store.
- `internal/adapters/gopass.Reader`: shells out to the `gopass` CLI for read-only project workflows. It maps Shelf paths like `app:token` to gopass paths like `app/token`, derives env names from paths unless `.shelf.json` provides `env`, and currently reports tag selectors as unsupported.

Future 1Password or Bitwarden integrations should enter under `internal/adapters/` so `internal/project` keeps resolving manifests without knowing the provider.

## Shelf vault adapter and persistence

`internal/adapters/shelfvault` owns the current local encrypted JSON vault implementation. It is both a source adapter and the current concrete repository used by app/secret workflows.

Current file responsibilities:

- `model.go`: `Data`, `Secret`, `Info`, `CurrentVersion`, `NewData`;
- `path.go`: secret path parsing and path-token validation;
- `validate.go`: secret validation and env-name validation;
- `store.go`: decrypted in-memory `Store` snapshot methods;
- `json.go`: strict JSON encode/decode for the plaintext model;
- `age.go`: Shelf vault file framing around `internal/crypto/age`;
- `vault.go`: vault file format detection and encrypted vault orchestration;
- `io.go`: legacy plaintext store load/save support used by migration tests and compatibility paths;
- `lock.go`: file locking for vault writes;
- `reader.go`: `source.Reader` implementation for project resolution;
- `status.go`: typed status records for vault status/check/doctor diagnostics.

There is intentionally no storage backend interface for writes yet. Project env resolution uses `internal/source.Reader`; additional read-only providers should implement that source boundary before any broader write/sync backend abstraction is introduced.

## Crypto boundary

`internal/crypto/age` owns direct `filippo.io/age` use: encrypt/decrypt helpers and X25519 identity read-or-create behavior. Shelf vault framing, headers, JSON decode, lock orchestration, and diagnostics stay in `internal/adapters/shelfvault`.

## Project workflows

`internal/project` owns Git project identity, `.shelf.json` schema/IO/validation, and manifest-to-env resolution.

Resolution order for env names:

1. manifest entry `env`;
2. secret object's `env`;
3. env name derived from the full secret path.

Prefix and tag manifest entries may expand to multiple secrets and cannot carry `env`. Tag selectors use AND semantics and are stored as value-free `tags` arrays. Required missing entries and duplicate env names are diagnostics; commands decide whether diagnostics are fatal.

## Secret edit workflow

`internal/secret` owns reusable editor-based secret updates. It converts a stored secret into editable JSON, invokes the configured editor through the CLI seam, validates the edited object, and ensures the temporary plaintext file is removed on normal exits where possible.

## Utilities

`internal/util` holds small shared helpers that are not domain concepts yet: atomic file replacement and env/shell/JSON binding formatting. If either area grows into a cohesive subsystem again, split it back out with that concrete pressure.

## Local manager

`internal/manager` is an on-demand local manager surface. Today it is implemented as loopback HTTP/Web, but the package name is intentionally not vault-only or Web-only so future config/project panels can live behind the same manager concept. The public entrypoint is `shelf manager`; there is no `shelf vault open` alias.

Safety boundaries:

- loopback-only address validation before server start;
- tokenized URL with token removal from the visible URL after first load;
- strict cookie, Host checks, and Origin checks;
- no-store responses for manager pages and API responses;
- metadata list/search/detail without values;
- explicit reveal endpoint for plaintext secret values;
- no external network dependency.

Public docs describe current behavior only. Planning state, phase history, and architecture refactor records stay under `.planning/`.
