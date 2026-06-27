# v0.1.0 Summary

## Released
2026-06-26

## What Shipped
Shelf Go v0.1.0 shipped the first public encrypted-vault CLI: age-encrypted portable storage, scoped `shelf setup`/`shelf vault`/`shelf secret`/`shelf project` command groups, value-free project manifests, project export/run workflows, localhost vault manager, release automation, and public documentation.

## Included Phases
- 001-encrypted-vault-core
- 002-migration-and-git-safety
- 003-project-workflow-compatibility
- 004-localhost-vault-manager
- 005-documentation-and-release-hardening
- 006-command-hierarchy-cutover
- 007-vault-ux-hardening
- 008-project-session-design
- 009-project-export-shell-default
- 010-vault-restore-recovery
- 011-edit-manager-safety
- 012-restore-simplification
- 013-architecture-package-boundaries
- 014-vault-secret-extraction
- 015-persistence-store-layout
- 016-first-release-readiness

## Completed Scope

### Roadmap Items
- Phases 001-003 established age-encrypted vault persistence, plaintext migration, git-safety diagnostics, and encrypted-vault compatibility for secret, export, project, and run workflows.
- Phase 004 added the on-demand localhost vault manager with metadata search, explicit reveal, and encrypted create/update/delete behavior.
- Phase 005 replaced stale docs with public documentation for encrypted vault usage, safety boundaries, and release readiness.
- Phases 006-008 completed command hierarchy cutover, vault diagnostics, and future project session design while keeping hook-based shell mutation out of scope.
- Phases 009-012 improved project export defaults, simplified encrypted `.bak` recovery, and hardened secret edit and manager safety behavior.
- Phases 013-015 extracted reusable runtime, project, vault, and secret workflows from `internal/cli`, centralized atomic writes, and clarified store file layout without adding speculative backend interfaces.
- Phase 016 added GoReleaser/GitHub release automation, Windows-compatible vault locking and release artifacts, CI vet, usage-focused README, changelog, and release verification.

### Requirements
- BASE-VAULT-01..09 — encrypted age vault configuration, load/save, migration, recovery artifacts, and plaintext avoidance completed.
- BASE-CLI-01..05 — secret, export, project manifest, project run, and encrypted-vault regression behavior completed.
- BASE-SAFE-01..07 — portable encrypted vault, non-secret config, doctor git safety, and localhost manager safety completed.
- CMD-01..08 — canonical scoped command hierarchy completed.
- VUX-01..04 — vault status/check/doctor diagnostics and docs completed.
- SES-01..04 — project activation/deactivation/shell semantics captured and intentionally left unimplemented.
- PUX-01..03 — `shelf project export` defaults to sourceable shell output; dotenv format remains out of scope.
- VREC-01..03 — manual last-write encrypted `.bak` recovery documented.
- SAFE-EDIT-01, SAFE-MGR-01, SAFE-DOC-01 — secret edit temp cleanup, manager token/host boundaries, and safety docs completed.
- ARCH-01..08 — app/project/vault/secret extraction, atomic writes, canonical validators, and store layout completed.
- REL-01..04 — GoReleaser artifacts, GitHub Actions release workflow, usage README, changelog, and release verification completed.

## Notable Decisions
- Use age encryption for the portable vault because it matches the target chezmoi workflow and keeps the source of truth as a git-safe encrypted file.
- Keep config and `.shelf.json` manifests value-free; plaintext may be intentionally exported but is not the source of truth.
- Use scoped command groups (`setup`, `vault`, `secret`, `project`) instead of retaining unpublished ambiguous top-level commands.
- Prefer explicit export/source and child-process workflows over hook-based parent-shell mutation for the first release.
- Keep recovery minimal: one encrypted last-write `.bak`, ordinary file copy, and `shelf vault status` verification instead of a dedicated restore command.
- Keep storage as age-encrypted JSON for v0.1.0; SQLite remains a future spike only if schema/search/history pressure becomes real.
- Keep `internal/cli` command-family oriented and move reusable behavior into feature/base packages without speculative backend interfaces.
- Publish v0.1.0 with GoReleaser archives and checksums for Linux, macOS, and Windows on amd64/x86_64 and arm64.

## Follow-ups
- Select the v0.1.1 milestone before adding new active requirements.
- Consider post-0.1 manager UI redesign for information architecture, reveal/edit UX, and visual polish.
- Consider native Windows smoke tests for `shelf setup`, secret set/get, and project run on a real Windows runner.
- Consider optional package-manager distribution after initial usage validates demand.
