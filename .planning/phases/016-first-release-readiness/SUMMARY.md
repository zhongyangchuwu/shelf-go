# Phase 16 Summary: First Release Readiness

## Result

Complete pending commit and tag. Shelf now has minimal repeatable release automation, a usage-oriented README, first-release changelog entries, and release verification evidence.

## Changes

- Added `.goreleaser.yaml` for Linux/macOS tarballs, Windows zip archives, and checksums.
- Added `.github/workflows/release.yml` for tag-triggered GoReleaser publishing.
- Added `go vet ./...` to CI before tests/build.
- Added `just vet`, `just release-check`, and `just release-snapshot` recipes.
- Added release-version ldflag hook in `internal/version` so GoReleaser artifacts report the release version.
- Replaced Unix-only `syscall.Flock` with `github.com/gofrs/flock` for portable vault write locks.
- Rewrote `README.md` around motivation and primary flows: install, initialization, secret use, project use, portability/recovery, safety.
- Updated `CHANGELOG.md` with `0.1.0 - 2026-06-25`.
- Recorded manager UI redesign as deferred post-0.1 work.

## Platform Scope

0.1.0 release artifacts target Linux, macOS, and Windows on amd64 and arm64.

## Verification

See `VERIFICATION.md`.

Passed:

```bash
go vet ./...
go test ./...
go test -race ./...
go test ./internal/store
GOOS=windows GOARCH=amd64 go test ./internal/store ./internal/manager ./internal/manifest ./internal/config
GOOS=windows GOARCH=amd64 go build ./cmd/shelf
go build -o ./bin/shelf ./cmd/shelf
go run github.com/goreleaser/goreleaser/v2@latest check
go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot
./dist/shelf_linux_amd64_v1/shelf --version
./dist/shelf_linux_amd64_v1/shelf --help
```

## Next

Commit these changes, then tag `v0.1.0` from a clean tree and confirm GitHub Release artifacts.
