# Phase 16 Plan: First Release Readiness

## Goal

Prepare Shelf Go for the first public release without expanding product scope.

## Scope

- Add minimal GoReleaser configuration for reproducible multi-platform CLI artifacts.
- Add tag-triggered GitHub Actions release workflow.
- Add `go vet` to CI.
- Update README from command inventory to usage-oriented onboarding.
- Update CHANGELOG for `0.1.0`.
- Record deferred manager UI improvement as post-0.1 work.
- Run release verification locally, including GoReleaser config/snapshot when available.

## Non-goals

- No Homebrew tap, Scoop, deb/rpm/apk, Docker image, signing, SBOM, notarization, or package-manager distribution for 0.1.0.
- No manager UI redesign in this phase.
- No feature changes beyond release automation and documentation.
- No storage backend or architecture refactor.

## Acceptance

- `README.md` explains why Shelf exists and walks through initialization, secret use, and project use.
- GoReleaser can validate the config and create a local snapshot release.
- CI runs tests, vet, and build.
- CHANGELOG has a `0.1.0` section.
- `.planning` records verification evidence and manager UI deferral.
