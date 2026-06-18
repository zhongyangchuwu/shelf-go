# Technology Stack

**Analysis Date:** 2026-06-16

## Languages

**Primary:**
- Go 1.26.4 - CLI application and all production code in `cmd/shelf/main.go` and `internal/**`.

**Secondary:**
- JSON - Local secret store format in `internal/store/model.go`, project manifest format in `internal/manifest/manifest.go`, and examples in `docs/data-spec.md`.
- YAML - Optional user config file parsed by `internal/config/config.go`.
- Just - Developer task runner commands in `justfile`.

## Runtime

**Environment:**
- Go runtime 1.26.4, declared in `go.mod`.
- Local command-line execution; the binary entry point is `cmd/shelf/main.go`.
- Unix-style filesystem behavior is used for file locks and permissions in `internal/store/lock.go` and `internal/store/io.go`.

**Package Manager:**
- Go modules
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- `github.com/spf13/cobra` v1.9.1 - Command tree, flags, command validation, completions, and execution in `internal/cli/root.go`, `internal/cli/secret.go`, `internal/cli/project.go`, `internal/cli/run.go`, and related command files.

**Testing:**
- Go standard `testing` package - Unit and command tests in `internal/cli/*_test.go`, `internal/store/io_test.go`, and `internal/manifest/manifest_test.go`.

**Build/Dev:**
- Go toolchain - Build, install, and test commands in `justfile`.
- `just` - Convenience task runner for `build`, `install`, `test`, and `tag` recipes in `justfile`.

## Key Dependencies

**Critical:**
- `filippo.io/age` v1.3.1 - Encrypts and decrypts the Shelf vault in `internal/store/vault.go`.
- `github.com/spf13/cobra` v1.9.1 - Defines the entire CLI command surface; use it for new commands under `internal/cli/`.
- `gopkg.in/yaml.v3` v3.0.1 - Parses optional runtime config from `~/.config/shelf/config.yaml` in `internal/config/config.go`.
- `golang.org/x/term` v0.44.0 - Reads hidden terminal input for `shelf secret add` in `internal/cli/secret.go`.

**Infrastructure:**
- `github.com/spf13/pflag` v1.0.6 - Indirect Cobra flag dependency.
- `github.com/inconshreveable/mousetrap` v1.1.0 - Indirect Cobra dependency.
- `golang.org/x/sys` v0.46.0 - Indirect terminal and syscall support.

## Configuration

**Environment:**
- Runtime config resolution is implemented in `internal/config/config.go`.
- Default config path: `~/.config/shelf/config.yaml` from `internal/config/config.go`.
- Default vault path: `~/.local/share/shelf/vault.age` from `internal/config/config.go`.
- `SHELF_CONFIG` overrides the config file path in `internal/config/config.go`.
- `SHELF_VAULT` overrides the encrypted vault path in `internal/config/config.go`.
- `EDITOR` supplies the default editor when config has no editor value in `internal/config/config.go`.
- `FPATH` / `fpath` are inspected by `shelf doctor` for zsh completion installation in `internal/cli/doctor.go`.

**Build:**
- `go.mod` declares module `github.com/zhongyangchuwu/shelf-go`, Go version `1.26.4`, and all module dependencies.
- `go.sum` pins module checksums.
- `justfile` provides:
  - `just build`: `go build -o ./bin/shelf ./cmd/shelf`
  - `just install`: `go install ./cmd/shelf`
  - `just test`: `go test ./...`
  - `just tag version`: creates a Git tag.

## Platform Requirements

**Development:**
- Install Go 1.26.4 or compatible toolchain as declared in `go.mod`.
- Install `just` only if using `justfile`; equivalent `go build`, `go install`, and `go test` commands work without it.
- Git is required for project-aware commands such as `shelf project id`, `shelf project init`, `shelf project explain`, and `shelf run` in `internal/cli/project.go` and `internal/cli/run.go`.
- A terminal is required for hidden interactive password input in `shelf secret add` from `internal/cli/secret.go`.
- A shell and editor are required for `shelf secret edit`, which invokes `sh -c "$SHELF_EDITOR \"$SHELF_EDIT_FILE\""` in `internal/cli/secret.go`.

**Production:**
- Deployment target is a local CLI binary named `shelf`.
- Build output defaults to `./bin/shelf` via `justfile`.
- No server runtime, container runtime, hosted platform, or daemon process is detected.

---

*Stack analysis: 2026-06-16*
