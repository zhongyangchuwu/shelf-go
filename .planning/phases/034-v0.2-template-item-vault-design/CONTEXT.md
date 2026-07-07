# Context: v0.2.0 Template Item Vault Design

## Goal

Design Shelf Go v0.2.0 around a template-based item model that can represent generic secrets, API credentials, and future developer-secret shapes without treating `group:key` paths as the product model.

This phase produces planning/design artifacts only. It does not implement the new vault model, migrate data, or change user-facing commands.

## Constraints

- Shelf remains local-first and single-developer focused.
- Core workflows must stay scriptable; interactive/TUI work is deferred until the data model is stable.
- Secret values must not appear in config, project manifests, list/detail outputs, logs, or planning artifacts.
- Project manifests remain value-free.
- Encryption can be ignored during the first design pass, but the design must keep an adapter boundary so age-encrypted portable files can be restored later.
- Storage redesign is allowed for v0.2.0, but the design must compare JSON and SQLite against portability, temp-file leakage, migration, and chezmoi/git-safe sync.
- Existing v0.1.x behavior is pre-release and may break, but migration/export safety must be planned before implementation deletes or rewrites old vault data.
- Do not expand into team sharing, hosted sync, browser extension/autofill, permanent daemon, or broad password-manager replacement.

## Decisions

- v0.2.0 should move from flat path-keyed `Secret` records toward a core model of `Item`, `Field`, `Template`, and adapters.
- API credential is a built-in template, not the root domain model.
- Generic secret remains a built-in template so simple env-secret workflows stay lightweight.
- Field definitions carry type, sensitivity, purpose, and bindings. This allows public fields such as provider, website, expiry, and scopes to appear in metadata views while concealed fields remain hidden.
- Environment export should bind to fields, not whole items. A field may have one or more env bindings.
- Storage should be selected behind a repository/codec boundary. Plain JSON is the default design candidate because it preserves portability and age wrapping; SQLite remains a comparison point, not the default.
- The v0.2 design should keep encryption out of the core model and attach encryption as a repository adapter later.
- Current `secret add` interactive prompt should not be evolved during this design; common CLI commands such as create/set/get/copy/list/info should be redesigned first.

## Open Questions

- Should canonical item identity be generated stable IDs only, or should human aliases/paths remain first-class selectors?
- Should v0.2 retain any compatibility path syntax for old `prefix:name` secrets, or only support it through migration aliases?
- Should project manifests bind canonical item IDs plus field IDs, or support value-free selectors such as title/env labels for human readability?
- Should built-in templates be compiled into Go code only, or should custom templates be loaded from value-free config later?
- Should v0.2 ship with only `generic_secret` and `api_credential`, or also include a database/url template if the field model makes it cheap?
- Should the Web manager be paused/removed during the v0.2 model cutover, or adapted after the CLI/core model stabilizes?

## Verification Expectations

- Design artifacts define a data model that can represent current generic secrets and API credentials with expiration/source metadata.
- Design artifacts explicitly compare JSON and SQLite storage options before implementation begins.
- Design artifacts specify value exposure rules for public vs concealed fields.
- Design artifacts specify migration safety from v1 flat secrets to v2 item records.
- Design artifacts specify project manifest implications and value-free guarantees.
- A later implementation plan can be executed without re-litigating the core model boundaries.
