# Architecture

Shelf Go is a local-first Go CLI. The runtime secret store is a local encrypted Shelf vault. External systems such as gopass are import sources, not runtime backends.

## Package layers

```text
cmd/shelf/                    process entry point

internal/cli/                 Cobra command tree, flags, argument validation, and text rendering
internal/manager/             local manager surface, currently loopback HTTP/Web

internal/app/                 runtime construction, workflow orchestration, import services, and version composition
internal/project/             project identity, .shelf.json schema/IO/validation, and binding resolution
internal/secret/              reusable secret workflows such as editor-based updates

internal/source/              project-resolution reader contract
internal/vault/               Shelf vault domain model, path/env/tag rules, in-memory store, and local reader
internal/vaultfile/           current encrypted JSON file vault implementation
internal/vaultcrypto/         vault encryption helpers; currently age-specific
internal/importer/gopass/     gopass CLI import client
internal/config/              runtime config resolution
internal/util/                small shared primitives: atomic write and env/shell/JSON binding formatting
```

The intended dependency direction is local surface to workflow to domain/persistence/support, enforced by `.go-arch-lint.yml`:

```text
Surface:
  cmd/shelf -> internal/cli
  internal/cli -> app, project, manager
  internal/manager -> app

Workflow:
  app -> config, project, secret, source, vault, vaultfile, importer/gopass, util
  project -> source, util
  secret -> vault

Domain/persistence/support:
  vault -> source, util
  vaultfile -> vault, vaultcrypto, util, flock
  vaultcrypto -> filippo.io/age
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
- `ValidateSecret`, `ParseValue`, env/tag validation;
- in-memory `Store` mutation/list/query methods;
- `Reader`, which adapts a local `Store` to `source.Reader` for project resolution.

The domain package does not know how data is persisted or encrypted. Future SQL/NoSQL persistence should depend on `internal/vault`, not replace its model prematurely.

## Vault file persistence

`internal/vaultfile` owns the current file implementation:

- strict JSON plaintext model encode/decode;
- `shelf-vault/v1` encrypted file framing;
- age sealing/opening through `internal/vaultcrypto`;
- file format detection;
- file lock and atomic write/backup behavior;
- vault status/check diagnostics;
- legacy plaintext store load/save used by migration paths.

This is the current production repository for local vault data. It is not the entire Shelf vault concept.

## Vault crypto

`internal/vaultcrypto` owns vault encryption helpers. Current exports are age-specific:

- `AgeIdentity`;
- `ReadOrCreateAgeIdentity`;
- `EncryptAge`;
- `DecryptAge`.

`vaultcrypto` does not own vault file headers, JSON format, locks, or status diagnostics. Those stay in `vaultfile`. Future GPG support should add GPG-specific helpers or a small crypto port after file-format framing is decided.

## Importers

`internal/importer/gopass` is a gopass CLI client for imports. It shells out to:

- `gopass list --flat`;
- `gopass show --password <path>`.

It is not a runtime backend and does not implement `source.Reader`.

`internal/app.ImportGopassForRuntime` maps gopass entries into the local Shelf vault:

- gopass `a/b/c` maps to Shelf `a/b:c`;
- entries without `/` are skipped as unmappable;
- imported passwords are stored as JSON strings;
- existing local secrets are skipped unless `--force` is passed;
- read failures abort before writing to avoid partial imports.

The CLI entrypoint is `shelf vault import gopass`.

## Source boundary

`internal/source` defines the read-side contract used by project env resolution. It exists so `internal/project` does not depend on the local vault store type. Today the runtime implementation is always the local vault reader from `internal/vault`.

Do not add runtime external backends to `source.Reader` without a new product decision. External managers currently import into the local vault first.

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

Command handlers should stay thin: parse flags, call feature/base packages, then render output through Cobra writers.

## Runtime construction

`internal/app` centralizes runtime and local vault loading:

- `LoadVault(configPath, vaultPath)` resolves config and constructs `*vaultfile.Vault`;
- `LoadRuntime(configPath, vaultPath)` loads a decrypted local vault `*vault.Store` snapshot;
- `LoadSecretReader(configPath, vaultPath)` loads the local vault and returns `vault.Reader`;
- `ReadVault(configPath, vaultPath, fn)` runs read-only local vault work;
- `UpdateVault(configPath, vaultPath, fn)` locks, loads, mutates, and encrypted-saves through `vaultfile.Vault.Update`;
- `String()` returns the application version string from release ldflags or Go build info.

## Project workflows

`internal/project` owns Git project identity, `.shelf.json` schema/IO/validation, and manifest-to-env resolution.

Resolution order for env names:

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
