# Architecture

Shelf Go is a local-first Go CLI. The durable store is an age-encrypted vault file; project bindings are value-free JSON manifests; the local manager is an on-demand loopback surface over the same vault operations.

## Package layers

```text
cmd/shelf/          process entry point

internal/cli/       Cobra command tree, flags, argument validation, and text rendering
internal/manager/   local manager surface, currently loopback HTTP/Web

internal/app/       runtime, vault/source construction, and version composition
internal/project/   project identity, .shelf.json schema/IO/validation, and binding resolution
internal/secret/    reusable secret workflows such as editor-based updates

internal/config/     runtime config resolution
internal/source/     backend-neutral secret reader contract and Shelf vault adapter
internal/vault/      encrypted vault core: model, JSON codec, age persistence, locking, diagnostics
internal/exportfmt/  env/shell/JSON export formatting
```

The intended dependency direction is local surface to workflow to kernel/support, enforced by `.go-arch-lint.yml`:

```text
Surface:
  cmd/shelf -> internal/cli
  internal/cli -> app, project, manager
  internal/manager -> app

Workflow:
  app -> config, vault, source, project, secret, exportfmt
  project -> source, vault, exportfmt
  source -> vault
  secret -> vault

Kernel/support:
  exportfmt -> vault
  config -> standard library + YAML
  vault -> standard library + age/flock dependencies
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

Command handlers should stay thin: parse flags, call feature/base packages, then render output through Cobra writers. Reusable behavior belongs outside `internal/cli` once it is not purely command presentation.

## Runtime and vault construction

`internal/app` centralizes runtime, vault loading, and version composition:

- `LoadVault(configPath, vaultPath)` resolves config and constructs `*vault.Vault`;
- `LoadRuntime(configPath, vaultPath)` loads a decrypted vault store snapshot;
- `LoadSecretReader(configPath, vaultPath)` adapts the current Shelf vault into the backend-neutral source reader used by project workflows;
- `ReadVault(configPath, vaultPath, fn)` runs read-only vault work;
- `UpdateVault(configPath, vaultPath, fn)` locks, loads, mutates, and encrypted-saves through `vault.Vault.Update`;
- `String()` returns the application version string from release ldflags or Go build info.

This keeps command files independent from vault construction and build-info details.

## Source boundary

`internal/source` defines the read-side contract for project env resolution. `source.Reader` exposes only exact lookup, prefix listing, and tag listing; it returns backend-neutral `source.Secret` values with string material plus optional env/description/tag metadata.

The current implementation is `source.VaultReader`, an adapter over `internal/vault.Store`. Future gopass, 1Password, or Bitwarden integrations should enter through this package so `internal/project` keeps resolving manifests without knowing the provider.

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
- `status.go`: typed status records for vault status/check/doctor diagnostics.

Atomic replacement remains in `internal/vault` for now because existing project manifest and config write paths reuse that primitive; it should move only with a dedicated write-primitive refactor.

There is intentionally no storage backend interface for writes yet. Project env resolution uses `internal/source.Reader`, with the current age-encrypted JSON vault provided through a Shelf vault adapter. Additional read-only providers such as gopass, 1Password, or Bitwarden should implement that source boundary before any broader write/sync backend abstraction is introduced.

## Project workflows

`internal/project` owns Git project identity, `.shelf.json` schema/IO/validation, and manifest-to-env resolution.

Resolution order for env names:

1. manifest entry `env`;
2. secret object's `env`;
3. env name derived from the full secret path.

Prefix and tag manifest entries may expand to multiple secrets and cannot carry `env`. Tag selectors use AND semantics and are stored as value-free `tags` arrays. Required missing entries and duplicate env names are diagnostics; commands decide whether diagnostics are fatal.

## Secret edit workflow

`internal/secret` owns reusable editor-based secret updates. It converts a stored secret into editable JSON, invokes the configured editor through the CLI seam, validates the edited object, and ensures the temporary plaintext file is removed on normal exits where possible.

## Export formatting

`internal/exportfmt` formats vault secrets and project bindings as env, shell, or JSON output. It is not UI rendering; manager UI assets stay in `internal/manager`.

## Local manager

`internal/manager` is an on-demand local manager surface. Today it is implemented as loopback HTTP/Web, but the package name is intentionally not vault-only or Web-only so future config/project panels can live behind the same manager concept. The public entrypoint is `shelf manager`; there is no `shelf vault open` alias.

Safety boundaries:

- loopback-only address validation before server start;
- tokenized URL with token removal from the visible URL after first load;
- strict cookie, Host checks, and Origin checks;
- no-store responses for manager pages and API responses;
- metadata list/search/detail without values;
- explicit POST reveal/copy flows for plaintext values;
- embedded local HTML/CSS/JS assets, no CDN or permanent daemon requirement.

Manager route handlers call app-layer secret workflows instead of mutating vault state directly.

## Public documentation boundary

Public docs describe current behavior only. Planning state, phase history, and architecture refactor records stay under `.planning/`.
