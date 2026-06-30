# State

## Current Position
- Phase: Vault Import Pivot
- Status: implementation-in-progress
- Active Artifact: .planning/phases/030-vault-import-pivot/PLAN.md
- Next Action: Refactor vault/domain/import/crypto packages, remove gopass runtime source selection, and add gopass-to-local-vault import.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the manager editing console and manager API hardening.
- Phase 19 implemented direct secret tag selection for `secret list` and `secret export`.
- Phase 20 implemented value-free project tag bindings for `.shelf.json`, `project add/list/rm`, and project explain/export/run resolution.
- Phase 21 completed script workflow consolidation.
- Phase 22 completed architecture repartition.
- Phase 23 completed documentation alignment.
- Phase 24 completed release hardening.
- Phase 25 completed project/session boundary refactor.
- Phase 26 completed app service extraction.
- Phase 27 completed CLI test rebalancing and boundary verification.
- Phase 28 completed architecture boundary lint and CLI adapter slimming.
- Phase 29 planning created `.planning/phases/029-backend-pluggability-architecture/CONTEXT.md` and `PLAN.md` for adding gopass as a read source and evaluating GPG as a Shelf vault crypto backend.
- Gopass read-source MVP implementation added config `source.type`, `source.gopass_command`, `internal/adapters/gopass.Reader`, and runtime source selection for project workflows.
- Phase 30 pivot decided that gopass is an import source, not a runtime backend; target package layout is `internal/vault`, `internal/vaultfile`, `internal/vaultcrypto`, and `internal/importer/gopass`.

## Updated
- 2026-06-30
