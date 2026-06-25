# Phase 16 Verification: First Release Readiness

## Result

Passed with one scoped platform decision: GoReleaser publishes Linux and macOS archives for 0.1.0. Windows artifacts are deferred because the current store lock uses Unix `syscall.Flock`, which does not compile for Windows.

## Checks

| Claim | Evidence | Result |
| --- | --- | --- |
| CI includes vet, tests, and build | `.github/workflows/ci.yml` has `go vet ./...`, `go test ./...`, and `go build ./cmd/shelf` | Passed |
| Release workflow publishes from version tags | `.github/workflows/release.yml` triggers on `v*`, uses `contents: write`, checks out full history, runs vet/test, then GoReleaser | Passed |
| GoReleaser config is valid | `go run github.com/goreleaser/goreleaser/v2@latest check` | Passed |
| Snapshot release builds release artifacts | `go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot` | Passed |
| Snapshot artifacts include checksums | `dist/checksums.txt` produced with Linux/macOS archives | Passed |
| Snapshot binary reports injected release version | `./dist/shelf_linux_amd64_v1/shelf --version` printed `0.1.0-SNAPSHOT-97ab92a ... linux/amd64` | Passed |
| CLI still builds locally | `go build -o ./bin/shelf ./cmd/shelf` | Passed |
| Go vet is clean | `go vet ./...` | Passed |
| Unit/integration suite still passes | `go test ./...` | Passed |
| Race suite still passes | `go test -race ./...` | Passed |

## GoReleaser Snapshot Artifacts Observed

```text
dist/shelf_0.1.0-SNAPSHOT-97ab92a_linux_amd64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-97ab92a_linux_arm64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-97ab92a_darwin_amd64.tar.gz
dist/shelf_0.1.0-SNAPSHOT-97ab92a_darwin_arm64.tar.gz
dist/checksums.txt
```

## Platform Decision

Initial Windows targets failed during snapshot release:

```text
internal/store/lock.go:21:20: undefined: syscall.Flock
internal/store/lock.go:21:50: undefined: syscall.LOCK_EX
internal/store/lock.go:32:17: undefined: syscall.Flock
internal/store/lock.go:32:49: undefined: syscall.LOCK_UN
```

Decision: 0.1.0 release artifacts target Linux and macOS only. Windows support requires a portable file-lock implementation and should be handled in a later compatibility phase.

## Remaining Release Steps

- Commit Phase 16 changes.
- Push release workflow.
- Tag `v0.1.0` from a clean tree.
- Confirm GitHub Release uploads archives and `checksums.txt`.
