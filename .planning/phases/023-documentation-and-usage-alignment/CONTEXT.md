# Context: Phase 23 Documentation and Usage Alignment

## Goal

Update user and developer documentation after the v0.1.1 manager, tag workflow, script workflow, and package architecture changes.

## Inputs

- Phase 18 implemented `shelf manager` as the local Web editing console with add/edit/rename/delete/reveal/copy/tag flows.
- Phase 19 implemented direct tag selection for `shelf secret list` and `shelf secret export`.
- Phase 20 implemented project tag bindings via `.shelf.json` `tags` entries and `shelf project add --tag`.
- Phase 21 consolidated install/release workflows under `scripts/install.sh` and `scripts/release.sh`.
- Phase 22 repartitioned packages and made `shelf manager` the only manager entrypoint.

## Requirements

- DOC-01: User-facing docs describe manager editing, direct tag list/export, and project tag bindings.
- DOC-02: Developer docs describe install/tag/release scripts and the final internal architecture.
- ARCH-01 and ARCH-02 remain reflected in docs.
- BOUND-01: Do not document fine-grained CLI metadata edit subcommands.
- BOUND-02: Keep storage documented as age-encrypted JSON; SQLite remains deferred.

## Scope

Update existing docs only. Do not add new product behavior.

## Candidate Files

- `README.md`
- `docs/getting-started.md`
- `docs/reference.md`
- `docs/troubleshooting.md`
- `docs/contributing.md`
- `docs/architecture.md`
- `CHANGELOG.md` only if required by release hardening; otherwise Phase 24 owns changelog.

## Verification Inputs

- Command flags and behavior should be checked against `internal/cli/*.go` and `internal/project/*.go`.
- Manager behavior should be checked against `internal/manager/server.go` and `internal/manager/ui.go`.
- Script behavior should be checked against `scripts/install.sh`, `scripts/release.sh`, and `justfile`.
