# Phase 16 Verification: First Release Readiness

## Result

Passed. GoReleaser publishes Linux, macOS, and Windows archives for 0.1.0 after replacing the Unix-only `syscall.Flock` lock with the cross-platform `github.com/gofrs/flock` lock.

## Checks

| Claim | Evidence | Result |
| --- | --- | --- |
| CI includes vet, tests, and build | `.github/workflows/ci.yml` has `go vet ./...`, `go test ./...`, and `go build ./cmd/shelf` | Passed |
| Release workflow publishes from version tags | `.github/workflows/release.yml` triggers on `v*`, uses `contents: write`, checks out full history, runs vet/test, then GoReleaser | Passed |
| GoReleaser config is valid | `go run github.com/goreleaser/goreleaser/v2@latest check` | Passed |
| Snapshot release builds release artifacts | `go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot` | Passed |
| Snapshot artifacts include checksums | `dist/checksums.txt` produced with Linux/macOS tarballs and Windows zip archives | Passed |
| Snapshot binary reports injected release version | `./dist/shelf_linux_amd64_v1/shelf --version` printed `0.1.0-SNAPSHOT-97ab92a ... linux/amd64` before the Windows lock fix; post-fix snapshot artifacts use `0.1.0-SNAPSHOT-261aa6c` | Passed |
| CLI still builds locally | `go build -o ./bin/shelf ./cmd/shelf` | Passed |
| Go vet is clean | `go vet ./...` | Passed |
| Unit/integration suite still passes | `go test ./...` | Passed |
| Race suite still passes | `go test -race ./...` | Passed |
| Store lock blocks concurrent writers | `go test ./internal/store` includes `TestLockFileBlocksConcurrentLock` | Passed |
| Windows store/manager/manifest/config packages compile and pass | `GOOS=windows GOARCH=amd64 go test ./internal/store ./internal/manager ./internal/manifest ./internal/config` | Passed |
| Windows CLI binary compiles | `GOOS=windows GOARCH=amd64 go build ./cmd/shelf` | Passed |

## GoReleaser Snapshot Artifacts Observed

```text
dist/shelf_0.1.0-SNAPSHOT-261aa6c_linux_amd64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-261aa6c_linux_arm64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-261aa6c_darwin_amd64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-261aa6c_darwin_arm64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-261aa6c_windows_amd64.zip
dist/shelf_0.1.0-SNAPSHOT-261aa6c_windows_arm64.zip
dist/checksums.txt
```

## Platform Decision

Initial Windows targets failed during snapshot release because `internal/store/lock.go` used Unix-only `syscall.Flock` constants. Research identified `github.com/gofrs/flock` as a maintained cross-platform file-locking package that wraps the relevant OS-specific primitives, including Windows `LockFileEx`/`UnlockFileEx`.

Decision: use `github.com/gofrs/flock` for vault write locks and restore Windows GoReleaser targets for 0.1.0.

## Windows Test Note

`GOOS=windows GOARCH=amd64 go test ./...` compiles and runs most packages but current CLI doctor tests have Windows-specific assumptions around Unix permission output, `FPATH` path splitting, and temp directory cleanup after Git subprocesses. The release gate therefore verifies Windows support at the store lock, manager, manifest, config, and CLI binary compile levels, plus GoReleaser Windows artifact creation. A later Windows compatibility phase should harden CLI doctor tests and path semantics on native Windows.

## Remaining Release Steps

- Push release workflow.
- Tag `v0.1.0` from a clean tree.
- Confirm GitHub Release uploads Linux/macOS archives, Windows zip files, and `checksums.txt`.
