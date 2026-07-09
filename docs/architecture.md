# Architecture

Shelf Go is a local-first Go CLI. The runtime secret store is a local encrypted Shelf vault. External systems such as gopass are import sources, not runtime backends.

## Package layers

```text
cmd/shelf/                    process entry point

internal/cli/                 Cobra command tree, flags, argument validation, completions, and text rendering
internal/manager/             local manager surface, currently v1 loopback HTTP/Web

internal/app/                 application service, runtime construction, workflow orchestration, manager gating, import services, and version composition
internal/project/             project identity, .shelf.json schema/IO/validation, and v1/v2 binding resolution
internal/secret/              reusable legacy secret workflows such as editor-based updates

internal/vaultmodel/          v2 item/field/template model, validation, profiles, and runtime indexes
internal/vaultjson/           v2 plaintext JSON codec/repository boundary; temporary cutover storage, not safe to sync
internal/vault/               v1 Shelf vault domain model, repository abstraction, path/env/tag rules, and in-memory store
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
  internal/cli -> app

Workflow:
  app -> config, project, secret, vault, jsonvault, importer/gopass, manager, util, vaultmodel, vaultjson
  project -> vault, util, vaultmodel
  secret -> vault

Domain/persistence/support:
  vaultjson -> vaultmodel, util
  jsonvault -> vault, age, util, flock
  age -> filippo.io/age
  importer/gopass -> standard library
  config -> YAML
  util -> standard library
```

Base packages must not import `internal/cli`. Feature packages should expose concrete functions and data types, not speculative backend interfaces. Automatic v1-to-v2 migration code is intentionally absent from the production package map until migration UX, review, and rollback contracts are designed.

## Product data model

Shelf's source of truth is the local Shelf vault. gopass, 1Password, Bitwarden, or other password managers should first enter as importers that copy data into the Shelf vault model.

Runtime commands read local vault data only:

- v2 `secret` item/field workflows use the temporary v2 JSON sidecar;
- v2 `project status/export/run` resolve concrete item/field bindings against v2 items;
- v1 manager list/reveal/edit/delete is available only while no matching valid v2 vault is present;
- v1 project/export/run behavior remains available for v1 manifests during cutover.

This avoids split-brain behavior where one command reads gopass while another command reads the local vault.

## Vault domain

`internal/vault` owns Shelf's vault data model and rules:

- `Data`, `Secret`, `Info`, `CurrentVersion`, `NewData`;
- `SecretID`, `ParseSecretID`, `ValidatePath`, `ValidateSecretID`;
- `ValidateSecret`, `ParseValue`, env/tag validation and env-name derivation;
- in-memory `Store` mutation/list/query methods;
- application-facing vault repository contracts: `Options`, `Repository`, `Provider`, and status `Report`.

The domain package defines the application-facing contract for an encrypted Shelf vault, including recipients and identity paths needed to open one. It does not know the file format, encryption algorithm, lock strategy, or project manifests. Future SQL/NoSQL persistence should implement the `vault.Repository` contract rather than replacing the model prematurely.

## V2 vault model and JSON repository

`internal/vaultmodel` owns the v2 item model:

- `Vault`, `Item`, `Field`, templates, profiles, sensitivity, and field purposes;
- validation for item IDs, field IDs, required fields, public env-name fields, and profile defaults;
- runtime indexes for item selectors and env-name lookups.

`internal/vaultjson` owns the current v2 JSON codec/repository boundary:

- strict JSON decode with unknown-field and trailing-content rejection;
- validated encode/save through `vaultmodel.Validate`;
- atomic save and backup behavior through `internal/util`.

This repository is plaintext during cutover and is not safe to commit or sync. It is a temporary boundary so age wrapping or a different storage backend can replace file persistence without changing item semantics. There is intentionally no production v1-to-v2 migration package in the current release scope.

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
- `App.OpenManager(configPath, vaultPath, addr)` gates v2 vaults, then opens the v1 local manager using a manager-specific service adapter;
- `String()` returns the application version string from release ldflags or Go build info.

V2 repository use is currently concrete: `internal/app` constructs `vaultjson.Repository` directly for v2 secret, project, and manager-gate workflows. A future storage phase should introduce a v2 provider/repository interface before adding encrypted v2 storage or alternate backends.

## Project workflows

`internal/project` owns Git project identity, `.shelf.json` schema/IO/validation, and manifest-to-env resolution.

Version 1 resolution uses a local `*vault.Store` snapshot supplied by `internal/app`. Version 2 resolution uses `vaultmodel.Vault` loaded through the v2 JSON repository boundary. Env name rules:

1. v1 exact path entries may carry manifest `env` overrides;
2. v2 project bindings carry explicit env names derived from item env-name public fields when added through CLI;
3. v1 prefix/tag entries and migrated bindings use concrete resolved env names.

V1 prefix and tag manifest entries may expand to multiple local vault secrets and cannot carry `env`. V2 bindings are concrete item/field/env records. Required missing entries and duplicate env names are diagnostics; commands decide whether diagnostics are fatal.

## Secret editing workflow

The first v2 secret CLI is non-interactive and writes fields from stdin. `internal/secret` remains a legacy v1 editing-workflow helper used by manager/app paths; new v2 editing logic should not be added there without renaming or splitting the package. Full-object v2 editing is deferred to a later TUI/wizard so template fields, env-name fields, and secret values can be edited without reintroducing raw JSON editor UX.

## Utilities

`internal/util` holds small shared helpers that are not domain concepts yet: atomic file replacement and env/shell/JSON binding formatting. If either area grows into a cohesive subsystem again, split it back out with that concrete pressure.

## Local manager

`internal/manager` is an on-demand local manager surface. Today it is implemented as a v1 loopback HTTP/Web editor, but the package name is intentionally not vault-only or Web-only so future config/project panels can live behind the same manager concept. CLI reaches it through `internal/app`; manager receives a narrow secret-service contract and does not import app.

During the v2 cutover, `App.OpenManager` checks the resolved v2 vault path before opening the v1 manager. If a v2 vault is present, Shelf returns a value-free gate error instead of opening a v1 path-shaped editor against stale or mismatched data.

Safety boundaries:

- loopback-only address validation before server start;
- tokenized URL with token removal from the visible URL after first load;
- strict cookie, Host checks, and Origin checks;
- no-store responses for manager pages and API responses;
- metadata list/search/detail without values;
- explicit reveal endpoint for plaintext secret values;
- no external network dependency.
