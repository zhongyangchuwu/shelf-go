# Context: Backend Pluggability Architecture

## Goal

Use a real second backend, gopass, to test whether Shelf's current architecture cleanly supports non-Shelf secret sources, then plan GPG support without forcing a broad speculative backend framework.

## Current Code Facts
- `internal/source.Reader` is backend-neutral and sufficient for project env resolution: exact get, prefix list, and tag list.
- `internal/app.LoadSecretReader` now selects `shelfvault` or `gopass` from config.
- Secret CRUD, manager editing, migration, status, and setup still use concrete `*shelfvault.Vault` / `*shelfvault.Store` types.
- `internal/config.Config` now has `source.type` and `source.gopass_command`; legacy `vault_path`, `recipients`, and `identity_paths` still configure the default Shelf vault.
- `internal/crypto/age` isolates direct age usage, but `shelfvault.VaultOptions` is still age-shaped.

## External Findings

- gopass supports multiple mounted stores in one namespace; each store can use different crypto and storage backends.
- gopass has pluggable crypto backends including GPG and age, and storage backends including git/fs variants.
- gopass exposes a Go API with `List`, `Get`, `Set`, and `Remove`, but the stable user contract is still the installed CLI/store behavior.
- GnuPG automation should use machine-readable / non-interactive flags such as `--batch`, `--status-fd`, and `--with-colons` when shelling out.

## Decisions

- Treat gopass support first as a `source.Reader` implementation for project env workflows.
- Do not generalize Shelf write/manager APIs until gopass read support reveals the real shape of metadata, path, and error semantics.
- Keep gopass native GPG support separate from a Shelf-local GPG crypto backend. They solve different problems.
- Plan Shelf-local GPG encryption only after a crypto port and file-format decision are explicit.
- The first gopass implementation uses CLI commands: `gopass list --flat` and `gopass show --password <path>`.
- Shelf path `group/key:field` maps to gopass path `group/key/field`; only project workflows use this mapping.
- gopass tag selectors are intentionally unsupported in the first slice and return diagnostics through project resolution.

## Risks

- gopass path semantics do not use Shelf's `group:key` grammar by default.
- gopass secrets may not carry Shelf metadata (`env`, `description`, `tags`) without a mapping convention.
- Using the gopass Go API may pull a large dependency surface; using the CLI creates process-boundary and parsing concerns.
- A GPG backend for Shelf vault could require new file framing and migration behavior; it must not silently reinterpret age vaults.

## Canonical References

- gopass repo/docs: https://github.com/gopasspw/gopass
- gopass backend docs: https://github.com/gopasspw/gopass/blob/master/docs/backends.md
- gopass Go API example: https://github.com/gopasspw/gopass/blob/master/docs/hacking.md
- GnuPG command docs: https://gnupg.org/documentation/manuals/gnupg24/gpg.1.html
- GnuPG configuration / batch mode docs: https://gnupg.org/documentation/manuals/gnupg/GPG-Configuration-Options.html
