# Capture: Phase 24 v0.1.1 Release Hardening

## Release Readiness Facts

- `CHANGELOG.md` now has a `0.1.1` section dated 2026-06-27.
- Final local verification passed:
  - `go test ./...`
  - `go vet ./...`
  - `./scripts/release.sh check`
  - `./scripts/release.sh snapshot`
- Snapshot artifacts include Linux, macOS, and Windows archives for amd64/x86_64 and arm64 targets.
- The snapshot version is `v0.1.0-next` until a real `v0.1.1` tag is created.

## Boundaries Preserved

- No SQLite/storage backend change in v0.1.1.
- No fine-grained CLI metadata command group.
- No compatibility alias for the old manager command.
- No hosted frontend, CDN, SPA requirement, or permanent daemon introduced.

## Next Action

After review, publish v0.1.1 by creating/pushing the release tag through the consolidated script flow:

```bash
./scripts/release.sh tag 0.1.1
git push origin v0.1.1
```

GitHub Actions release workflow should then run GoReleaser with `args: release --clean`.
