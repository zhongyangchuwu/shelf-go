# Shelf Go MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Go rewrite MVP as a fast local secret manager with direct export, editable JSON-backed storage, and shell completion support.

**Architecture:** A Cobra command tree sits on top of a small core store/config layer. The store owns JSON load/save, path validation, atomic writes, and secret object normalization. CLI commands are thin adapters that call the store and render output, including `completion` script generation plus runtime completion hooks for path and flag suggestions.

**Tech Stack:** Go, Cobra, stdlib `encoding/json`, stdlib filesystem APIs, `gopkg.in/yaml.v3` for config loading, and Go test.

---

### Task 1: Bootstrap the Cobra command tree and completions

**Files:**
- Create: `cmd/shelf/main.go`
- Create: `internal/cli/root.go`
- Create: `internal/cli/completion.go`
- Create: `internal/cli/root_test.go`
- Modify: `go.mod`

**Files:**
- Create: `internal/cli/flags.go`

- [ ] **Step 1: Write the failing test**

```go
func TestRootCommandIncludesCompletionAndSecretNamespace(t *testing.T) {
    // verify root command wiring, completion command presence,
    // and that `secret` and `project id` are registered.
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/cli -run TestRootCommandIncludesCompletionAndSecretNamespace -v`

Expected: fail because the command tree does not exist yet.

- [ ] **Step 3: Write the minimal CLI scaffold**

```go
func NewRootCmd() *cobra.Command {
    // root command, persistent flags, subcommands, completion command
}
```

- [ ] **Step 4: Add shell completion generation and runtime completion hooks**

Use Cobra completion support:

```go
cmd.Root().GenBashCompletionV2(os.Stdout, true)
cmd.Root().GenZshCompletion(os.Stdout)
cmd.Root().GenFishCompletion(os.Stdout, true)
cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
```

Use runtime completion APIs where needed:

```go
ValidArgsFunction
RegisterFlagCompletionFunc
cobra.NoFileCompletions
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/cli -run TestRootCommandIncludesCompletionAndSecretNamespace -v`

Expected: PASS.

---

### Task 2: Implement config and JSON store loading/saving

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/store/store.go`
- Create: `internal/store/model.go`
- Create: `internal/store/path.go`
- Create: `internal/store/io.go`
- Create: `internal/store/validate.go`
- Create: `internal/store/store_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadAndSaveJSONStore(t *testing.T) {
    // create a temp secrets.json, load it, mutate it, save it, and reload it.
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/store -run TestLoadAndSaveJSONStore -v`

Expected: fail because storage types and loader do not exist yet.

- [ ] **Step 3: Implement config/data path resolution**

Support the documented priorities for `--config`, `--data`, `SHELF_CONFIG`, `SHELF_DATA`, and config file data path selection.

- [ ] **Step 4: Implement JSON store model and atomic save/load**

Implement a plain JSON store with:

```json
{
  "version": 1,
  "secrets": { ... }
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/store -run TestLoadAndSaveJSONStore -v`

Expected: PASS.

---

### Task 3: Implement secret commands and export

**Files:**
- Create: `internal/cli/secret.go`
- Create: `internal/cli/export.go`
- Create: `internal/secret/path.go`
- Create: `internal/secret/ops.go`
- Create: `internal/cli/secret_test.go`
- Create: `internal/cli/export_test.go`

- [ ] **Step 1: Write the failing tests**

```go
func TestSecretSetGetListInfoEdit(t *testing.T) { ... }
func TestExportShellEnvJson(t *testing.T) { ... }
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/cli -run 'TestSecretSetGetListInfoEdit|TestExportShellEnvJson' -v`

Expected: fail because commands and store ops do not exist yet.

- [ ] **Step 3: Implement secret path parsing and secret object validation**

Validate:

- exactly one `:` in the full path
- namespace segments non-empty
- key non-empty
- env name shape `^[A-Za-z_][A-Za-z0-9_]*$`
- tags are unique strings
- secret values are JSON-compatible

- [ ] **Step 4: Implement commands**

Implement:

```bash
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>
shelf export <path-or-prefix> --format shell|env|json
```

Use `$EDITOR` for `secret edit` and JSON validation on save.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/cli -run 'TestSecretSetGetListInfoEdit|TestExportShellEnvJson' -v`

Expected: PASS.

---

### Task 4: Add project id and completion coverage

**Files:**
- Create: `internal/cli/project.go`
- Create: `internal/cli/completion_test.go`
- Modify: `README.md`

- [ ] **Step 1: Write the failing test**

```go
func TestProjectIDNormalizesGitRemote(t *testing.T) { ... }
func TestCompletionScriptsAreGenerated(t *testing.T) { ... }
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/cli -run 'TestProjectIDNormalizesGitRemote|TestCompletionScriptsAreGenerated' -v`

Expected: fail because project identity and completion code are incomplete.

- [ ] **Step 3: Implement `project id`**

Resolve the current worktree root, read the default remote, normalize it to `host/owner/repo`, and fail cleanly when no Git context exists.

- [ ] **Step 4: Verify completion scripts and dynamic completion hooks**

Ensure the root command can generate Bash, Zsh, Fish, and PowerShell completion scripts, and that secret path / format completion is wired through Cobra.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/cli -run 'TestProjectIDNormalizesGitRemote|TestCompletionScriptsAreGenerated' -v`

Expected: PASS.

---

### Task 5: Final integration check

**Files:**
- All files added above

- [ ] **Step 1: Run the package tests**

Run: `go test ./...`

Expected: all tests pass.

- [ ] **Step 2: Run a real CLI smoke check**

Run:

```bash
go run ./cmd/shelf --help
go run ./cmd/shelf completion zsh
go run ./cmd/shelf secret list
```

Expected: help renders, completion script prints, and command wiring is functional.

- [ ] **Step 3: Keep README aligned**

Ensure the README still describes the MVP scope, the JSON store, and the completion-friendly CLI surface.
