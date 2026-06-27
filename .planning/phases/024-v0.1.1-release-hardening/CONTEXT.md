# Context: Phase 24 v0.1.1 Release Hardening

## Goal

Prepare v0.1.1 for release after manager, tag workflows, scripts, architecture cleanup, and docs alignment are complete.

## Inputs

- Phase 18: manager editing console and manager API hardening complete.
- Phase 19: direct secret tag selection complete.
- Phase 20: project tag bindings complete.
- Phase 21: install/release workflows consolidated under `scripts/`.
- Phase 22: architecture package repartition complete.
- Phase 23: user/developer docs aligned.

## Requirements

- REL-011-01: release readiness is checked only after architecture and documentation cleanup are complete.
- BOUND-01: no fine-grained CLI metadata edit command groups such as `secret meta` or `secret tag`.
- BOUND-02: current age-encrypted JSON vault format remains; no SQLite implementation or spike.

## Release Hardening Checks

- `CHANGELOG.md` has a `0.1.1` section.
- `go test ./...` passes.
- `go vet ./...` passes.
- `./scripts/release.sh check` passes.
- `./scripts/release.sh snapshot` passes.
- Release evidence records no storage format change and SQLite deferral.

## Scope

This phase prepares release readiness and records evidence. Publishing/tagging the release is a separate action after the committed hardening work is reviewed.
