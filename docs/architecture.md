# Architecture

Shelf Go is a local-first Go CLI. The runtime secret store is a local encrypted Shelf vault. External systems such as gopass are import sources, not runtime backends.

## Package layers

```text
cmd/shelf/                    process entry point

internal/cli/                 Cobra command tree, flags, argument validation, and text rendering
internal/manager/             local manager surface, currently loopback HTTP/Web

internal/app/                 application service, runtime construction, workflow orchestration, import services, and version composition
internal/project/             project identity, .shelf.json schema/IO/validation, and binding resolution
internal/secret/              reusable secret workflows such as editor-based updates

internal/vault/               Shelf vault domain model, repository abstraction, path/env/tag rules, and in-memory store
internal/jsonvault/           current shelf-vault/v1 encrypted JSON implementation
internal/age/                 age encryption/decryption and identity helpers
internal/importer/gopass/     gopass CLI import client
internal/config/              runtime config resolution
internal/util/                small shared primitives: atomic write and env/shell/JSON binding formatting
```

The intended dependency direction is local surface to workflow to domain/persistence/support, enforced by `.go-arch-lint.yml`:

```text
Surface:
  cmd/shelf -> internal/cli, internal/app
  internal/cli -> app, manager
  internal/manager -> app

Workflow:
  app -> config, project, secret, vault, jsonvault, importer/gopass, util
  project -> vault, util
  secret -> vault

Domain/persistence/support:
  jsonvault -> vault, age, util, flock
  age -> filippo.io/age
  importer/gopass -> standard library
  config -> YAML
  util -> standard library
```

Base packages must not import `internal/cli` or `internal/manager`. Feature packages should expose concrete functions and data types, not speculative backend interfaces.

## Product data model

Shelf's source of truth is the local Shelf vault. gopass, 1Password, Bitwarden, or other password managers should first enter as importers that copy data into the Shelf vault model.

Runtime commands read local vault data only:

- `secret` CRUD;
- manager list/reveal/edit/delete;
- `project explain/export`;
- `shelf run`;
- `secret export`.

This avoids split-brain behavior where one command reads gopass while another command reads the local vault.

## Vault domain

`internal/vault` owns Shelf's vault data model and rules:

- `Data`, `Secret`, `Info`, `CurrentVersion`, `NewData`;
- `SecretID`, `ParseSecretID`, `ValidatePath`, `ValidateSecretID`;
- `ValidateSecret`, `ParseValue`, env/tag validation and env-name derivation;
- in-memory `Store` mutation/list/query methods;
- application-facing vault repository contracts: `Options`, `Repository`, `Provider`, and status `Report`.

The domain package defines the application-facing contract for an encrypted Shelf vault, including recipients and identity paths needed to open one. It does not know the file format, encryption algorithm, lock strategy, or project manifests. Future SQL/NoSQL persistence should implement the `vault.Repository` contract rather than replacing the model prematurely.

## JSON vault persistence

`internal/jsonvault` owns the current `shelf-vault/v1` encrypted JSON implementation:

- strict JSON plaintext model encode/decode;
- `shelf-vault/v1` encrypted file framing;
- age sealing/opening through `internal/age`;
- file format detection;
- file lock and atomic write/backup behavior;
- vault status/check diagnostics behind the `vault.Provider` contract;
- legacy plaintext store load/save used by migration paths.

This is the current production repository for local vault data. It is not the entire Shelf vault concept and should not be treated as the generic file-storage abstraction for future formats.

## Age helper

`internal/age` owns age encryption/decryption and identity helpers:

- `Identity`;
- `ReadOrCreateIdentity`;
- `Encrypt`;
- `Decrypt`.

`internal/age` does not own vault file headers, JSON format, locks, or status diagnostics. Those stay in `jsonvault`. Future GPG support should start as its own helper or a small crypto port after file-format framing is decided.

## Importers

`internal/importer/gopass` is a gopass CLI client for imports. It shells out to:

- `gopass list --flat`;
- `gopass show --password <path>`.

It is not a runtime backend. Importers copy secrets into the local Shelf vault before project/runtime commands can use them.

`internal/app.ImportGopassForRuntime` maps gopass entries into the local Shelf vault:

- gopass `a/b/c` maps to Shelf `a/b:c`;
- entries without `/` are skipped as unmappable;
- imported passwords are stored as JSON strings;
- existing local secrets are skipped unless `--force` is passed;
- read failures abort before writing to avoid partial imports.

The CLI entrypoint is `shelf vault import gopass`.

## Command layer

`internal/cli` owns user-facing command shape:

- root command setup and global flags;
- `setup` / `vault` lifecycle commands;
- `vault import gopass`;
- `manager` local manager command;
- `secret` commands;
- `project` commands;
- `doctor`;
- shell completion.

Command handlers should stay thin: parse flags, call an injected `*app.App`, then render output through Cobra writers. Project/run orchestration belongs in `internal/app`, not in CLI command handlers. `app.NewDefault` wires the current `jsonvault.Provider{}` implementation; CLI and `cmd/shelf` do not import `internal/jsonvault`.

## Runtime construction

`internal/app` centralizes runtime and local vault loading behind the `internal/vault` repository abstraction:

- `app.NewDefault()` constructs the default app using the current `jsonvault.Provider{}` implementation;
- `App.LoadVault(configPath, vaultPath)` resolves config and opens a `vault.Repository` through the app provider;
- `App.LoadRuntime(configPath, vaultPath)` loads a decrypted local vault `*vault.Store` snapshot;
- `App.ReadVault(configPath, vaultPath, fn)` runs read-only local vault work;
- `App.UpdateVault(configPath, vaultPath, fn)` locks, loads, mutates, and saves through `vault.Repository.Update`;
- `App.ProjectRun(req)` resolves the current project manifest against local vault data and runs the child process;
- `String()` returns the application version string from release ldflags or Go build info.

## Project workflows

`internal/project` owns Git project identity, `.shelf.json` schema/IO/validation, and manifest-to-env resolution.

Resolution uses a local `*vault.Store` snapshot supplied by `internal/app`. Resolution order for env names:

1. manifest entry `env`;
2. secret object's `env`;
3. env name derived from the full secret path.

Prefix and tag manifest entries may expand to multiple local vault secrets and cannot carry `env`. Required missing entries and duplicate env names are diagnostics; commands decide whether diagnostics are fatal.

## Secret edit workflow

`internal/secret` owns reusable editor-based secret updates. It converts a stored secret into editable JSON, invokes the configured editor through the CLI seam, validates the edited object, and ensures the temporary plaintext file is removed on normal exits where possible.

## Utilities

`internal/util` holds small shared helpers that are not domain concepts yet: atomic file replacement and env/shell/JSON binding formatting. If either area grows into a cohesive subsystem again, split it back out with that concrete pressure.

## Local manager

`internal/manager` is an on-demand local manager surface. Today it is implemented as loopback HTTP/Web, but the package name is intentionally not vault-only or Web-only so future config/project panels can live behind the same manager concept.

Safety boundaries:

- loopback-only address validation before server start;
- tokenized URL with token removal from the visible URL after first load;
- strict cookie, Host checks, and Origin checks;
- no-store responses for manager pages and API responses;
- metadata list/search/detail without values;
- explicit reveal endpoint for plaintext secret values;
- no external network dependency.
