# State

## Current Position
- Phase: Secret CLI v2 Workflows
- Status: executing
- Active Artifact: .planning/phases/038-secret-cli-v2-workflows/PLAN.md
- Next Action: Refactor v2 env bindings into env-name fields, add profile registry support, then implement the first-version v2 secret CLI from `CLI-SPEC.md`.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- v0.1.1 planning and implementation phases through Phase 33 completed architecture cleanup, app/CLI boundaries, jsonvault abstraction, manager boundary cleanup, and full verification.
- Recent product discussion selected a v0.2 direction: template-based items and fields, API credential as a built-in template, env bindings on fields, repository/codec adapters, and JSON as the initial storage recommendation over SQLite unless later evidence changes the storage decision.
- Created `.planning/phases/034-v0.2-template-item-vault-design/CONTEXT.md` for v0.2 design context.
- Created `.planning/phases/034-v0.2-template-item-vault-design/PLAN.md` for v0.2 model/storage/migration planning.
- Completed `.planning/phases/034-v0.2-template-item-vault-design/SPEC.md`, `STORAGE-DECISION.md`, `MIGRATION.md`, `SUMMARY.md`, `REVIEW.md`, `VERIFICATION.md`, and `CAPTURE.md`.
- Phase 34 is complete; Phase 35 is the next implementation phase and must implement the v2 model/validation core before storage or CLI cutover.
- Created `.planning/phases/035-vault-model-validation-core/CONTEXT.md` and `PLAN.md` for the v2 model/validation implementation phase.
- Completed Phase 35: `internal/vaultmodel` model types, built-in templates, validation, runtime indexes, focused tests, and architecture boundary registration.
- Verification passed for Phase 35: `go test ./internal/vaultmodel`; `go test ./internal/vault ./internal/project ./internal/app ./internal/cli ./internal/vaultmodel`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Created `.planning/phases/036-json-codec-repository-boundary/CONTEXT.md` and `PLAN.md` for the next implementation phase.
- Completed Phase 36: `internal/vaultjson` strict JSON codec, plaintext repository, atomic save, backup behavior, missing-file errors, focused tests, and architecture boundary registration.
- Verification passed for Phase 36: `go test ./internal/vaultjson`; `go test ./internal/vaultmodel ./internal/vaultjson`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 37: `internal/vaultmigrate` converts v1 flat secrets to v2 `generic_secret` items, produces deterministic path mappings, converts v1 project manifests to value-free item/field/env bindings, and validates converted output.
- Verification passed for Phase 37: `go test ./internal/vaultmigrate`; `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Created `.planning/phases/038-secret-cli-v2-workflows/CONTEXT.md` and `PLAN.md` for the next implementation phase.

## Updated
- 2026-07-07
