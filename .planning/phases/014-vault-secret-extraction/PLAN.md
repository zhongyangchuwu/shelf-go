# Phase 14 Plan: Vault Diagnostics and Secret Workflow Extraction

## Goal

Move reusable vault diagnostic rules and `secret edit` workflow out of `internal/cli` while preserving command behavior and plaintext cleanup guarantees.

## Scope

- Add `internal/vault` for status/check/doctor diagnostic records and vault load/tracking guidance.
- Add `internal/secret` for editable secret JSON and temp-file/editor lifecycle.
- Keep `internal/cli` command-family oriented; do not split one file per subcommand.
- Keep `internal/gitutil` deferred unless vault extraction makes shared git helpers necessary. In this phase vault can own its git tracking check.

## Non-goals

- No command behavior or output changes.
- No manager HTTP changes.
- No atomicfile extraction.
- No store backend interface.

## Design

`internal/vault` returns typed diagnostics:

```go
type Level string
const (LevelOK Level = "ok"; LevelWarn Level = "warn"; LevelFail Level = "fail")
type Check struct { Level Level; Name string; Detail string }
type Report []Check
func Status(runtime config.Runtime) Report
func Doctor(runtime config.Runtime) Report
func HasFailures(Report) bool
```

`internal/secret` exposes edit workflow:

```go
type EditRequest struct { Path, Editor string; Stdin io.Reader; Stdout, Stderr io.Writer }
func Edit(st *store.Store, req EditRequest) error
```

## Implementation steps

1. Create `internal/vault/status.go` and move vault diagnostic rules from `internal/cli/manager.go` and `internal/cli/doctor.go`.
2. Update `vault status` and `doctor` commands to render `vault.Report` through CLI diagnostics.
3. Create `internal/secret/edit.go` and move editable-secret JSON/temp/editor workflow from `internal/cli/secret.go`.
4. Update `secret edit` command to call `secret.Edit`.
5. Run focused vault/doctor/secret edit tests and full test suite.

## Verification

- `go test ./internal/vault ./internal/secret ./internal/cli -run 'Test(Vault|Doctor|SecretEdit)'`
- `go test ./...`

## Completion criteria

- ARCH-04 complete: vault diagnostics reusable outside CLI.
- ARCH-05 complete: secret edit workflow reusable outside CLI.
- Existing CLI output and behavior unchanged.
