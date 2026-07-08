# State

## Current Position
- Phase: Project Manifest v2 Bindings
- Status: ready
- Active Artifact: .planning/phases/039-project-manifest-v2-bindings/PLAN.md
- Next Action: Plan and implement v2 project manifest item/field/env bindings so project status/export/run can resolve v2 vault items without storing secret values.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- v0.1.1 planning and implementation phases through Phase 33 completed architecture cleanup, app/CLI boundaries, jsonvault abstraction, manager boundary cleanup, and full verification.
- Recent product discussion selected a v0.2 direction: template-based items and fields, API credential as a built-in template, env names as public fields, repository/codec adapters, and JSON as the initial storage recommendation over SQLite unless later evidence changes the storage decision.
- Completed Phase 34: template item vault design, storage decision, migration plan, and validation artifacts.
- Completed Phase 35: `internal/vaultmodel` model types, built-in templates, validation, runtime indexes, and architecture boundary registration.
- Verification passed for Phase 35: `go test ./internal/vaultmodel`; `go test ./internal/vault ./internal/project ./internal/app ./internal/cli ./internal/vaultmodel`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 36: `internal/vaultjson` strict JSON codec, plaintext repository, atomic save, backup behavior, missing-file errors, focused tests, and architecture boundary registration.
- Verification passed for Phase 36: `go test ./internal/vaultjson`; `go test ./internal/vaultmodel ./internal/vaultjson`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 37: `internal/vaultmigrate` converts v1 flat secrets to v2 `generic_secret` items, produces deterministic path mappings, converts v1 project manifests to value-free item/field/env bindings, and validates converted output.
- Verification passed for Phase 37: `go test ./internal/vaultmigrate`; `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 38: v2 secret CLI workflows create, set from stdin, get, list long, info, and remove over v2 items; env names are public fields; profiles include `openai`; value-free outputs and ambiguity errors are tested.
- Verification passed for Phase 38: `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate ./internal/app ./internal/cli`; `go test ./internal/app -run 'Test.*V2'`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.

## Updated
- 2026-07-08
