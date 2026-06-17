# External Integrations

**Analysis Date:** 2026-06-16

## APIs & External Services

**External APIs:**
- Not detected - The Go code does not import HTTP API SDKs or call external web services. No `net/http` client usage is present in `cmd/shelf/main.go` or `internal/**`.

**Command-Line Tools:**
- Git - Used to discover project roots, normalize project identity, and check whether the data file is tracked.
  - SDK/Client: `os/exec` calls to `git` in `internal/cli/project.go`, `internal/cli/run.go`, and `internal/cli/doctor.go`.
  - Auth: Git uses the user's local Git configuration; Shelf does not manage Git credentials.
- User editor - Used by `shelf secret edit` to edit a JSON representation of one secret.
  - SDK/Client: `os/exec` call to `sh -c "$SHELF_EDITOR \"$SHELF_EDIT_FILE\""` in `internal/cli/secret.go`.
  - Auth: Not applicable.
- Child commands - `shelf run -- command args...` executes arbitrary local commands with resolved secrets injected into the environment.
  - SDK/Client: `exec.Command(args[0], args[1:]...)` in `internal/cli/run.go`.
  - Auth: Secrets are passed as environment variables derived from local store entries.

**Provider Secret Namespaces:**
- AI provider names such as OpenAI, Anthropic, DeepSeek, and OpenRouter appear as example/user-managed secret paths in `internal/cli/project_test.go`, `internal/cli/secret_test.go`, `README.md`, and `docs/data-spec.md`.
  - SDK/Client: Not detected; these are data examples and manifest entries, not API clients.
  - Auth: User-defined env names such as `OPENAI_API_KEY` or `OPENROUTER_API_KEY` can be stored in local secret metadata and exported by `internal/render/export.go`.

## Data Storage

**Databases:**
- Local JSON file store, not a database.
  - Connection: `SHELF_DATA`, `--data`, optional config `data`, or default `~/.local/share/shelf/secrets.json` resolved by `internal/config/config.go`.
  - Client: Custom file-backed store in `internal/store/io.go` and data model in `internal/store/model.go`.
- Store writes use temp-file plus rename behavior in `internal/store/io.go`.
- Mutating commands take an exclusive lock at `<data-path>.lock` through `internal/store/lock.go` and `internal/cli/root.go`.
- The data file is written with user-only file mode `0600` and parent directories with `0700` in `internal/store/io.go`.

**File Storage:**
- Local filesystem only.
- Runtime config is optional YAML at `~/.config/shelf/config.yaml` by default, parsed in `internal/config/config.go`.
- Project manifests are `.shelf.json` files in Git project roots, modeled in `internal/manifest/manifest.go` and read/written in `internal/manifest/io.go`.
- The project manifest file is written with mode `0644` in `internal/manifest/io.go` and should not contain secret values.

**Caching:**
- None detected.

## Authentication & Identity

**Auth Provider:**
- Custom local secret management.
  - Implementation: Secrets are stored as JSON values with metadata in `internal/store/model.go`, resolved by path/prefix through `internal/store/io.go`, and exported/injected by `internal/render/export.go`, `internal/cli/project.go`, and `internal/cli/run.go`.
- No OAuth, JWT, SSO, session, or remote identity provider integration is detected.

**Project Identity:**
- Git remote/project identity is derived locally by `projectID()` in `internal/cli/project.go`.
- `shelf run` requires a Git root and `.shelf.json` manifest before injecting project-bound secrets in `internal/cli/run.go`.

## Monitoring & Observability

**Error Tracking:**
- None detected.

**Logs:**
- CLI output uses command stdout/stderr through Cobra command writers in `internal/cli/**`.
- Diagnostics are printed by `shelf doctor` in `internal/cli/doctor.go` and by project manifest resolution in `internal/cli/project.go` and `internal/cli/run.go`.
- No structured logging framework is detected.

## CI/CD & Deployment

**Hosting:**
- Not applicable - Shelf Go is a local CLI binary.

**CI Pipeline:**
- None detected - no `.github/`, `.gitlab-ci.yml`, `.circleci/`, or Jenkinsfile files were found.

**Release/Install:**
- Local build command is `go build -o ./bin/shelf ./cmd/shelf` in `justfile`.
- Local install command is `go install ./cmd/shelf` in `justfile`.
- Tagging helper is `git tag v{{version}}` in `justfile`.

## Environment Configuration

**Required env vars:**
- None required for default CLI operation.

**Optional env vars:**
- `SHELF_CONFIG` - Overrides config file path in `internal/config/config.go`.
- `SHELF_DATA` - Overrides secret store path in `internal/config/config.go`.
- `EDITOR` - Fallback editor for `shelf secret edit` in `internal/config/config.go`.
- `FPATH` / `fpath` - Used by `shelf doctor` to check zsh completion installation in `internal/cli/doctor.go`.
- User-defined exported variables such as `OPENAI_API_KEY` or `OPENROUTER_API_KEY` are metadata values stored in secret records and are injected by `internal/cli/run.go`.

**Secrets location:**
- Secrets are stored in the local JSON data file resolved by `internal/config/config.go`.
- Project `.shelf.json` manifests store references to secret paths/prefixes only; they do not store secret values.
- No `.env` files were detected in the repository scan.

## Webhooks & Callbacks

**Incoming:**
- None detected.

**Outgoing:**
- None detected.

---

*Integration audit: 2026-06-16*
