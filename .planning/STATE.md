# State

## Current Position
- Phase: Item Vault Storage and Crypto Boundary
- Status: planned
- Active Artifact: .planning/phases/042-item-vault-storage-crypto-boundary/PLAN.md
- Next Action: Execute Phase 42: introduce an item-vault provider boundary, add age-wrapped encrypted item-vault storage, update setup key choice/generation, and remove temporary `v2` default storage naming before release.

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
- Completed Phase 37: an internal `vaultmigrate` adapter converted v1 flat secrets to v2 `generic_secret` items and project bindings, but the Phase 41 audit removed that package before release because automatic migration requires a reviewed user-facing backup, dry-run, review, and rollback workflow.
- Verification passed for Phase 37 at the time: `go test ./internal/vaultmigrate`; `go test ./internal/vaultmodel ./internal/vaultjson ./internal/vaultmigrate`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 38: v2 secret CLI workflows create, set from stdin, get, list long, info, and remove over v2 items; env names are public fields; profiles include `openai`; value-free outputs and ambiguity errors are tested.
- Verification passed for Phase 38: `go test ./internal/vaultmodel ./internal/vaultjson ./internal/app ./internal/cli`; `go test ./internal/app -run 'Test.*V2'`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 39: `.shelf.json` version 2 supports concrete item/field/env bindings; project status/add/list/rm/export/run dispatch v1 vs v2 manifests; v2 diagnostics stay value-free.
- Verification passed for Phase 39: `go test ./internal/project ./internal/app ./internal/cli`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 40: `shelf manager` is explicitly gated when a resolved valid v2 vault exists, preventing the v1 Web editor from operating over v2 item workflows while preserving v1 manager behavior.
- Verification passed for Phase 40: `go test ./internal/app ./internal/cli ./internal/manager`; `go test ./...`; `go vet ./...`; `go-arch-lint check`.
- Completed Phase 41 audit: removed `internal/vaultmigrate`, deferred public migration automation, fixed stale security/architecture docs, and verified `go test ./...`, `go vet ./...`, and `go-arch-lint check`.
- Planned Phase 42: release-quality item-vault storage should use age encryption by default, keep encryption decoupled from `vaultmodel` and JSON codec, offer setup choice between existing and generated age identities, and defer Web/TUI/project expansion until CLI/model/file storage are stable.

## Updated
- 2026-07-09
