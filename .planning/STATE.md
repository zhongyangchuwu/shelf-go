# State

## Current Position
- Phase: v0.1.1 Release Ready
- Status: ready-to-tag
- Active Artifact: .planning/phases/027-cli-test-rebalancing-and-boundary-verification/VERIFICATION.md
- Next Action: Review the release-ready commits, then tag and push v0.1.1 with `./scripts/release.sh tag 0.1.1` and `git push origin v0.1.1`.

## Blockers
- None

## Recent Evidence
- v0.1.0 is published as a non-draft, non-prerelease GitHub Release at tag `v0.1.0`.
- Completed v0.1.0 phase history is archived under `.planning/archive/releases/v0.1.0/`.
- Phase 17 produced the Web manager design contract in `.planning/phases/017-web-manager-design/UI-SPEC.md`.
- Phase 18 implemented the manager editing console and manager API hardening.
- Phase 19 implemented direct secret tag selection for `secret list` and `secret export`.
- Phase 20 implemented value-free project tag bindings for `.shelf.json`, `project add/list/rm`, and project explain/export/run resolution.
- Phase 21 completed script workflow consolidation: install now runs through `scripts/install.sh`, release check/snapshot/tag now run through one `scripts/release.sh` command surface, and `justfile` remains a thin task runner.
- Phase 22 completed architecture repartition:
  - `shelf manager` replaced `shelf vault open` with no alias.
  - Vault core moved to `internal/vault`.
  - Project manifest schema/IO/validation moved into `internal/project`.
  - Version composition moved into `internal/app`.
  - Export formatting moved into `internal/exportfmt`.
  - Final internal package set is `app`, `cli`, `config`, `exportfmt`, `manager`, `project`, `secret`, and `vault`.
- Phase 23 completed documentation alignment for manager editing, direct tag workflows, project tag bindings, scripted workflows, and final package layout.
- Phase 24 completed release hardening: changelog updated, `go test ./...`, `go vet ./...`, `./scripts/release.sh check`, and `./scripts/release.sh snapshot` passed.
- v0.1.1 still defers SQLite/storage redesign to v0.2.0 and keeps the current age-encrypted JSON vault format.
- Phase 25 planning added optional pre-tag CLI boundary refactor scope: project/session business rules move into `internal/project`, later app orchestration extraction moves into `internal/app`, and CLI tests narrow to command contracts.
- Phase 25 completed project/session boundary refactor: project entry construction, child env merge, and env override warnings now live in `internal/project`; `go test ./internal/project`, `go test ./internal/cli -run 'Test(Project|Run)'`, and `go test ./...` passed.
- Phase 26 completed app service extraction: direct export, plaintext migration, setup file helpers, and manager loopback/token helpers now live in `internal/app`; `go test ./internal/app`, focused CLI tests, and `go test ./...` passed.
- Phase 27 completed CLI test rebalancing and boundary verification: CLI tests were narrowed to command contracts and smoke paths, domain/app tests own behavior rules, boundary greps passed, and `go test ./internal/project`, `go test ./internal/app`, `go test ./internal/cli`, and `go test ./...` passed.

## Updated
- 2026-06-28
