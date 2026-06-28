# Plan: Phase 25 CLI Project Boundary Refactor

## Objective

Narrow project-related CLI files to adapter responsibilities by moving project/session business rules and their behavior-rule tests into `internal/project`.

## Scope

- Move project entry construction rules out of `internal/cli/project.go`.
- Move project environment utilities out of `internal/cli/run.go`.
- Keep Cobra command definitions, flags, completions, stdout/stderr routing, and child process execution in CLI.
- Add or move tests so `internal/project` owns behavior-rule coverage.
- Preserve CLI behavior exactly.

## Non-Goals

- No command rename, alias, or new command.
- No vault file format or manifest schema change.
- No `internal/app` export/setup/migrate extraction; Phase 26 owns that.
- No broad one-file-per-command split in `internal/cli`.
- No new fine-grained CLI secret metadata command group.

## Inputs Read

- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `internal/cli/project.go`
- `internal/cli/run.go`
- `internal/cli/project_test.go`
- `internal/project/manifest.go`
- `internal/project/resolve.go`
- `internal/vault/store.go`

## Tasks

1. Add project entry-building service APIs in `internal/project`.
   - Introduce request/result shapes for path, prefix, and tag selector entry construction.
   - Preserve existing error semantics expected by CLI tests.
   - Validate exact path existence, prefix matches, tag selector matches, `--env` limitations, and optional required-state handling.
2. Move project environment utilities into `internal/project`.
   - Move child env merge behavior from `childEnv` into a package-level project utility.
   - Move parent env override warning generation into `internal/project`.
   - Keep exit-code mapping and `exec.Command` in CLI.
3. Refactor CLI project/run handlers to call project services.
   - `project add` should delegate entry construction and manifest mutation rules.
   - `project explain` and `project run --dry-run` should call project env warning utilities.
   - `project run` should call project child-env merge utility before executing the child process.
4. Add direct `internal/project` tests.
   - Cover path/prefix/tag entry building, optional entries, duplicate entries, missing exact secret, empty prefix, empty tag selector, and invalid env usage.
   - Cover required/optional diagnostics, prefix/tag expansion, env conflicts, env merge, and override warnings.
5. Trim or update CLI tests only where behavior-rule coverage moved.
   - Keep CLI tests for command output, stdout/stderr routing, missing manifest/outside git wording, completions, no-leak assertions, and smoke workflows.
   - Avoid deleting CLI coverage until equivalent domain tests exist.
6. Format and verify.
   - Run `gofmt` on changed Go files.
   - Run focused project/run tests and then full package tests before phase completion.
7. Update phase summary, verification, capture, and root planning state when execution completes.

## Acceptance Criteria

- `internal/cli/project.go` no longer constructs project entries by directly encoding path/prefix/tag business rules in the Cobra handler.
- `internal/cli/run.go` no longer owns reusable child environment merge or parent override warning rules.
- `internal/project` exposes cohesive APIs for entry construction and project environment behavior without importing Cobra or CLI.
- Project behavior-rule tests run in `internal/project` without `runShelf`, temp git repos, or Cobra unless the behavior actually requires file/git integration.
- CLI tests still cover user-visible command contracts and no secret values leak through project explain/list paths.
- Existing CLI behavior remains unchanged.

## Verification

- `gofmt` on changed Go files.
- `go test ./internal/project`
- `go test ./internal/cli -run 'Test(Project|Run)'`
- `go test ./...`

## Risks

- Moving error construction can change user-facing error strings; preserve current CLI contract or update tests only when behavior is intentionally unchanged but ownership moved.
- Env merge behavior is security-sensitive because `project run` injects secrets into child processes; direct tests must cover parent replacement and new variable insertion.
- Reducing CLI tests too early can lose end-to-end coverage; move coverage only after domain tests prove the same branch.
- Project entry APIs should not overfit CLI flag names; use domain terms such as selector, env override, optional, and tags.
